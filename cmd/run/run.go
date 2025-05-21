// cmd/run/run.go
package run

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

func RunCmd(a *app.App, cfg *config.Config) *cobra.Command {
	var (
		noCache bool
		file    string
	)

	cmd := &cobra.Command{
		Use:     "run [targets...]",
		Short:   "Execute one or more Makefile targets",
		Long:    "Run given targets from a local or remote Makefile. Uses cache unless --no-cache is set; use --file to point to an alternate Makefile (path or OCI reference).",
		Example: "  remake run all build test\n  remake run --file=ghcr.io/user/test.mk:v1 test",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.NoCache = noCache
			return a.Run(context.Background(), file, args)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	cmd.Flags().StringVarP(&file, "file", "f", "makefile", "Makefile (local path or OCI reference) to use")
	return cmd
}
