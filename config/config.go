package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	BaseDir         string
	ConfigFile      string
	CacheDir        string
	DefaultMakefile string
	DefaultRegistry string
	Version         string
	NoCache         bool
}

func InitConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(home, ".remake")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(baseDir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	viper.SetDefault("baseDir", baseDir)
	viper.SetDefault("configFile", configFile)
	viper.SetDefault("cacheDir", filepath.Join(baseDir, "cache"))
	viper.SetDefault("defaultMakefile", "makefile")
	viper.SetDefault("defaultRegistry", "ghcr.io")
	viper.SetDefault("version", "dev")
	viper.SetDefault("noCache", false)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		viper.Set("registries", map[string]interface{}{})
		if err := viper.WriteConfigAs(configFile); err != nil {
			return nil, err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		BaseDir:         viper.GetString("baseDir"),
		ConfigFile:      viper.GetString("configFile"),
		CacheDir:        viper.GetString("cacheDir"),
		DefaultMakefile: viper.GetString("defaultMakefile"),
		DefaultRegistry: viper.GetString("defaultRegistry"),
		Version:         viper.GetString("version"),
		NoCache:         viper.GetBool("noCache"),
	}
	return cfg, nil
}

func SaveConfig() error {
	return viper.WriteConfig()
}

func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}

func (c *Config) PrintConfig() error {
	configFilePath := viper.GetString("configFile")

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	fmt.Print(string(content))
	return nil
}
