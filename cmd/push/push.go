// cmd/push/push.go
package push

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

func PushCmd(cfg *config.Config) *cobra.Command {
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
			cfg.NoCache = noCache
			return app.New(cfg).Push(context.Background(), ref, file)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	cmd.Flags().StringVarP(&file, "file", "f", "makefile", "Local Makefile to push")
	return cmd
}
