package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// restoreRunStubs resets package-level variables after tests
func restoreRunStubs(origFile string, origNoCache bool, origDefault func() string, origRun func(string, []string, bool) error) {
	runFile = origFile
	runNoCache = origNoCache
	defaultMakefileFn = origDefault
	runFn = origRun
}

func TestRunCmd_WithFileFlag(t *testing.T) {
	origFile, origNoCache := runFile, runNoCache
	origDef, origRun := defaultMakefileFn, runFn
	defer restoreRunStubs(origFile, origNoCache, origDef, origRun)

	called := false
	var gotFile string
	var gotTargets []string
	var gotCache bool

	// stub runFn
	runFn = func(f string, targets []string, useCache bool) error {
		called = true
		gotFile, gotTargets, gotCache = f, targets, useCache
		return nil
	}

	runFile = "Custom.mk"
	runNoCache = true

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"run", "-f", "Custom.mk", "--no-cache", "build", "test"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected runFn to be called")
	}
	if gotFile != "Custom.mk" || strings.Join(gotTargets, ",") != "build,test" || gotCache {
		t.Errorf("unexpected args: file=%q targets=%v cache=%v", gotFile, gotTargets, gotCache)
	}
}

func TestRunCmd_DefaultFileMissing(t *testing.T) {
	origFile, origNoCache := runFile, runNoCache
	origDef, origRun := defaultMakefileFn, runFn
	defer restoreRunStubs(origFile, origNoCache, origDef, origRun)

	runFile = ""
	defaultMakefileFn = func() string { return "" }
	runNoCache = false

	rootCmd.SetArgs([]string{"run", "build"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "no Makefile found; specify with -f") {
		t.Fatalf("expected missing Makefile error, got %v", err)
	}
}

func TestRunCmd_DefaultFileThenErrorAndSuccess(t *testing.T) {
	origFile, origNoCache := runFile, runNoCache
	origDef, origRun := defaultMakefileFn, runFn
	defer restoreRunStubs(origFile, origNoCache, origDef, origRun)

	dir := t.TempDir()
	os.Chdir(dir)
	const name = "Makefile"
	if err := os.WriteFile(name, []byte("all:\n\techo hi\n"), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}
	defaultMakefileFn = func() string { return name }

	// success
	called := false
	runFn = func(f string, targets []string, useCache bool) error {
		called = true
		return nil
	}
	rootCmd.SetArgs([]string{"run", "build"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error on success, got %v", err)
	}
	if !called {
		t.Fatal("expected runFn called on success")
	}

	// error
	runFn = func(f string, targets []string, useCache bool) error {
		return errors.New("oops")
	}
	rootCmd.SetArgs([]string{"run", "build"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "oops") {
		t.Fatalf("expected run error, got %v", err)
	}
}

func init() {
	// reset flags for reuse
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	runCmd.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
}
