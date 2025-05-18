package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/util"
	"github.com/spf13/cobra"
)

var templateFile string
var templateNoCache bool

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Print the fully-resolved Makefile without executing it",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}

		file := templateFile
		var err error

		if file == "" {
			file = config.GetDefaultMakefile()
		} else {
			fetcher, err := util.GetFetcher(file)
			if err == nil {
				file, err = fetcher.Fetch(file, !templateNoCache)
				if err != nil {
					return err
				}
			}
		}

		cacheDir := config.GetCacheDir()
		gen := filepath.Join(cacheDir, "Makefile.generated")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}

		if _, err = run.Template(file, gen, !templateNoCache); err != nil { // Parámetro añadido
			return err
		}

		f, err := os.Open(gen)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = io.Copy(os.Stdout, f); err != nil {
			return err
		}

		return os.Remove(gen)
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringVarP(&templateFile, "file", "f", "", "Makefile to template (can be local or remote)")
	templateCmd.Flags().BoolVar(&templateNoCache, "no-cache", false, "Skip cache")
}
