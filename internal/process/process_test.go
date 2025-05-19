package process

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestRun(t *testing.T) {
	tmp := t.TempDir()
	makeContent := "test:\n\techo test"
	makePath := filepath.Join(tmp, "Makefile")
	if err := os.WriteFile(makePath, []byte(makeContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		src      string
		targets  []string
		useCache bool
		wantErr  bool
	}{
		{
			name:     "Local file success",
			src:      makePath,
			targets:  []string{"test"},
			useCache: false,
			wantErr:  false,
		},
		{
			name:     "Local file not found",
			src:      "nonexistent",
			targets:  []string{"test"},
			useCache: false,
			wantErr:  true,
		},
		{
			name:     "Invalid make command",
			src:      makePath,
			targets:  []string{"invalid"},
			useCache: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.src, tt.targets, tt.useCache)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchSource(t *testing.T) {
	tmp := t.TempDir()
	viper.Set("cacheDir", tmp)

	localFile := filepath.Join(tmp, "local.mk")
	if err := os.WriteFile(localFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		ref      string
		useCache bool
		wantErr  bool
	}{
		{
			name:     "Local file exists",
			ref:      localFile,
			useCache: false,
			wantErr:  false,
		},
		{
			name:     "Local file relative path",
			ref:      "local.mk",
			useCache: false,
			wantErr:  true,
		},
		{
			name:     "Remote file error",
			ref:      "http://nonexistent.example.com/Makefile",
			useCache: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchSource(tt.ref, tt.useCache)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("fetchSource() returned empty path")
			}
		})
	}
}
