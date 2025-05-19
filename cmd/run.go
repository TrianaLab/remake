package cmd

import (
	"fmt"

	"github.com/TrianaLab/remake/internal/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runFile    string
	runNoCache bool
)

var runCmd = &cobra.Command{
	Use:   "run [flags] [targets]",
	Short: "Run a Makefile target from a local or remote source.",
	Long: `Run executes the specified Makefile target. The Makefile can be:
• A local file on disk
• An public HTTP(s) URL (e.g. https://example.com/makefile)
• An OCI registry reference (e.g. oci://registry.example.com/repo:tag)
Downloaded artifacts are cached by default under the configured cache directory.
Use --no-cache to bypass the cache and force a fresh download.
All output and errors from make are streamed to stdout/stderr.`,
	Example: `  # Run a target from the default Makefile
  remake run build
  
  # Run a target from a remote Makefile URL
  remake run oci://ghcr.io/user/package:version test
  
  Run a target without using the cache
  remake run --no-cache oci://ghcr.io/user/package:version test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		src := runFile
		if src == "" {
			src = viper.GetString("defaultMakefile")
			if src == "" {
				return fmt.Errorf("no Makefile specified; use -f flag or config")
			}
		}
		return process.Run(src, args, !runNoCache)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Makefile path or remote reference")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "disable cache")
	viper.BindPFlag("defaultMakefile", runCmd.Flags().Lookup("file"))
}
