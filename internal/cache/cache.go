package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/google/go-containerregistry/pkg/name"
)

type CacheRepository interface {
	Push(ctx context.Context, reference string, data []byte) error
	Pull(ctx context.Context, reference string) (string, error)
}

type LocalCache struct {
	cfg *config.Config
}

func New(cfg *config.Config) CacheRepository {
	return &LocalCache{cfg: cfg}
}

func (c *LocalCache) Push(ctx context.Context, reference string, data []byte) error {
	cacheFile, err := c.cachePath(c.cfg.CacheDir, reference)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cacheFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cacheFile, data, 0o644)
}

func (c *LocalCache) Pull(ctx context.Context, reference string) (string, error) {
	cacheFile, err := c.cachePath(c.cfg.CacheDir, reference)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(cacheFile); err != nil {
		return "", err
	}
	return cacheFile, nil
}

func (c *LocalCache) cachePath(base, reference string) (string, error) {
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
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
