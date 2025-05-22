// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// THE SOFTWARE.

package cache

import (
	"context"

	"github.com/TrianaLab/remake/config"
)

// CacheRepository defines the interface for caching Makefile artifacts.
// Implementations may store artifacts locally, over HTTP, or in OCI registries.
type CacheRepository interface {
	// Push stores the given data bytes under the specified reference key.
	// The reference typically matches an OCI artifact reference or URL.
	Push(ctx context.Context, reference string, data []byte) error

	// Pull retrieves a cached artifact by reference and returns the
	// local filesystem path where the data is stored.
	Pull(ctx context.Context, reference string) (string, error)
}

// NewCache constructs a CacheRepository based on the reference type.
// It inspects the reference string and returns an HTTP-based cache
// or an OCI repository-based cache. Returns nil for unsupported types.
func NewCache(cfg *config.Config, reference string) CacheRepository {
	switch cfg.ParseReference(reference) {
	case config.ReferenceHTTP:
		return NewHTTPCache(cfg)
	case config.ReferenceOCI:
		return NewOCIRepository(cfg)
	}
	return nil
}
