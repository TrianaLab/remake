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
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var publishFile string

// publishCmd publishes a Makefile as an OCI artifact
var publishCmd = &cobra.Command{
	Use:   "publish <remote_ref>",
	Short: "Publish a Makefile as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// load config
		if err := config.InitConfig(); err != nil {
			return err
		}
		// normalize ref
		ref := args[0]
		if !strings.HasPrefix(ref, "oci://") {
			ref = "oci://" + viper.GetString("default_registry") + "/" + ref
		}
		// determine file
		file := runFile
		if file == "" {
			file = config.GetDefaultMakefile()
		}
		if file == "" {
			return fmt.Errorf("no Makefile or makefile found; specify with --file")
		}
		// oras push
		ociRef := strings.TrimPrefix(ref, "oci://")
		cmdArgs := []string{
			"push", ociRef,
			"--artifact-type", "application/x-makefile",
			fmt.Sprintf("%s:application/x-makefile", file),
		}
		orasCmd := exec.Command("oras", cmdArgs...)
		orasCmd.Stdout = os.Stdout
		orasCmd.Stderr = os.Stderr
		orasCmd.Stdin = os.Stdin
		if err := orasCmd.Run(); err != nil {
			return fmt.Errorf("oras push failed: %w", err)
		}
		fmt.Printf("✅ Published %s to %s", file, ref)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringVarP(&publishFile, "file", "f", "", "Makefile to publish (default: Makefile or makefile)")
}
