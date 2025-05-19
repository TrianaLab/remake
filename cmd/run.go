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
	Short: "Run a Makefile target",
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
