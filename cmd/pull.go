package cmd

import (
	"context"
	"fmt"
	"regexp"
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

var (
	pullFile     string
	pullInsecure bool
)

var pullCmd = &cobra.Command{
	Use:   "pull <remote_ref>",
	Short: "Pull a Makefile OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}

		rawRef := args[0]
		refForValidation := rawRef
		if strings.HasPrefix(rawRef, "oci://") {
			refForValidation = strings.TrimPrefix(rawRef, "oci://")
		}
		if !regexp.MustCompile(`^[^/]+/[^/]+`).MatchString(refForValidation) {
			return fmt.Errorf("invalid reference: %s", rawRef)
		}

		ref := strings.TrimPrefix(rawRef, "oci://")
		if !strings.HasSuffix(ref, ":latest") && !strings.Contains(strings.SplitN(ref, "/", 2)[1], ":") {
			ref += ":latest"
		}
		if !strings.HasPrefix(rawRef, "oci://") {
			ref = viper.GetString("default_registry") + "/" + ref
		}

		ctx := context.Background()
		fs, err := file.New(".")
		if err != nil {
			return err
		}
		defer fs.Close()

		parts := strings.SplitN(ref, "/", 2)
		host := parts[0]
		repoAndTag := parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoName := rt[0]
		tag := "latest"
		if len(rt) == 2 {
			tag = rt[1]
		}
		repoRef := host + "/" + repoName
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}

		if pullInsecure {
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

		if _, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to pull artifact: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullFile, "file", "f", "", "Destination Makefile (default: Makefile or makefile)")
	pullCmd.Flags().BoolVar(&pullInsecure, "insecure", false, "Allow plain HTTP for registry")
}
