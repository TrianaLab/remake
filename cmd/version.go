package cmd

import (
	"fmt"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func versionCmd(app *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Prints the CLI version",
		Long:    "Prints the current version of the Remake CLI.",
		Example: "  remake version",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(app.Cfg.Version)
			return nil
		},
	}
	return cmd
}
