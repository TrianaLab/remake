package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestTemplateCommand_Execute(t *testing.T) {
	// Save original values to restore after test
	originalFile := templateFile
	originalNoCache := templateNoCache
	defer func() {
		templateFile = originalFile
		templateNoCache = originalNoCache
		viper.Reset()
	}()

	// Create test directory and file
	tmp := t.TempDir()
	makeContent := "test:\n\techo test"
	makePath := filepath.Join(tmp, "Makefile")
	if err := os.WriteFile(makePath, []byte(makeContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		setupFunc func()
		wantErr   bool
	}{
		{
			name: "success with file flag",
			setupFunc: func() {
				templateFile = makePath
				templateNoCache = false
			},
			wantErr: false,
		},
		{
			name: "success with config default",
			setupFunc: func() {
				templateFile = ""
				templateNoCache = false
				viper.Set("defaultMakefile", makePath)
			},
			wantErr: false,
		},
		{
			name: "error no file specified",
			setupFunc: func() {
				templateFile = ""
				templateNoCache = false
				viper.Reset()
			},
			wantErr: true,
		},
		{
			name: "error nonexistent file",
			setupFunc: func() {
				templateFile = "nonexistent"
				templateNoCache = false
			},
			wantErr: true,
		},
		{
			name: "with no-cache flag",
			setupFunc: func() {
				templateFile = makePath
				templateNoCache = true
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			// Setup test state
			tt.setupFunc()

			cmd := &cobra.Command{}
			err := templateCmd.RunE(cmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("templateCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateInit(t *testing.T) {
	// Clean up viper state before and after test
	viper.Reset()
	defer viper.Reset()

	// Test that the command is properly initialized
	if templateCmd.Use != "template [flags]" {
		t.Errorf("Expected Use to be 'template [flags]', got %s", templateCmd.Use)
	}

	// Test flag registration
	fileFlag := templateCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("file flag not registered")
	}
	if fileFlag.Shorthand != "f" {
		t.Errorf("Expected file flag shorthand to be 'f', got %s", fileFlag.Shorthand)
	}

	noCacheFlag := templateCmd.Flags().Lookup("no-cache")
	if noCacheFlag == nil {
		t.Error("no-cache flag not registered")
	}

	// Test viper binding
	err := viper.BindPFlag("defaultMakefile", templateCmd.Flags().Lookup("file"))
	if err != nil {
		t.Errorf("Failed to bind flag to viper: %v", err)
	}

	// Set a value to verify binding
	templateCmd.Flags().Set("file", "test.mk")
	if viper.GetString("defaultMakefile") != "test.mk" {
		t.Error("defaultMakefile not properly bound to viper")
	}
}
