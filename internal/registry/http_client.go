package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	httpClient *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{httpClient: http.DefaultClient}
}

func (h *HTTPClient) Login(ctx context.Context, registry, user, pass string) error {
	return nil
}

func (h *HTTPClient) Push(ctx context.Context, reference, path string) error {
	return nil
}

func (h *HTTPClient) Pull(ctx context.Context, reference string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reference, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", reference, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d when fetching %s", resp.StatusCode, reference)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
