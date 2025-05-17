package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/TrianaLab/remake/config"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var (
	pullOutput  string
	pullNoCache bool
)

// pullCmd descarga un Makefile desde un artefacto OCI usando la librería ORAS Go
var pullCmd = &cobra.Command{
	Use:   "pull <remote_ref>",
	Short: "Pull a Makefile into cache (OCI, HTTP or local)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Inicializar config
		if err := config.InitConfig(); err != nil {
			return err
		}

		// 1.1) Limpiar cache si se pidió
		cacheDir := config.GetCacheDir()
		if pullNoCache {
			os.RemoveAll(cacheDir)
		}

		// 2) Normalizar ref
		rawRef := args[0]
		hasOCI := strings.HasPrefix(rawRef, "oci://")
		raw := rawRef
		if hasOCI {
			raw = strings.TrimPrefix(rawRef, "oci://")
		}
		// Añadir latest si falta tag
		name := raw[strings.LastIndex(raw, "/")+1:]
		if !strings.Contains(name, ":") {
			raw += ":latest"
			name += ":latest"
		}
		// Extraer solo nombre de archivo sin tag
		fileName := strings.SplitN(name, ":", 2)[0]
		// Preponer registry
		if !hasOCI {
			raw = viper.GetString("default_registry") + "/" + raw
		}
		// Construir referencia final
		reference := "oci://" + raw

		// 3) Conectar con OCI remoto
		parts := strings.SplitN(raw, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid reference: %s", reference)
		}
		host := parts[0]
		repoAndTag := parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoPath := rt[0]
		tag := rt[1]

		repoRef := host + "/" + repoPath
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}
		// credenciales opcionales
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

		// 4) Crear file store de destino
		dir := filepath.Join(cacheDir, repoPath, tag)
		fs, err := file.New(dir)
		if err != nil {
			return err
		}
		defer fs.Close()

		// 5) Copiar artefacto remoto al file store
		ctx := context.Background()
		_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
		if err != nil {
			return fmt.Errorf("failed to pull artifact: %w", err)
		}

		// 6) Ruta al Makefile descargado
		localPath := filepath.Join(dir, fileName)

		// 7) Mover o imprimir
		if pullOutput != "" {
			return os.Rename(localPath, pullOutput)
		}
		fmt.Println(localPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", "", "Output file path (default prints cache path)")
	pullCmd.Flags().BoolVar(&pullNoCache, "no-cache", false, "Skip local cache and always fetch remote Makefile")
}
