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
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/cache"
	"github.com/TrianaLab/remake/internal/client"
)

type fakeClient struct {
	pullFunc func(ctx context.Context, reference string) ([]byte, error)
	pushFunc func(ctx context.Context, reference, path string) error
}

func (f *fakeClient) Login(ctx context.Context, registry, user, pass string) error {
	return nil
}

func (f *fakeClient) Push(ctx context.Context, reference, path string) error {
	return f.pushFunc(ctx, reference, path)
}

func (f *fakeClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	return f.pullFunc(ctx, reference)
}

type fakeCache struct {
	pushFunc func(ctx context.Context, reference string, data []byte) error
	pullFunc func(ctx context.Context, reference string) (string, error)
}

func (f *fakeCache) Push(ctx context.Context, reference string, data []byte) error {
	return f.pushFunc(ctx, reference, data)
}

func (f *fakeCache) Pull(ctx context.Context, reference string) (string, error) {
	return f.pullFunc(ctx, reference)
}

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

func TestStorePushOCICacheSuccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mk*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("all:\n\techo ok")
	tmpFile.Close()

	cfg := &config.Config{DefaultRegistry: "ghcr.io"}
	s := &ArtifactStore{cfg: cfg}
	newClient = func(cfg *config.Config, ref string) client.Client {
		return &fakeClient{
			pushFunc: func(ctx context.Context, reference, path string) error {
				return nil
			},
		}
	}
	newCache = func(cfg *config.Config, reference string) cache.CacheRepository {
		return &fakeCache{
			pushFunc: func(ctx context.Context, reference string, data []byte) error {
				if !bytes.Contains(data, []byte("echo ok")) {
					t.Error("data content mismatch")
				}
				return nil
			},
		}
	}
	err = s.Push(context.Background(), "oci://ghcr.io/test/repo:latest", tmpFile.Name())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStorePullCacheHit(t *testing.T) {
	cfg := &config.Config{NoCache: false}
	s := &ArtifactStore{cfg: cfg}
	expectedPath := "/tmp/dummy.mk"

	newCache = func(cfg *config.Config, reference string) cache.CacheRepository {
		return &fakeCache{
			pullFunc: func(ctx context.Context, reference string) (string, error) {
				return expectedPath, nil
			},
		}
	}
	path, err := s.Pull(context.Background(), "oci://ghcr.io/test/repo:latest")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}
}

func TestStorePullClientError(t *testing.T) {
	cfg := &config.Config{NoCache: true}
	s := &ArtifactStore{cfg: cfg}
	newCache = func(cfg *config.Config, reference string) cache.CacheRepository {
		return &fakeCache{}
	}
	newClient = func(cfg *config.Config, reference string) client.Client {
		return &fakeClient{
			pullFunc: func(ctx context.Context, reference string) ([]byte, error) {
				return nil, errors.New("pull error")
			},
		}
	}
	_, err := s.Pull(context.Background(), "oci://ghcr.io/test/repo:latest")
	if err == nil || err.Error() != "pull error" {
		t.Errorf("expected pull error, got %v", err)
	}
}

func TestStorePullCachePushError(t *testing.T) {
	cfg := &config.Config{NoCache: true}
	s := &ArtifactStore{cfg: cfg}
	newClient = func(cfg *config.Config, reference string) client.Client {
		return &fakeClient{
			pullFunc: func(ctx context.Context, reference string) ([]byte, error) {
				return []byte("data"), nil
			},
		}
	}
	newCache = func(cfg *config.Config, reference string) cache.CacheRepository {
		return &fakeCache{
			pushFunc: func(ctx context.Context, reference string, data []byte) error {
				return errors.New("cache push error")
			},
		}
	}
	_, err := s.Pull(context.Background(), "oci://ghcr.io/test/repo:latest")
	if err == nil || err.Error() != "cache push error" {
		t.Errorf("expected cache push error, got %v", err)
	}
}
