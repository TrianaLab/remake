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

package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient provides basic HTTP(S) access for fetching remote Makefile artifacts.
// It implements the Client interface with no-op Login and Push methods.
type HTTPClient struct {
	httpClient *http.Client
}

// NewHTTPClient returns a new HTTPClient using the default HTTP client.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{httpClient: http.DefaultClient}
}

// Login is a no-op for HTTPClient since HTTP references do not require authentication.
func (h *HTTPClient) Login(ctx context.Context, registry, user, pass string) error {
	return nil
}

// Push is a no-op for HTTPClient as pushing over HTTP is not supported.
func (h *HTTPClient) Push(ctx context.Context, reference, path string) error {
	return nil
}

// Pull performs an HTTP GET request to fetch the artifact data from the given URL.
// It returns the raw response body bytes or an error on non-200 status codes or failures.
func (h *HTTPClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reference, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for %s: %w", reference, err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", reference, err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching %s", resp.StatusCode, reference)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body for %s: %w", reference, err)
	}

	return data, nil
}
