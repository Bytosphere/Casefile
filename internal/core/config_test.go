package core

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_NewConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()

	config, err := NewConfig(tmpDir, "Default")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("expected config to be non-nil")
	}
	if config.Profile != "Default" {
		t.Errorf("expected profile to be 'Default', got %q", config.Profile)
	}
	if config.Data.Provider != "" {
		t.Errorf("expected provider to be empty, got %q", config.Data.Provider)
	}
	if config.Data.Intent != "" {
		t.Errorf("expected intent to be empty, got %q", config.Data.Intent)
	}

	// Verify config file was created with correct filename
	configPath := filepath.Join(tmpDir, "casefile.config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file to exist: %v", err)
	}
}

func TestConfig_NewConfig_ConfigAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the config once
	_, err := NewConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("failed to create initial config: %v", err)
	}

	// Try to create again - should return ErrConfigExists
	_, err = NewConfig(tmpDir, "Default")

	if !errors.Is(err, ErrConfigExists) {
		t.Fatalf("expected ErrConfigExists, got %v", err)
	}
}

func TestConfig_LoadConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// First create a config
	_, err := NewConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Load the config
	config, err := LoadConfig(tmpDir, "Default")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("expected config to be non-nil")
	}
	if config.Profile != "Default" {
		t.Errorf("expected profile to be 'Default', got %q", config.Profile)
	}
}

func TestConfig_LoadConfig_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to load a config that doesn't exist
	_, err := LoadConfig(tmpDir, "Default")

	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestConfig_NewConfig_InvalidPath(t *testing.T) {
	// Use an invalid path that cannot be created
	_, err := NewConfig("/root/.casefile-test-invalid-path", "Default")

	// This should return an error (permission denied or similar)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestConfig_LoadConfig_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config
	config, err := NewConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	// Verify the config was created with default values
	if config.Profile != "Default" {
		t.Errorf("expected profile 'Default', got %q", config.Profile)
	}

	// Load the config
	loadedConfig, err := LoadConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the loaded config matches
	if loadedConfig.Profile != config.Profile {
		t.Errorf("expected profile %q, got %q", config.Profile, loadedConfig.Profile)
	}
	if loadedConfig.Data.Provider != config.Data.Provider {
		t.Errorf("expected provider %q, got %q", config.Data.Provider, loadedConfig.Data.Provider)
	}
	if loadedConfig.Data.Intent != config.Data.Intent {
		t.Errorf("expected intent %q, got %q", config.Data.Intent, loadedConfig.Data.Intent)
	}
}
