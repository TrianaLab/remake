package util

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FetchOCI pulls a Makefile artifact from an OCI registry using `oras` and caches it under .remake/cache
func FetchOCI(ociRef string) (string, error) {
	ref := strings.TrimPrefix(ociRef, "oci://")
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ociRef)))
	outPath := filepath.Join(".remake", "cache", hash+".mk")

	// Return cached
	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	// Pull via oras into temp dir
	tmpDir := filepath.Join(".remake", "cache", "tmp-"+hash)
	os.MkdirAll(tmpDir, 0o755)

	cmd := exec.Command("oras", "pull", ref, "-o", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	// Find .mk file
	files, err := filepath.Glob(filepath.Join(tmpDir, "*.mk"))
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("no .mk found in OCI artifact: %s", ociRef)
	}

	// Move to cache
	os.MkdirAll(filepath.Dir(outPath), 0o755)
	if err := os.Rename(files[0], outPath); err != nil {
		return "", err
	}

	// Cleanup
	os.RemoveAll(tmpDir)

	return outPath, nil
}
