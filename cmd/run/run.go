package run

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func RunCmd(a *app.App) *cobra.Command {
	var reference, target string
	cmd := &cobra.Command{
		Use: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.Run(context.Background(), reference, target)
		},
	}
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "artifact reference")
	cmd.Flags().StringVarP(&target, "target", "t", "", "make target")
	return cmd
}
