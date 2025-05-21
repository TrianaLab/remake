package app

import (
	"context"
	"fmt"
	"os"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/store"
)

type App struct {
	store  store.Store
	runner run.Runner
	cfg    *config.Config
}

func New(cfg *config.Config) *App {
	return &App{store: store.New(cfg), runner: run.New(cfg), cfg: cfg}
}

func (a *App) Login(ctx context.Context, registry, user, pass string) error {
	return a.store.Login(ctx, registry, user, pass)
}

func (a *App) Push(ctx context.Context, reference, path string) error {
	return a.store.Push(ctx, reference, path)
}

func (a *App) Pull(ctx context.Context, reference string) error {
	path, err := a.store.Pull(ctx, reference)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func (a *App) Run(ctx context.Context, reference string, targets []string) error {
	path, err := a.store.Pull(ctx, reference)
	if err != nil {
		return err
	}
	return a.runner.Run(ctx, path, targets)
}
