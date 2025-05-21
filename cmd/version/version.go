package version

import (
	"fmt"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

func VersionCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the CLI version",
		Long:    "Prints the current version of the Remake CLI.",
		Example: "  remake version",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cfg.Version)
			return nil
		},
	}
	return cmd
}
