package client

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
func (h *HTTPClient) Pull(ctx context.Context, reference string) (data []byte, err error) {
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

	// Read body using named return variable so defer CloseErr can override
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body for %s: %w", reference, err)
	}

	return
}
