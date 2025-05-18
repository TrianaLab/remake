package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "remake",
	Short: "CLI for running Makefiles with remote includes",
	Long:  "remake is a tool to wrap Makefiles as OCI artifacts and resolve remote includes.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
}
