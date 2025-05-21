// cmd/login/login.go
package login

import (
	"context"
	"fmt"
	"os"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func LoginCmd(a *app.App, cfg *config.Config) *cobra.Command {
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
			// 1) elegir registry
			registry := cfg.DefaultRegistry
			if len(args) == 1 {
				registry = args[0]
			}
			// 2) credenciales desde flags o config
			user := usernameFlag
			pass := passwordFlag
			if user == "" || pass == "" {
				key := config.NormalizeKey(registry)
				if user == "" {
					user = viper.GetString("registries." + key + ".username")
				}
				if pass == "" {
					pass = viper.GetString("registries." + key + ".password")
				}
			}
			// 3) si siguen vacías, pedir interactivamente
			if user == "" {
				fmt.Fprint(os.Stderr, "Username: ")
				fmt.Scanln(&user)
			}
			if pass == "" {
				fmt.Fprint(os.Stderr, "Password: ")
				bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(os.Stderr)
				if err != nil {
					return err
				}
				pass = string(bytePass)
			}
			// 4) invocar login
			if err := a.Login(context.Background(), registry, user, pass); err != nil {
				return err
			}
			// 5) reporte exitoso
			fmt.Println("Login succeeded ✅")
			return nil
		},
	}

	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for the registry")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for the registry")
	return cmd
}
