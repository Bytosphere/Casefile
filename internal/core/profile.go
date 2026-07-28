package core

import (
	"errors"
	"os"
	"strings"
	"unicode"
)

const configSuffix = ".config.yaml"

var ErrProfileNotFound = errors.New("profile not found")

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

// Profile represents a named configuration profile.
type Profile struct {
	name   string
	config Config
}

// LoadProfile loads the specified profile of the program. Providing a value of "default" loads the
// default profile.
func LoadProfile(path, name string) (*Profile, error) {
	// Load the configuration file for this profile.
	config, err := LoadConfig(path, name)
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	return &Profile{
		name:   name,
		config: *config,
	}, nil
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) Config() Config {
	return p.config
}
