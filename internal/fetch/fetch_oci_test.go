package fetch

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func calculateSHA256(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", hash)
}

func TestOCIFetch_CacheHit(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	ref := "oci://host/repo:tag"
	host := "host"
	repo := "repo"
	tag := "tag"
	cachePath := filepath.Join(tmp, host, repo, tag)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		t.Fatal(err)
	}
	fileName := "artifact.bin"
	fullPath := filepath.Join(cachePath, fileName)
	if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	fetcher := &OCIFetcher{}
	path, err := fetcher.Fetch(ref, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != fullPath {
		t.Fatalf("expected %s, got %s", fullPath, path)
	}
}

func TestOCIFetch_InvalidRef(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &OCIFetcher{}
	_, err := fetcher.Fetch("oci://invalid", false)
	if err == nil || !strings.Contains(err.Error(), "invalid OCI reference") {
		t.Fatalf("expected invalid OCI reference error, got %v", err)
	}
}

func TestOCIFetch_FileNewError(t *testing.T) {
	tmp := t.TempDir()
	block := filepath.Join(tmp, "block")
	if err := os.WriteFile(block, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	viper.Set("cacheDir", block)
	fetcher := &OCIFetcher{}
	_, err := fetcher.Fetch("oci://host/repo:tag", false)
	if err == nil {
		t.Fatal("expected file.New error, got nil")
	}
}

func TestOCIFetch_NewRepoError(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &OCIFetcher{}
	_, err := fetcher.Fetch("oci:///repo:tag", false)
	if err == nil {
		t.Fatal("expected NewRepository error, got nil")
	}
}

func TestOCIFetch_CopyError_NoCreds(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	ref := "oci://doesnotexist/repo:tag"
	fetcher := &OCIFetcher{}
	_, err := fetcher.Fetch(ref, false)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch OCI artifact") {
		t.Fatalf("expected failed to fetch OCI artifact error, got %v", err)
	}
}

func TestOCIFetch_CopyError_WithCreds(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	viper.Set("registries.doesnotexist.username", "user")
	viper.Set("registries.doesnotexist.password", "pass")
	ref := "oci://doesnotexist/repo:tag"
	fetcher := &OCIFetcher{}
	_, err := fetcher.Fetch(ref, false)
	if err == nil || !strings.Contains(err.Error(), "failed to fetch OCI artifact") {
		t.Fatalf("expected failed to fetch OCI artifact error, got %v", err)
	}
}
