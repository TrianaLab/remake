package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig_HomeDirError(t *testing.T) {
	origHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", origHome)

	err := InitConfig()
	if err == nil || !strings.Contains(err.Error(), "cannot find home directory") {
		t.Fatalf("expected home-dir error, got %v", err)
	}
}

func TestInitConfig_MkdirAllError(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "notadir")
	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	os.Setenv("HOME", filePath)
	viper.Reset()

	err := InitConfig()
	if err == nil || !strings.Contains(err.Error(), "cannot create config directory") {
		t.Fatalf("expected MkdirAll error, got %v", err)
	}
}

func TestInitConfig_ReadConfigError(t *testing.T) {
	// Point HOME to a fresh temp dir
	home := t.TempDir()
	os.Setenv("HOME", home)
	// Fully reset Viper’s state
	viper.Reset()

	// Pre-create ~/.remake/config.yaml with a syntactically invalid YAML
	configDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("mkdir configDir failed: %v", err)
	}
	cfgFile := filepath.Join(configDir, "config.yaml")
	badContent := "foo: [unclosed" // missing closing bracket
	if err := os.WriteFile(cfgFile, []byte(badContent), 0600); err != nil {
		t.Fatalf("write invalid config failed: %v", err)
	}

	// Reset again so Viper picks up the existing file and tries to parse it
	viper.Reset()

	// Now InitConfig should hit the ReadInConfig error path
	err := InitConfig()
	if err == nil || !strings.Contains(err.Error(), "error reading config file") {
		t.Fatalf("expected read-in-config error, got %v", err)
	}
}

func TestInitConfig_CreatesDefaultConfig(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	viper.Reset()

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	cfgDir := filepath.Join(home, ".remake")
	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if fi, err := os.Stat(cfgDir); err != nil || !fi.IsDir() {
		t.Fatalf("expected config dir %q, got err %v", cfgDir, err)
	}
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("expected config file %q, got err %v", cfgFile, err)
	}
	if !strings.Contains(string(data), "registries: {}") {
		t.Errorf("expected default registries in config, got %q", string(data))
	}
	if viper.ConfigFileUsed() != cfgFile {
		t.Errorf("viper did not read config file, used %q", viper.ConfigFileUsed())
	}
}

func TestSaveConfig_WritesConfig(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	viper.Reset()

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}
	viper.Set("foo", "bar")
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	cfgFile := filepath.Join(home, ".remake", "config.yaml")
	viper.Reset()
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("viper.ReadInConfig failed: %v", err)
	}
	if viper.GetString("foo") != "bar" {
		t.Errorf("expected foo=bar, got %q", viper.GetString("foo"))
	}
}

func TestGetDefaultMakefile(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	if got := GetDefaultMakefile(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	os.WriteFile("makefile", []byte{}, 0644)
	if got := GetDefaultMakefile(); got != "makefile" {
		t.Errorf("expected 'makefile', got %q", got)
	}
	os.Remove("makefile")
	os.WriteFile("Makefile", []byte{}, 0644)
	if got := GetDefaultMakefile(); got != "Makefile" {
		t.Errorf("expected 'Makefile', got %q", got)
	}
}

func TestGetCacheDir(t *testing.T) {
	home := t.TempDir()
	os.Setenv("HOME", home)
	want := filepath.Join(home, ".remake", "cache")
	if got := GetCacheDir(); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	os.Unsetenv("HOME")
	got := GetCacheDir()
	if !strings.HasSuffix(got, filepath.Join(".remake", "cache")) {
		t.Errorf("expected fallback ending in .remake/cache, got %q", got)
	}
}

func TestNormalizeKey(t *testing.T) {
	input := "example.com:5000"
	want := "example_com:5000"
	if got := NormalizeKey(input); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
