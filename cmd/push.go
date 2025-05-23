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
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

// pushCmd returns the Cobra command for uploading a local Makefile artifact
// to an OCI registry. The Makefile is read from the specified file and
// pushed under the provided reference (e.g., registry/repo:tag).
func pushCmd(app *app.App) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "push <reference>",
		Short: "Upload a Makefile artifact to an OCI registry",
		Long: `Push a local Makefile artifact to the given OCI reference.
By default, the command reads from a file named 'makefile' in the current
directory. Use the -f flag to specify a different filename or path.

The <reference> syntax is registry host followed by repository and tag,
for example: ghcr.io/myorg/myrepo:1.0.0`,
		Example: `  # Push default makefile to GitHub Container Registry
  remake push ghcr.io/myorg/myrepo:latest

  # Push a development makefile
  remake push ghcr.io/myorg/myrepo:dev -f Makefile.dev

  # Push to default registry with custom file
  remake push myorg/myrepo:v2 -f ./ci/Makefile.ci`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			return app.Push(context.Background(), ref, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "makefile",
		"Path to the local Makefile to upload (default 'makefile')")
	return cmd
}
