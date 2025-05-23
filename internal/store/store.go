package store

import (
	"context"
	"fmt"
	"os"

	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/cache"
	"github.com/TrianaLab/remake/internal/client"
)

// overrideable constructors for testing
var (
	newClient      = client.NewClient
	newCache       = cache.NewCache
	parseReference = func(cfg *config.Config, ref string) config.ReferenceType {
		return cfg.ParseReference(ref)
	}
)

// Store defines the interface for login, push, and pull operations
// on Makefile artifacts across local, HTTP, or OCI backends.
type Store interface {
	// Login authenticates against the given registry endpoint.
	Login(ctx context.Context, registry, user, pass string) error

	// Push uploads the local Makefile at path to the specified reference.
	Push(ctx context.Context, reference, path string) error

	// Pull retrieves a Makefile artifact by reference and returns
	// the local filesystem path where it is stored.
	Pull(ctx context.Context, reference string) (string, error)
}

// ArtifactStore implements the Store interface by delegating to
// registry clients and cache repositories based on reference type.
type ArtifactStore struct {
	cfg *config.Config
}

// New returns a new Store implementation using the provided configuration.
func New(cfg *config.Config) Store {
	return &ArtifactStore{cfg: cfg}
}

// Login authenticates to the remote registry using the configured registry client.
func (s *ArtifactStore) Login(ctx context.Context, reg, user, pass string) error {
	c := newClient(s.cfg, reg)
	return c.Login(ctx, reg, user, pass)
}

// Push uploads and caches a Makefile artifact based on its reference type.
// For OCI references, it pushes to the registry and then caches the data locally.
// HTTP and local references are not supported for push operations.
func (s *ArtifactStore) Push(ctx context.Context, reference, path string) error {
	switch parseReference(s.cfg, reference) {
	case config.ReferenceHTTP:
		return fmt.Errorf("pushing to HTTP(s) references is not supported")
	case config.ReferenceLocal:
		return fmt.Errorf("pushing local references is not supported")
	case config.ReferenceOCI:
		c := newClient(s.cfg, reference)
		if err := c.Push(ctx, reference, path); err != nil {
			return err
		}
		// Read file data for caching
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		cacheRepo := newCache(s.cfg, reference)
		return cacheRepo.Push(ctx, reference, data)
	default:
		return fmt.Errorf("unknown reference type for %s", reference)
	}
}

// Pull retrieves a Makefile artifact, using cache when enabled.
// For local references, it returns the path directly. For other types,
// it attempts to read from cache (unless NoCache is set), otherwise fetches
// from the registry and then caches the result.
func (s *ArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	switch parseReference(s.cfg, reference) {
	case config.ReferenceLocal:
		return reference, nil
	default:
		// Attempt cache lookup
		cacheRepo := newCache(s.cfg, reference)
		if !s.cfg.NoCache {
			if path, err := cacheRepo.Pull(ctx, reference); err == nil {
				return path, nil
			}
		}
		c := newClient(s.cfg, reference)
		data, err := c.Pull(ctx, reference)
		if err != nil {
			return "", err
		}
		// Cache and return
		if err := cacheRepo.Push(ctx, reference, data); err != nil {
			return "", err
		}
		return cacheRepo.Pull(ctx, reference)
	}
}
