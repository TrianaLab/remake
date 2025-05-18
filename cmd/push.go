package cmd

import (
	"fmt"
	"path/filepath"
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

var pushFile string

var pushCmd = &cobra.Command{
	Use:   "push <remote_ref>",
	Short: "Push a Makefile as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}
		raw := args[0]
		hasOCI := strings.HasPrefix(raw, "oci://")
		ref := raw
		if hasOCI {
			ref = strings.TrimPrefix(raw, "oci://")
		}
		name := ref[strings.LastIndex(ref, "/")+1:]
		if !strings.Contains(name, ":") {
			ref += ":latest"
		}
		if !hasOCI {
			ref = viper.GetString("default_registry") + "/" + ref
		}
		fullRef := "oci://" + ref
		parts := strings.SplitN(ref, "/", 2)
		host, repoAndTag := parts[0], parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoPath, tag := rt[0], rt[1]
		filePath := pushFile
		if filePath == "" {
			filePath = config.GetDefaultMakefile()
		}
		if filePath == "" {
			return fmt.Errorf("no Makefile found; specify with --file flag")
		}
		ctx := cmd.Context()
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
		username := viper.GetString(fmt.Sprintf("registries.%s.username", host))
		password := viper.GetString(fmt.Sprintf("registries.%s.password", host))
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(host, auth.Credential{
				Username: username,
				Password: password,
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
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Makefile to push")
}
