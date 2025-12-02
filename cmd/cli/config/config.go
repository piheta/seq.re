package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/piheta/seq.re/cmd/cli/models"
	"gopkg.in/yaml.v3"
)

// GetPath returns the configuration file path
func GetPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "seqre", "config")
}

// Load reads the configuration from disk
func Load() (models.Config, error) {
	configPath := GetPath()
	if configPath == "" {
		return models.Config{}, nil
	}

	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		if os.IsNotExist(err) {
			return models.Config{}, nil
		}
		return models.Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg models.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return models.Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// Save writes the configuration to disk
func Save(cfg models.Config) error {
	configPath := GetPath()
	if configPath == "" {
		return errors.New("could not determine config path")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// GetServerURL returns the configured server URL or default
func GetServerURL() string {
	if url := os.Getenv("SEQRE_SERVER"); url != "" {
		return url
	}

	cfg, err := Load()
	if err != nil {
		slog.Warn("Failed to load config", slog.String("error", err.Error()))
	}
	if cfg.Server != "" {
		return cfg.Server
	}

	return "http://localhost:8080"
}
