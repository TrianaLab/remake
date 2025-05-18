package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// restoreTemplateStubs resets package-level vars after each test
func restoreTemplateStubs(origFile string, origNoCache bool,
	origDef func() string, origCache func() string,
	origRender func(string, string, bool) error,
) {
	templateFile = origFile
	templateNoCache = origNoCache
	templateDefaultMakefileFn = origDef
	cacheDirFn = origCache
	renderFn = origRender
}

func TestTemplateCmd_NoFileNoDefault(t *testing.T) {
	// Stub out hooks
	origFile, origNoCache := templateFile, templateNoCache
	origDef, origCache, origRender := templateDefaultMakefileFn, cacheDirFn, renderFn
	defer restoreTemplateStubs(origFile, origNoCache, origDef, origCache, origRender)

	templateFile = ""
	templateDefaultMakefileFn = func() string { return "" }
	cacheDirFn = func() string { return "" } // not used
	renderFn = func(src, out string, useCache bool) error { return nil }

	// Run command
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"template"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no Makefile found; specify with -f flag") {
		t.Fatalf("expected missing Makefile error, got %v", err)
	}
}

func TestTemplateCmd_RenderError(t *testing.T) {
	origFile, origNoCache := templateFile, templateNoCache
	origDef, origCache, origRender := templateDefaultMakefileFn, cacheDirFn, renderFn
	defer restoreTemplateStubs(origFile, origNoCache, origDef, origCache, origRender)

	templateFile = ""
	templateDefaultMakefileFn = func() string { return "Makefile" }
	// Provide some cache dir
	tmp := os.TempDir()
	cacheDirFn = func() string { return tmp }
	renderFn = func(src, out string, useCache bool) error {
		return errors.New("render failed")
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"template"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "render failed") {
		t.Fatalf("expected render error, got %v", err)
	}
}

func TestTemplateCmd_OpenError(t *testing.T) {
	origFile, origNoCache := templateFile, templateNoCache
	origDef, origCache, origRender := templateDefaultMakefileFn, cacheDirFn, renderFn
	defer restoreTemplateStubs(origFile, origNoCache, origDef, origCache, origRender)

	templateFile = ""
	templateDefaultMakefileFn = func() string { return "Makefile" }
	cacheDirTmp := t.TempDir()
	cacheDirFn = func() string { return cacheDirTmp }
	// render succeeds but does not actually write the file
	renderFn = func(src, out string, useCache bool) error { return nil }

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"template"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("expected open-file error, got %v", err)
	}
}

func captureStdout(f func()) (string, error) {
	// keep copy of real stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	// run the function
	f()

	// restore stdout and read buffer
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func TestTemplateCmd_Success(t *testing.T) {
	// stub hooks as before...
	origFile, origNoCache := templateFile, templateNoCache
	origDef, origCache, origRender := templateDefaultMakefileFn, cacheDirFn, renderFn
	defer restoreTemplateStubs(origFile, origNoCache, origDef, origCache, origRender)

	// prepare a real Makefile and cache directory
	tmp := t.TempDir()
	os.Chdir(tmp)
	const name = "Makefile"
	if err := os.WriteFile(name, []byte("foo"), 0644); err != nil {
		t.Fatalf("write source Makefile: %v", err)
	}
	templateDefaultMakefileFn = func() string { return name }
	cacheTmp := t.TempDir()
	cacheDirFn = func() string { return cacheTmp }

	// stub renderFn to write the generated file
	renderFn = func(src, out string, useCache bool) error {
		return os.WriteFile(out, []byte("rendered content"), 0644)
	}

	// capture os.Stdout around Execute
	rootCmd.SetArgs([]string{"template"})
	output, err := captureStdout(func() {
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
	if err != nil {
		t.Fatalf("could not capture stdout: %v", err)
	}

	if !strings.Contains(output, "rendered content") {
		t.Errorf("stdout missing rendered content: %q", output)
	}
	// generated file was removed
	if _, err := os.Stat(filepath.Join(cacheTmp, "Makefile.generated")); !os.IsNotExist(err) {
		t.Errorf("expected generated file removed, but still exists")
	}
}

func init() {
	// Clear any flag state between test runs
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	templateCmd.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
}
