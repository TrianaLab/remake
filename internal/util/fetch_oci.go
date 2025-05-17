package util

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

// FetchOCI descarga un artifact OCI (vía oras), extrae UN archivo y lo cachea
func FetchOCI(ociRef string) (string, error) {
	ref := strings.TrimPrefix(ociRef, "oci://")
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(ociRef)))
	cacheDir := config.GetCacheDir()
	outPath := filepath.Join(cacheDir, hash+".mk")

	// cached?
	if _, err := os.Stat(outPath); err == nil {
		return outPath, nil
	}

	tmpDir := filepath.Join(cacheDir, "tmp-"+hash)
	os.MkdirAll(tmpDir, 0o755)

	// pull con oras
	cmd := exec.Command("oras", "pull", ref, "-o", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	// busca *cualquier* fichero dentro de tmpDir
	entries, err := os.ReadDir(tmpDir)
	if err != nil || len(entries) == 0 {
		return "", fmt.Errorf("no files found in OCI artifact: %s", ociRef)
	}
	// elige el primero (o podrías filtrar por ext .mk)
	src := filepath.Join(tmpDir, entries[0].Name())

	// mover a cache
	os.MkdirAll(cacheDir, 0o755)
	if err := os.Rename(src, outPath); err != nil {
		return "", err
	}
	os.RemoveAll(tmpDir)

	return outPath, nil
}
