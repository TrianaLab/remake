package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var (
	pushFile     string
	pushInsecure bool
)

// pushCmd pushes a local Makefile as an OCI artifact.
var pushCmd = &cobra.Command{
	Use:   "push <oci://endpoint/repo:tag>",
	Short: "Push a Makefile as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]
		if !strings.HasPrefix(raw, "oci://") {
			return fmt.Errorf("reference must start with oci://")
		}
		ref := strings.TrimPrefix(raw, "oci://")

		parts := strings.SplitN(ref, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid reference: %s", raw)
		}
		host, repoTag := parts[0], parts[1]

		// ensure tag
		if !strings.Contains(repoTag, ":") {
			repoTag += ":latest"
		}

		// determine file to push
		filePath := pushFile
		if filePath == "" {
			filePath = viper.GetString("defaultMakefile")
			if filePath == "" {
				return fmt.Errorf("no Makefile found; specify with -f flag")
			}
		}

		// prepare file store
		dir := filepath.Dir(filePath)
		fs, err := file.New(dir)
		if err != nil {
			return err
		}
		defer fs.Close()

		// add Makefile layer
		desc, err := fs.Add(context.Background(), filePath, "application/x-makefile", "")
		if err != nil {
			return err
		}

		// pack manifest using OCI image-spec
		manifestDesc, err := oras.PackManifest(context.Background(), fs, oras.PackManifestVersion1_1,
			"application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
		if err != nil {
			return err
		}

		// tag manifest
		tag := strings.SplitN(repoTag, ":", 2)[1]
		if err := fs.Tag(context.Background(), manifestDesc, tag); err != nil {
			return err
		}

		// push to remote
		repoRef := fmt.Sprintf("%s/%s", host, strings.SplitN(repoTag, ":", 2)[0])
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}
		if pushInsecure {
			repo.PlainHTTP = true
		}

		// authentication
		key := config.NormalizeKey(host)
		username := viper.GetString(fmt.Sprintf("registries.%s.username", key))
		password := viper.GetString(fmt.Sprintf("registries.%s.password", key))
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(host, auth.Credential{
				Username: username,
				Password: password,
			}),
		}

		if _, err := oras.Copy(context.Background(), fs, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to push artifact: %w", err)
		}

		fmt.Printf("Pushed %s to %s%s/%s\n", filePath, "oci://", host, repoTag)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Makefile to push (default: Makefile or makefile)")
}
