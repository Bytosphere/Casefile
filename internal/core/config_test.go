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
	if config.Provider.Name != "" {
		t.Errorf("expected provider name to be empty, got %q", config.Provider.Name)
	}
	if config.Intent != "" {
		t.Errorf("expected intent to be empty, got %q", config.Intent)
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
	// Verify the config has expected default values
	if config.Provider.Name != "" {
		t.Errorf("expected provider name to be empty, got %q", config.Provider.Name)
	}
	if config.Intent != "" {
		t.Errorf("expected intent to be empty, got %q", config.Intent)
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
	if config.Provider.Name != "" {
		t.Errorf("expected provider name to be empty, got %q", config.Provider.Name)
	}
	if config.Intent != "" {
		t.Errorf("expected intent to be empty, got %q", config.Intent)
	}

	// Load the config
	loadedConfig, err := LoadConfig(tmpDir, "Default")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the loaded config matches
	if loadedConfig.Provider.Name != config.Provider.Name {
		t.Errorf("expected provider name %q, got %q", config.Provider.Name, loadedConfig.Provider.Name)
	}
	if loadedConfig.Intent != config.Intent {
		t.Errorf("expected intent %q, got %q", config.Intent, loadedConfig.Intent)
	}
}
