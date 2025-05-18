package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	configDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.WriteFile(configFile, []byte("registries: {}\n"), 0600); err != nil {
			return fmt.Errorf("cannot create default config: %w", err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

func SaveConfig() error {
	return viper.WriteConfig()
}

func GetDefaultMakefile() string {
	entries, err := os.ReadDir(".")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.Name() == "makefile" {
			return "makefile"
		}
	}
	for _, e := range entries {
		if e.Name() == "Makefile" {
			return "Makefile"
		}
	}
	return ""
}

func GetCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".remake", "cache")
	}
	return filepath.Join(home, ".remake", "cache")
}

func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}
