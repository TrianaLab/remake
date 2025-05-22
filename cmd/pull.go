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
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

// pullCmd returns the Cobra command for retrieving a Makefile artifact
// from an OCI registry. It downloads the artifact into the local cache
// (unless bypassed) and prints its contents to stdout.
func pullCmd(app *app.App) *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:   "pull <reference>",
		Short: "Download and display a Makefile artifact",
		Long: `Download the specified Makefile artifact from an OCI registry and print its
contents to stdout. By default, the artifact is stored in a local cache directory
(e.g., ~/.remake/cache) and subsequent pulls use the cache unless the --no-cache
flag is specified.

If the <reference> does not include a registry host (e.g., myorg/myrepo:tag),
the default registry from configuration is used.`,
		Example: `  # Pull latest Makefile from GitHub Container Registry
  remake pull ghcr.io/myorg/myrepo:latest

  # Pull from default registry (from config)
  remake pull myorg/myrepo:latest

  # Force re-download and bypass cache
  remake pull ghcr.io/myorg/myrepo:latest --no-cache`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			app.Cfg.NoCache = noCache
			return app.Pull(context.Background(), ref)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false,
		"Bypass the local cache and always fetch from the registry")
	return cmd
}
