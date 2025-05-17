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

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/util"
	"github.com/spf13/cobra"
)

var (
	pullOutput  string
	pullNoCache bool
)

// pullCmd pulls a Makefile (local, HTTP, or OCI) into cache
var pullCmd = &cobra.Command{
	Use:   "pull <remote_ref>",
	Short: "Pull a Makefile into cache (assumes ghcr.io and :latest)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Inicializar config para que default_registry esté disponible
		if err := config.InitConfig(); err != nil {
			return err
		}
		// 1.1. si piden no-cache, limpio la caché
		if pullNoCache {
			os.RemoveAll(config.GetCacheDir())
		}
		// 2) Resolver y cachear el makefile remoto/local
		ref := args[0]
		local, err := util.FetchMakefile(ref)
		if err != nil {
			return err
		}
		// 3) Si se indicó -o, moverlo; si no, imprimir la ruta
		if pullOutput != "" {
			return os.Rename(local, pullOutput)
		}
		fmt.Println(local)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", "", "Output file path (default prints cache path)")
	pullCmd.Flags().BoolVar(&pullNoCache, "no-cache", false, "Skip local cache and always fetch remote Makefile")
}
