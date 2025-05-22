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
// THE SOFTWARE.

package cmd

import (
	"fmt"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

// versionCmd returns the Cobra command for displaying the CLI version.
// It prints the semantic version of this Remake CLI binary as defined in configuration.
func versionCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the Remake CLI version",
		Long: `Display the current semantic version of the Remake CLI tool.
This is useful for verifying the installed version and checking for updates.
`,
		Example: `  # Print version to stdout
  remake version

  # Use version in scripts or logs
  if [ "$(remake version)" != "v1.2.3" ]; then echo "Update available"; fi`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(app.Cfg.Version)
			return nil
		},
	}
	return cmd
}
