package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/internal/fetch"
)

// fakeFetcher mocks fetch.Fetcher
type fakeFetcher struct {
	path string
	err  error
}

func (f *fakeFetcher) Fetch(ref string, useCache bool) (string, error) {
	return f.path, f.err
}

func TestArgsValidation(t *testing.T) {
	if err := pullCmd.Args(pullCmd, []string{}); err == nil {
		t.Error("Expected error for no arguments")
	}
	if err := pullCmd.Args(pullCmd, []string{"a", "b"}); err == nil {
		t.Error("Expected error for too many arguments")
	}
}

func TestGetFetcherError(t *testing.T) {
	orig := pullGetFetcher
	defer func() { pullGetFetcher = orig }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return nil, errors.New("bad fetcher")
	}
	err := pullCmd.RunE(pullCmd, []string{"any"})
	if err == nil || !strings.Contains(err.Error(), "bad fetcher") {
		t.Fatalf("Expected bad fetcher error, got %v", err)
	}
}

func TestFetchError(t *testing.T) {
	orig := pullGetFetcher
	defer func() { pullGetFetcher = orig }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return &fakeFetcher{"", errors.New("fetch failed")}, nil
	}
	err := pullCmd.RunE(pullCmd, []string{"any"})
	if err == nil || !strings.Contains(err.Error(), "fetch failed") {
		t.Fatalf("Expected fetch failed error, got %v", err)
	}
}

func TestCopyToFile(t *testing.T) {
	// mock fetcher returns temp file path
	tmp := t.TempDir()
	srcPath := tmp + "/src.mk"
	os.WriteFile(srcPath, []byte("hello"), 0644)

	outPath := tmp + "/out.mk"
	orig := pullGetFetcher
	defer func() { pullGetFetcher = orig }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return &fakeFetcher{srcPath, nil}, nil
	}
	origFile := pullFile
	defer func() { pullFile = origFile }()
	pullFile = outPath

	err := pullCmd.RunE(pullCmd, []string{"ref"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	data, _ := os.ReadFile(outPath)
	if string(data) != "hello" {
		t.Errorf("File content = %q; want 'hello'", string(data))
	}
}

func TestArtifactNotFoundError(t *testing.T) {
	tmp := t.TempDir()
	outPath := tmp + "/out.mk"
	origFetcher := pullGetFetcher
	defer func() { pullGetFetcher = origFetcher }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return &fakeFetcher{"", nil}, nil
	}
	origFile := pullFile
	defer func() { pullFile = origFile }()
	pullFile = outPath

	err := pullCmd.RunE(pullCmd, []string{"ref"})
	if err == nil || !strings.Contains(err.Error(), "artifact not found in cache") {
		t.Fatalf("Expected artifact not found error, got %v", err)
	}
}

func TestPrintSavedPath(t *testing.T) {
	// mock fetcher returns non-empty path
	tmp := t.TempDir()
	srcPath := tmp + "/cache.mk"
	os.WriteFile(srcPath, []byte("data"), 0644)

	origFetcher := pullGetFetcher
	defer func() { pullGetFetcher = origFetcher }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return &fakeFetcher{srcPath, nil}, nil
	}

	origFile := pullFile
	defer func() { pullFile = origFile }()
	pullFile = ""

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	err := pullCmd.RunE(pullCmd, []string{"ref"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()
	if !strings.HasPrefix(out, fmt.Sprintf("Saved to %s", srcPath)) {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestPrintFetchedCurrentDir(t *testing.T) {
	origFetcher := pullGetFetcher
	defer func() { pullGetFetcher = origFetcher }()
	pullGetFetcher = func(ref string) (fetch.Fetcher, error) {
		return &fakeFetcher{"", nil}, nil
	}

	origFile := pullFile
	defer func() { pullFile = origFile }()
	pullFile = ""

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	err := pullCmd.RunE(pullCmd, []string{"ref"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "Fetched to current directory") {
		t.Errorf("Unexpected output: %s", buf.String())
	}
}
