package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var orasCopy = oras.Copy

// newRepository creates and configures a remote.Repository with credentials.
func newRepository(hostURL string, insecure bool) (*remote.Repository, error) {
	repo, err := remote.NewRepository(hostURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository for %s: %w", hostURL, err)
	}
	repo.PlainHTTP = insecure

	parts := strings.SplitN(hostURL, "/", 2)
	domain := parts[0]
	key := config.NormalizeKey(domain)
	userKey := fmt.Sprintf("registries.%s.username", key)
	passKey := fmt.Sprintf("registries.%s.password", key)
	username := viper.GetString(userKey)
	password := viper.GetString(passKey)
	if username != "" && password != "" {
		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(domain, auth.Credential{Username: username, Password: password}),
		}
	} else {
		repo.Client = &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache()}
	}
	return repo, nil
}

// Login validates credentials against the registry and saves them.
func Login(host, username, password string) error {
	reg, err := remote.NewRegistry(host)
	if err != nil {
		return fmt.Errorf("invalid registry %s: %w", host, err)
	}
	reg.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(host, auth.Credential{Username: username, Password: password}),
	}
	if err := reg.Ping(context.Background()); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	key := config.NormalizeKey(host)
	viper.Set(fmt.Sprintf("registries.%s.username", key), username)
	viper.Set(fmt.Sprintf("registries.%s.password", key), password)
	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// Pull downloads an OCI artifact to destDir using ORAS copy.
func Pull(ctx context.Context, host, repo, ref, destDir string, insecure bool) (string, error) {
	// Ensure destDir is not an existing file
	if info, err := os.Stat(destDir); err == nil && !info.IsDir() {
		return "", fmt.Errorf("failed to create file store: %s is not a directory", destDir)
	}

	hostURL := fmt.Sprintf("%s/%s", host, repo)
	repoClient, err := newRepository(hostURL, insecure)
	if err != nil {
		return "", err
	}
	fs, err := file.New(destDir)
	if err != nil {
		return "", fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	desc, err := orasCopy(ctx, repoClient, ref, fs, ref, oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("failed to pull artifact: %w", err)
	}
	_ = desc

	entries, err := os.ReadDir(destDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			return filepath.Join(destDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no artifact found in %s", destDir)
}

// Push uploads a file as an OCI artifact using ORAS copy.
func Push(ctx context.Context, host, repo, ref, filePath string, insecure bool) error {
	hostURL := fmt.Sprintf("%s/%s", host, repo)
	repoClient, err := newRepository(hostURL, insecure)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}
	fs, err := file.New(dir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	fileName := filepath.Base(filePath)
	if _, err := fs.Add(ctx, fileName, "application/octet-stream", ""); err != nil {
		return fmt.Errorf("failed to add file to store: %w", err)
	}

	if _, err := orasCopy(ctx, fs, fileName, repoClient, ref, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("failed to push artifact: %w", err)
	}

	return nil
}
