package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
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

func TestRunCmd_SuccessLocal(t *testing.T) {
	work := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("Makefile", []byte("all:\n\t@echo done\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runFile = ""
	runNoCache = false
	if err := runCmd.RunE(runCmd, []string{"all"}); err != nil {
		t.Errorf("runCmd.RunE() success = %v; want nil", err)
	}
}

func TestPushCmd_NoArgs(t *testing.T) {
	if err := pushCmd.RunE(pushCmd, []string{}); err == nil {
		t.Error("pushCmd without args should fail, err = nil")
	}
}

func TestPushCmd_InvalidRef(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	pushFile = "Makefile"
	err := pushCmd.RunE(pushCmd, []string{"noSlash"})
	if err == nil || !strings.Contains(err.Error(), "invalid reference") {
		t.Errorf("expected 'invalid reference', got %v", err)
	}
}

func TestPullCmd_InvalidRef(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := pullCmd.RunE(pullCmd, []string{"solo_un_elemento"})
	if err == nil || !strings.Contains(err.Error(), "invalid reference") {
		t.Errorf("expected 'invalid reference', got %v", err)
	}
}

func TestCommandFlags(t *testing.T) {
	if pullCmd.Flag("file") == nil {
		t.Error("pullCmd should have 'file' flag")
	}
	if pushCmd.Flag("file") == nil {
		t.Error("pushCmd should have 'file' flag")
	}
	if runCmd.Flag("file") == nil || runCmd.Flag("no-cache") == nil {
		t.Error("runCmd should have 'file' and 'no-cache' flags")
	}
	if renderCmd.Flag("file") == nil || renderCmd.Flag("no-cache") == nil {
		t.Error("renderCmd should have 'file' and 'no-cache' flags")
	}
	if loginCmd.Flags().Lookup("username") == nil || loginCmd.Flags().Lookup("password") == nil {
		t.Error("loginCmd should have 'username' and 'password' flags")
	}
}

func TestLoginCmd_ArgsValidation(t *testing.T) {
	if err := loginCmd.Args(loginCmd, []string{"a", "b"}); err == nil {
		t.Error("loginCmd.Args() with 2 args should fail")
	}
}
func TestPullCmd_TooManyArgs(t *testing.T) {
	if err := pullCmd.Args(pullCmd, []string{"a", "b"}); err == nil {
		t.Error("pullCmd.Args() with 2 args should fail")
	}
}
func TestRunCmd_TooFewArgs(t *testing.T) {
	if err := runCmd.Args(runCmd, []string{}); err == nil {
		t.Error("runCmd.Args() with no args should fail")
	}
}

func TestLoginCmd_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	regHandler := registry.New()
	ts := httptest.NewServer(regHandler)
	defer ts.Close()

	hostPort := strings.TrimPrefix(ts.URL, "http://")

	loginCmd.Flags().Set("username", "u")
	loginCmd.Flags().Set("password", "p")
	loginCmd.Flags().Set("insecure", "true")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := loginCmd.RunE(loginCmd, []string{hostPort})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("loginCmd.RunE() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	want := "Connected to " + hostPort + " successfully"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("loginCmd output = %q; want %q", buf.String(), want)
	}
}

func TestRenderCmd_Remote(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("REMOTE\n"))
	}))
	defer ts.Close()

	renderFile = ts.URL + "/Makefile"
	renderNoCache = true

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
	if buf.String() != "REMOTE\n" {
		t.Errorf("renderCmd output = %q; want %q", buf.String(), "REMOTE\n")
	}
}

func TestPushCmd_PushError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dir := t.TempDir()
	makefile := filepath.Join(dir, "Makefile")
	os.WriteFile(makefile, []byte("all:\n\techo hi\n"), 0644)
	pushFile = makefile

	err := pushCmd.RunE(pushCmd, []string{"example.com/repo"})
	if err == nil || !strings.Contains(err.Error(), "failed to push artifact") {
		t.Errorf("expected 'failed to push artifact', got %v", err)
	}
}

func TestPullCmd_PullError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := pullCmd.RunE(pullCmd, []string{"example.com/repo"})
	if err == nil || !strings.Contains(err.Error(), "failed to pull artifact") {
		t.Errorf("expected 'failed to pull artifact', got %v", err)
	}
}

func TestRootCmd_HasSubCommands(t *testing.T) {
	names := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"login", "pull", "push", "render", "run", "version"} {
		if !names[want] {
			t.Errorf("rootCmd missing %q", want)
		}
	}
}
