package cmd

import (
	"fmt"
	"os"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

var (
	initConfigFunc = config.InitConfig
	exitFunc       = os.Exit
)

// rootCmd is the base command for remake.
var rootCmd = &cobra.Command{
	Use:   "remake",
	Short: "CLI for running Makefiles with remote includes",
	Long:  "remake is a tool to wrap Makefiles as OCI artifacts and resolve remote includes.",
	// PersistentPreRun ensures configuration is initialized once for all subcommands.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initConfigFunc(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitFunc(1)
	}
}

func init() {
}
