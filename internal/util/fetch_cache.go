package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

// CacheFetcher checks for a cached artifact without fetching it.
// If useCache is false, it always returns "" and nil.
// If the artifact is found in cache, returns its local path, otherwise "" and nil.
// On real errors, returns "" and the error.
type CacheFetcher struct{}

func (c *CacheFetcher) Fetch(ref string, useCache bool) (string, error) {
	if !useCache {
		return "", nil
	}

	// Normalize reference to cache path: remove scheme
	n := strings.TrimPrefix(ref, "oci://")

	// Split host and rest
	parts := strings.SplitN(n, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid reference format: %s", ref)
	}
	host := parts[0]
	repoTag := parts[1]

	// Extract tag
	tag := "latest"
	if idx := strings.LastIndex(repoTag, ":"); idx != -1 {
		tag = repoTag[idx+1:]
		repoTag = repoTag[:idx]
	}

	// Build cache directory: <cacheDir>/<host>/<repo>/<tag>
	cacheDir := filepath.Join(config.GetCacheDir(), host, repoTag, tag)

	// Look for any file in that folder
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Not cached
			return "", nil
		}
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			return filepath.Join(cacheDir, e.Name()), nil
		}
	}

	// No file found
	return "", nil
}
