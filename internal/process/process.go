package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/TrianaLab/remake/internal/fetch"
)

// Run fetches a Makefile (local or remote), then executes `make` with given targets.
func Run(src string, targets []string, useCache bool) error {
	// Fetch source if remote
	local, err := fetchSource(src, useCache)
	if err != nil {
		return err
	}
	// Build make command
	dest := local
	args := append([]string{"-f", dest}, targets...)
	cmd := exec.CommandContext(context.Background(), "make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make failed: %w", err)
	}
	return nil
}

// fetchSource uses util.Fetcher to obtain a local path for a given reference or returns local file
func fetchSource(ref string, useCache bool) (string, error) {
	// try remote fetcher
	fetcher, err := fetch.GetFetcher(ref)
	if err == nil {
		// remote source
		local, err := fetcher.Fetch(ref, useCache)
		if err != nil {
			return "", fmt.Errorf("failed to fetch source %q: %w", ref, err)
		}
		return local, nil
	}
	// fallback to local file
	if _, err := os.Stat(ref); os.IsNotExist(err) {
		rel := filepath.Clean(ref)
		if _, err2 := os.Stat(rel); err2 != nil {
			return "", fmt.Errorf("source file not found: %s", ref)
		}
		return rel, nil
	}
	return ref, nil
}
