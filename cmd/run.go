package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/util"
)

var runFile string
var runNoCache bool

var runCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run make target with remote includes resolved",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// ensure make is available
		if _, err := exec.LookPath("make"); err != nil {
			return fmt.Errorf("make not found in PATH")
		}

		// load configuration and credentials
		if err := config.InitConfig(); err != nil {
			return err
		}

		// clear cache if requested
		if runNoCache {
			os.RemoveAll(config.GetCacheDir())
		}

		// determine Makefile path, fetching remote if needed
		file := runFile
		if strings.HasPrefix(file, "oci://") || strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
			localPath, err := util.FetchMakefile(file)
			if err != nil {
				return err
			}
			file = localPath
		}
		if file == "" {
			file = config.GetDefaultMakefile()
		}

		// execute make with generated file
		cacheDir := config.GetCacheDir()
		gen := filepath.Join(cacheDir, "Makefile.generated")
		err := run.Run(args, file)
		_ = os.Remove(gen)
		return err
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Makefile to use")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "Skip cache")
}
