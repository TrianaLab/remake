package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote"
)

func TestLogin_InvalidEndpoint(t *testing.T) {
	cmd := loginCmd
	cmd.SetArgs([]string{"invalid/endpoint"})
	err := cmd.RunE(cmd, []string{"invalid/endpoint"})
	if err == nil || !strings.Contains(err.Error(), "invalid registry invalid/endpoint") {
		t.Fatalf("expected invalid registry error, got %v", err)
	}
}

func TestLogin_ReadUsernameError(t *testing.T) {
	loginUsername = ""
	loginPassword = "pass"
	loginInsecure = false

	inputReader = func() *bufio.Reader { return bufio.NewReader(&errReader{}) }

	err := loginCmd.RunE(loginCmd, []string{"host"})
	if err == nil || !strings.Contains(err.Error(), "read error") {
		t.Fatalf("expected read error, got %v", err)
	}
}

type errReader struct{}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestLogin_ReadPasswordError(t *testing.T) {
	loginUsername = "user"
	loginPassword = ""
	loginInsecure = false

	passwordReader = func(fd int) ([]byte, error) { return nil, errors.New("pw error") }

	err := loginCmd.RunE(loginCmd, []string{"host"})
	if err == nil || !strings.Contains(err.Error(), "pw error") {
		t.Fatalf("expected pw error, got %v", err)
	}
}

func TestLogin_RegistryCreationError(t *testing.T) {
	loginUsername = "user"
	loginPassword = "pass"
	loginInsecure = false

	newRegistry = func(endpoint string) (*remote.Registry, error) {
		return nil, fmt.Errorf("new error")
	}

	err := loginCmd.RunE(loginCmd, []string{"host"})
	if err == nil || !strings.Contains(err.Error(), "invalid registry host: new error") {
		t.Fatalf("expected newRegistry error, got %v", err)
	}
}

func TestLogin_PingError(t *testing.T) {
	loginUsername = "user"
	loginPassword = "pass"
	loginInsecure = true

	newRegistry = func(endpoint string) (*remote.Registry, error) {
		return &remote.Registry{}, nil
	}

	err := loginCmd.RunE(loginCmd, []string{"host"})
	if err == nil || !strings.Contains(err.Error(), "login failed") {
		t.Fatalf("expected ping failed, got %v", err)
	}
}

func TestLogin_SaveConfigError(t *testing.T) {
	loginUsername = "user"
	loginPassword = "pass"
	loginInsecure = true

	// start fake registry server with HTTP
	rServer := httptest.NewServer(registry.New())
	defer rServer.Close()
	endpoint := strings.TrimPrefix(rServer.URL, "http://")

	newRegistry = remote.NewRegistry

	saveConfig = func() error { return errors.New("save error") }

	tmp := t.TempDir()
	if err := os.Setenv("HOME", tmp); err != nil {
		t.Fatal(err)
	}

	viper.Reset()

	err := loginCmd.RunE(loginCmd, []string{endpoint})
	if err == nil || !strings.Contains(err.Error(), "save error") {
		t.Fatalf("expected save error, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	loginUsername = "user"
	loginPassword = "pass"
	loginInsecure = true

	// start fake registry server with HTTP
	rServer := httptest.NewServer(registry.New())
	defer rServer.Close()
	endpoint := strings.TrimPrefix(rServer.URL, "http://")

	newRegistry = remote.NewRegistry
	// stub saveConfig to succeed
	saveConfig = func() error { return nil }

	// capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// set HOME for config path
	tmp := t.TempDir()
	if err := os.Setenv("HOME", tmp); err != nil {
		t.Fatal(err)
	}
	viper.Reset()

	err := loginCmd.RunE(loginCmd, []string{endpoint})
	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(out), fmt.Sprintf("Connected to %s successfully", endpoint)) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func resetGlobals() {
	loginUsername = ""
	loginPassword = ""
	loginInsecure = false
	newRegistry = remote.NewRegistry
	saveConfig = config.SaveConfig
	passwordReader = term.ReadPassword
	inputReader = func() *bufio.Reader { return bufio.NewReader(os.Stdin) }
}

func TestMain(m *testing.M) {
	code := m.Run()
	resetGlobals()
	os.Exit(code)
}
