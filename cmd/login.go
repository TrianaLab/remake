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

// loginCmd returns the Cobra command for authenticating the CLI
// against an OCI registry. It supports interactive prompt fallback
// when credentials are not provided via flags or config.
func loginCmd(app *app.App) *cobra.Command {
	var (
		usernameFlag string
		passwordFlag string
	)

	cmd := &cobra.Command{
		Use:   "login [registry]",
		Short: "Authenticate with an OCI registry (e.g., ghcr.io)",
		Long: `Authenticate the Remake CLI with the specified OCI registry.
If no registry is provided, the default registry from configuration is used.
Credentials can be supplied via flags, config file, or interactively prompted.

Examples of registries:
  - GitHub Container Registry: ghcr.io (default)
  - Docker Hub: docker.io
  - Private registry: registry.example.com`,
		Example: `  # Login using flags
  remake login ghcr.io -u myuser -p mypass

  # Login and prompt for password
  remake login ghcr.io -u myuser

  # Login to default registry (from config)
  remake login`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := app.Cfg.DefaultRegistry
			if len(args) == 1 {
				registry = args[0]
			}
			return app.Login(context.Background(), registry, usernameFlag, passwordFlag)
		},
	}

	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for the OCI registry")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password or token for the OCI registry")
	return cmd
}
