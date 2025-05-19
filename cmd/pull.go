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
	Short: "Pull an artifact and print its content.",
	Long: `Pull retrieves a Makefile artifact from the specified endpoint, which can be:
• An HTTP(s) URL pointing to a raw Makefile
• An OCI registry reference (e.g. oci://registry.example.com/repo:tag)
The command will cache downloaded artifacts under the configured cache directory to speed up
subsequent pulls. Use the --no-cache flag to bypass the cache and force a fresh download.
The content of the Makefile is printed to stdout for inspection or piping into other tools.`,
	Example: `  # Pull a Makefile over HTTP
  remake pull https://example.com/Makefile

  # Pull a Makefile from an OCI registry
  remake pull oci://registry.example.com/myrepo:latest

  # Pull a Makefile without using cache
  remake pull --no-cache https://example.com/Makefile`,
	Args: cobra.ExactArgs(1),
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
