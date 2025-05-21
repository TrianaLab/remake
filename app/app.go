package app

import (
	"context"

	"github.com/TrianaLab/remake/internal/artifactstore"
	"github.com/TrianaLab/remake/internal/process"
)

type App struct {
	store  artifactstore.ArtifactStore
	runner process.Runner
}

func New(store artifactstore.ArtifactStore, runner process.Runner) *App {
	return &App{store: store, runner: runner}
}

func (a *App) Login(ctx context.Context, registry, user, pass string) error {
	return a.store.Login(ctx, registry, user, pass)
}

func (a *App) Push(ctx context.Context, reference, path string) error {
	return a.store.Push(ctx, reference, path)
}

func (a *App) Pull(ctx context.Context, reference string) (string, error) {
	return a.store.Pull(ctx, reference)
}

func (a *App) Run(ctx context.Context, reference string, targets []string) error {
	path, err := a.store.Pull(ctx, reference)
	if err != nil {
		return err
	}
	return a.runner.Run(ctx, path, targets)
}
