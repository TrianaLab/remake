package cmd

import (
	"fmt"
	"os"

	"github.com/TrianaLab/remake/internal/fetch"
	"github.com/spf13/cobra"
)

var (
	pullNoCache bool
)

// pullCmd pulls an artifact (OCI or HTTP), optionally bypasses cache, and prints its content.
var pullCmd = &cobra.Command{
	Use:   "pull <endpoint>",
	Short: "Pull an artifact and print its content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]

		// Obtain appropriate fetcher (OCI or HTTP)
		fetcher, err := fetch.GetFetcher(ref)
		if err != nil {
			return err
		}

		// Determine cache usage
		useCache := !pullNoCache

		// Fetch artifact
		path, err := fetcher.Fetch(ref, useCache)
		if err != nil {
			return err
		}

		// Read file content
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Print content
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVar(&pullNoCache, "no-cache", false, "Force download and ignore cache")
}
