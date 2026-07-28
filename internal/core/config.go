package core

import (
	"errors"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

var ErrConfigExists = errors.New("config already exists")
var ErrConfigNotFound = errors.New("config file not found")

type ProviderConfig struct {
	Name    string `yaml:"name"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base-url"`
	APIKey  string `yaml:"api-key"`
}

// Config is the root user configuration for the state.
type Config struct {
	Provider ProviderConfig `yaml:"provider"`
	Intent   string         `yaml:"intent"`
}

// NewConfig creates a new configuration file. Will fail if the configuration already exists.
func NewConfig(path, profile string) (*Config, error) {
	configFilename := configFilenameForProfile(profile)
	configPath := filepath.Join(path, configFilename)

	// Check if config already exists.
	if _, err := os.Stat(configPath); err == nil {
		return nil, ErrConfigExists
	}

	config := Config{}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig loads the YAML configuration file for the specified profile.
func LoadConfig(path, profile string) (*Config, error) {
	configFilename := configFilenameForProfile(profile)
	configPath := filepath.Join(path, configFilename)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
