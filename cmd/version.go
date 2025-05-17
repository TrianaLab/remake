package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "dev"

// versionCmd prints the current version of the CLI
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of remake",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("remake version %s", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
