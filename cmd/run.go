package cmd

import (
	"fmt"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/spf13/cobra"
)

var (
	runFile    string
	runNoCache bool
)

var runCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run make target with remote includes resolved",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := runFile
		if file == "" {
			file = config.GetDefaultMakefile()
			if file == "" {
				return fmt.Errorf("no Makefile found; specify with -f")
			}
		}
		return run.Run(file, args, !runNoCache)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Makefile to use")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "Skip cache")
}
