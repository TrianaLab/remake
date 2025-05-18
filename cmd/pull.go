package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var pullOutput string
var pullNoCache bool

var pullCmd = &cobra.Command{
	Use:   "pull <remote_ref>",
	Short: "Pull a Makefile into cache (OCI, HTTP or local)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}
		cacheDir := config.GetCacheDir()
		if pullNoCache {
			os.RemoveAll(cacheDir)
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
			name += ":latest"
		}
		fileName := strings.SplitN(name, ":", 2)[0]
		if !hasOCI {
			ref = viper.GetString("default_registry") + "/" + ref
		}

		parts := strings.SplitN(ref, "/", 2)
		host, repoAndTag := parts[0], parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoPath, tag := rt[0], rt[1]

		dir := filepath.Join(cacheDir, repoPath, tag)

		localPath := filepath.Join(dir, fileName)
		if !pullNoCache {
			if _, err := os.Stat(localPath); err == nil {
				if pullOutput != "" {
					return os.Rename(localPath, pullOutput)
				}
				fmt.Println(localPath)
				return nil
			}
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

		fs, err := file.New(dir)
		if err != nil {
			return err
		}
		defer fs.Close()
		ctx := cmd.Context()
		_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
		if err != nil {
			return fmt.Errorf("failed to pull artifact: %w", err)
		}

		localPath = filepath.Join(dir, fileName)
		if pullOutput != "" {
			return os.Rename(localPath, pullOutput)
		}
		fmt.Println(localPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", "", "Output file path")
	pullCmd.Flags().BoolVar(&pullNoCache, "no-cache", false, "Skip cache")
}
