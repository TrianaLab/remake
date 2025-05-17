package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCycleDetection(t *testing.T) {
	d := t.TempDir()
	a := filepath.Join(d, "a.mk")
	b := filepath.Join(d, "b.mk")
	os.WriteFile(a, []byte("include b.mk"), 0644)
	os.WriteFile(b, []byte("include a.mk"), 0644)
	visited := make(map[string]bool)
	if err := processFile(a, visited, filepath.Join(d, "o.mk")); err == nil {
		t.Errorf("expected cycle error")
	}
}

func TestNoInclude(t *testing.T) {
	d := t.TempDir()
	f := filepath.Join(d, "f.mk")
	os.WriteFile(f, []byte("all:\n\techo ok\n"), 0644)
	visited := make(map[string]bool)
	out := filepath.Join(d, "o.mk")
	if err := processFile(f, visited, out); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "echo ok") {
		t.Errorf("missing content: %s", string(data))
	}
}
