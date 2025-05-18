package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// restoreRootStubs resets initConfigFunc and exitFunc to originals
func restoreRootStubs(origInit func() error, origExit func(int)) {
	initConfigFunc = origInit
	exitFunc = origExit
}

func TestPersistentPreRun_Success(t *testing.T) {
	origInit, origExit := initConfigFunc, exitFunc
	defer restoreRootStubs(origInit, origExit)
	initConfigFunc = func() error { return nil }

	// dummy command to trigger PersistentPreRunE
	dummy := &cobra.Command{
		Use:  "dummy",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	rootCmd.AddCommand(dummy)
	defer rootCmd.RemoveCommand(dummy)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"dummy"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPersistentPreRun_Failure(t *testing.T) {
	origInit, origExit := initConfigFunc, exitFunc
	defer restoreRootStubs(origInit, origExit)
	initConfigFunc = func() error { return errors.New("boom") }

	dummy := &cobra.Command{
		Use:  "dummy2",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	rootCmd.AddCommand(dummy)
	defer rootCmd.RemoveCommand(dummy)

	rootCmd.SetArgs([]string{"dummy2"})
	err := rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to initialize config: boom") {
		t.Fatalf("expected initConfig error, got %v", err)
	}
}

func TestExecute_ExitOnError(t *testing.T) {
	origInit, origExit := initConfigFunc, exitFunc
	defer restoreRootStubs(origInit, origExit)

	exitCode := 0
	exitFunc = func(code int) { exitCode = code }
	initConfigFunc = func() error { return nil }

	// failing subcommand
	fail := &cobra.Command{
		Use: "fail",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("fail")
		},
	}
	rootCmd.AddCommand(fail)
	defer rootCmd.RemoveCommand(fail)

	Execute()
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}
