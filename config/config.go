package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var DefaultRegistry string

func BaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".remake")
}

func ConfigFile() string {
	return filepath.Join(BaseDir(), "config.yaml")
}

func InitConfig() error {
	if err := os.MkdirAll(BaseDir(), 0o755); err != nil {
		return err
	}
	viper.SetConfigFile(ConfigFile())
	viper.SetConfigType("yaml")

	viper.SetDefault("cacheDir", filepath.Join(BaseDir(), "cache"))
	viper.SetDefault("defaultMakefile", "makefile")
	viper.SetDefault("insecure", false)
	viper.SetDefault("defaultRegistry", "ghcr.io")

	if _, err := os.Stat(ConfigFile()); os.IsNotExist(err) {
		viper.Set("registries", map[string]interface{}{})
		if err := viper.WriteConfigAs(ConfigFile()); err != nil {
			return err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	DefaultRegistry = viper.GetString("defaultRegistry")
	return nil
}

func SaveConfig() error {
	return viper.WriteConfig()
}

func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}
