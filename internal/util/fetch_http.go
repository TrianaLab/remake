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

func FetchHTTP(url string) (string, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))
	cacheDir := config.GetCacheDir()
	dest := filepath.Join(cacheDir, hash+".mk")
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}
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
