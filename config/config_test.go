package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name    string
		setupFn func() string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Success with new config",
			setupFn: func() string {
				tmp := t.TempDir()
				os.Setenv("HOME", tmp)
				return tmp
			},
			wantErr: false,
		},
		{
			name: "Success with existing config",
			setupFn: func() string {
				tmp := t.TempDir()
				os.Setenv("HOME", tmp)
				remakeDir := filepath.Join(tmp, ".remake")
				configFile := filepath.Join(remakeDir, "config.yaml")
				os.MkdirAll(remakeDir, 0700)
				os.WriteFile(configFile, []byte("registries: {}\n"), 0600)
				return tmp
			},
			wantErr: false,
		},
		{
			name: "Error on MkdirAll",
			setupFn: func() string {
				tmp := t.TempDir()
				os.Setenv("HOME", tmp)
				remakeDir := filepath.Join(tmp, ".remake")
				os.WriteFile(remakeDir, []byte("block"), 0644)
				return tmp
			},
			wantErr: true,
			errMsg:  "cannot create config directory",
		},
		{
			name: "Error on WriteFile",
			setupFn: func() string {
				tmp := t.TempDir()
				os.Setenv("HOME", tmp)
				remakeDir := filepath.Join(tmp, ".remake")
				os.MkdirAll(remakeDir, 0500)
				return tmp
			},
			wantErr: true,
			errMsg:  "cannot create default config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			homeDir := tt.setupFn()
			defer os.RemoveAll(homeDir)

			err := InitConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("InitConfig() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if !tt.wantErr {
				if viper.GetString("defaultMakefile") != "makefile" {
					t.Error("default value for defaultMakefile not set")
				}
				if viper.GetBool("insecure") {
					t.Error("default value for insecure not set")
				}
				cacheDir := viper.GetString("cacheDir")
				if !strings.Contains(cacheDir, ".remake/cache") {
					t.Errorf("unexpected cacheDir: %s", cacheDir)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}

	viper.Set("test", "value")

	if err := SaveConfig(); err != nil {
		t.Errorf("SaveConfig() error = %v", err)
	}

	configFile := filepath.Join(tmp, ".remake", "config.yaml")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "test: value") {
		t.Error("saved config does not contain expected value")
	}
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "Simple domain",
			endpoint: "example.com",
			want:     "example_com",
		},
		{
			name:     "Multiple dots",
			endpoint: "registry.example.com",
			want:     "registry_example_com",
		},
		{
			name:     "No dots",
			endpoint: "localhost",
			want:     "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeKey(tt.endpoint); got != tt.want {
				t.Errorf("NormalizeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
