package util

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

func FetchOCI(ociRef string) (string, error) {
	raw := strings.TrimPrefix(ociRef, "oci://")
	if !strings.Contains(raw, ":") {
		raw += ":latest"
	}
	parts := strings.SplitN(raw, "/", 2)
	host, repoAndTag := parts[0], parts[1]
	rt := strings.SplitN(repoAndTag, ":", 2)
	repoPath, tag := rt[0], rt[1]
	repoRef := host + "/" + repoPath
	cacheDir := config.GetCacheDir()
	os.MkdirAll(cacheDir, 0o755)
	fs, err := file.New(cacheDir)
	if err != nil {
		return "", err
	}
	defer fs.Close()
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return "", err
	}
	username := viper.GetString(fmt.Sprintf("registries.%s.username", host))
	password := viper.GetString(fmt.Sprintf("registries.%s.password", host))
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(host, auth.Credential{
			Username: username,
			Password: password,
		}),
	}
	ctx := context.Background()
	if _, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions); err != nil {
		return "", fmt.Errorf("failed to download OCI artifact: %w", err)
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".mk") {
			return filepath.Join(cacheDir, e.Name()), nil
		}
	}
	for _, e := range entries {
		if !e.IsDir() {
			return filepath.Join(cacheDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no file found in OCI artifact: %s", ociRef)
}
