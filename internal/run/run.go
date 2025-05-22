package run

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
)

type Runner interface {
	Run(ctx context.Context, path string, targets []string) error
}

type ExecRunner struct {
	cfg *config.Config
}

func New(cfg *config.Config) Runner {
	return &ExecRunner{cfg: cfg}
}

func (r *ExecRunner) Run(ctx context.Context, path string, targets []string) error {
	cmd := exec.CommandContext(ctx, "make", "-f", path)
	cmd.Args = append(cmd.Args, targets...)
	cmd.Dir = filepath.Dir(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
