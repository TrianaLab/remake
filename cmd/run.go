package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/util"
	"github.com/spf13/cobra"
)

var runFile string
var runNoCache bool

var runCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run make target with remote includes resolved",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := exec.LookPath("make"); err != nil {
			return fmt.Errorf("make not found in PATH")
		}

		if err := config.InitConfig(); err != nil {
			return err
		}

		file := runFile
		var err error

		if file == "" {
			file = config.GetDefaultMakefile()
		} else {
			fetcher, err := util.GetFetcher(file)
			if err == nil {
				file, err = fetcher.Fetch(file, !runNoCache)
				if err != nil {
					return err
				}
			}
		}

		err = run.Run(args, file, !runNoCache) // Añadido parámetro useCache
		if err != nil {
			return err
		}

		gen := filepath.Join(config.GetCacheDir(), "Makefile.generated")
		_ = os.Remove(gen)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Makefile to use (can be local or remote)")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "Skip cache")
}
