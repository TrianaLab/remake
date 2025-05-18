package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfigAndCacheDirAndRegistry(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	cfg := filepath.Join(tempHome, ".remake", "config.yaml")
	if _, err := os.Stat(cfg); err != nil {
		t.Errorf("expected config.yaml in %s, but err = %v", cfg, err)
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
		t.Errorf("no files, GetDefaultMakefile() = %q; want \"\"", got)
	}

	if err := os.WriteFile("makefile", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if got := GetDefaultMakefile(); got != "makefile" {
		t.Errorf("with makefile, GetDefaultMakefile() = %q; want %q", got, "makefile")
	}
}

func TestSaveConfig(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	viper.Set("default_registry", "example.com")

	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	cfgPath := filepath.Join(tempHome, ".remake", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("couldn't read %s: %v", cfgPath, err)
	}
	if !bytes.Contains(data, []byte("default_registry: example.com")) {
		t.Errorf("config.yaml doesn't contain modified default_registry; content=\n%s", data)
	}
}
