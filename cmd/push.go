package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/TrianaLab/remake/config"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var pushFile string

// pushCmd publica un Makefile como un artefacto OCI usando la librería ORAS Go
var pushCmd = &cobra.Command{
	Use:   "push <remote_ref>",
	Short: "Push a Makefile as an OCI artifact",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Inicializar configuración
		if err := config.InitConfig(); err != nil {
			return err
		}

		// 2) Normalizar referencia y tag por defecto 'latest'
		rawRef := args[0]
		hasOCI := strings.HasPrefix(rawRef, "oci://")
		raw := rawRef
		if hasOCI {
			raw = strings.TrimPrefix(rawRef, "oci://")
		}
		// Añadir "latest" si no hay tag tras el último '/'
		name := raw[strings.LastIndex(raw, "/")+1:]
		if !strings.Contains(name, ":") {
			raw += ":latest"
		}
		// Preponer registry si faltaba
		if !hasOCI {
			defaultReg := viper.GetString("default_registry")
			raw = defaultReg + "/" + raw
		}
		ref := "oci://" + raw

		// 3) Extraer host, repositorio y tag
		parts := strings.SplitN(raw, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid reference: %s", ref)
		}
		host := parts[0]
		repoAndTag := parts[1]
		rt := strings.SplitN(repoAndTag, ":", 2)
		repoPath := rt[0]
		tag := rt[1]

		// 4) Determinar Makefile a publicar
		filePath := pushFile
		if filePath == "" {
			filePath = config.GetDefaultMakefile()
		}
		if filePath == "" {
			return fmt.Errorf("no Makefile or makefile found; specify with --file")
		}

		// 5) Crear file store y añadir el Makefile
		ctx := context.Background()
		fs, err := file.New(filepath.Dir(filePath))
		if err != nil {
			return err
		}
		defer fs.Close()

		desc, err := fs.Add(ctx, filePath, "application/x-makefile", "")
		if err != nil {
			return err
		}

		// 6) Empaquetar el manifiesto OCI con la capa única
		manifestDesc, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1,
			"application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
		if err != nil {
			return err
		}

		// 7) Etiquetar el manifiesto localmente
		if err := fs.Tag(ctx, manifestDesc, tag); err != nil {
			return err
		}

		// 8) Configurar repositorio remoto
		repoRef := host + "/" + repoPath
		repo, err := remote.NewRepository(repoRef)
		if err != nil {
			return err
		}
		// 9) Autenticación si existe en config
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

		// 10) Push del artefacto
		if _, err := oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions); err != nil {
			return fmt.Errorf("failed to push artifact: %w", err)
		}

		fmt.Printf("✅ Pushed %s to %s\n", filePath, ref)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Makefile to push (default: Makefile or makefile)")
}
