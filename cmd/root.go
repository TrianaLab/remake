package cmd

import (
	"github.com/TrianaLab/remake/app"
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

func Execute(a *app.App, cfg *config.Config) error {
	rootCmd.AddCommand(
		login.LoginCmd(a, cfg),
		push.PushCmd(a, cfg),
		pull.PullCmd(a, cfg),
		run.RunCmd(a, cfg),
		version.VersionCmd(cfg),
	)
	return rootCmd.Execute()
}
