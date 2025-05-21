package push

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func PushCmd(a *app.App) *cobra.Command {
	var reference, path string
	cmd := &cobra.Command{
		Use: "push",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.Push(context.Background(), reference, path)
		},
	}
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "artifact reference")
	cmd.Flags().StringVarP(&path, "path", "p", "", "artifact path")
	return cmd
}
