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

// runCmd returns the Cobra command for executing one or more Makefile targets.
// It can operate on a local Makefile or pull a remote artifact from an OCI registry.
// By default, the Makefile is cached locally and subsequent runs reuse the cache
// unless --no-cache is specified. Additional flags can be passed to make via --make-flag.
func runCmd(app *app.App) *cobra.Command {
	var (
		noCache   bool
		file      string
		makeFlags []string
	)

	cmd := &cobra.Command{
		Use:   "run [targets...]",
		Short: "Execute Makefile targets from local or remote artifact",
		Long: `Execute specified targets from a Makefile. By default, the CLI looks for a
local file named 'makefile' in the current directory. To run a Makefile stored
as an OCI artifact, use the -f flag with a reference (e.g., ghcr.io/myorg/myrepo:latest).

The command uses a local cache directory (e.g., ~/.remake/cache) to avoid repeated
downloads; use --no-cache to force re-download. Any flags provided via
--make-flag are forwarded directly to the make process.`,
		Example: `  # Run default targets 'all' and 'test' from local Makefile
  remake run all test

  # Pass custom flags to make
  remake run --make-flag -j4 --make-flag --silent build

  # Execute target from remote Makefile artifact, bypassing cache
  remake run -f ghcr.io/myorg/myrepo:latest --no-cache deploy`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Cfg.NoCache = noCache
			return app.Run(context.Background(), file, makeFlags, args)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false,
		"Bypass the local cache and always fetch Makefile artifact")
	cmd.Flags().StringVarP(&file, "file", "f", "makefile",
		"Makefile path or OCI reference to use (default 'makefile')")
	cmd.Flags().StringArrayVar(&makeFlags, "make-flag", nil,
		"Flags to pass through to the make process")
	return cmd
}
