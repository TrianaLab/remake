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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitAndSaveConfig(t *testing.T) {
	tmpHome := os.TempDir() + "/homecfg"
	os.Setenv("HOME", tmpHome)
	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig error: %v", err)
	}
	if cfg.BaseDir != filepath.Join(tmpHome, ".remake") {
		t.Errorf("unexpected BaseDir: %s", cfg.BaseDir)
	}
	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}
}

func TestParseReference(t *testing.T) {
	tmp := os.TempDir() + "/file.txt"
	os.WriteFile(tmp, []byte("x"), 0o644)
	defer os.Remove(tmp)

	cfg := &Config{}
	if cfg.ParseReference("http://example.com") != ReferenceHTTP {
		t.Error("expected HTTP")
	}
	if cfg.ParseReference(tmp) != ReferenceLocal {
		t.Error("expected Local")
	}
	if cfg.ParseReference("repo:tag") != ReferenceOCI {
		t.Error("expected OCI")
	}
}

func TestNormalizeKey(t *testing.T) {
	out := NormalizeKey("a.b.c")
	if out != "a_b_c" {
		t.Errorf("unexpected normalize: %s", out)
	}
}
