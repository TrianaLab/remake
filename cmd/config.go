// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

// configCmd returns the Cobra command for showing the current CLI
// configuration values. It prints out settings such as the default
// OCI registry endpoint, cache location, and any stored credentials.
func configCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display the current Remake CLI configuration",
		Long: `Display the current Remake CLI configuration, including:
  - Default OCI registry endpoint (e.g., registry.example.com)
  - Local cache directory
  - Stored credentials for registries

Use this to verify or export your active settings before running other commands.`,
		Example: `  # Print configuration to stdout
  remake config

  # Save current configuration to a file
  remake config > config.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Cfg.PrintConfig()
		},
	}
	return cmd
}
