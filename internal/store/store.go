package store

import (
	"context"
	"os"
	"path/filepath"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/cache"
	"github.com/TrianaLab/remake/internal/registry"
)

type Store interface {
	Login(ctx context.Context, registry, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) (string, error)
}

type ArtifactStore struct {
	client registry.Client
	cache  cache.CacheRepository
	cfg    *config.Config
}

func New(cfg *config.Config) Store {
	return &ArtifactStore{client: registry.New(cfg), cache: cache.New(cfg), cfg: cfg}
}

func (s *ArtifactStore) Login(ctx context.Context, registry, user, pass string) error {
	return s.client.Login(ctx, registry, user, pass)
}

func (s *ArtifactStore) Push(ctx context.Context, reference, path string) error {
	if err := s.client.Push(ctx, reference, path); err != nil {
		return err
	}
	return nil
}

func (s *ArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	if !s.cfg.NoCache {
		path, err := s.cache.Pull(ctx, reference)
		if err == nil {
			return path, nil
		}
	}

	data, err := s.client.Pull(ctx, reference)
	if err != nil {
		return "", err
	}

	if !s.cfg.NoCache {
		_ = s.cache.Push(ctx, reference, data)
		return s.cache.Pull(ctx, reference)
	}

	tmpFile, err := os.CreateTemp("", filepath.Base(reference)+"-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}
