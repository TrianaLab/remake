package store

import (
	"context"
	"fmt"
	"os"

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
	cfg *config.Config
}

func New(cfg *config.Config) Store {
	return &ArtifactStore{cfg: cfg}
}

func (s *ArtifactStore) Login(ctx context.Context, reg, user, pass string) error {
	client := registry.NewClient(s.cfg, reg)
	return client.Login(ctx, reg, user, pass)
}

func (s *ArtifactStore) Push(ctx context.Context, reference, path string) error {
	switch s.cfg.ParseReference(reference) {
	case config.ReferenceHTTP:
		return fmt.Errorf("pushing to HTTP(s) references is not supported")
	case config.ReferenceLocal:
		return fmt.Errorf("pushing local references is not supported")
	case config.ReferenceOCI:
		client := registry.NewClient(s.cfg, reference)
		if err := client.Push(ctx, reference, path); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		cacheRepo := cache.NewCache(s.cfg, reference)
		return cacheRepo.Push(ctx, reference, data)
	default:
		return fmt.Errorf("unknown reference type for %s", reference)
	}
}

func (s *ArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	switch s.cfg.ParseReference(reference) {
	case config.ReferenceLocal:
		return reference, nil
	default:
		cache := cache.NewCache(s.cfg, reference)
		if !s.cfg.NoCache {
			if path, err := cache.Pull(ctx, reference); err == nil {
				return path, nil
			}
		}
		client := registry.NewClient(s.cfg, reference)
		data, err := client.Pull(ctx, reference)
		if err != nil {
			return "", err
		}
		if err := cache.Push(ctx, reference, data); err != nil {
			return "", err
		}
		return cache.Pull(ctx, reference)
	}
}
