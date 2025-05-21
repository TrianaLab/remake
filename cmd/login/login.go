// cmd/login/login.go
package login

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

func LoginCmd(cfg *config.Config) *cobra.Command {
	var (
		usernameFlag string
		passwordFlag string
	)
	cmd := &cobra.Command{
		Use:     "login [registry]",
		Short:   "Authenticate against an OCI registry",
		Long:    "Log in to the given OCI registry (defaults to the configured defaultRegistry).",
		Example: "  remake login ghcr.io -u myuser -p mypass\n  remake login             # toma defaultRegistry y pide credenciales si hacen falta",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := cfg.DefaultRegistry
			if len(args) == 1 {
				registry = args[0]
			}
			return app.New(cfg).Login(context.Background(), registry, usernameFlag, passwordFlag)
		},
	}

	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for the registry")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for the registry")
	return cmd
}
