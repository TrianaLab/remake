package cmd

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func pullCmd(app *app.App) *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:     "pull <reference>",
		Short:   "Fetch a remote Makefile artifact",
		Long:    "Download the remote Makefile artifact and print its contents to stdout. Uses local cache unless --no-cache is set.",
		Example: "  remake pull ghcr.io/myorg/myrepo:latest\n  remake pull myorg/myrepo\n  remake pull ghcr.io/myorg/myrepo:latest --no-cache",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			app.Cfg.NoCache = noCache
			return app.Pull(context.Background(), ref)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	return cmd
}
