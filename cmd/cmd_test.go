package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderCmd_Local(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	work := t.TempDir()
	mf := filepath.Join(work, "Makefile")
	content := "OK\n"
	_ = os.WriteFile(mf, []byte(content), 0644)

	renderFile = mf
	renderNoCache = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := renderCmd.RunE(renderCmd, []string{})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("renderCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.String() != content {
		t.Errorf("renderCmd output = %q; want %q", buf.String(), content)
	}
}

func TestRunCmd_NoMake(t *testing.T) {
	t.Setenv("PATH", "")
	err := runCmd.RunE(runCmd, []string{"foo"})
	if err == nil || !strings.Contains(err.Error(), "make not found") {
		t.Errorf("RunE no-make = %v; want error de make not found", err)
	}
}

func TestVersionCmd(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	versionCmd.Run(versionCmd, []string{})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "remake version dev") {
		t.Errorf("versionCmd = %q; want 'remake version dev'", buf.String())
	}
}

func TestPushCmd_NoFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := pushCmd.RunE(pushCmd, []string{"remote/ref"})
	if err == nil || !strings.Contains(err.Error(), "no Makefile found") {
		t.Errorf("pushCmd no-file = %v; want error de no Makefile", err)
	}
}

func TestPullCmd_ArgsValidation(t *testing.T) {
	if err := pullCmd.Args(pullCmd, []string{}); err == nil {
		t.Error("pullCmd.Args() con 0 args debía fallar")
	}
}

func TestLoginCmd_InvalidRegistry(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	loginUsername = "u"
	loginPassword = "p"

	err := loginCmd.RunE(loginCmd, []string{"$$bad"})
	if err == nil || !strings.Contains(err.Error(), "invalid registry") {
		t.Errorf("loginCmd invalid = %v; want invalid registry", err)
	}
}

func TestLoginCmd_PingFail(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	loginUsername = "u"
	loginPassword = "p"
	err := loginCmd.RunE(loginCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "login failed") {
		t.Errorf("loginCmd ping fail = %v; want login failed", err)
	}
}
