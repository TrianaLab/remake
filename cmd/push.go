package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/config"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var (
	pushFile     string
	pushInsecure bool
)

var pushCmd = &cobra.Command{
	Use:   "push <remote_ref>",
	Short: "Push a Makefile as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}

		rawRef := args[0]
		refStr := strings.TrimPrefix(rawRef, "oci://")
		if !regexp.MustCompile(`^[^/]+/[^/]+`).MatchString(refStr) {
			return fmt.Errorf("invalid reference: %s", rawRef)
		}

		parts0 := strings.SplitN(refStr, "/", 2)
		hostPart := parts0[0]
		pathPart := parts0[1]

		if !strings.Contains(pathPart, ":") {
			pathPart += ":latest"
		}

		refWithoutOCI := hostPart + "/" + pathPart

		if !strings.HasPrefix(rawRef, "oci://") && !pushInsecure {
			refWithoutOCI = viper.GetString("default_registry") + "/" + refWithoutOCI
		}
		fullRef := "oci://" + refWithoutOCI

		parts := strings.SplitN(refWithoutOCI, "/", 2)
		host := parts[0]
		repoAndTag := parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoPath, tag := rt[0], rt[1]

		filePath := pushFile
		if filePath == "" {
			filePath = config.GetDefaultMakefile()
		}
		if filePath == "" {
			return fmt.Errorf("no Makefile found; specify with --file flag")
		}

		ctx := context.Background()
		fs, err := file.New(filepath.Dir(filePath))
		if err != nil {
			return err
		}
		defer fs.Close()
		desc, err := fs.Add(ctx, filePath, "application/x-makefile", "")
		if err != nil {
			return err
		}

		manifestDesc, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1,
			"application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
		if err != nil {
			return err
		}
		if err := fs.Tag(ctx, manifestDesc, tag); err != nil {
			return err
		}

		repoRef := host + "/" + repoPath
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}
		if pushInsecure || strings.Contains(host, ":") {
			repo.PlainHTTP = true
		}
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(host, auth.Credential{
				Username: viper.GetString(fmt.Sprintf("registries.%s.username", host)),
				Password: viper.GetString(fmt.Sprintf("registries.%s.password", host)),
			}),
		}
		if _, err := oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to push artifact: %w", err)
		}

		fmt.Printf("Pushed %s to %s\n", filePath, fullRef)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Makefile to push (default: Makefile or makefile)")
	pushCmd.Flags().BoolVar(&pushInsecure, "insecure", false, "Allow plain HTTP for registry")
}
