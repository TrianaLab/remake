package artifactstore

import "context"

type OCIRepository interface {
	Login(ctx context.Context, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference, dest string) error
}

type CacheRepository interface {
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) (string, error)
}

type ArtifactStore interface {
	Login(ctx context.Context, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) (string, error)
}

type DefaultArtifactStore struct {
	oci   OCIRepository
	cache CacheRepository
}

func NewDefaultArtifactStore(oci OCIRepository, cache CacheRepository) ArtifactStore {
	return &DefaultArtifactStore{oci: oci, cache: cache}
}

func (s *DefaultArtifactStore) Login(ctx context.Context, user, pass string) error {
	return s.oci.Login(ctx, user, pass)
}

func (s *DefaultArtifactStore) Push(ctx context.Context, reference, path string) error {
	if err := s.oci.Push(ctx, reference, path); err != nil {
		return err
	}
	return s.cache.Push(ctx, reference, path)
}

func (s *DefaultArtifactStore) Pull(ctx context.Context, reference string) (string, error) {
	path, err := s.cache.Pull(ctx, reference)
	if err == nil {
		return path, nil
	}
	dest := reference
	if err := s.oci.Pull(ctx, reference, dest); err != nil {
		return "", err
	}
	_ = s.cache.Push(ctx, reference, dest)
	return dest, nil
}
