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
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/google/go-containerregistry/pkg/name"
)

// OCIRepository implements CacheRepository for OCI artifact references.
// It stores blobs under a directory structure based on registry and repository,
// uses SHA-256 digests for content addressing, and symlinks tags and digests under refs.
type OCIRepository struct {
	cfg *config.Config
}

// NewOCIRepository returns a new OCIRepository using the provided configuration.
func NewOCIRepository(cfg *config.Config) CacheRepository {
	return &OCIRepository{cfg: cfg}
}

// Push caches the given data bytes as an OCI blob and creates a tag reference.
// It parses the reference (with default registry), computes SHA-256 digest,
// writes the blob under 'cacheDir/registry/repo/blobs', and symlinks under 'refs/<tag|digest>'.
func (c *OCIRepository) Push(ctx context.Context, reference string, data []byte) error {
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	domain := ref.Context().RegistryStr()
	repo := ref.Context().RepositoryStr()

	// Ensure blob directory exists
	blobDir := filepath.Join(c.cfg.CacheDir, domain, repo, "blobs")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		return err
	}

	// Write data to temporary file then rename to digest name
	blobPath := filepath.Join(blobDir, digest)
	tmp := blobPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmp, blobPath); err != nil {
		return err
	}

	// Create or update reference symlink under refs directory
	refDir := filepath.Join(c.cfg.CacheDir, domain, repo, "refs")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return err
	}
	tagPath := filepath.Join(refDir, ref.Identifier())
	os.Remove(tagPath)
	if err := os.Symlink(blobPath, tagPath); err != nil {
		return err
	}
	return nil
}

// Pull retrieves a cached artifact path for the given OCI reference.
// It looks for a symlink under 'refs/<identifier>' first. If missing and the reference
// is a digest, it checks the blob directly. Returns an error on cache miss.
func (c *OCIRepository) Pull(ctx context.Context, reference string) (string, error) {
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return "", err
	}
	domain := ref.Context().RegistryStr()
	repo := ref.Context().RepositoryStr()

	// Attempt to resolve tag or digest symlink under refs
	tagPath := filepath.Join(c.cfg.CacheDir, domain, repo, "refs", ref.Identifier())
	if info, err := os.Lstat(tagPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(tagPath)
			if err != nil {
				return "", err
			}
			return target, nil
		}
		return tagPath, nil
	}

	// If reference is a digest and no symlink, check blob path directly
	if dig, ok := ref.(name.Digest); ok {
		blobPath := filepath.Join(c.cfg.CacheDir, domain, repo, "blobs", dig.DigestStr())
		if _, err := os.Stat(blobPath); err == nil {
			return blobPath, nil
		}
	}

	return "", fmt.Errorf("cache miss for %s", reference)
}
