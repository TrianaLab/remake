package fetch

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestCacheFetch_NoCache(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &CacheFetcher{}
	path, err := fetcher.Fetch("any", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %s", path)
	}
}

func TestCacheFetch_HTTPFound(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	urlStr := "http://example.com/file"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(urlStr)))
	host := "example.com"
	dir := filepath.Join(tmp, host)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, hash+".mk")
	f, err := os.Create(dest)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	fetcher := &CacheFetcher{}
	path, err := fetcher.Fetch(urlStr, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != dest {
		t.Fatalf("expected %s, got %s", dest, path)
	}
}

func TestCacheFetch_HTTPNotFound(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	urlStr := "https://example.org/other"
	fetcher := &CacheFetcher{}
	path, err := fetcher.Fetch(urlStr, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %s", path)
	}
}

func TestCacheFetch_InvalidOCI(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &CacheFetcher{}
	_, err := fetcher.Fetch("oci://invalid", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expected := "invalid reference format: oci://invalid"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestCacheFetch_OCINotFound(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	ref := "oci://host/repo:tag"
	fetcher := &CacheFetcher{}
	path, err := fetcher.Fetch(ref, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "" {
		t.Fatalf("expected empty path, got %s", path)
	}
}

func TestCacheFetch_OCIFound(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	host := "registry.io"
	repo := "myrepo"
	tag := "v1"
	cachePath := filepath.Join(tmp, host, repo, tag)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		t.Fatal(err)
	}
	fileName := "artifact.bin"
	fullPath := filepath.Join(cachePath, fileName)
	f, err := os.Create(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	ref := fmt.Sprintf("oci://%s/%s:%s", host, repo, tag)
	fetcher := &CacheFetcher{}
	path, err := fetcher.Fetch(ref, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != fullPath {
		t.Fatalf("expected %s, got %s", fullPath, path)
	}
}
