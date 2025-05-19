package fetch

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type CacheFetcher struct{}

func (c *CacheFetcher) Fetch(ref string, useCache bool) (string, error) {
	if !useCache {
		return "", nil
	}

	parsed, err := url.Parse(ref)
	if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		host := parsed.Host
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ref)))
		dir := filepath.Join(viper.GetString("cacheDir"), host)
		dest := filepath.Join(dir, hash+".mk")
		if _, err := os.Stat(dest); err == nil {
			return dest, nil
		} else if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	n := strings.TrimPrefix(ref, "oci://")
	parts := strings.SplitN(n, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid reference format: %s", ref)
	}

	host := parts[0]
	repoTag := parts[1]
	tag := "latest"
	if idx := strings.LastIndex(repoTag, ":"); idx != -1 {
		tag = repoTag[idx+1:]
		repoTag = repoTag[:idx]
	}

	cachePath := filepath.Join(viper.GetString("cacheDir"), host, repoTag, tag)
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			return filepath.Join(cachePath, e.Name()), nil
		}
	}
	return "", nil
}
