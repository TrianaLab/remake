package fetch

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// HTTPFetcher retrieves HTTP-hosted Makefiles, honoring cache if enabled.
type HTTPFetcher struct{}

func (h *HTTPFetcher) Fetch(ref string, useCache bool) (string, error) {
	// First, consult cache
	cacheFetcher := &CacheFetcher{}
	if path, err := cacheFetcher.Fetch(ref, useCache); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	// Not cached or cache disabled: proceed with download
	parsed, err := url.Parse(ref)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", err
	}

	host := parsed.Host
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ref)))
	dest := filepath.Join(viper.GetString("cacheDir"), host, hash+".mk")

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", err
	}

	resp, err := http.Get(ref)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d fetching %s", resp.StatusCode, ref)
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
