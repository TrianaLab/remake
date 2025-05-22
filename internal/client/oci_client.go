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
	client := &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(registry, auth.Credential{Username: user, Password: pass}),
	}
	reg.Client = client
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
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return err
	}
	repoRef := ref.Context()
	repo, err := remote.NewRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return err
	}
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

	dir := filepath.Dir(path)
	fs, err := file.New(dir)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := fs.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	mediaType := "application/vnd.remake.file"
	fileDesc, err := fs.Add(ctx, path, mediaType, "")
	if err != nil {
		return fmt.Errorf("adding file to store: %w", err)
	}

	artifactType := "application/vnd.remake.artifact"
	opts := oras.PackManifestOptions{Layers: []v1.Descriptor{fileDesc}}
	manifestDesc, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	if err != nil {
		return fmt.Errorf("packing manifest: %w", err)
	}
	if manifestDesc.Digest.String() == "" {
		return fmt.Errorf("invalid manifest descriptor: empty digest")
	}

	tag := ref.Identifier()
	if err := fs.Tag(ctx, manifestDesc, tag); err != nil {
		return fmt.Errorf("tagging manifest: %w", err)
	}

	if _, err := oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("pushing to remote: %w", err)
	}

	return nil
}

// Pull downloads the artifact data for the given reference from the OCI registry.
// It retrieves the manifest and returns the contents of the first layer (Makefile data).
func (c *OCIClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	ref, err := name.ParseReference(reference, name.WithDefaultRegistry(c.cfg.DefaultRegistry))
	if err != nil {
		return nil, err
	}

	repoRef := ref.Context()
	repo, err := remote.NewRepository(repoRef.RegistryStr() + "/" + repoRef.RepositoryStr())
	if err != nil {
		return nil, err
	}

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

	store := memory.New()
	if _, err := oras.Copy(ctx, repo, ref.Identifier(), store, ref.Identifier(), oras.DefaultCopyOptions); err != nil {
		return nil, err
	}

	manifestDesc, err := store.Resolve(ctx, ref.Identifier())
	if err != nil {
		return nil, err
	}

	manifestBytes, err := content.FetchAll(ctx, store, manifestDesc)
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
	data, err := content.FetchAll(ctx, store, layerDesc)
	if err != nil {
		return nil, err
	}

	return data, nil
}
