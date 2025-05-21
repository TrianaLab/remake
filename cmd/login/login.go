package login

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func LoginCmd(a *app.App) *cobra.Command {
	var user, pass string
	cmd := &cobra.Command{
		Use: "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.Login(context.Background(), user, pass)
		},
	}
	cmd.Flags().StringVarP(&user, "user", "u", "", "username")
	cmd.Flags().StringVarP(&pass, "pass", "p", "", "password")
	return cmd
}
