package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
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
		t.Error("pullCmd.Args() with 0 should fail")
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

func TestPullCmd_Success_LocalRegistry(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	reg := registry.New()
	ts := httptest.NewServer(reg)
	defer ts.Close()
	host := strings.TrimPrefix(ts.URL, "http://")

	dir := t.TempDir()
	mf := filepath.Join(dir, "Makefile")
	wantContent := "HELLO\n"
	os.WriteFile(mf, []byte(wantContent), 0644)

	ctx := context.Background()
	fsStore, err := file.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	desc, err := fsStore.Add(ctx, "Makefile", "application/x-makefile", "")
	if err != nil {
		t.Fatal(err)
	}
	manifestDesc, err := oras.PackManifest(ctx, fsStore, oras.PackManifestVersion1_1,
		"application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
	if err != nil {
		t.Fatal(err)
	}
	if err := fsStore.Tag(ctx, manifestDesc, "latest"); err != nil {
		t.Fatal(err)
	}
	repoRef := host + "/myrepo"
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		t.Fatal(err)
	}
	repo.PlainHTTP = true
	repo.Client = &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache()}
	if _, err := oras.Copy(ctx, fsStore, "latest", repo, "latest", oras.DefaultCopyOptions); err != nil {
		t.Fatal(err)
	}

	os.Remove("Makefile")

	pullCmd.Flags().Set("insecure", "true")
	uri := "oci://" + host + "/myrepo:latest"
	if err := pullCmd.RunE(pullCmd, []string{uri}); err != nil {
		t.Fatalf("pullCmd.RunE() error = %v", err)
	}

	if err := pullCmd.RunE(pullCmd, []string{uri}); err != nil {
		t.Fatalf("pullCmd.RunE() error = %v", err)
	}
}

func TestRunCmd_RemoteFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "all:\n\t@echo R\n")
	}))
	defer srv.Close()

	runFile = srv.URL + "/Makefile"
	runCmd.Flags().Set("no-cache", "true")

	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := runCmd.RunE(runCmd, []string{"all"}); err != nil {
		t.Fatalf("runCmd.RunE() = %v", err)
	}
	w.Close()
	os.Stdout = oldOut
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "R") {
		t.Errorf("got %q, want output R", buf.String())
	}
}

func TestRenderCmd_DefaultFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	work := t.TempDir()
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	content := "HELLO\n"
	if err := os.WriteFile("makefile", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	renderFile = ""
	renderNoCache = false

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := renderCmd.RunE(renderCmd, []string{}); err != nil {
		t.Fatalf("renderCmd default error = %v", err)
	}
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	if buf.String() != content {
		t.Errorf("renderCmd default = %q; want %q", buf.String(), content)
	}
}

func TestRenderCmd_FetchError(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	renderFile = ts.URL + "/makefile"
	renderNoCache = true

	err := renderCmd.RunE(renderCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "HTTP error 404") {
		t.Errorf("expected HTTP error 404, got %v", err)
	}
}

func TestPushCmd_Success_LocalRegistry(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	reg := registry.New()
	ts := httptest.NewServer(reg)
	defer ts.Close()
	host := strings.TrimPrefix(ts.URL, "http://")

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	content := "all:\n\techo OK\n"
	if err := os.WriteFile("Makefile", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	fsStore, err := file.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer fsStore.Close()
	desc, err := fsStore.Add(ctx, "Makefile", "application/x-makefile", "")
	if err != nil {
		t.Fatal(err)
	}
	manifestDesc, err := oras.PackManifest(ctx, fsStore, oras.PackManifestVersion1_1,
		"application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
	if err != nil {
		t.Fatal(err)
	}
	if err := fsStore.Tag(ctx, manifestDesc, "latest"); err != nil {
		t.Fatal(err)
	}

	pushFile = "Makefile"
	pushCmd.Flags().Set("insecure", "true")
	err = pushCmd.RunE(pushCmd, []string{host + "/myrepo"})
	if err != nil {
		t.Fatalf("pushCmd success failed: %v", err)
	}
}

func TestRunCmd_RemoteFetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	runFile = srv.URL + "/Makefile"
	runCmd.Flags().Set("no-cache", "true")

	err := runCmd.RunE(runCmd, []string{"all"})
	if err == nil || !strings.Contains(err.Error(), "HTTP error 404") {
		t.Errorf("expected HTTP error 404, got %v", err)
	}
}

func TestRunCmd_FileFlag(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	work := t.TempDir()
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}
	content := "all:\n\techo FLAG\n"
	if err := os.WriteFile("custom.mk", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	runFile = "custom.mk"
	runCmd.Flags().Set("file", "custom.mk")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := runCmd.RunE(runCmd, []string{"all"}); err != nil {
		t.Fatalf("runCmd file flag failed: %v", err)
	}
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "FLAG") {
		t.Errorf("runCmd file flag output = %q; want contains FLAG", buf.String())
	}
}

func TestPullCmd_InitConfigError(t *testing.T) {
	pullInitConfig = func() error { return errors.New("init err") }
	err := pullCmd.RunE(pullCmd, []string{"example.com/myrepo:latest"})

	if err == nil || !strings.Contains(err.Error(), "init err") {
		t.Fatalf("expected init err, got %v", err)
	}
}

