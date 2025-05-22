package registry

import (
	"context"

	"github.com/TrianaLab/remake/config"
)

type Client interface {
	Login(ctx context.Context, registry, user, pass string) error
	Push(ctx context.Context, reference, path string) error
	Pull(ctx context.Context, reference string) ([]byte, error)
}

func NewClient(cfg *config.Config, reference string) Client {
	switch cfg.ParseReference(reference) {
	case config.ReferenceHTTP:
		return NewHTTPClient()
	case config.ReferenceOCI:
		return NewOCIClient(cfg)
	default:
		return NewOCIClient(cfg)
	}
}
