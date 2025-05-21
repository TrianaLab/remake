package main

import (
	"log"

	"github.com/TrianaLab/remake/app"
	"github.com/TrianaLab/remake/cmd"
	"github.com/TrianaLab/remake/config"
	"github.com/TrianaLab/remake/internal/artifactstore"
	"github.com/TrianaLab/remake/internal/cache"
	"github.com/TrianaLab/remake/internal/process"
	"github.com/TrianaLab/remake/internal/registry"
)

func main() {
	// Inicializar configuración global
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}
	// Crear cliente de registry y repositorio de cache
	regClient := registry.NewDefaultClient()
	cacheRepo := cache.NewLocalCache(config.BaseDir())

	// Crear ArtifactStore y ProcessRunner
	store := artifactstore.NewDefaultArtifactStore(regClient, cacheRepo)
	runner := process.NewExecRunner()
	a := app.New(store, runner)

	// Ejecutar comando CLI
	if err := cmd.Execute(a); err != nil {
		log.Fatal(err)
	}
}
