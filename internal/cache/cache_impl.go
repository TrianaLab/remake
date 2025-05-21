package cache

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/google/go-containerregistry/pkg/name"
)

// LocalCache implements CacheRepository using filesystem under cacheDir
type LocalCache struct {
	cfg *config.Config
}

// NewLocalCache returns a filesystem cache, using cacheDir from configuration
func NewLocalCache(cfg *config.Config) CacheRepository {
	return &LocalCache{cfg: cfg}
}

// Push stores the referenced file into cache
func (c *LocalCache) Push(ctx context.Context, reference, path string) error {
	cacheFile, err := cachePath(c.cfg.CacheDir, reference)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o755); err != nil {
		return err
	}
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

// Pull retrieves the cached file path for the reference
func (c *LocalCache) Pull(ctx context.Context, reference string) (string, error) {
	cacheFile, err := cachePath(c.cfg.CacheDir, reference)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(cacheFile); err != nil {
		return "", fmt.Errorf("cache miss for %s: %w", reference, err)
	}
	return cacheFile, nil
}

func cachePath(base, reference string) (string, error) {
	ref, err := name.ParseReference(reference)
	if err != nil {
		return "", fmt.Errorf("invalid reference %s: %w", reference, err)
	}
	domain := ref.Context().RegistryStr()
	repo := ref.Context().RepositoryStr()
	var tagOrDigest string
	if tagged, ok := ref.(name.Tag); ok {
		tagOrDigest = tagged.TagStr()
	} else if digested, ok := ref.(name.Digest); ok {
		tagOrDigest = digested.DigestStr()
	}
	return filepath.Join(base, domain, repo, tagOrDigest), nil
}
