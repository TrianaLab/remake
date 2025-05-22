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

package store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
)

func TestStoreLoginHTTP(t *testing.T) {
	cfg := &config.Config{}
	s := New(cfg)
	if err := s.Login(context.Background(), "http://example.com", "user", "pass"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStorePushErrors(t *testing.T) {
	cfg := &config.Config{}
	s := New(cfg)
	// HTTP reference not supported
	if err := s.Push(context.Background(), "http://example.com", "path"); err == nil {
		t.Error("expected error for HTTP push")
	}
	// Local reference not supported
	tmp, _ := os.CreateTemp("", "f*")
	tmp.WriteString("x")
	tmp.Close()
	defer os.Remove(tmp.Name())
	if err := s.Push(context.Background(), tmp.Name(), "path"); err == nil {
		t.Error("expected error for local push")
	}
	// OCI missing file
	ref := "reg.io/repo:tag"
	if err := s.Push(context.Background(), ref, "nofile"); err == nil {
		t.Error("expected error for missing file push")
	}
}

func TestStorePullLocal(t *testing.T) {
	// Create a local file
	tmp, err := os.CreateTemp("", "f*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmp.WriteString("x")
	tmp.Close()
	defer os.Remove(tmp.Name())

	cfg := &config.Config{}
	s := New(cfg)
	path, err := s.Pull(context.Background(), tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != tmp.Name() {
		t.Errorf("expected path %q, got %q", tmp.Name(), path)
	}
}

func TestStorePullHTTP(t *testing.T) {
	// Setup HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer server.Close()

	// Prepare cache directory
	tmpDir, err := os.MkdirTemp("", "storecache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{CacheDir: tmpDir, NoCache: true}
	s := New(cfg)

	// Pre-create refs/latest symlink to allow cache push
	u, _ := url.Parse(server.URL)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{cfg.CacheDir, u.Host}, segments...)
	refDir := filepath.Join(append(base, "refs")...)
	os.MkdirAll(refDir, 0o755)
	os.Symlink("dummy", filepath.Join(refDir, "latest"))

	// Pull should fetch, cache, and return file path
	path, err := s.Pull(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read cached file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
}
