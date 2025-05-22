package cmd

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func pushCmd(app *app.App) *cobra.Command {
	var (
		noCache bool
		file    string
	)

	cmd := &cobra.Command{
		Use:     "push <reference>",
		Short:   "Upload a local Makefile artifact",
		Long:    "Push a Makefile (default 'makefile') to the specified OCI reference. Uses cache unless --no-cache is set.",
		Example: "  remake push ghcr.io/myorg/myrepo:latest\n  remake push ghcr.io/myorg/myrepo:latest --file=Makefile.dev --no-cache",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			app.Cfg.NoCache = noCache
			return app.Push(context.Background(), ref, file)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	cmd.Flags().StringVarP(&file, "file", "f", "makefile", "Local Makefile to push")
	return cmd
}
