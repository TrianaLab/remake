/*
Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/spf13/cobra"
)

var (
	runFile    string
	runNoCache bool
)

// runCmd resolves includes and executes make target
var runCmd = &cobra.Command{
	Use:   "run <target>",
	Short: "Run make target with remote includes resolved",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. ensure make exists
		if _, err := exec.LookPath("make"); err != nil {
			return fmt.Errorf("make not found in PATH")
		}
		// 2. load config
		config.InitConfig()
		// 2.1. si piden no-cache, limpio la caché
		if runNoCache {
			os.RemoveAll(config.GetCacheDir())
		}
		// 3. determine file
		file := pushFile
		if file == "" {
			file = config.GetDefaultMakefile()
		}
		// 4. run process
		return run.Run(args, file)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Makefile to use (default: Makefile or makefile)")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "Skip local cache and always fetch remote includes")
}
