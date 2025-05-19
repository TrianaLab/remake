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

	remakeDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(remakeDir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	configFile := filepath.Join(remakeDir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	viper.SetDefault("cacheDir", filepath.Join(remakeDir, "cache"))
	viper.SetDefault("defaultMakefile", "makefile")
	viper.SetDefault("insecure", false)

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

func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}
