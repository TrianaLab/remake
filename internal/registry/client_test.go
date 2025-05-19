package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/viper"
	authpkg "oras.land/oras-go/v2/registry/remote/auth"
)

func TestNewRepository_NoCreds(t *testing.T) {
	viper.Reset()
	repo, err := newRepository("example.com/myrepo", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !repo.PlainHTTP {
		t.Error("expected PlainHTTP true")
	}
	if _, ok := repo.Client.(*authpkg.Client); !ok {
		t.Errorf("expected auth.Client, got %T", repo.Client)
	}
}

func TestNewRepository_WithCreds(t *testing.T) {
	viper.Reset()
	key := config.NormalizeKey("example.com")
	viper.Set(fmt.Sprintf("registries.%s.username", key), "user")
	viper.Set(fmt.Sprintf("registries.%s.password", key), "pass")
	repo, err := newRepository("example.com/myrepo", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	client, ok := repo.Client.(*authpkg.Client)
	if !ok {
		t.Fatalf("expected auth.Client, got %T", repo.Client)
	}
	if client.Credential == nil {
		t.Error("expected credential to be set")
	}
}

func TestNewRepository_InvalidHost(t *testing.T) {
	viper.Reset()
	_, err := newRepository("::::", false)
	if err == nil {
		t.Fatal("expected error for invalid host URL")
	}
}

func TestLogin_InvalidHost(t *testing.T) {
	viper.Reset()
	err := Login("::::", "u", "p")
	if err == nil {
		t.Fatal("expected error for invalid registry login")
	}
}

func TestPull_FileStoreError(t *testing.T) {
	f, err := os.CreateTemp("", "file")
	if err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	path := f.Name()
	f.Close()
	_, err = Pull(context.Background(), "example.com", "repo", "tag", path, false)
	if err == nil || !strings.Contains(err.Error(), "is not a directory") {
		t.Errorf("expected file store creation error, got %v", err)
	}
}

func TestPull_CopyError(t *testing.T) {
	d := t.TempDir()
	_, err := Pull(context.Background(), "example.com", "repo", "tag", d, false)
	if err == nil || !strings.Contains(err.Error(), "failed to pull artifact") {
		t.Errorf("expected pull artifact error, got %v", err)
	}
}

func TestPush_InvalidHost(t *testing.T) {
	err := Push(context.Background(), "::::", "repo", "tag", "file", false)
	if err == nil {
		t.Fatal("expected error for invalid host in Push")
	}
}

func TestPush_AddError(t *testing.T) {
	d := t.TempDir()
	nonexistent := filepath.Join(d, "nofile")
	err := Push(context.Background(), "example.com", "repo", "tag", nonexistent, false)
	if err == nil || !strings.Contains(err.Error(), "failed to add file to store") {
		t.Errorf("expected add file to store error, got %v", err)
	}
}
