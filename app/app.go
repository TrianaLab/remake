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
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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

// App is the main application object that coordinates
// interaction with the artifact store and the process runner.
type App struct {
	store  store.Store    // store handles artifact push/pull and caching
	runner run.Runner     // runner executes Makefile targets
	Cfg    *config.Config // Cfg holds global configuration settings
}

// New creates a new App given a configuration.
// It initializes the underlying store and runner.
func New(cfg *config.Config) *App {
	return &App{
		store:  store.New(cfg),
		runner: run.New(cfg),
		Cfg:    cfg,
	}
}

// Login authenticates against the specified OCI registry.
// It tries flags first, then configuration, and finally prompts
// interactively for missing username or password.
func (a *App) Login(ctx context.Context, registry, user, pass string) error {
	// Load credentials from config if not provided
	if user == "" || pass == "" {
		key := config.NormalizeKey(registry)
		if user == "" {
			user = viper.GetString("registries." + key + ".username")
		}
		if pass == "" {
			pass = viper.GetString("registries." + key + ".password")
		}
	}

	// Prompt for username if still empty
	if user == "" {
		fmt.Fprint(os.Stderr, "Username: ")
		fmt.Scanln(&user)
	}

	// Prompt for password if still empty
	if pass == "" {
		fmt.Fprint(os.Stderr, "Password: ")
		bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		pass = string(bytePass)
	}

	// Perform login and report result
	if err := a.store.Login(ctx, registry, user, pass); err != nil {
		return err
	}
	fmt.Println("Login succeeded ✅")
	return nil
}

// Push uploads a local Makefile artifact to the given OCI reference.
// reference should be in the form "registry/repo:tag".
func (a *App) Push(ctx context.Context, reference, path string) error {
	return a.store.Push(ctx, reference, path)
}

// Pull fetches a remote Makefile artifact and prints its contents to stdout.
// It first retrieves the file from cache or, on cache miss, from the registry.
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

// Run pulls the specified Makefile (from cache or registry) and executes
// the given targets using the configured process runner.
func (a *App) Run(ctx context.Context, reference string, makeFlags, targets []string) error {
	path, err := a.store.Pull(ctx, reference)
	if err != nil {
		return err
	}
	return a.runner.Run(ctx, path, makeFlags, targets)
}
