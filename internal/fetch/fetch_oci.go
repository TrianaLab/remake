package fetch

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/internal/registry"
	"github.com/spf13/viper"
)

type OCIFetcher struct{}

func (o *OCIFetcher) Fetch(ref string, useCache bool) (string, error) {
	cache := &CacheFetcher{}
	if path, err := cache.Fetch(ref, useCache); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	clean := strings.TrimPrefix(ref, "oci://")
	parts := strings.SplitN(clean, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid OCI reference: %s", ref)
	}
	host, repoTag := parts[0], parts[1]
	var repoName, tag string
	if strings.Contains(repoTag, ":") {
		rt := strings.SplitN(repoTag, ":", 2)
		repoName, tag = rt[0], rt[1]
	} else {
		repoName = repoTag
		tag = "latest"
	}

	destDir := filepath.Join(viper.GetString("cacheDir"), host, repoName, tag)

	insecure := viper.GetBool("insecure")
	return registry.Pull(context.Background(), host, repoName, tag, destDir, insecure)
}
