package cache

import (
	"context"

	"github.com/TrianaLab/remake/config"
)

type CacheRepository interface {
	Push(ctx context.Context, reference string, data []byte) error
	Pull(ctx context.Context, reference string) (string, error)
}

func NewCache(cfg *config.Config, reference string) CacheRepository {
	switch cfg.ParseReference(reference) {
	case config.ReferenceHTTP:
		return NewHTTPCache(cfg)
	case config.ReferenceOCI:
		return NewOCIRepository(cfg)
	}
	return nil
}
