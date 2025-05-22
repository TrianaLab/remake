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
	"testing"

	"github.com/TrianaLab/remake/config"
)

type fakeStore struct {
	loginCalled bool
	loginErr    error
	pushArgs    []string
	pushErr     error
	pullPath    string
	pullErr     error
}

func (f *fakeStore) Login(ctx context.Context, registry, user, pass string) error {
	f.loginCalled = true
	return f.loginErr
}

func (f *fakeStore) Push(ctx context.Context, reference, path string) error {
	f.pushArgs = []string{reference, path}
	return f.pushErr
}

func (f *fakeStore) Pull(ctx context.Context, reference string) (string, error) {
	return f.pullPath, f.pullErr
}

type fakeRunner struct {
	runArgs []interface{}
	runErr  error
}

func (f *fakeRunner) Run(ctx context.Context, path string, makeFlags, targets []string) error {
	f.runArgs = []interface{}{path, makeFlags, targets}
	return f.runErr
}

func capture(f func()) (string, string) {
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr

	f()

	wOut.Close()
	wErr.Close()
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	os.Stdout, os.Stderr = origOut, origErr
	return string(outBytes), string(errBytes)
}

func TestLoginSuccess(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStore{}
	app := &App{store: fs, runner: &fakeRunner{}, Cfg: cfg}

	out, _ := capture(func() {
		if err := app.Login(context.Background(), "reg", "user", "pass"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !fs.loginCalled {
		t.Error("expected Login to be called")
	}
	if out != "Login succeeded ✅\n" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestLoginError(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStore{loginErr: errors.New("fail")}
	app := &App{store: fs, runner: &fakeRunner{}, Cfg: cfg}

	_, _ = capture(func() {
		if err := app.Login(context.Background(), "reg", "user", "pass"); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPushDelegation(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStore{}
	app := &App{store: fs, runner: &fakeRunner{}, Cfg: cfg}

	if err := app.Push(context.Background(), "ref", "path"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fs.pushArgs) != 2 || fs.pushArgs[0] != "ref" || fs.pushArgs[1] != "path" {
		t.Errorf("unexpected push args: %v", fs.pushArgs)
	}
}

func TestPullDelegation(t *testing.T) {
	tmpFile := os.TempDir() + "/testfile.txt"
	os.WriteFile(tmpFile, []byte("hello"), 0o644)
	defer os.Remove(tmpFile)

	cfg := &config.Config{}
	fs := &fakeStore{pullPath: tmpFile}
	app := &App{store: fs, runner: &fakeRunner{}, Cfg: cfg}

	out, _ := capture(func() {
		if err := app.Pull(context.Background(), "ref"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if out != "hello" {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRunDelegation(t *testing.T) {
	cfg := &config.Config{}
	fs := &fakeStore{pullPath: "Makefile"}
	fr := &fakeRunner{}
	app := &App{store: fs, runner: fr, Cfg: cfg}

	err := app.Run(context.Background(), "ref", []string{"-j4"}, []string{"build"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.runArgs) != 3 {
		t.Error("expected runner.Run to be called with three args")
	}
}
