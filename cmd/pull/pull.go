package pull

import (
	"context"
	"fmt"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func PullCmd(a *app.App) *cobra.Command {
	var reference string
	cmd := &cobra.Command{
		Use:  "pull",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := a.Pull(context.Background(), args[0])
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "artifact reference")
	return cmd
}
