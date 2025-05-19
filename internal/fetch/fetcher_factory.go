package fetch

import (
	"fmt"
	"strings"
)

func GetFetcher(ref string) (Fetcher, error) {
	if strings.HasPrefix(ref, "oci://") {
		return &OCIFetcher{}, nil
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return &HTTPFetcher{}, nil
	}
	return nil, fmt.Errorf("unsupported reference: %s", ref)
}
