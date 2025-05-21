package artifactstore

import (
	"context"

	"github.com/TrianaLab/remake/config"
)

// OCIRepository defines operations against an OCI registry
// CacheRepository defines local cache operations
// ArtifactStore composes registry and cache behaviors

type (
	OCIRepository interface {
		Login(ctx context.Context, registry, user, pass string) error
		Push(ctx context.Context, reference, path string) error
		Pull(ctx context.Context, reference string) (string, error)
	}

	CacheRepository interface {
		Push(ctx context.Context, reference, path string) error
		Pull(ctx context.Context, reference string) (string, error)
	}

	ArtifactStore interface {
		Login(ctx context.Context, registry, user, pass string) error
		Push(ctx context.Context, reference, path string) error
		Pull(ctx context.Context, reference string) (string, error)
	}
)

// DefaultArtifactStore is the default implementation of ArtifactStore
// It delegates to OCIRepository and optionally caches with CacheRepository

type DefaultArtifactStore struct {
	oci   OCIRepository
	cache CacheRepository
	cfg   *config.Config
}

// NewDefaultArtifactStore constructs a new ArtifactStore
func NewDefaultArtifactStore(oci OCIRepository, cache CacheRepository, cfg *config.Config) ArtifactStore {
	return &DefaultArtifactStore{oci: oci, cache: cache, cfg: cfg}
}

// Login proxies to the OCI registry client
func (s *DefaultArtifactStore) Login(ctx context.Context, registry, user, pass string) error {
	return s.oci.Login(ctx, registry, user, pass)
}

// Push uploads via OCI and then caches unless insecure
func (s *DefaultArtifactStore) Push(ctx context.Context, reference, path string) error {
	// always attempt push to registry
	if err := s.oci.Push(ctx, reference, path); err != nil {
		return err
	}
	// cache only if not insecure
	if !s.cfg.NoCache {
		return s.cache.Push(ctx, reference, path)
	}
	return nil
}

// internal/artifactstore/default_artifactstore.go
func (s *DefaultArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	// si no es insecure, probar cache primero
	if !s.cfg.NoCache {
		if path, err := s.cache.Pull(ctx, reference); err == nil {
			return path, nil
		}
	}

	// tiramos contra el registry y recibimos path al fichero
	filePath, err := s.oci.Pull(ctx, reference)
	if err != nil {
		return "", err
	}

	// cachear si procede
	if !s.cfg.NoCache {
		_ = s.cache.Push(ctx, reference, filePath)
	}
	return filePath, nil
}
