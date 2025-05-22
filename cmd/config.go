package cmd

import (
	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func configCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Prints the CLI configuration",
		Long:    "Prints the current configuration of the Remake CLI.",
		Example: "  remake config",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.Cfg.PrintConfig()
		},
	}
	return cmd
}
