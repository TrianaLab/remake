package cache

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/viper"

	"github.com/TrianaLab/remake/config"
)

type LocalCache struct {
	dir string
}

func NewLocalCache(_ string) CacheRepository {
	if err := config.InitConfig(); err != nil {
		panic(err)
	}
	dir := viper.GetString("cacheDir")
	return &LocalCache{dir: dir}
}

func (c *LocalCache) Push(ctx context.Context, reference, path string) error {
	cacheFile, err := cachePath(c.dir, reference)
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

func (c *LocalCache) Pull(ctx context.Context, reference string) (string, error) {
	cacheFile, err := cachePath(c.dir, reference)
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
