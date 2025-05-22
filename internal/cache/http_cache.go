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

	"github.com/TrianaLab/remake/config"
)

type HTTPCache struct {
	cfg *config.Config
}

func NewHTTPCache(cfg *config.Config) CacheRepository {
	return &HTTPCache{cfg: cfg}
}

func (c *HTTPCache) Push(ctx context.Context, reference string, data []byte) error {
	u, err := url.Parse(reference)
	if err != nil {
		return err
	}
	encoded := url.PathEscape(reference)
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	baseDir := filepath.Join(c.cfg.CacheDir, "http", u.Host, encoded)
	blobDir := filepath.Join(baseDir, "blobs")
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
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmp, blobPath); err != nil {
		return err
	}
	refDir := filepath.Join(baseDir, "refs")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return err
	}
	link := filepath.Join(refDir, "latest")
	os.Remove(link)
	if err := os.Symlink(blobPath, link); err != nil {
		return err
	}
	return nil
}

func (c *HTTPCache) Pull(ctx context.Context, reference string) (string, error) {
	u, err := url.Parse(reference)
	if err != nil {
		return "", err
	}
	encoded := url.PathEscape(reference)
	refLink := filepath.Join(c.cfg.CacheDir, "http", u.Host, encoded, "refs", "latest")
	if info, err := os.Lstat(refLink); err == nil && info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(refLink)
		if err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("cache miss for %s", reference)
}
