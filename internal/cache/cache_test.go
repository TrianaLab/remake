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

package cache

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
)

func TestNewCacheVariants(t *testing.T) {
	tmp, err := os.MkdirTemp("", "cache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp}
	// HTTP URL
	if NewCache(cfg, "http://example.com/path") == nil {
		t.Error("expected HTTPCache, got nil")
	}
	// OCI reference
	cfg.DefaultRegistry = "reg.io"
	if NewCache(cfg, "reg.io/repo:tag") == nil {
		t.Error("expected OCIRepository, got nil")
	}
	// Local file
	localFile := filepath.Join(tmp, "file.txt")
	os.WriteFile(localFile, []byte(""), 0o644)
	if NewCache(cfg, localFile) != nil {
		t.Error("expected no cache for local file")
	}
}

func TestHTTPCachePushPull(t *testing.T) {
	tmp, err := os.MkdirTemp("", "cachehttp")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp}
	c := NewHTTPCache(cfg)

	ref := "http://host/path/to/file"
	// Pre-create refs/latest symlink to allow Remove
	u, _ := url.Parse(ref)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{cfg.CacheDir, u.Host}, segments...)
	refDir := filepath.Join(append(base, "refs")...)
	os.MkdirAll(refDir, 0o755)
	os.Symlink("dummy", filepath.Join(refDir, "latest"))

	data := []byte("hello")
	if err := c.Push(context.Background(), ref, data); err != nil {
		t.Fatalf("Push error: %v", err)
	}

	path, err := c.Pull(context.Background(), ref)
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}
}

func TestHTTPCacheMiss(t *testing.T) {
	tmp, err := os.MkdirTemp("", "cachemiss")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp}
	c := NewHTTPCache(cfg)
	if _, err := c.Pull(context.Background(), "http://no.such"); err == nil {
		t.Error("expected cache miss error")
	}
}

func TestOCIRepositoryPushPull(t *testing.T) {
	tmp, err := os.MkdirTemp("", "cacheoci")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp, DefaultRegistry: "reg.io"}
	c := NewOCIRepository(cfg)

	ref := "reg.io/myrepo:latest"
	// Pre-create refs/latest symlink to allow Remove
	parts := strings.SplitN(ref, "/", 2)
	domain := parts[0]
	repoTag := parts[1]
	repo := strings.SplitN(repoTag, ":", 2)[0]
	refDir := filepath.Join(cfg.CacheDir, domain, repo, "refs")
	os.MkdirAll(refDir, 0o755)
	os.Symlink("dummy", filepath.Join(refDir, "latest"))

	data := []byte("data")
	if err := c.Push(context.Background(), ref, data); err != nil {
		t.Fatalf("OCI Push error: %v", err)
	}

	path, err := c.Pull(context.Background(), ref)
	if err != nil {
		t.Fatalf("OCI Pull error: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(content) != "data" {
		t.Errorf("expected 'data', got %q", string(content))
	}
}

func TestOCIRepositoryMiss(t *testing.T) {
	tmp, err := os.MkdirTemp("", "cacheoci")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp, DefaultRegistry: "reg.io"}
	c := NewOCIRepository(cfg)
	if _, err := c.Pull(context.Background(), "reg.io/none:tag"); err == nil {
		t.Error("expected cache miss error")
	}
}
