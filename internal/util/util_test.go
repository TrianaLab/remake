package util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeRef(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "x.mk")
	if err := os.WriteFile(f, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(f)
	if got := NormalizeRef(f); got != abs {
		t.Errorf("NormalizeRef(local) = %q; want %q", got, abs)
	}

	if got := NormalizeRef("http://foo"); got != "http://foo" {
		t.Errorf("NormalizeRef(http) = %q", got)
	}

	if got := NormalizeRef("oci://repo/name"); got != "oci://repo/name:latest" {
		t.Errorf("NormalizeRef(oci) = %q", got)
	}

	if got := NormalizeRef("bar"); got != "oci://ghcr.io/bar:latest" {
		t.Errorf("NormalizeRef(default) = %q", got)
	}
}

func TestFetchHTTP_Caching(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "hola")
	}))
	defer srv.Close()

	url := srv.URL + "/f"
	first, err := FetchHTTP(url)
	if err != nil {
		t.Fatalf("FetchHTTP() error = %v", err)
	}
	data, _ := os.ReadFile(first)
	if string(data) != "hola" {
		t.Errorf("content = %q; want %q", data, "hola")
	}

	second, err := FetchHTTP(url)
	if err != nil {
		t.Fatalf("FetchHTTP(cache) error = %v", err)
	}
	if first != second {
		t.Errorf("cache path changed: %q vs %q", first, second)
	}
}

func TestFetchMakefile_LocalAndOCI(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	local := filepath.Join(tempHome, "l.mk")
	if err := os.WriteFile(local, []byte("L"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := FetchMakefile(local)
	if err != nil || got != local {
		t.Errorf("FetchMakefile(local) = %q, %v; want %q, nil", got, err, local)
	}

	cache := filepath.Join(tempHome, ".remake", "cache", "repo", "v1")
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(cache, "foo.mk")
	if err := os.WriteFile(f, []byte("O"), 0644); err != nil {
		t.Fatal(err)
	}
	ref := "oci://repo:v1"
	got2, err := FetchMakefile(ref)
	if err != nil {
		t.Fatalf("FetchMakefile(oci) error = %v", err)
	}
	if got2 != f {
		t.Errorf("FetchMakefile(oci) = %q; want %q", got2, f)
	}
}

func TestFetchHTTP_ErrorStatus(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := FetchHTTP(srv.URL + "/f")
	if err == nil || !strings.Contains(err.Error(), "HTTP error 404") {
		t.Errorf("expected HTTP error 404, got %v", err)
	}
}

func TestFetchMakefile_Invalid(t *testing.T) {
	_, err := FetchMakefile("file_no_exists")
	if err == nil {
		t.Error("expected error invalid reference, but err = nil")
	}
}
