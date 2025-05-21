package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config almacena la configuración de Remake
type Config struct {
	CacheDir        string
	DefaultMakefile string
	DefaultRegistry string
	Version         string
	NoCache         bool
}

// InitConfig prepara Viper, establece defaults y carga el archivo de configuración
func InitConfig() (*Config, error) {
	// determinar directorio base ~/.remake
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}

	// configurar Viper para usar ~/.remake/config.yaml
	configFile := filepath.Join(baseDir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// valores por defecto
	viper.SetDefault("cacheDir", filepath.Join(baseDir, "cache"))
	viper.SetDefault("defaultMakefile", "makefile")
	viper.SetDefault("defaultRegistry", "ghcr.io")
	viper.SetDefault("version", "dev")
	viper.SetDefault("noCache", false)

	// inicializar archivo if no existe
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		viper.Set("registries", map[string]interface{}{})
		if err := viper.WriteConfigAs(configFile); err != nil {
			return nil, err
		}
	}

	// leer configuración
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// mapear en struct
	cfg := &Config{
		CacheDir:        viper.GetString("cacheDir"),
		DefaultMakefile: viper.GetString("defaultMakefile"),
		DefaultRegistry: viper.GetString("defaultRegistry"),
		Version:         viper.GetString("version"),
		NoCache:         viper.GetBool("noCache"),
	}
	return cfg, nil
}

// SaveConfig persiste los cambios en config.yaml
func SaveConfig() error {
	return viper.WriteConfig()
}

// NormalizeKey reemplaza puntos por guiones bajos en un endpoint
func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}
