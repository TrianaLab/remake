package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

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
	viper.SetDefault("registries", map[string]map[string]string{})
	viper.SetDefault("default_registry", "ghcr.io")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SafeWriteConfig()
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

func SaveConfig() error {
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

func GetDefaultMakefile() string {
	if _, err := os.Stat("Makefile"); err == nil {
		return "Makefile"
	}
	if _, err := os.Stat("makefile"); err == nil {
		return "makefile"
	}
	return ""
}

func GetDefaultRegistry() string {
	return viper.GetString("default_registry")
}

func GetCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".remake", "cache")
	}
	return filepath.Join(home, ".remake", "cache")
}
