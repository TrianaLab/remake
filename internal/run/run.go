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

package run

import (
	"context"
	"os"
	"os/exec"

	"github.com/TrianaLab/remake/config"
)

// Runner defines the interface for executing Makefile targets.
// Implementations of Runner take a path to a Makefile, flags for make,
// and target names to execute.
type Runner interface {
	// Run executes the provided make targets using the Makefile at path.
	// makeFlags are passed to the make command, and targets specify which targets to run.
	Run(ctx context.Context, path string, makeFlags, targets []string) error
}

// ExecRunner is the default Runner implementation that invokes the system 'make' command.
// It constructs the command line to include the '-f' flag for the Makefile path, any
// additional flags, and the specified targets. Standard output and error are forwarded.
type ExecRunner struct {
	cfg *config.Config
}

// New returns a Runner implementation based on the given configuration.
// Currently, it returns an ExecRunner that runs the 'make' binary.
func New(cfg *config.Config) Runner {
	return &ExecRunner{cfg: cfg}
}

// Run executes the make command with the specified Makefile path, flags, and targets.
// It builds arguments as: make -f <path> <makeFlags...> <targets...>.
// The command's stdout and stderr are connected to the current process.
func (r *ExecRunner) Run(ctx context.Context, path string, makeFlags, targets []string) error {
	// Build make command arguments
	args := []string{"-f", path}
	args = append(args, makeFlags...)
	args = append(args, targets...)

	// Execute 'make' with context
	cmd := exec.CommandContext(ctx, "make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
