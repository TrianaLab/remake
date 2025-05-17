package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFetchMakefileLocal(t *testing.T) {
	tmp := t.TempDir()
	tfile := filepath.Join(tmp, "t.mk")
	os.WriteFile(tfile, []byte("x"), 0644)
	res, err := FetchMakefile(tfile)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res == "" {
		t.Error("expected local path")
	}
}