func TestLoginCmd_loginInitConfigError(t *testing.T) {
	loginInitConfig = func() error { return errors.New("init err") }
	err := loginCmd.RunE(loginCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "init err") {
		t.Fatalf("expected init err, got %v", err)
	}
}

func TestLoginCmd_InvalidRegistryError(t *testing.T) {
	loginInitConfig = func() error { return nil }
	newRegistry = func(endpoint string) (*remote.Registry, error) {
		return nil, errors.New("bad registry")
	}
	err := loginCmd.RunE(loginCmd, []string{"foo"})
	if err == nil || !strings.Contains(err.Error(), "invalid registry foo") {
		t.Fatalf("expected invalid registry foo, got %v", err)
	}
}

func TestLoginCmd_PasswordReadError(t *testing.T) {
	loginUsername = "user"
	loginPassword = ""
	loginInsecure = true
	loginInitConfig = func() error { return nil }
	saveConfig = func() error { return nil }
	newRegistry = remote.NewRegistry
	passwordReader = func(fd int) ([]byte, error) {
		return nil, errors.New("pw error")
	}
	err := loginCmd.RunE(loginCmd, []string{"example.com"})
	if err == nil || !strings.Contains(err.Error(), "pw error") {
		t.Fatalf("expected pw error, got %v", err)
	}
}

func TestLoginCmd_SaveConfigError(t *testing.T) {
	loginInitConfig = func() error { return nil }
	inputReader = func() *bufio.Reader {
		return bufio.NewReader(strings.NewReader("user\n"))
	}
	passwordReader = func(fd int) ([]byte, error) {
		return []byte("pass"), nil
	}
	rServer := httptest.NewServer(registry.New())
	defer rServer.Close()
	endpoint := strings.TrimPrefix(rServer.URL, "http://")
	newRegistry = remote.NewRegistry
	saveConfig = func() error { return errors.New("save err") }
	err := loginCmd.RunE(loginCmd, []string{endpoint})
	if err == nil || !strings.Contains(err.Error(), "save err") {
		t.Fatalf("expected save err, got %v", err)
	}
}

func TestLoginCmd_SuccessInteractive(t *testing.T) {
	loginInitConfig = func() error { return nil }
	saveConfig = func() error { return nil }
	inputReader = func() *bufio.Reader {
		return bufio.NewReader(strings.NewReader("user\n"))
	}
	passwordReader = func(fd int) ([]byte, error) {
		return []byte("pass"), nil
	}
	rServer := httptest.NewServer(registry.New())
	defer rServer.Close()
	endpoint := strings.TrimPrefix(rServer.URL, "http://")
	newRegistry = remote.NewRegistry
	loginInsecure = true
	old := os.Stdout
	rPipe, wPipe, _ := os.Pipe()
	os.Stdout = wPipe
	err := loginCmd.RunE(loginCmd, []string{endpoint})
	wPipe.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	out, _ := io.ReadAll(rPipe)
	want := fmt.Sprintf("Connected to %s successfully", endpoint)
	if !strings.Contains(string(out), want) {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestLoginCmd_UsernameReadSuccess(t *testing.T) {
	loginUsername = ""
	loginPassword = ""
	loginInitConfig = func() error { return nil }
	saveConfig = func() error { return nil }
	inputReader = func() *bufio.Reader {
		return bufio.NewReader(strings.NewReader("alice_user\n"))
	}
	passwordReader = func(fd int) ([]byte, error) {
		return []byte("pass123"), nil
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()
	hostPort := strings.TrimPrefix(srv.URL, "http://")
	newRegistry = func(endpoint string) (*remote.Registry, error) {
		r, err := remote.NewRegistry(endpoint)
		if err != nil {
			return nil, err
		}
		r.PlainHTTP = true
		return r, nil
	}

	old := os.Stdout
	rp, wp, _ := os.Pipe()
	os.Stdout = wp

	err := loginCmd.RunE(loginCmd, []string{hostPort})
	wp.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	outBytes, _ := io.ReadAll(rp)
	out := string(outBytes)
	if !strings.Contains(out, "Username: ") {
		t.Errorf("missing Username prompt, got %q", out)
	}
	if loginUsername != "alice_user" {
		t.Errorf("expected loginUsername 'alice_user', got %q", loginUsername)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}
func TestLoginCmd_UsernameReadError(t *testing.T) {
	loginUsername = ""
	loginPassword = ""
	loginInitConfig = func() error { return nil }
	saveConfig = func() error { return nil }
	inputReader = func() *bufio.Reader {
		errR := &errReader{}
		return bufio.NewReader(errR)
	}
	passwordReader = func(fd int) ([]byte, error) { return []byte("pass123"), nil }
	newRegistry = func(endpoint string) (*remote.Registry, error) {
		r := &remote.Registry{}
		return r, nil
	}

	err := loginCmd.RunE(loginCmd, []string{"example.com"})
	if err == nil || !strings.Contains(err.Error(), "read error") {
		t.Fatalf("expected read error, got %v", err)
	}
}
