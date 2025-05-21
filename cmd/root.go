package cmd

import (
	"log"

	"github.com/TrianaLab/remake/cmd/login"
	"github.com/TrianaLab/remake/cmd/pull"
	"github.com/TrianaLab/remake/cmd/push"
	"github.com/TrianaLab/remake/cmd/run"
	"github.com/TrianaLab/remake/cmd/version"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "remake",
	Short: "Remake CLI",
}

func Execute() error {

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(
		login.LoginCmd(cfg),
		push.PushCmd(cfg),
		pull.PullCmd(cfg),
		run.RunCmd(cfg),
		version.VersionCmd(cfg),
	)
	return rootCmd.Execute()
}
