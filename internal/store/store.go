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
	cfg   *config.Config
	cache cache.CacheRepository
}

func New(cfg *config.Config) Store {
	return &ArtifactStore{cfg: cfg, cache: cache.New(cfg)}
}

func (s *ArtifactStore) Login(ctx context.Context, registry, user, pass string) error {
	client := s.getClient(registry)
	return client.Login(ctx, registry, user, pass)
}

func (s *ArtifactStore) Push(ctx context.Context, reference, path string) error {
	switch s.cfg.ParseReference(reference) {
	case config.ReferenceHTTP:
		return fmt.Errorf("pushing to HTTP(s) references is not supported")
	case config.ReferenceLocal:
		return fmt.Errorf("pushing local references is not supported")
	case config.ReferenceOCI:
		client := s.getClient(reference)
		if err := client.Push(ctx, reference, path); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return s.cache.Push(ctx, reference, data)
	default:
		return fmt.Errorf("unknown reference type for %s", reference)
	}
}

func (s *ArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	switch s.cfg.ParseReference(reference) {
	case config.ReferenceLocal:
		return reference, nil
	case config.ReferenceHTTP:
		client := s.getClient(reference)
		data, err := client.Pull(ctx, reference)
		if err != nil {
			return "", err
		}
		if err := s.cache.Push(ctx, reference, data); err != nil {
			return "", err
		}
		return s.cache.Pull(ctx, reference)
	case config.ReferenceOCI:
		if !s.cfg.NoCache {
			if path, err := s.cache.Pull(ctx, reference); err == nil {
				return path, nil
			}
		}
		client := s.getClient(reference)
		data, err := client.Pull(ctx, reference)
		if err != nil {
			return "", err
		}
		if err := s.cache.Push(ctx, reference, data); err != nil {
			return "", err
		}
		return s.cache.Pull(ctx, reference)
	default:
		return "", fmt.Errorf("unknown reference type for %s", reference)
	}
}

func (s *ArtifactStore) getClient(reference string) registry.Client {
	t := s.cfg.ParseReference(reference)
	switch t {
	case config.ReferenceHTTP:
		return registry.NewHTTPClient()
	case config.ReferenceOCI:
		return registry.NewOCIClient(s.cfg)
	default:
		return registry.NewOCIClient(s.cfg)
	}
}
