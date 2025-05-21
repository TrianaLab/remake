package cache

import "context"

type CacheRepository interface {
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) (string, error)
}
