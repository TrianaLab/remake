package registry

import (
	"context"
)

type Client interface {
	Login(ctx context.Context, registry, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) ([]byte, error)
}
