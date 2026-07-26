package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"go.yaml.in/yaml/v3"
)

const configSuffix = "casefile.config.yaml"

var ErrConfigExists = errors.New("config already exists")
var ErrConfigNotFound = errors.New("config file not found")

type ConfigData struct {
	Provider string `yaml:"provider"`
	Intent   string `yaml:"intent"`
}

// Config represents the main user configuration for the program.
type Config struct {
	Profile string
	Data    ConfigData
}

// profileFromFilename extracts the profile from the config filename.
// For example: "casefile.config.yaml" -> "Default", "casefile-local.config.yaml" -> "Local"
func profileFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, configSuffix)
	if name == "casefile" {
		return "default"
	}

	// Extract the profile name.
	profile := strings.TrimPrefix(name, "casefile-")
	if profile == "" {
		return "default"
	}

	runes := []rune(profile)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// configFilenameForProfile returns the config filename for the given profile.
func configFilenameForProfile(profile string) string {
	profileLower := strings.ToLower(profile)
	if profileLower == "default" {
		return "casefile.config.yaml"
	}
	return "casefile-" + profileLower + configSuffix
}

// NewConfig creates a new configuration file. Will fail if the configuration already exists.
func NewConfig(path, profile string) (*Config, error) {
	configFilename := configFilenameForProfile(profile)
	configPath := filepath.Join(path, configFilename)

	// Check if config already exists.
	if _, err := os.Stat(configPath); err == nil {
		return nil, ErrConfigExists
	}

	config := &Config{
		Profile: profile,
		Data: ConfigData{
			Provider: "",
			Intent:   "",
		},
	}

	data, err := yaml.Marshal(config.Data)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, err
	}

	return config, nil
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

	var configData ConfigData
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, err
	}

	return &Config{
		Profile: profile,
		Data:    configData,
	}, nil
}

// ListProfiles returns all available config profiles in the given directory.
func ListProfiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), configSuffix) {
			profiles = append(profiles, profileFromFilename(entry.Name()))
		}
	}

	if len(profiles) == 0 {
		return nil, ErrConfigNotFound
	}

	return profiles, nil
}
