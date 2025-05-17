package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

// NormalizeRef aplica default registry y tag `latest` si falta
func NormalizeRef(ref string) string {
	// 1) HTTP URL
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	// 2) OCI con esquema
	if strings.HasPrefix(ref, "oci://") {
		name := strings.TrimPrefix(ref, "oci://")
		if !strings.Contains(name, ":") {
			name = name + ":latest"
		}
		return "oci://" + name
	}
	// 3) Ruta local existente
	if _, err := os.Stat(ref); err == nil {
		abs, err := filepath.Abs(ref)
		if err == nil {
			return abs
		}
		return ref
	}
	// 4) Destinar a default registry (shorthand)
	defaultReg := config.GetDefaultRegistry() // ghcr.io por defecto
	name := defaultReg + "/" + ref
	if !strings.Contains(ref, ":") {
		name = name + ":latest"
	}
	return "oci://" + name
}

// FetchMakefile resuelve un ref local, HTTP(S) u OCI y devuelve la ruta cacheada o local
func FetchMakefile(ref string) (string, error) {
	nref := NormalizeRef(ref)
	switch {
	case strings.HasPrefix(nref, "http://"), strings.HasPrefix(nref, "https://"):
		return FetchHTTP(nref)
	case strings.HasPrefix(nref, "oci://"):
		return FetchOCI(nref)
	default:
		// local absolute
		if _, err := os.Stat(nref); err == nil {
			return nref, nil
		}
		return "", fmt.Errorf("invalid reference or file does not exist: %s", ref)
	}
}
