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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

// HTTPCache implements CacheRepository for HTTP(S) references.
// It stores blobs under a directory structure based on the URL host and path,
// using SHA-256 digests for content addressing and symbolic links for references.
type HTTPCache struct {
	cfg *config.Config
}

// NewHTTPCache returns a new HTTPCache configured with the given settings.
func NewHTTPCache(cfg *config.Config) CacheRepository {
	return &HTTPCache{cfg: cfg}
}

// Push stores the given data bytes at a cache path derived from the reference URL.
// It computes a SHA-256 digest, writes the blob under 'cacheDir/host/.../blobs',
// and creates or updates a 'latest' symlink under 'cacheDir/host/.../refs'.
func (c *HTTPCache) Push(ctx context.Context, reference string, data []byte) error {
	u, err := url.Parse(reference)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	baseElems := append([]string{c.cfg.CacheDir, u.Host}, segments...)

	blobDir := filepath.Join(append(baseElems, "blobs")...)
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		return err
	}
	blobPath := filepath.Join(blobDir, digest)
	tmp := blobPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, blobPath); err != nil {
		return err
	}

	refDir := filepath.Join(append(baseElems, "refs")...)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return err
	}
	latest := filepath.Join(refDir, "latest")
	if err := os.Remove(latest); err != nil {
		return err
	}
	if err := os.Symlink(blobPath, latest); err != nil {
		return err
	}
	return nil
}

// Pull retrieves the cached path for the given reference URL.
// It reads the 'latest' symlink under 'cacheDir/host/.../refs' and returns its target.
// Returns an error if the cache entry does not exist or is invalid.
func (c *HTTPCache) Pull(ctx context.Context, reference string) (string, error) {
	u, err := url.Parse(reference)
	if err != nil {
		return "", err
	}
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	baseElems := append([]string{c.cfg.CacheDir, u.Host}, segments...)

	link := filepath.Join(append(append(baseElems, "refs"), "latest")...)
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("cache miss for %s", reference)
	}
	target, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	return target, nil
}
