package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

func NormalizeRef(ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}

	if strings.HasPrefix(ref, "oci://") {
		name := strings.TrimPrefix(ref, "oci://")
		defaultReg := config.GetDefaultRegistry()
		if defaultReg == "" {
			defaultReg = "ghcr.io"
		}
		if !strings.Contains(name, "/") {
			name = defaultReg + "/" + name
		}
		if !strings.Contains(name, ":") {
			name += ":latest"
		}
		return "oci://" + name
	}

	if _, err := os.Stat(ref); err == nil {
		if abs, err := filepath.Abs(ref); err == nil {
			return abs
		}
	}

	defaultReg := config.GetDefaultRegistry()
	if defaultReg == "" {
		defaultReg = "ghcr.io"
	}
	name := ref
	if !strings.Contains(name, ":") {
		name += ":latest"
	}
	return "oci://" + defaultReg + "/" + name
}

func FetchMakefile(ref string) (string, error) {
	nref := NormalizeRef(ref)
	switch {
	case strings.HasPrefix(nref, "http://"), strings.HasPrefix(nref, "https://"):
		return FetchHTTP(nref)
	case strings.HasPrefix(nref, "oci://"):
		return FetchOCI(nref)
	default:
		if _, err := os.Stat(nref); err == nil {
			return nref, nil
		}
		return "", fmt.Errorf("invalid reference or file not found: %s", ref)
	}
}
