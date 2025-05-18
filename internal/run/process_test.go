package run

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
)

// writeFile writes a test file.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

// prependFakeMake creates a fake `make` binary in PATH.
func prependFakeMake(t *testing.T, script string) func() {
	t.Helper()
	tmpBin := t.TempDir()
	makePath := filepath.Join(tmpBin, "make")
	if err := os.WriteFile(makePath, []byte(script), 0755); err != nil {
		t.Fatalf("creating fake make: %v", err)
	}
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpBin+string(os.PathListSeparator)+origPath)
	return func() { os.Setenv("PATH", origPath) }
}

func TestRender_SimpleFile(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	writeFile(t, dir, "Makefile", "all:\n\techo hi\n")
	out := filepath.Join(dir, "out.mk")
	if err := Render("Makefile", out, true); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	data, _ := os.ReadFile(out)
	if !bytes.Contains(data, []byte("all:")) || !bytes.Contains(data, []byte("\techo hi")) {
		t.Errorf("output missing lines: %q", string(data))
	}
}

func TestRender_BlockIncludeAndInline(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	main := `include:
 - one.mk
other: val
include two.mk

end`
	writeFile(t, dir, "main.mk", main)
	writeFile(t, dir, "one.mk", "ONE\n")
	writeFile(t, dir, "two.mk", "TWO\n")
	out := filepath.Join(dir, "out.mk")
	if err := Render(filepath.Join(dir, "main.mk"), out, true); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	data, _ := os.ReadFile(out)
	got := string(data)
	if !strings.Contains(got, "ONE\n") || !strings.Contains(got, "TWO\n") || !strings.Contains(got, "other: val\n") {
		t.Errorf("includes not in output: %q", got)
	}
}

func TestRender_CyclicInclude(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	writeFile(t, dir, "A.mk", "include:\n - B.mk\n")
	writeFile(t, dir, "B.mk", "include A.mk\n")
	out := filepath.Join(dir, "out.mk")
	src := filepath.Join(dir, "A.mk")
	err := Render(src, out, true)
	if err == nil || !strings.Contains(err.Error(), "cyclic include detected") {
		t.Fatalf("expected cyclic include error, got %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	writeFile(t, dir, "Makefile", "all:\n\techo HELLO\n")
	// fake make: always succeed
	script := `#!/bin/sh
exit 0
`
	restore := prependFakeMake(t, script)
	defer restore()

	err := Run(filepath.Join(dir, "Makefile"), []string{"all"}, true)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	cache := config.GetCacheDir()
	gen := filepath.Join(cache, "Makefile.generated")
	if _, err := os.Stat(gen); !os.IsNotExist(err) {
		t.Errorf("expected generated file removed, still exists: %v", gen)
	}
	// test cleanup of checksum name
	csgen := fmt.Sprintf("%x.mk.generated", sha256.Sum256([]byte(filepath.Join(dir, "Makefile"))))
	csPath := filepath.Join(cache, csgen)
	if _, err := os.Stat(csPath); !os.IsNotExist(err) {
		t.Errorf("expected checksum-generated file removed, still exists: %v", csPath)
	}
}

func TestRun_MakeError(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(dir)

	writeFile(t, dir, "Makefile", "all:\n\texit 1\n")
	script := `#!/bin/sh
exit 1
`
	restore := prependFakeMake(t, script)
	defer restore()

	err := Run(filepath.Join(dir, "Makefile"), []string{"all"}, true)
	if err == nil {
		t.Fatalf("expected error from failed make, got nil")
	}
}
