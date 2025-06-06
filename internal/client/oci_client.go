// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// THE SOFTWARE.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/viper"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/TrianaLab/remake/config"
)

// These vars allows us to override functions in tests.
var (
	newRepository  = remote.NewRepository
	newFileStore   = file.New
	packManifest   = oras.PackManifest
	copyFunc       = oras.Copy
	contentFetcher = content.FetchAll
	absPathFunc    = filepath.Abs
)

// OCIClient provides an implementation of Client for OCI registries.
// It uses oras and go-containerregistry to authenticate, push, and pull artifacts.
type OCIClient struct {
	cfg *config.Config
}

// NewOCIClient returns a new OCIClient initialized with the given configuration.
func NewOCIClient(cfg *config.Config) Client {
	return &OCIClient{cfg: cfg}
}

// Login authenticates to the specified OCI registry using the provided credentials.
// Successful login is persisted in the configuration file for future operations.
func (c *OCIClient) Login(ctx context.Context, registry, user, pass string) error {
	reg, err := remote.NewRegistry(registry)
	if err != nil {
		return err
	}
	clientAuth := &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(registry, auth.Credential{Username: user, Password: pass}),
	}
	reg.Client = clientAuth
	if err := reg.Ping(ctx); err != nil {
		return err
	}
	key := config.NormalizeKey(registry)
	viper.Set("registries."+key+".username", user)
	viper.Set("registries."+key+".password", pass)
	return viper.WriteConfig()
}

// Push uploads the local file at path as an OCI artifact to the given reference.
// It tags the artifact with the reference identifier and pushes it to the remote repository.
func (c *OCIClient) Push(ctx context.Context, reference, path string) error {
	// Validate and parse reference
	if strings.Contains(reference, "://") && !strings.HasPrefix(reference, "oci://") {
		return fmt.Errorf("invalid OCI reference: %s", reference)
	}
	raw := strings.ToLower(strings.TrimPrefix(reference, "oci://"))
	ref, err := name.ParseReference(raw, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return err
	}
	repoRef := ref.Context()

	repo, err := newRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return err
	}

	// Authenticate if credentials present
	key := config.NormalizeKey(repoRef.RegistryStr())
	user := viper.GetString("registries." + key + ".username")
	pass := viper.GetString("registries." + key + ".password")
	if user != "" {
		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: auth.StaticCredential(repoRef.RegistryStr(), auth.Credential{Username: user, Password: pass}),
		}
	}

	// Resolve absolute path and split directory
	absPath, err := absPathFunc(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path %s: %w", path, err)
	}
	dir := filepath.Dir(absPath)

	// Prepare a file store rooted at the file's directory
	fs, err := newFileStore(dir)
	if err != nil {
		return fmt.Errorf("creating file store: %w", err)
	}
	defer func() { _ = fs.Close() }()

	// Add the file using its absolute path to ensure tests find it
	mediaType := "application/vnd.remake.file"
	fileDesc, err := fs.Add(ctx, absPath, mediaType, "")
	if err != nil {
		return fmt.Errorf("adding file to store: %w", err)
	}

	// Pack manifest using injected function
	artifactType := "application/vnd.remake.artifact"
	opts := oras.PackManifestOptions{Layers: []v1.Descriptor{fileDesc}}
	manifestDesc, err := packManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return fmt.Errorf("packing manifest: %w", err)
	}
	if manifestDesc.Digest.String() == "" {
		return fmt.Errorf("invalid manifest descriptor: empty digest")
	}

	tag := ref.Identifier()
	_ = fs.Tag(ctx, manifestDesc, tag)

	// Push to remote using injected function
	if _, err := copyFunc(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}
	return nil
}

// Pull downloads the artifact data for the given reference from the OCI registry.
// It retrieves the manifest and returns the contents of the first layer (Makefile data).
func (c *OCIClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	if strings.Contains(reference, "://") && !strings.HasPrefix(reference, "oci://") {
		return nil, fmt.Errorf("invalid OCI reference: %s", reference)
	}
	raw := strings.ToLower(strings.TrimPrefix(reference, "oci://"))
	ref, err := name.ParseReference(raw, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return nil, err
	}
	repoRef := ref.Context()
	repo, err := newRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return nil, err
	}
	key := config.NormalizeKey(repoRef.RegistryStr())
	user := viper.GetString("registries." + key + ".username")
	pass := viper.GetString("registries." + key + ".password")
	if user != "" || pass != "" {
		repo.Client = &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache(),
			Credential: auth.StaticCredential(repoRef.RegistryStr(), auth.Credential{Username: user, Password: pass})}
	}

	store := memory.New()
	manifestDesc, err := copyFunc(ctx, repo, ref.Identifier(), store, ref.Identifier(), oras.DefaultCopyOptions)
	if err != nil {
		return nil, err
	}

	manifestBytes, err := contentFetcher(ctx, store, manifestDesc)
	if err != nil {
		return nil, err
	}
	var manifest v1.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("no layers found in artifact %s", reference)
	}

	layerDesc := manifest.Layers[0]
	data, err := contentFetcher(ctx, store, layerDesc)
	if err != nil {
		return nil, err
	}
	return data, nil
}
