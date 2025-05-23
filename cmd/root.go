// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// THE SOFTWARE.

package cmd

import (
	"log"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
)

// allow tests to override InitConfig and Fatal
var initConfigFunc = config.InitConfig
var fatalFunc = log.Fatal

// rootCmd is the base command for the Remake CLI. It initializes configuration,
// creates the application instance, and registers all subcommands.
var rootCmd = &cobra.Command{
	Use:   "remake",
	Short: "Remake CLI - wrapper for Makefiles as OCI artifacts",
	Long: `Remake is a CLI tool to package, distribute, and run Makefiles
as OCI artifacts. It supports pushing local Makefiles to container registries,
pulling cached artifacts, and executing targets locally or remotely.

Available commands:
  login    Authenticate to an OCI registry
  push     Upload a Makefile artifact
  pull     Download and display a Makefile artifact
  run      Execute Makefile targets
  version  Show the CLI version
  config   Display current CLI configuration`,
	Example: `  # Display help for all commands
  remake --help

  # Authenticate to GitHub Container Registry
  remake login ghcr.io

  # Push a Makefile artifact and then run a target
  remake push ghcr.io/myorg/myrepo:latest
  remake run ghcr.io/myorg/myrepo:latest build`,
}

// Execute initializes configuration and executes the root command.
// It loads settings, creates the application, registers subcommands,
// silences usage output on error, and runs the Cobra command tree.
func Execute() error {
	cfg, err := initConfigFunc()
	if err != nil {
		fatalFunc(err)
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

	// Prevent Cobra from printing usage on error
	rootCmd.SilenceUsage = true

	return rootCmd.Execute()
}
