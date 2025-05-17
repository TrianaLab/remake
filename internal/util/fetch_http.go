package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// FetchHTTP downloads a remote Makefile via HTTP(S) and caches it under .remake/cache
func FetchHTTP(url string) (string, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))
	dest := filepath.Join(".remake", "cache", hash+".mk")

	// Return cached if exists
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	// Ensure cache dir
	os.MkdirAll(filepath.Dir(dest), 0o755)

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
