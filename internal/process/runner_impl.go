package process

import (
	"context"
	"os"
	"os/exec"
)

type ExecRunner struct{}

func NewExecRunner() Runner {
	return &ExecRunner{}
}

func (r *ExecRunner) Run(ctx context.Context, path string, targets []string) error {
	cmd := exec.CommandContext(ctx, "make", targets...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
