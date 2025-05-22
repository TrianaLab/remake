package cmd

import (
	"log"

	"github.com/TrianaLab/remake/app"
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

	a := app.New(cfg)

	rootCmd.AddCommand(
		loginCmd(a),
		pushCmd(a),
		pullCmd(a),
		runCmd(a),
		versionCmd(a),
		configCmd(a),
	)
	return rootCmd.Execute()
}
