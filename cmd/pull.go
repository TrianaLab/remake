package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

var (
	pullFile     string
	pullInsecure bool
	// allow overriding for tests
	pullInitConfig = config.InitConfig
)

// pullCmd downloads a Makefile OCI artifact (or HTTP URL) into the working directory.
var pullCmd = &cobra.Command{
	Use:   "pull <remote_ref>",
	Short: "Pull a Makefile OCI artifact or HTTP URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Init config
		if err := pullInitConfig(); err != nil {
			return err
		}

		rawRef := args[0]

		// 2) Handle HTTP/HTTPS
		if strings.HasPrefix(rawRef, "http://") || strings.HasPrefix(rawRef, "https://") {
			// Download via HTTP
			resp, err := http.Get(rawRef)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("HTTP error %d", resp.StatusCode)
			}
			// write to file
			out := pullFile
			if out == "" {
				out = filepath.Base(rawRef)
			}
			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(f, resp.Body)
			return err
		}

		// 3) OCI reference: strip prefix
		oci := false
		ref := rawRef
		if strings.HasPrefix(rawRef, "oci://") {
			oci = true
			ref = rawRef[len("oci://"):]
		}

		// 4) Split host/path
		parts0 := strings.SplitN(ref, "/", 2)
		if len(parts0) != 2 {
			return fmt.Errorf("invalid reference: %s", rawRef)
		}
		host := parts0[0]
		path := parts0[1]

		// 5) Extract tag
		tag := "latest"
		if idx := strings.LastIndex(path, ":"); idx != -1 {
			tag = path[idx+1:]
			path = path[:idx]
		}

		// 6) Default registry
		if !oci {
			host = viper.GetString("default_registry")
		}

		// 7) Prepare local store
		ctx := context.Background()
		fs, err := file.New(".")
		if err != nil {
			return err
		}
		defer fs.Close()

		// 8) Remote repository
		repoRef := host + "/" + path
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}
		if pullInsecure {
			repo.PlainHTTP = true
		}
		repo.Client = &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache()}

		// 9) Pull
		if _, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to pull artifact: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullFile, "file", "f", "", "Destination Makefile or output file")
	pullCmd.Flags().BoolVar(&pullInsecure, "insecure", false, "Allow plain HTTP for registry")
}
