package fetch

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

// OCIFetcher retrieves OCI artifacts, honoring cache and optional credentials.
type OCIFetcher struct{}

func (o *OCIFetcher) Fetch(ref string, useCache bool) (string, error) {
	// First, consult cache
	cacheFetcher := &CacheFetcher{}
	if path, err := cacheFetcher.Fetch(ref, useCache); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	// Not cached or cache disabled: proceed remote
	clean := strings.TrimPrefix(ref, "oci://")
	parts := strings.SplitN(clean, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid OCI reference: %s", ref)
	}
	host, repoTag := parts[0], parts[1]

	tag := "latest"
	if idx := strings.LastIndex(repoTag, ":"); idx != -1 {
		tag = repoTag[idx+1:]
		repoTag = repoTag[:idx]
	}

	cacheDir := filepath.Join(viper.GetString("cacheDir"), host, repoTag, tag)
	fs, err := file.New(cacheDir)
	if err != nil {
		return "", err
	}
	defer fs.Close()

	repoRef := fmt.Sprintf("%s/%s", host, repoTag)
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return "", err
	}

	if viper.GetBool("insecure") {
		repo.PlainHTTP = true
	}

	// Optionally set credentials if available
	key := config.NormalizeKey(host)
	username := viper.GetString(fmt.Sprintf("registries.%s.username", key))
	password := viper.GetString(fmt.Sprintf("registries.%s.password", key))
	if username != "" && password != "" {
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: auth.StaticCredential(host, auth.Credential{
				Username: username,
				Password: password,
			}),
		}
	} else {
		repo.Client = &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache()}
	}

	ctx := context.Background()
	if _, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions); err != nil {
		return "", fmt.Errorf("failed to fetch OCI artifact: %w", err)
	}

	// Return first file in cache dir
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			return filepath.Join(cacheDir, e.Name()), nil
		}
	}

	return "", fmt.Errorf("no file found after fetching %s", ref)
}
