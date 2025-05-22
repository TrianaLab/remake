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
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	domain := ref.Context().RegistryStr()
	repo := ref.Context().RepositoryStr()

	blobDir := filepath.Join(c.cfg.CacheDir, domain, repo, "blobs")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		return err
	}

	blobPath := filepath.Join(blobDir, digest)
	tmpPath := blobPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmpPath, blobPath); err != nil {
		return err
	}

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

func (c *LocalCache) Pull(ctx context.Context, reference string) (string, error) {
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return "", err
	}
	domain := ref.Context().RegistryStr()
	repo := ref.Context().RepositoryStr()

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

	if dig, ok := ref.(name.Digest); ok {
		blobPath := filepath.Join(c.cfg.CacheDir, domain, repo, "blobs", dig.DigestStr())
		if _, err := os.Stat(blobPath); err == nil {
			return blobPath, nil
		}
	}

	return "", fmt.Errorf("cache miss for %s", reference)
}
