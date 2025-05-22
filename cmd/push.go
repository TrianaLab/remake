package cmd

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func pushCmd(app *app.App) *cobra.Command {
	var (
		file string
	)

	cmd := &cobra.Command{
		Use:     "push <reference>",
		Short:   "Upload a local Makefile artifact",
		Long:    "Push a Makefile (default 'makefile') to the specified OCI reference.",
		Example: "  remake push ghcr.io/myorg/myrepo:latest\n  remake push ghcr.io/myorg/myrepo:latest -f Makefile.dev",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			return app.Push(context.Background(), ref, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "makefile", "Local Makefile to push")
	return cmd
}
