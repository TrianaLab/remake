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
	// inicializar configuración centralizada
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	// crear cliente OCI y repositorio de cache con cfg tipado
	regClient := registry.NewDefaultClient(cfg)
	cacheRepo := cache.NewLocalCache(cfg)

	// componer store y runner
	store := artifactstore.NewDefaultArtifactStore(regClient, cacheRepo, cfg)
	runner := process.NewExecRunner()
	a := app.New(store, runner)

	// ejecutar CLI
	if err := cmd.Execute(a); err != nil {
		log.Fatal(err)
	}
}
