package command

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"casefile/internal/core"
)

func TestRunInit_CreatesState(t *testing.T) {
	tmpDir := t.TempDir()

	output := captureStdout(t, func() {
		if err := runInit(nil, []string{tmpDir}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	if _, err := os.Stat(filepath.Join(tmpDir, ".casefile")); err != nil {
		t.Errorf("expected .casefile directory to be created, got %v", err)
	}
	if !strings.Contains(output, "state initialized") {
		t.Errorf("expected output to mention state initialization, got %q", output)
	}
}

func TestRunInit_AlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()

	if err := runInit(nil, []string{tmpDir}); err != nil {
		t.Fatalf("expected no error on first init, got %v", err)
	}

	err := runInit(nil, []string{tmpDir})

	if !errors.Is(err, core.ErrStateExists) {
		t.Fatalf("expected ErrStateExists, got %v", err)
	}
}
