// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ReferenceType enumerates the types of Makefile references.
// It can be an HTTP URL, a local filesystem path, or an OCI artifact reference.
type ReferenceType int

const (
	// ReferenceHTTP indicates the reference is an HTTP or HTTPS URL.
	ReferenceHTTP ReferenceType = iota

	// ReferenceLocal indicates the reference points to a local file.
	ReferenceLocal

	// ReferenceOCI indicates the reference is an OCI registry artifact.
	ReferenceOCI
)

// Config holds all settings for the Remake CLI, including directories,
// default values, and runtime flags.
type Config struct {
	// BaseDir is the root directory for Remake CLI data (e.g., ~/.remake).
	BaseDir string

	// ConfigFile is the path to the YAML configuration file.
	ConfigFile string

	// CacheDir is the directory where pulled artifacts are stored.
	CacheDir string

	// DefaultMakefile is the filename used when none is specified.
	DefaultMakefile string

	// DefaultRegistry is the OCI registry host used when none is provided.
	DefaultRegistry string

	// Version is the semantic version string of the CLI.
	Version string

	// NoCache disables cache usage when set to true.
	NoCache bool
}

// userHomeDir allows us to override os.UserHomeDir in tests.
var userHomeDir = os.UserHomeDir

// buildVersion is populated via -ldflags at build time.
var buildVersion = "dev"

// InitConfig initializes directory structure and loads configuration from
// ~/.remake/config.yaml, applying defaults for all settings.
func InitConfig() (*Config, error) {
	home, err := userHomeDir()
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

	// Set default values
	viper.SetDefault("baseDir", baseDir)
	viper.SetDefault("configFile", configFile)
	viper.SetDefault("cacheDir", filepath.Join(baseDir, "cache"))
	viper.SetDefault("defaultMakefile", "makefile")
	viper.SetDefault("defaultRegistry", "ghcr.io")
	viper.SetDefault("noCache", false)

	// Create default config file if it does not exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		viper.Set("registries", map[string]interface{}{})
		if err := viper.WriteConfigAs(configFile); err != nil {
			return nil, err
		}
	}

	// Read configuration into viper
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Populate Config struct from viper
	cfg := &Config{
		BaseDir:         viper.GetString("baseDir"),
		ConfigFile:      viper.GetString("configFile"),
		CacheDir:        viper.GetString("cacheDir"),
		DefaultMakefile: viper.GetString("defaultMakefile"),
		DefaultRegistry: viper.GetString("defaultRegistry"),
		Version:         buildVersion,
		NoCache:         viper.GetBool("noCache"),
	}
	return cfg, nil
}

// ParseReference determines the ReferenceType for a given string.
// It returns ReferenceHTTP for URLs, ReferenceLocal for existing files,
// and ReferenceOCI otherwise.
func (c *Config) ParseReference(ref string) ReferenceType {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ReferenceHTTP
	}
	if _, err := os.Stat(ref); err == nil {
		return ReferenceLocal
	}
	return ReferenceOCI
}

// SaveConfig writes any in-memory changes back to the config file.
func SaveConfig() error {
	return viper.WriteConfig()
}

// NormalizeKey transforms an endpoint string into a valid key
// used for storing credentials in viper (dots replaced by underscores).
func NormalizeKey(endpoint string) string {
	return strings.ReplaceAll(endpoint, ".", "_")
}

// PrintConfig outputs the raw contents of the configuration file to stdout.
func (c *Config) PrintConfig() error {
	configFilePath := viper.GetString("configFile")
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	fmt.Print(string(content))
	return nil
}
