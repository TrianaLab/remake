package cmd

import (
	"fmt"

	"github.com/TrianaLab/remake/internal/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	templateFile    string
	templateNoCache bool
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template [flags]",
	Short: "Print a Makefile with resolved includes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine source
		src := templateFile
		if src == "" {
			src = viper.GetString("defaultMakefile")
			if src == "" {
				return fmt.Errorf("no Makefile specified; use -f flag or config")
			}
		}
		// Execute
		return process.Template(src, !templateNoCache)
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringVarP(&templateFile, "file", "f", "", "Makefile path or remote reference")
	templateCmd.Flags().BoolVar(&templateNoCache, "no-cache", false, "disable cache")
	viper.BindPFlag("defaultMakefile", templateCmd.Flags().Lookup("file"))
}
