package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FetchMakefile resolves a reference that can be local path, HTTP(S) URL, or OCI reference.
// Returns the local file path to the makefile (cached for remote)
func FetchMakefile(ref string) (string, error) {
	// HTTP(S)
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return FetchHTTP(ref)
	}
	// OCI
	if strings.HasPrefix(ref, "oci://") {
		return FetchOCI(ref)
	}
	// Otherwise assume local path
	if _, err := os.Stat(ref); err == nil {
		abs, err := filepath.Abs(ref)
		if err != nil {
			return ref, nil
		}
		return abs, nil
	}
	return "", errors.New("invalid reference or file does not exist: " + ref)
}
