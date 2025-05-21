package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
	return cmd
}

var Version = "dev"
