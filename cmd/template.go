package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/spf13/cobra"
)

var (
	templateFile              string
	templateNoCache           bool
	templateDefaultMakefileFn = config.GetDefaultMakefile
	cacheDirFn                = config.GetCacheDir
	renderFn                  = run.Render
)

// templateCmd prints the fully-resolved Makefile without executing it.
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Print the fully-resolved Makefile without executing it",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// file to render
		file := templateFile
		if file == "" {
			file = templateDefaultMakefileFn()
			if file == "" {
				return fmt.Errorf("no Makefile found; specify with -f flag")
			}
		}

		// generate into cache
		cacheDir := cacheDirFn()
		out := filepath.Join(cacheDir, "Makefile.generated")
		// ensure cache dir
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}

		// render with includes resolved
		if err := renderFn(file, out, !templateNoCache); err != nil {
			return err
		}

		// open and copy to stdout
		f, err := os.Open(out)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(os.Stdout, f); err != nil {
			return err
		}

		// cleanup
		_ = os.Remove(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringVarP(&templateFile, "file", "f", "", "Makefile to render (default: detect makefile)")
	templateCmd.Flags().BoolVar(&templateNoCache, "no-cache", false, "Skip cache")
}
