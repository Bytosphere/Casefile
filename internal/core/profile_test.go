package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestProfile_Name(t *testing.T) {
	profile := Profile{name: "test-profile", config: Config{}}
	if profile.Name() != "test-profile" {
		t.Errorf("expected name to be 'test-profile', got %q", profile.Name())
	}
}

func TestProfile_Config(t *testing.T) {
	config := Config{Provider: ProviderConfig{Name: "openai"}, Intent: "test"}
	profile := Profile{name: "test", config: config}

	got := profile.Config()
	if got.Provider.Name != "openai" {
		t.Errorf("expected config.Provider.Name to be 'openai', got %q", got.Provider.Name)
	}
	if got.Intent != "test" {
		t.Errorf("expected config.Intent to be 'test', got %q", got.Intent)
	}
}

func TestLoadProfile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// First create a config
	_, err := NewConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Load the profile
	profile, err := LoadProfile(tmpDir, "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile to be non-nil")
	}
	if profile.Name() != "default" {
		t.Errorf("expected profile name to be 'default', got %q", profile.Name())
	}
}

func TestLoadProfile_ProfileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load a profile that doesn't exist
	_, err := LoadProfile(tmpDir, "default")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestProfileFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"casefile.config.yaml", "default"},
		{"casefile-local.config.yaml", "Local"},
		{"casefile-development.config.yaml", "Development"},
		{"casefile-prod.config.yaml", "Prod"},
		{"casefile-test.config.yaml", "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := profileFromFilename(tt.filename)
			if got != tt.expected {
				t.Errorf("profileFromFilename(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

func TestConfigFilenameForProfile(t *testing.T) {
	tests := []struct {
		profile  string
		expected string
	}{
		{"default", "casefile.config.yaml"},
		{"Default", "casefile.config.yaml"},
		{"DEFAULT", "casefile.config.yaml"},
		{"local", "casefile-local.config.yaml"},
		{"Local", "casefile-local.config.yaml"},
		{"development", "casefile-development.config.yaml"},
		{"prod", "casefile-prod.config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			got := configFilenameForProfile(tt.profile)
			if got != tt.expected {
				t.Errorf("configFilenameForProfile(%q) = %q, want %q", tt.profile, got, tt.expected)
			}
		})
	}
}

func TestListProfiles_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple config files
	_, err := NewConfig(tmpDir, "default")
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	_, err = NewConfig(tmpDir, "local")
	if err != nil {
		t.Fatalf("failed to create local config: %v", err)
	}

	// List profiles
	profiles, err := ListProfiles(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}

	// Check that both profiles are present
	foundDefault := false
	foundLocal := false
	for _, p := range profiles {
		if p == "default" {
			foundDefault = true
		}
		if p == "Local" { // profileFromFilename capitalizes the first letter
			foundLocal = true
		}
	}
	if !foundDefault {
		t.Error("expected 'default' profile to be in list")
	}
	if !foundLocal {
		t.Error("expected 'Local' profile to be in list")
	}
}

func TestListProfiles_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to list profiles in an empty directory
	_, err := ListProfiles(tmpDir)
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestListProfiles_IgnoresNonConfigFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config file
	_, err := NewConfig(tmpDir, "default")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Create some non-config files
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create readme: %v", err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "data.json"), []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("failed to create json: %v", err)
	}

	// List profiles - should only return the config file
	profiles, err := ListProfiles(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestLoadProfile_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config file with specific values
	configPath := filepath.Join(tmpDir, "casefile.config.yaml")
	err := os.WriteFile(configPath, []byte("provider:\n  name: openai\nintent: testing\n"), 0644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Load the profile
	profile, err := LoadProfile(tmpDir, "default")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the profile name
	if profile.Name() != "default" {
		t.Errorf("expected name 'default', got %q", profile.Name())
	}

	// Verify the config data
	config := profile.Config()
	if config.Provider.Name != "openai" {
		t.Errorf("expected provider 'openai', got %q", config.Provider.Name)
	}
	if config.Intent != "testing" {
		t.Errorf("expected intent 'testing', got %q", config.Intent)
	}
}
