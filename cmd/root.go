package cmd

import (
	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/cmd/login"
	"github.com/TrianaLab/remake/cmd/pull"
	"github.com/TrianaLab/remake/cmd/push"
	"github.com/TrianaLab/remake/cmd/run"
	"github.com/TrianaLab/remake/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "remake",
	Short: "Remake CLI",
}

func Execute(a *app.App) error {
	rootCmd.AddCommand(
		login.LoginCmd(a),
		push.PushCmd(a),
		pull.PullCmd(a),
		run.RunCmd(a),
		version.VersionCmd(),
	)
	return rootCmd.Execute()
}
