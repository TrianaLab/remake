package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRunCommand_Execute(t *testing.T) {
	// Save original values to restore after tests
	originalFile := runFile
	originalNoCache := runNoCache
	defer func() {
		runFile = originalFile
		runNoCache = originalNoCache
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
		args      []string
		setupFunc func()
		wantErr   bool
	}{
		{
			name: "success with file flag",
			args: []string{"test"},
			setupFunc: func() {
				runFile = makePath
				runNoCache = false
			},
			wantErr: false,
		},
		{
			name: "success with default configuration",
			args: []string{"test"},
			setupFunc: func() {
				runFile = ""
				runNoCache = false
				viper.Set("defaultMakefile", makePath)
			},
			wantErr: false,
		},
		{
			name: "error without specified file",
			args: []string{"test"},
			setupFunc: func() {
				runFile = ""
				runNoCache = false
				viper.Reset()
			},
			wantErr: true,
		},
		{
			name: "error with non-existent file",
			args: []string{"test"},
			setupFunc: func() {
				runFile = "archivo_inexistente"
				runNoCache = false
			},
			wantErr: true,
		},
		{
			name: "with no-cache flag",
			args: []string{"test"},
			setupFunc: func() {
				runFile = makePath
				runNoCache = true
			},
			wantErr: false,
		},
		{
			name: "without specified targets",
			args: []string{},
			setupFunc: func() {
				runFile = makePath
				runNoCache = false
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
			err := runCmd.RunE(cmd, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("runCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunInit(t *testing.T) {
	// Clean up viper state before and after test
	viper.Reset()
	defer viper.Reset()

	// Test that the command is initialized correctly
	if runCmd.Use != "run [flags] [targets]" {
		t.Errorf("Expected Use to be 'run [flags] [targets]', got %s", runCmd.Use)
	}

	// Test flag registration
	fileFlag := runCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("file flag not registered")
	}
	if fileFlag.Shorthand != "f" {
		t.Errorf("Expected shorthand for file flag to be 'f', got %s", fileFlag.Shorthand)
	}

	noCacheFlag := runCmd.Flags().Lookup("no-cache")
	if noCacheFlag == nil {
		t.Error("no-cache flag not registered")
	}

	// Test viper binding
	err := viper.BindPFlag("defaultMakefile", runCmd.Flags().Lookup("file"))
	if err != nil {
		t.Errorf("Failed to bind flag to viper: %v", err)
	}

	// Set a value to verify binding
	runCmd.Flags().Set("file", "test.mk")
	if viper.GetString("defaultMakefile") != "test.mk" {
		t.Error("defaultMakefile not properly bound to viper")
	}
}
