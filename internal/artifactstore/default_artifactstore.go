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
		Login(ctx context.Context, user, pass string) error
		Push(ctx context.Context, reference, path string) error
		Pull(ctx context.Context, reference, dest string) error
	}

	CacheRepository interface {
		Push(ctx context.Context, reference, path string) error
		Pull(ctx context.Context, reference string) (string, error)
	}

	ArtifactStore interface {
		Login(ctx context.Context, user, pass string) error
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
func (s *DefaultArtifactStore) Login(ctx context.Context, user, pass string) error {
	return s.oci.Login(ctx, user, pass)
}

// Push uploads via OCI and then caches unless insecure
func (s *DefaultArtifactStore) Push(ctx context.Context, reference, path string) error {
	// always attempt push to registry
	if err := s.oci.Push(ctx, reference, path); err != nil {
		return err
	}
	// cache only if not insecure
	if !s.cfg.Insecure {
		return s.cache.Push(ctx, reference, path)
	}
	return nil
}

// Pull tries cache first, then registry, and caches on success unless insecure
func (s *DefaultArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	// if not insecure, attempt cache
	if !s.cfg.Insecure {
		if path, err := s.cache.Pull(ctx, reference); err == nil {
			return path, nil
		}
	}
	// fetch from registry
	dest := reference
	if err := s.oci.Pull(ctx, reference, dest); err != nil {
		return "", err
	}
	// cache fetched artifact if not insecure
	if !s.cfg.Insecure {
		_ = s.cache.Push(ctx, reference, dest)
	}
	return dest, nil
}
