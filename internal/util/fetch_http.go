package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
)

// HTTPFetcher retrieves HTTP-hosted Makefiles, honoring cache if enabled.
type HTTPFetcher struct{}

func (h *HTTPFetcher) Fetch(url string, useCache bool) (string, error) {
	// First, consult cache
	cacheFetcher := &CacheFetcher{}
	if path, err := cacheFetcher.Fetch(url, useCache); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	// Not cached or cache disabled: proceed with download
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))
	dest := filepath.Join(config.GetCacheDir(), hash+".mk")

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d fetching %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	return dest, nil
}
