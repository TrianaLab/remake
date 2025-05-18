package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/util"
	"github.com/spf13/cobra"
)

var renderFile string
var renderNoCache bool

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Print the fully-resolved Makefile without executing it",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}
		if renderNoCache {
			os.RemoveAll(config.GetCacheDir())
		}
		file := renderFile
		if strings.HasPrefix(file, "oci://") || strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
			local, err := util.FetchMakefile(file)
			if err != nil {
				return err
			}
			file = local
		}
		if file == "" {
			file = config.GetDefaultMakefile()
		}
		cacheDir := config.GetCacheDir()
		gen := filepath.Join(cacheDir, "Makefile.generated")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return err
		}
		if _, err := run.Render(file, gen); err != nil {
			return err
		}
		f, err := os.Open(gen)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(os.Stdout, f); err != nil {
			return err
		}
		return os.Remove(gen)
	},
}

func init() {
	rootCmd.AddCommand(renderCmd)
	renderCmd.Flags().StringVarP(&renderFile, "file", "f", "", "Makefile to render")
	renderCmd.Flags().BoolVar(&renderNoCache, "no-cache", false, "Skip cache")
}
