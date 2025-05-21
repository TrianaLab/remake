package registry

import "context"

type Client interface {
	Login(ctx context.Context, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference, dest string) error
}
