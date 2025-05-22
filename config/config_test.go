// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software are
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS be LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// THE SOFTWARE.

package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// capture redirects stdout and stderr for testing.
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

// TestInitConfigCreatesDefaultFile verifies that InitConfig creates the config file
// when none exists and applies defaults.
func TestInitConfigCreatesDefaultFile(t *testing.T) {
	viper.Reset()

	tmpHome := filepath.Join(os.TempDir(), "homecfg_create")
	os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpHome)

	os.Setenv("HOME", tmpHome)
	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig error: %v", err)
	}
	expectedBase := filepath.Join(tmpHome, ".remake")
	if cfg.BaseDir != expectedBase {
		t.Errorf("unexpected BaseDir: got %q want %q", cfg.BaseDir, expectedBase)
	}
	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	// SaveConfig should succeed
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}
}

// TestInitConfigUsesExistingFile ensures InitConfig reads an existing file
// without overwriting and PrintConfig outputs its contents.
func TestInitConfigUsesExistingFile(t *testing.T) {
	viper.Reset()

	tmpHome := filepath.Join(os.TempDir(), "homecfg_existing")
	os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpHome)

	os.Setenv("HOME", tmpHome)
	// First call creates defaults and file
	cfg1, err := InitConfig()
	if err != nil {
		t.Fatalf("first InitConfig error: %v", err)
	}
	// Write valid YAML sentinel to existing config file
	sentinel := "baseDir: sentinel\nregistries:\n  foo: bar\n"
	if err := os.WriteFile(cfg1.ConfigFile, []byte(sentinel), 0o644); err != nil {
		t.Fatalf("failed to write sentinel: %v", err)
	}
	// Second call should read existing file without overwrite
	cfg2, err := InitConfig()
	if err != nil {
		t.Fatalf("second InitConfig error: %v", err)
	}
	// PrintConfig should show sentinel exactly
	out, _ := capture(func() {
		if err := cfg2.PrintConfig(); err != nil {
			t.Fatalf("PrintConfig error: %v", err)
		}
	})
	if out != sentinel {
		t.Errorf("unexpected PrintConfig output: got %q want %q", out, sentinel)
	}
}

// TestInitConfigMkdirAllError covers MkdirAll failure in InitConfig.
func TestInitConfigMkdirAllError(t *testing.T) {
	viper.Reset()

	tmp := filepath.Join(os.TempDir(), "home_ro")
	os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o555); err != nil {
		t.Fatalf("failed to create tmp home: %v", err)
	}
	defer os.RemoveAll(tmp)

	os.Setenv("HOME", tmp)
	if _, err := InitConfig(); err == nil {
		t.Fatal("expected error from MkdirAll")
	}
}

// TestInitConfigWriteError covers WriteConfigAs failure in InitConfig.
func TestInitConfigWriteError(t *testing.T) {
	viper.Reset()

	tmpHome := filepath.Join(os.TempDir(), "homecfg_writeerr")
	os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpHome)

	os.Setenv("HOME", tmpHome)
	baseDir := filepath.Join(tmpHome, ".remake")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("failed to create baseDir: %v", err)
	}
	// make baseDir read-only to force WriteConfigAs error
	if err := os.Chmod(baseDir, 0o555); err != nil {
		t.Fatalf("failed to chmod baseDir: %v", err)
	}
	if _, err := InitConfig(); err == nil {
		t.Fatal("expected error from WriteConfigAs")
	}
}

// TestInitConfigReadConfigError covers ReadInConfig failure by making the file unreadable.
func TestInitConfigReadConfigError(t *testing.T) {
	viper.Reset()

	tmpHome := filepath.Join(os.TempDir(), "homecfg_readerr")
	os.RemoveAll(tmpHome)
	defer os.RemoveAll(tmpHome)

	os.Setenv("HOME", tmpHome)
	// create valid first config
	cfg1, err := InitConfig()
	if err != nil {
		t.Fatalf("first InitConfig error: %v", err)
	}
	// make config file unreadable
	if err := os.Chmod(cfg1.ConfigFile, 0o000); err != nil {
		t.Fatalf("failed to chmod config file: %v", err)
	}
	// second InitConfig should fail at ReadInConfig
	if _, err := InitConfig(); err == nil {
		t.Fatal("expected error from ReadInConfig on unreadable file")
	}
}

// TestPrintConfigError ensures PrintConfig returns an error when the file path is invalid.
func TestPrintConfigError(t *testing.T) {
	viper.Reset()
	viper.Set("configFile", "/nonexistent/file")
	cfg := &Config{}
	if err := cfg.PrintConfig(); err == nil {
		t.Fatal("expected error from PrintConfig")
	}
}

// TestParseReference covers HTTP, absolute local, relative local, and default OCI.
func TestParseReference(t *testing.T) {
	cfg := &Config{}

	// HTTP URL
	if got := cfg.ParseReference("http://example.com"); got != ReferenceHTTP {
		t.Errorf("expected HTTP, got %v", got)
	}

	// absolute path
	abs := filepath.Join(os.TempDir(), "f.txt")
	os.WriteFile(abs, []byte("x"), 0o644)
	defer os.Remove(abs)
	if got := cfg.ParseReference(abs); got != ReferenceLocal {
		t.Errorf("expected Local for absolute path, got %v", got)
	}

	// relative path
	rel := "f_rel.txt"
	os.WriteFile(rel, []byte("y"), 0o644)
	defer os.Remove(rel)
	if got := cfg.ParseReference(rel); got != ReferenceLocal {
		t.Errorf("expected Local for relative path, got %v", got)
	}

	// missing file defaults to OCI
	if got := cfg.ParseReference("no_such_file_123"); got != ReferenceOCI {
		t.Errorf("expected OCI for missing file, got %v", got)
	}
}

// TestNormalizeKey verifies that dots are replaced by underscores.
func TestNormalizeKey(t *testing.T) {
	out := NormalizeKey("a.b.c")
	if out != "a_b_c" {
		t.Errorf("unexpected normalize: got %q want %q", out, "a_b_c")
	}
}

// TestInitConfigUserHomeDirError covers the error return path of os.UserHomeDir.
func TestInitConfigUserHomeDirError(t *testing.T) {
	viper.Reset()

	// Mock the home-dir call to always error
	orig := userHomeDir
	userHomeDir = func() (string, error) {
		return "", errors.New("cannot get home")
	}
	defer func() { userHomeDir = orig }()

	os.Setenv("HOME", "/some/where") // no effect, uso mock
	if _, err := InitConfig(); err == nil {
		t.Fatal("expected InitConfig to return error when UserHomeDir fails")
	}
}
