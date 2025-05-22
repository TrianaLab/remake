package cmd

import (
	"context"

	"github.com/TrianaLab/remake/app"
	"github.com/spf13/cobra"
)

func runCmd(app *app.App) *cobra.Command {
	var (
		noCache   bool
		file      string
		makeFlags []string
	)

	cmd := &cobra.Command{
		Use:     "run [targets...]",
		Short:   "Execute one or more Makefile targets",
		Long:    "Run given targets from a local or remote Makefile. Supports passing additional flags to make via --make-flag. Uses cache unless --no-cache is set; use --file to point to an alternate Makefile (path or OCI reference).",
		Example: "  # run default targets\n  remake run all build test\n\n  # pass make flags\n  remake run --make-flag -w --make-flag --silent all test\n\n  # specify remote Makefile\n  remake run -f ghcr.io/myorg/myrepo:latest test",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Cfg.NoCache = noCache
			return app.Run(context.Background(), file, makeFlags, args)
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	cmd.Flags().StringVarP(&file, "file", "f", "makefile", "Makefile (local path or OCI reference) to use")
	cmd.Flags().StringArrayVar(&makeFlags, "make-flag", nil, "Flags to pass through to make")
	return cmd
}
