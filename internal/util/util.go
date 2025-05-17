package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
)

// NormalizeRef aplica ghcr.io como registry por defecto y añade :latest si falta.
// Prefija con oci:// para artefactos OCI y deja HTTP/HTTPS intactos.
func NormalizeRef(ref string) string {
	// HTTP(S) no se toca
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	// Si ya es OCI
	if strings.HasPrefix(ref, "oci://") {
		name := strings.TrimPrefix(ref, "oci://")
		if !strings.Contains(name, ":") {
			name += ":latest"
		}
		return "oci://" + name
	}
	// Si existe localmente, devolver ruta absoluta
	if _, err := os.Stat(ref); err == nil {
		if abs, err := filepath.Abs(ref); err == nil {
			return abs
		}
		return ref
	}
	// Shorthand: repo[:tag] → oci://ghcr.io/repo[:tag]
	defaultReg := config.GetDefaultRegistry()
	name := ref
	if !strings.Contains(name, ":") {
		name += ":latest"
	}
	return "oci://" + defaultReg + "/" + name
}

// FetchMakefile normaliza y descarga:
// - HTTP/HTTPS → FetchHTTP
// - OCI        → FetchOCI
// - local      → ruta absoluta
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
		return "", fmt.Errorf("invalid reference or file does not exist: %s", ref)
	}
}
