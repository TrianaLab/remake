package run

import (
	"context"
	"os"
	"os/exec"

	"github.com/TrianaLab/remake/config"
)

type Runner interface {
	Run(ctx context.Context, path string, makeFlags, targets []string) error
}

type ExecRunner struct {
	cfg *config.Config
}

func New(cfg *config.Config) Runner {
	return &ExecRunner{cfg: cfg}
}

func (r *ExecRunner) Run(ctx context.Context, path string, makeFlags, targets []string) error {
	args := []string{"-f", path}
	args = append(args, makeFlags...)
	args = append(args, targets...)
	cmd := exec.CommandContext(ctx, "make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
