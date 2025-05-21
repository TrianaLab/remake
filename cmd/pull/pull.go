// cmd/pull/pull.go
package pull

import (
	"context"
	"fmt"
	"os"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

func PullCmd(a *app.App, cfg *config.Config) *cobra.Command {
	var noCache bool

	cmd := &cobra.Command{
		Use:     "pull <reference>",
		Short:   "Fetch a remote Makefile artifact",
		Long:    "Download the remote Makefile artifact and print its contents to stdout. Uses local cache unless --no-cache is set.",
		Example: "  remake pull ghcr.io/myorg/myrepo:latest\n  remake pull ghcr.io/myorg/myrepo:latest --no-cache",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			// seteo de bypass de cache
			cfg.NoCache = noCache
			// obtenemos ruta local (o descargamos)
			path, err := a.Pull(context.Background(), ref)
			if err != nil {
				return err
			}
			// leemos y volcamos contenido
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
			return nil
		},
	}

	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Bypass local cache")
	return cmd
}
