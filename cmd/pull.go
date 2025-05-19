package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/TrianaLab/remake/internal/fetch"
	"github.com/spf13/cobra"
)

var (
	pullFile       string
	pullNoCache    bool
	pullInsecure   bool
	pullGetFetcher = fetch.GetFetcher
)

// pullCmd downloads a Makefile OCI artifact or HTTP URL.
var pullCmd = &cobra.Command{
	Use:   "pull <oci://endpoint/repo:tag|http(s)://...>",
	Short: "Pull a Makefile artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		// Select fetcher
		fetcher, err := pullGetFetcher(ref)
		if err != nil {
			return err
		}

		// Fetch (cache or remote)
		path, err := fetcher.Fetch(ref, !pullNoCache)
		if err != nil {
			return err
		}

		// If a destination file is specified, copy from path
		if pullFile != "" {
			if path == "" {
				return fmt.Errorf("artifact not found in cache and remote fetch disabled")
			}
			src, err := os.Open(path)
			if err != nil {
				return err
			}
			defer src.Close()

			dst, err := os.Create(pullFile)
			if err != nil {
				return err
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return err
			}

			fmt.Printf("Saved to %s (copied from %s)\n", pullFile, path)
			return nil
		}

		// Otherwise, print cache path or success message
		if path != "" {
			fmt.Printf("Saved to %s\n", path)
		} else {
			fmt.Printf("Fetched to current directory\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullFile, "file", "f", "", "Destination file (default: cache file)")
	pullCmd.Flags().BoolVar(&pullNoCache, "no-cache", false, "Force download and ignore cache")
	pullCmd.Flags().BoolVar(&pullInsecure, "insecure", false, "Allow insecure HTTP registry")
}
