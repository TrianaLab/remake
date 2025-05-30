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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fakeStore implements store.Store for testing commands
// It stubs Login, Push, Pull.
type fakeStore struct {
	loginErr error
	pushErr  error
	pullErr  error
	pullPath string
}

func (f *fakeStore) Login(ctx context.Context, registry, user, pass string) error {
	return f.loginErr
}

func (f *fakeStore) Push(ctx context.Context, reference, path string) error {
	return f.pushErr
}

func (f *fakeStore) Pull(ctx context.Context, reference string) (string, error) {
	return f.pullPath, f.pullErr
}

// fakeRunner implements run.Runner for testing
// Captures invocation details.
type fakeRunner struct {
	executed  bool
	err       error
	path      string
	makeFlags []string
	targets   []string
}

func (f *fakeRunner) Run(ctx context.Context, path string, makeFlags []string, targets []string) error {
	f.executed = true
	f.path = path
	f.makeFlags = makeFlags
	f.targets = targets
	return f.err
}

// setUnexportedField injects a value into an unexported field via reflection.
func setUnexportedField(obj interface{}, name string, value interface{}) {
	v := reflect.ValueOf(obj).Elem()
	f := v.FieldByName(name)
	rv := reflect.ValueOf(value)
	ptr := unsafe.Pointer(f.UnsafeAddr())
	fv := reflect.NewAt(f.Type(), ptr).Elem()
	fv.Set(rv)
}

// captureCmdOutput captures stdout for any Cobra command.
func captureCmdOutput(cmd *cobra.Command, args []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.SetArgs(args)
	err := cmd.Execute()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out), err
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func TestVersionCmd(t *testing.T) {
	cfg := &config.Config{Version: "v1.2.3"}
	a := app.New(cfg)
	c := versionCmd(a)
	out, err := captureCmdOutput(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Remake version: v1.2.3\n" {
		t.Errorf("expected version 'Remake version: v1.2.3', got %q", out)
	}
}

func TestConfigCmd(t *testing.T) {
	cfg, err := config.InitConfig()
	if err != nil {
		t.Fatalf("InitConfig error: %v", err)
	}
	a := app.New(cfg)
	c := configCmd(a)
	out, err := captureCmdOutput(c, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty config output")
	}
}

func TestLoginCmdWithFlags(t *testing.T) {
	cfg, _ := config.InitConfig()
	a := app.New(cfg)
	fs := &fakeStore{loginErr: nil}
	setUnexportedField(a, "store", fs)
	c := loginCmd(a)
	_, err := captureCmdOutput(c, []string{"myreg", "-u", "user", "-p", "pass"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPushCmdErrorPropagation(t *testing.T) {
	cfg, _ := config.InitConfig()
	a := app.New(cfg)
	fs := &fakeStore{pushErr: errors.New("push failed")}
	setUnexportedField(a, "store", fs)
	c := pushCmd(a)
	c.SilenceUsage = true
	c.SilenceErrors = true

	_, err := captureCmdOutput(c, []string{"ref", "-f", "path"})
	if err == nil || err.Error() != "push failed" {
		t.Fatalf("expected error 'push failed', got %v", err)
	}
}

func TestPullCmdHTTP(t *testing.T) {
	// Start test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	// Prepare temp cache dir and Viper
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	viper.SetConfigType("yaml")

	// Capture output of Execute which registers subcommands
	output := captureOutput(func() {
		// Execute with args: pull --no-cache <url>
		os.Args = []string{"remake", "pull", "--no-cache", srv.URL}
		if err := Execute(); err != nil {
			t.Fatalf("Execute error: %v", err)
		}
	})

	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestRunCmdExecution(t *testing.T) {
	tmp := os.TempDir() + "/Makefile_test"
	_ = os.WriteFile(tmp, []byte("all:\n\techo ok"), 0o644)
	defer func() { _ = os.Remove(tmp) }()

	cfg, _ := config.InitConfig()
	a := app.New(cfg)
	fs := &fakeStore{pullPath: tmp}
	fr := &fakeRunner{err: nil}
	setUnexportedField(a, "store", fs)
	setUnexportedField(a, "runner", fr)
	c := runCmd(a)
	_, err := captureCmdOutput(c, []string{"all"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fr.executed {
		t.Error("expected runner to execute")
	}
}

// TestExecute_InitConfigFatal covers the log.Fatal path in Execute()
func TestExecute_InitConfigFatal(t *testing.T) {
	// swap out initConfigFunc and fatalFunc
	origInit := initConfigFunc
	origFatal := fatalFunc
	defer func() {
		initConfigFunc = origInit
		fatalFunc = origFatal
	}()

	// make InitConfig return an error
	initConfigFunc = func() (*config.Config, error) {
		return nil, errors.New("boom init")
	}

	called := false
	var fatalErr error
	// override fatalFunc to capture its args
	fatalFunc = func(v ...any) {
		called = true
		if len(v) > 0 {
			if errArg, ok := v[0].(error); ok {
				fatalErr = errArg
			} else {
				fatalErr = fmt.Errorf("%v", v[0])
			}
		}
		// do not os.Exit in tests
	}

	// run Execute; it should call fatalFunc("boom init")
	_ = Execute()

	if !called {
		t.Fatal("expected fatalFunc to be called")
	}
	if fatalErr == nil || fatalErr.Error() != "boom init" {
		t.Fatalf("expected fatalErr \"boom init\", got %v", fatalErr)
	}
}
