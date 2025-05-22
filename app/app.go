package app

import (
	"context"
	"fmt"
	"os"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/run"
	"github.com/TrianaLab/remake/internal/store"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

type App struct {
	store  store.Store
	runner run.Runner
	Cfg    *config.Config
}

func New(cfg *config.Config) *App {
	return &App{store: store.New(cfg), runner: run.New(cfg), Cfg: cfg}
}

func (a *App) Login(ctx context.Context, registry, user, pass string) error {
	if user == "" || pass == "" {
		key := config.NormalizeKey(registry)
		if user == "" {
			user = viper.GetString("registries." + key + ".username")
		}
		if pass == "" {
			pass = viper.GetString("registries." + key + ".password")
		}
	}
	if user == "" {
		fmt.Fprint(os.Stderr, "Username: ")
		fmt.Scanln(&user)
	}
	if pass == "" {
		fmt.Fprint(os.Stderr, "Password: ")
		bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		pass = string(bytePass)
	}
	err := a.store.Login(ctx, registry, user, pass)
	if err != nil {
		return err
	}
	fmt.Println("Login succeeded ✅")
	return nil
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

func (a *App) Run(ctx context.Context, reference string, makeFlags, targets []string) error {
	path, err := a.store.Pull(ctx, reference)
	if err != nil {
		return err
	}
	return a.runner.Run(ctx, path, makeFlags, targets)
}
