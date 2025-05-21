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
			// Choose registry
			registry := cfg.DefaultRegistry
			if len(args) == 1 {
				registry = args[0]
			}
			// Credentials from config
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
			// If config credentials are empty, request them
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
			// Run login
			if err := app.New(cfg).Login(context.Background(), registry, user, pass); err != nil {
				return err
			}
			// Output
			fmt.Println("Login succeeded ✅")
			return nil
		},
	}

	cmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username for the registry")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password for the registry")
	return cmd
}
