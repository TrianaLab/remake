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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
)

var (
	originalCreateFile = createFile
	originalCopyData   = copyData
	origCloseFile      = closeFile
	origSymlink        = symlink
	origReadLink       = readLink
)

func restoreHTTPFactories() {
	createFile = originalCreateFile
	copyData = originalCopyData
	closeFile = origCloseFile
	symlink = origSymlink
	readLink = origReadLink
}

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

func TestHTTPCachePushInvalidURL(t *testing.T) {
	cfg := &config.Config{CacheDir: os.TempDir()}
	c := NewHTTPCache(cfg)
	err := c.Push(context.Background(), "://invalid-url", []byte("data"))
	if err == nil {
		t.Error("expected error for invalid URL on Push")
	}
}

func TestHTTPCachePullInvalidURL(t *testing.T) {
	cfg := &config.Config{CacheDir: os.TempDir()}
	c := NewHTTPCache(cfg)
	_, err := c.Pull(context.Background(), "://invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL on Pull")
	}
}

func TestHTTPCachePullNotSymlink(t *testing.T) {
	tmp, err := os.MkdirTemp("", "httptestnonsymlink")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	cfg := &config.Config{CacheDir: tmp}
	c := NewHTTPCache(cfg)
	ref := "http://example.com/some/path/file"
	u, _ := url.Parse(ref)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{cfg.CacheDir, u.Host}, segments...)
	refDir := filepath.Join(append(base, "refs")...)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatalf("failed to mkdir refs: %v", err)
	}
	latest := filepath.Join(refDir, "latest")
	// Create a regular file instead of a symlink
	if err := os.WriteFile(latest, []byte("plain"), 0o644); err != nil {
		t.Fatalf("failed to write regular file: %v", err)
	}
	_, err = c.Pull(context.Background(), ref)
	if err == nil {
		t.Error("expected cache miss error for non-symlink file")
	}
}

func TestHTTPCachePushMkdirAllBlobDirError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "notadir")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	cfg := &config.Config{CacheDir: tmpFilePath}
	c := NewHTTPCache(cfg)

	err = c.Push(context.Background(), "http://host/path", []byte("data"))
	if err == nil {
		t.Error("expected error for MkdirAll blobDir")
	}
}

func TestHTTPCachePushRemoveError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "removeerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{CacheDir: tmpDir}
	urlStr := "http://example.org/baz"
	data := []byte("123")

	u, _ := url.Parse(urlStr)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	baseElems := append([]string{tmpDir, u.Host}, segments...)

	refDir := filepath.Join(append(baseElems, "refs")...)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatalf("failed to mkdir refs: %v", err)
	}

	latest := filepath.Join(refDir, "latest")
	if err := os.MkdirAll(latest, 0o755); err != nil {
		t.Fatalf("failed to mkdir latest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(latest, "inner"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	c := NewHTTPCache(cfg)
	if err := c.Push(context.Background(), urlStr, data); err == nil {
		t.Error("expected error for os.Remove on non-empty dir")
	}
}

func TestHTTPCachePushCreateTempError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "createtmp")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{CacheDir: tmpDir}
	urlStr := "http://example.com/foo"
	data := []byte("x")

	u, _ := url.Parse(urlStr)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{tmpDir, u.Host}, segments...)

	blobDir := filepath.Join(append(base, "blobs")...)
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		t.Fatalf("cannot mkdir blobDir: %v", err)
	}

	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	tmpPath := filepath.Join(blobDir, digest+".tmp")
	if err := os.MkdirAll(tmpPath, 0o755); err != nil {
		t.Fatalf("cannot create tmp path: %v", err)
	}

	c := NewHTTPCache(cfg)
	if err := c.Push(context.Background(), urlStr, data); err == nil {
		t.Error("expected error for os.Create on temp file path")
	}
}

func TestHTTPCachePushMkdirAllRefsError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "refsdirerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{CacheDir: tmpDir}
	urlStr := "http://localhost/bar"
	data := []byte("z")

	u, _ := url.Parse(urlStr)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{tmpDir, u.Host}, segments...)

	refDir := filepath.Join(append(base, "refs")...)
	parent := filepath.Dir(refDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatalf("cannot mkdir parent: %v", err)
	}
	if err := os.WriteFile(refDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write file at refDir: %v", err)
	}

	c := NewHTTPCache(cfg)
	if err := c.Push(context.Background(), urlStr, data); err == nil {
		t.Error("expected error for MkdirAll refs")
	}
}

func TestHTTPCachePushCopyError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	copyData = func(dst io.Writer, src io.Reader) (int64, error) {
		return 0, fmt.Errorf("copy failure")
	}
	defer restoreHTTPFactories()

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)
	err = c.Push(context.Background(), "http://h/x", []byte("d"))
	if err == nil || !strings.Contains(err.Error(), "copy failure") {
		t.Errorf("expected copy failure, got %v", err)
	}
}

func TestHTTPCachePushCloseError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "closeerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	createFile = func(name string) (*os.File, error) {
		return os.NewFile(uintptr(0xffff), name), nil
	}
	defer restoreHTTPFactories()

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)
	err = c.Push(context.Background(), "http://h/x", []byte("ok"))
	if err == nil {
		t.Errorf("expected close failure, got nil")
	}
}

func TestHTTPCachePushRenameError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renameerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)

	ref := "http://host/path"
	data := []byte("zz")
	u, _ := url.Parse(ref)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	base := append([]string{tmpDir, u.Host}, segments...)

	blobDir := filepath.Join(append(base, "blobs")...)
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		t.Fatalf("mkdir blobDir: %v", err)
	}

	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	if err := os.Mkdir(filepath.Join(blobDir, digest), 0o755); err != nil {
		t.Fatalf("mkdir existing digest dir: %v", err)
	}

	err = c.Push(context.Background(), ref, data)
	if err == nil {
		t.Errorf("expected rename error, got nil")
	}
}

func TestHTTPCachePushCloseErrorBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "closeerr2")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	closeFile = func(f *os.File) error { return fmt.Errorf("close failed") }
	defer restoreHTTPFactories()

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)
	err = c.Push(context.Background(), "http://host/close", []byte("data"))
	if err == nil || !strings.Contains(err.Error(), "close failed") {
		t.Errorf("expected close failed, got %v", err)
	}
}

func TestHTTPCachePushSymlinkErrorBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlinkerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	symlink = func(oldname, newname string) error { return fmt.Errorf("symlink failed") }
	defer restoreHTTPFactories()

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)
	err = c.Push(context.Background(), "http://host/sym", []byte("data"))
	if err == nil || !strings.Contains(err.Error(), "symlink failed") {
		t.Errorf("expected symlink failed, got %v", err)
	}
}

func TestHTTPCachePullReadlinkErrorBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "readlinkerr")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ref := "http://host/link"
	u, _ := url.Parse(ref)
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	parts := append([]string{tmpDir, u.Host}, segments...)
	parts = append(parts, "refs")
	refDir := filepath.Join(parts...)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatalf("failed to mkdir refs: %v", err)
	}
	link := filepath.Join(refDir, "latest")
	if err := os.Symlink("dummy", link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	readLink = func(name string) (string, error) { return "", fmt.Errorf("readlink failed") }
	defer restoreHTTPFactories()

	cfg := &config.Config{CacheDir: tmpDir}
	c := NewHTTPCache(cfg)
	_, err = c.Pull(context.Background(), ref)
	if err == nil || !strings.Contains(err.Error(), "readlink failed") {
		t.Errorf("expected readlink failed, got %v", err)
	}
}
