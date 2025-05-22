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
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmp, blobPath); err != nil {
		return err
	}

	refDir := filepath.Join(append(baseElems, "refs")...)
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		return err
	}
	latest := filepath.Join(refDir, "latest")
	os.Remove(latest)
	if err := os.Symlink(blobPath, latest); err != nil {
		return err
	}
	return nil
}

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
