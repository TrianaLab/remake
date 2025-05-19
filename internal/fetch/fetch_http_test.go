package fetch

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestHTTPFetch_CacheHit(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	urlStr := "http://example.com/file"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(urlStr)))
	parsed, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}
	host := parsed.Host
	dir := filepath.Join(tmp, host)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("unable to create cache directory: %v", err)
	}
	dest := filepath.Join(dir, hash+".mk")
	if err := os.WriteFile(dest, []byte("cached"), 0644); err != nil {
		t.Fatalf("unable to write cache file: %v", err)
	}
	fetcher := &HTTPFetcher{}
	path, err := fetcher.Fetch(urlStr, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != dest {
		t.Fatalf("expected %s, got %s", dest, path)
	}
}

func TesHTTPFetch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "hello world")
	}))
	defer srv.Close()

	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &HTTPFetcher{}
	path, err := fetcher.Fetch(srv.URL, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to read fetched file: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected content \"hello world\", got %q", string(data))
	}
}

func TestHTTPFetch_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &HTTPFetcher{}
	_, err := fetcher.Fetch(srv.URL, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP error 500") {
		t.Fatalf("expected HTTP error 500, got %v", err)
	}
}

func TestHTTPFetch_GetError(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &HTTPFetcher{}
	_, err := fetcher.Fetch("http://127.0.0.1:0", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHTTPFetch_MkdirError(t *testing.T) {
	tmp := t.TempDir()
	d := filepath.Join(tmp, "cachefile")
	if err := os.WriteFile(d, []byte{}, 0644); err != nil {
		t.Fatalf("unable to create blocking file: %v", err)
	}
	viper.Set("cacheDir", d)
	fetcher := &HTTPFetcher{}
	_, err := fetcher.Fetch("http://example.com", false)
	if err == nil {
		t.Fatal("expected mkdir error, got nil")
	}
}

func TestHTTPFetch_CreateFileError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "test content")
	}))
	defer srv.Close()

	tmp := t.TempDir()
	if err := os.Chmod(tmp, 0444); err != nil {
		t.Fatalf("unable to change directory permissions: %v", err)
	}

	viper.Set("cacheDir", tmp)
	fetcher := &HTTPFetcher{}
	_, err := fetcher.Fetch(srv.URL, false)
	if err == nil {
		t.Fatal("expected file creation error, got nil")
	}
}

func TestHTTPFetch_CopyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial "))
		panic("connection reset")
	}))
	defer srv.Close()

	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)
	fetcher := &HTTPFetcher{}
	_, err := fetcher.Fetch(srv.URL, false)
	if err == nil {
		t.Fatal("expected copy error, got nil")
	}
}
