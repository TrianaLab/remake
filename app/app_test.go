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

package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/TrianaLab/remake/config"
	"github.com/creack/pty"
	"github.com/spf13/viper"
)

// capture redirects stdout and stderr for testing.
func capture(f func()) (string, string) {
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	f()

	_ = wOut.Close()
	_ = wErr.Close()
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	os.Stdout, os.Stderr = origOut, origErr
	return string(outBytes), string(errBytes)
}

type fakeStoreArgs struct {
	loginCalled                   bool
	loginErr                      error
	registryArg, userArg, passArg string
	pushArgs                      []string
	pushErr                       error
	pullPath                      string
	pullErr                       error
}

func (f *fakeStoreArgs) Login(ctx context.Context, registry, user, pass string) error {
	f.loginCalled = true
	f.registryArg, f.userArg, f.passArg = registry, user, pass
	return f.loginErr
}

func (f *fakeStoreArgs) Push(ctx context.Context, reference, path string) error {
	f.pushArgs = []string{reference, path}
	return f.pushErr
}

func (f *fakeStoreArgs) Pull(ctx context.Context, reference string) (string, error) {
	return f.pullPath, f.pullErr
}

type fakeRunnerErr struct {
	runArgs []interface{}
	runErr  error
}

func (f *fakeRunnerErr) Run(ctx context.Context, path string, makeFlags, targets []string) error {
	f.runArgs = []interface{}{path, makeFlags, targets}
	return f.runErr
}

// TestLoginLoadsFromConfig ensures credentials load from Viper when flags are empty.
func TestLoginLoadsFromConfig(t *testing.T) {
	viper.Reset()
	key := config.NormalizeKey("myreg")
	viper.Set("registries."+key+".username", "cfgUser")
	viper.Set("registries."+key+".password", "cfgPass")

	cfg := &config.Config{}
	fs := &fakeStoreArgs{}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	out, _ := capture(func() {
		if err := app.Login(context.Background(), "myreg", "", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !fs.loginCalled {
		t.Error("expected Login to be called")
	}
	if fs.registryArg != "myreg" || fs.userArg != "cfgUser" || fs.passArg != "cfgPass" {
		t.Errorf("wrong args: got registry=%q user=%q pass=%q", fs.registryArg, fs.userArg, fs.passArg)
	}
	if out != "Login succeeded ✅\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

// TestLoginErrorPropagates ensures Login returns store errors.
func TestLoginErrorPropagates(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{loginErr: errors.New("login failed")}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	_, _ = capture(func() {
		if err := app.Login(context.Background(), "reg", "u", "p"); err == nil {
			t.Fatal("expected error")
		}
	})
}

// TestPushError ensures Push returns store errors.
func TestPushError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{pushErr: errors.New("push fail")}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	if err := app.Push(context.Background(), "ref", "path"); err == nil || err.Error() != "push fail" {
		t.Fatalf("expected push fail, got %v", err)
	}
}

// TestPullStoreError ensures Pull returns store errors.
func TestPullStoreError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{pullErr: errors.New("pull fail")}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	if err := app.Pull(context.Background(), "ref"); err == nil || err.Error() != "pull fail" {
		t.Fatalf("expected pull fail, got %v", err)
	}
}

// TestPullReadFileError ensures Pull returns file-read errors.
func TestPullReadFileError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{pullPath: "/non/existent/file"}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	if err := app.Pull(context.Background(), "ref"); err == nil {
		t.Fatal("expected read file error")
	}
}

// TestPullPrintsFileContent ensures Pull prints the file contents to stdout.
func TestPullPrintsFileContent(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "testfile.txt")
	content := "hello world"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	cfg := &config.Config{}
	fs := &fakeStoreArgs{pullPath: tmpFile}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	out, _ := capture(func() {
		if err := app.Pull(context.Background(), "ref"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if out != content {
		t.Errorf("unexpected output: %q", out)
	}
}

// TestRunPullError ensures Run returns errors from Pull.
func TestRunPullError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{pullErr: errors.New("pull err")}
	fr := &fakeRunnerErr{}
	app := &App{store: fs, runner: fr, Cfg: cfg}

	if err := app.Run(context.Background(), "ref", nil, []string{"t"}); err == nil || err.Error() != "pull err" {
		t.Fatalf("expected pull err, got %v", err)
	}
}

// TestRunRunnerError ensures Run returns errors from runner.Run.
func TestRunRunnerError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{pullPath: "Makefile"}
	fr := &fakeRunnerErr{runErr: errors.New("run err")}
	app := &App{store: fs, runner: fr, Cfg: cfg}

	if err := app.Run(context.Background(), "ref", []string{"-j4"}, []string{"build"}); err == nil || err.Error() != "run err" {
		t.Fatalf("expected run err, got %v", err)
	}
}

// TestNewInitializesFields ensures New sets up store, runner, and config.
func TestNewInitializesFields(t *testing.T) {
	cfg := &config.Config{}
	app := New(cfg)

	if app.Cfg != cfg {
		t.Error("expected Cfg to be the provided config")
	}
	if app.store == nil {
		t.Error("expected store to be initialized")
	}
	if app.runner == nil {
		t.Error("expected runner to be initialized")
	}
}

// TestLoginPromptUsernameSuccess covers interactive username prompt.
func TestLoginPromptUsernameSuccess(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	// simulate user input for username
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("inputUser\n")
	_ = w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, errOut := capture(func() {
		if err := app.Login(context.Background(), "reg", "", "passValue"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !fs.loginCalled {
		t.Error("expected Login to be called")
	}
	if fs.userArg != "inputUser" || fs.passArg != "passValue" {
		t.Errorf("wrong credentials: got user=%q pass=%q", fs.userArg, fs.passArg)
	}
	if errOut != "Username: " {
		t.Errorf("unexpected stderr: %q", errOut)
	}
}

// TestLoginPromptUsernameError covers Scanln failure.
func TestLoginPromptUsernameError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	// empty stdin to force Scanln error
	r, w, _ := os.Pipe()
	_ = w.Close()
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, errOut := capture(func() {
		if err := app.Login(context.Background(), "reg", "", "pass"); err == nil {
			t.Fatal("expected error from Scanln")
		}
	})

	if errOut != "Username: " {
		t.Errorf("unexpected stderr: %q", errOut)
	}
}

// TestLoginPromptPasswordSuccess covers interactive password prompt using a pty.
func TestLoginPromptPasswordSuccess(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	// skip username prompt by providing user
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("failed to open pty: %v", err)
	}
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()

	// write password to master
	go func() {
		_, _ = master.Write([]byte("secretPass\n"))
	}()

	origStdin := os.Stdin
	os.Stdin = slave
	defer func() { os.Stdin = origStdin }()

	_, errOut := capture(func() {
		if err := app.Login(context.Background(), "reg", "userValue", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !fs.loginCalled {
		t.Error("expected Login to be called")
	}
	if fs.userArg != "userValue" || fs.passArg != "secretPass" {
		t.Errorf("wrong credentials: got user=%q pass=%q", fs.userArg, fs.passArg)
	}
	if errOut != "Password: \n" {
		t.Errorf("unexpected stderr: %q", errOut)
	}
}

// TestLoginPromptPasswordError covers ReadPassword failure.
func TestLoginPromptPasswordError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStoreArgs{}
	app := &App{store: fs, runner: &fakeRunnerErr{}, Cfg: cfg}

	// skip username prompt
	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	_ = r.Close()
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	_, _ = capture(func() {
		if err := app.Login(context.Background(), "reg", "u", ""); err == nil {
			t.Fatal("expected error from ReadPassword")
		}
	})
}
