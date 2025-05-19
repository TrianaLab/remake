package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current Remake CLI version.",
	Long: `Version prints the Remake CLI version and build metadata (commit hash, build date). 
Useful for debugging and ensuring compatibility.`,
	Example: ` # Show Remake version
  remake version`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
