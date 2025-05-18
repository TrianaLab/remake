package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigAndCacheDirAndRegistry(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	cfg := filepath.Join(tempHome, ".remake", "config.yaml")
	if _, err := os.Stat(cfg); err != nil {
		t.Errorf("se esperaba config.yaml en %s, pero err = %v", cfg, err)
	}

	if got := GetDefaultRegistry(); got != "ghcr.io" {
		t.Errorf("GetDefaultRegistry() = %q; want %q", got, "ghcr.io")
	}

	wantCache := filepath.Join(tempHome, ".remake", "cache")
	if cd := GetCacheDir(); cd != wantCache {
		t.Errorf("GetCacheDir() = %q; want %q", cd, wantCache)
	}
}

func TestGetDefaultMakefile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if got := GetDefaultMakefile(); got != "" {
		t.Errorf("sin ficheros, GetDefaultMakefile() = %q; want \"\"", got)
	}

	if err := os.WriteFile("makefile", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if got := GetDefaultMakefile(); got != "makefile" {
		t.Errorf("con makefile, GetDefaultMakefile() = %q; want %q", got, "makefile")
	}
}
