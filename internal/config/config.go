// internal/config/config.go
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Output            string `yaml:"output"`
	DefaultConnection string `yaml:"default_connection"`
	Editor            string `yaml:"editor"`
	Theme             string `yaml:"theme"`
	VersionsToKeep    int    `yaml:"versions_to_keep"`
}

func DefaultConfig() *Config {
	return &Config{
		Output:         "json",
		Theme:          "retro",
		VersionsToKeep: 50,
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ddctl")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ddctl")
}

func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

func DBPath() string {
	return filepath.Join(Dir(), "ddctl.db")
}
