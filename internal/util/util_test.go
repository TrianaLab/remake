package util

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
)

// TestCacheFetcher covers all branches of CacheFetcher.Fetch
func TestCacheFetcher(t *testing.T) {
	cf := &CacheFetcher{}
	// useCache = false
	if p, err := cf.Fetch("oci://h/r:t", false); err != nil || p != "" {
		t.Errorf("expected empty when cache disabled, got %q, %v", p, err)
	}
	// invalid reference
	if _, err := cf.Fetch("invalid", true); err == nil || !strings.Contains(err.Error(), "invalid reference format") {
		t.Errorf("expected invalid reference error, got %v", err)
	}
	// no cache dir
	home := t.TempDir()
	os.Setenv("HOME", home)
	if p, err := cf.Fetch("oci://h/r:t", true); err != nil || p != "" {
		t.Errorf("expected no cache found, got %q, %v", p, err)
	}
	// cache hit
	home = t.TempDir()
	os.Setenv("HOME", home)
	cacheDir := filepath.Join(config.GetCacheDir(), "h", "r", "t")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	filePath := filepath.Join(cacheDir, "f.txt")
	os.WriteFile(filePath, []byte("data"), 0644)
	if p, err := cf.Fetch("oci://h/r:t", true); err != nil || !strings.HasSuffix(p, "f.txt") {
		t.Errorf("expected cache hit, got %q, %v", p, err)
	}
}

// TestHTTPFetcher covers HTTPFetcher.Fetch
func TestHTTPFetcher(t *testing.T) {
	hf := &HTTPFetcher{}
	// simulate HTTP 404
	serverErr := httptest.NewServer(http.NotFoundHandler())
	defer serverErr.Close()
	os.Setenv("HOME", t.TempDir())
	if _, err := hf.Fetch(serverErr.URL, false); err == nil || !strings.Contains(err.Error(), "HTTP error 404") {
		t.Errorf("expected HTTP error, got %v", err)
	}
	// simulate success
	body := "hello"
	serverOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body)
	}))
	defer serverOK.Close()
	os.Setenv("HOME", t.TempDir())
	path, err := hf.Fetch(serverOK.URL, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != body {
		t.Errorf("expected %q, got %q", body, data)
	}
}

// TestOCIFetcher_ErrorAndSuccess minimal tests
func TestOCIFetcher_InvalidAndSuccess(t *testing.T) {
	ofetch := &OCIFetcher{}
	// invalid ref
	if _, err := ofetch.Fetch("badref", true); err == nil {
		t.Error("expected invalid OCI reference error")
	}
	// success path: stub CacheFetcher to hit cache
	// create cache file
	home := t.TempDir()
	os.Setenv("HOME", home)
	ref := "oci://h/r:tag"
	cacheDir := filepath.Join(config.GetCacheDir(), "h", "r", "tag")
	os.MkdirAll(cacheDir, 0755)
	filePath := filepath.Join(cacheDir, "a.txt")
	os.WriteFile(filePath, []byte("ok"), 0644)
	path, err := ofetch.Fetch(ref, true)
	if err != nil {
		t.Fatalf("expected cache hit, got error %v", err)
	}
	if !strings.HasSuffix(path, "a.txt") {
		t.Errorf("expected a.txt, got %q", path)
	}
}

// TestGetFetcher covers GetFetcher
func TestGetFetcher(t *testing.T) {
	if _, err := GetFetcher("ftp://x"); err == nil {
		t.Error("expected unsupported ref error")
	}
	if f, err := GetFetcher("oci://x/y"); err != nil || fmt.Sprintf("%T", f) != "*util.OCIFetcher" {
		t.Errorf("expected OCIFetcher, got %T, %v", f, err)
	}
	if f, err := GetFetcher("https://x"); err != nil || fmt.Sprintf("%T", f) != "*util.HTTPFetcher" {
		t.Errorf("expected HTTPFetcher, got %T, %v", f, err)
	}
}
