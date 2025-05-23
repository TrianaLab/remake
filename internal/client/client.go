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

package client

import (
	"context"

	"github.com/TrianaLab/remake/config"
)

// Client defines the interface for interacting with remote artifact
// registries over HTTP or OCI protocols. It supports login, push, and pull operations.
type Client interface {
	// Login authenticates against the given registry endpoint using username and password.
	Login(ctx context.Context, registry, user, pass string) error

	// Push uploads the local file at path to the specified reference
	// (e.g., registry/repo:tag) in the remote registry.
	Push(ctx context.Context, reference, path string) error

	// Pull downloads the artifact identified by reference from the registry
	// and returns its raw data bytes.
	Pull(ctx context.Context, reference string) ([]byte, error)
}

// NewClient constructs a Client implementation based on the reference type.
// HTTP(S) references use an HTTP client; OCI references use an OCI client.
// Default fallback is OCI client for any non-HTTP references.
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
