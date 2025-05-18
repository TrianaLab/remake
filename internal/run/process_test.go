package run

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "a")
	if err := os.WriteFile(f, []byte("XD"), 0644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := inlineFile(&buf, f); err != nil {
		t.Fatalf("inlineFile() error = %v", err)
	}
	if buf.String() != "XD" {
		t.Errorf("inlineFile = %q; want %q", buf.String(), "XD")
	}
}

func TestProcessFile_IndentAndPlain(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.mk")
	_ = os.WriteFile(src, []byte(" line1\nplain\n"), 0644)

	out := filepath.Join(tmp, "out.mk")
	if err := processFile(src, map[string]bool{}, out); err != nil {
		t.Fatalf("processFile() error = %v", err)
	}
	got, _ := os.ReadFile(out)
	if !strings.Contains(string(got), "\tline1") || !strings.Contains(string(got), "plain\n") {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestProcessFile_IncludeBlock(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "main.mk")
	inc := filepath.Join(tmp, "inc.mk")
	_ = os.WriteFile(inc, []byte("A\n"), 0644)
	_ = os.WriteFile(src, []byte("include:\n- "+inc+"\nB\n"), 0644)

	out := filepath.Join(tmp, "out.mk")
	if err := processFile(src, map[string]bool{}, out); err != nil {
		t.Fatalf("include block error = %v", err)
	}
	got, _ := os.ReadFile(out)
	if !bytes.Contains(got, []byte("A\n")) || !bytes.Contains(got, []byte("B\n")) {
		t.Errorf("include not injected: %s", got)
	}
}

func TestProcessFile_Cyclic(t *testing.T) {
	tmp := t.TempDir()
	a := filepath.Join(tmp, "a.mk")
	b := filepath.Join(tmp, "b.mk")
	_ = os.WriteFile(a, []byte("include:\n- "+b+"\n"), 0644)
	_ = os.WriteFile(b, []byte("include:\n- "+a+"\n"), 0644)

	err := processFile(a, map[string]bool{}, filepath.Join(tmp, "o"))
	if err == nil || !strings.Contains(err.Error(), "cyclic include") {
		t.Errorf("expected cyclic include, got %v", err)
	}
}

func TestRender(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "r.mk")
	_ = os.WriteFile(src, []byte("Z\n"), 0644)
	out := filepath.Join(tmp, "o.mk")

	path, err := Render(src, out)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if path != out {
		t.Errorf("Render path = %q; want %q", path, out)
	}
	got, _ := os.ReadFile(out)
	if string(got) != "Z\n" {
		t.Errorf("Render = %q", got)
	}
}

func TestRun_Success(t *testing.T) {
	cwd := t.TempDir()
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("Makefile", []byte("all:\n\techo OK\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Run([]string{"all"}, "Makefile"); err != nil {
		t.Errorf("Run() success failed: %v", err)
	}
}

func TestRun_FileNotFound(t *testing.T) {
	err := Run([]string{"all"}, "no_exists.mk")
	if err == nil || !strings.Contains(err.Error(), "read no_exists.mk") {
		t.Errorf("expected read error, got %v", err)
	}
}

func TestInlineFile_Error(t *testing.T) {
	var buf bytes.Buffer
	err := inlineFile(&buf, "no_exists")
	if err == nil || !strings.HasPrefix(err.Error(), "inline read no_exists") {
		t.Errorf("expected read error, got %v", err)
	}
}
