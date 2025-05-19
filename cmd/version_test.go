package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout temporarily redirects os.Stdout and returns what was printed.
func captureVersionStdout(f func()) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func TestVersionCmd_Default(t *testing.T) {
	// Default version is "dev" :contentReference[oaicite:0]{index=0}:contentReference[oaicite:1]{index=1}
	out, err := captureVersionStdout(func() {
		rootCmd.SetArgs([]string{"version"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("captureStdout error: %v", err)
	}
	expected := "version: dev\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestVersionCmd_Custom(t *testing.T) {
	// Override version variable
	orig := version
	version = "1.2.3"
	defer func() { version = orig }()

	out, err := captureVersionStdout(func() {
		rootCmd.SetArgs([]string{"version"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("captureStdout error: %v", err)
	}
	if !strings.Contains(out, "version: 1.2.3") {
		t.Errorf("expected output to contain custom version, got %q", out)
	}
}
