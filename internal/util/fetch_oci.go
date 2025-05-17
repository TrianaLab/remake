package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/viper"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// FetchOCI descarga un artifact OCI usando la librería ORAS Go, extrae el primer archivo y lo cachea
func FetchOCI(ociRef string) (string, error) {
	// Inicializar Viper (configurar default_registry y credenciales)
	// Se asume que config.InitConfig() ya fue llamado por el comando superior

	// Normalizar ref y tag
	raw := strings.TrimPrefix(ociRef, "oci://")
	if !strings.Contains(raw, ":") {
		raw += ":latest"
	}
	// Separar host y repo:tag
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("referencia OCI inválida: %s", ociRef)
	}
	host := parts[0]
	repoAndTag := parts[1]
	rt := strings.SplitN(repoAndTag, ":", 2)
	repoPath := rt[0]
	tag := rt[1]
	repoRef := host + "/" + repoPath

	// Directorio de caché: .remake/cache/oci/<sha256(ref)>
	cacheDir := config.GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}

	// Crear file store en cacheDir
	fs, err := file.New(cacheDir)
	if err != nil {
		return "", err
	}
	defer fs.Close()

	// Repositorio remoto
	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return "", err
	}
	// Autenticación estática si está en config
	username := viper.GetString(fmt.Sprintf("registries.%s.username", host))
	password := viper.GetString(fmt.Sprintf("registries.%s.password", host))
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(host, auth.Credential{
			Username: username,
			Password: password,
		}),
	}

	// Copiar capas del repositorio remoto al file store
	ctx := context.Background()
	_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("error al descargar artifact OCI: %w", err)
	}

	// Buscar el primer archivo .mk en cacheDir

	dirs, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", err
	}
	for _, entry := range dirs {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mk") {
			return filepath.Join(cacheDir, entry.Name()), nil
		}
	}
	// Si no hay .mk, devolver el primer archivo
	for _, entry := range dirs {
		if !entry.IsDir() {
			return filepath.Join(cacheDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no se encontró archivo en OCI artifact: %s", ociRef)
}
