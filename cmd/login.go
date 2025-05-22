package cmd

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func loginCmd(app *app.App) *cobra.Command {
	var (
		usernameFlag string
		passwordFlag string
	)
	cmd := &cobra.Command{
		Use:     "login [registry]",
		Short:   "Authenticate against an OCI registry",
		Long:    "Log in to the given OCI registry (defaults to the configured defaultRegistry).",
		Example: "  remake login ghcr.io -u myuser -p mypass\n  remake login ghcr.io\n  remake login",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := app.Cfg.DefaultRegistry
			if len(args) == 1 {
				registry = args[0]
			}
			return app.Login(context.Background(), registry, usernameFlag, passwordFlag)
		},
	}

	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for the registry")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for the registry")
	return cmd
}
