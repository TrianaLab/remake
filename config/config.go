package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitConfig initializes Viper and loads (or creates) the config file at ~/.remake/config.yaml
func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}
	configDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Default structure
	viper.SetDefault("registries", map[string]map[string]string{})
	viper.SetDefault("default_registry", "ghcr.io")

	// Read or create config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create empty one
			if err := viper.SafeWriteConfig(); err != nil {
				// SafeWriteConfig fails if file exists; ignore
			}
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

// SaveConfig writes current Viper settings to the config file
func SaveConfig() error {
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

// GetDefaultRegistry returns the configured default OCI registry (e.g., ghcr.io)
func GetDefaultRegistry() string {
	return viper.GetString("default_registry")
}

// GetDefaultMakefile returns "Makefile" or "makefile" if it exists in current directory, else empty string
func GetDefaultMakefile() string {
	if _, err := os.Stat("Makefile"); err == nil {
		return "Makefile"
	}
	if _, err := os.Stat("makefile"); err == nil {
		return "makefile"
	}
	return ""
}

// GetCacheDir devuelve la ruta al directorio de cache (~/.remake/cache)
func GetCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// En caso de fallo, por seguridad usar cwd
		return filepath.Join(".remake", "cache")
	}
	return filepath.Join(home, ".remake", "cache")
}
