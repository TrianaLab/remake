package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of remake",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
