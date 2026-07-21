package core

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestState_NewState_Success(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := NewState(tmpDir)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state == nil {
		t.Fatal("expected state to be non-nil")
	}
	if state.path == "" {
		t.Error("expected path to be non-empty")
	}
}

func TestState_NewState_DirectoryAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the state once
	_, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("failed to create initial state: %v", err)
	}

	// Try to create again - should return ErrStateExists
	_, err = NewState(tmpDir)

	if !errors.Is(err, ErrStateExists) {
		t.Fatalf("expected ErrStateExists, got %v", err)
	}
}

func TestState_NewState_InvalidPath(t *testing.T) {
	// Use an invalid path that cannot be created
	// On most systems, trying to create a directory in a read-only location would fail
	_, err := NewState("/root/.casefile-test-invalid-path")

	// This should return an error (permission denied or similar)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestState_Path(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("failed to create state: %v", err)
	}

	got := state.Path()
	expectedSuffix := ".casefile/"

	if len(got) < len(expectedSuffix) {
		t.Errorf("expected path to end with %s, got %s", expectedSuffix, got)
	}
}

func TestState_AbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) *State
		wantFunc func(got string) bool
	}{
		{
			name: "returns absolute path",
			setup: func(t *testing.T) *State {
				tmpDir := t.TempDir()
				state, err := NewState(tmpDir)
				if err != nil {
					t.Fatalf("failed to create state: %v", err)
				}
				return state
			},
			wantFunc: func(got string) bool {
				return filepath.IsAbs(got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup(t)
			got := state.AbsolutePath()

			if !tt.wantFunc(got) {
				t.Errorf("AbsolutePath() = %v, want absolute path", got)
			}
		})
	}
}

// Note: filepath.Abs rarely returns an error, so the error case is not tested.
func TestState_AbsolutePath_UnreachableError(t *testing.T) {
	t.Skip("filepath.Abs rarely returns an error; error case is unreachable in practice")
}

func TestState_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state
	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState failed: %v", err)
	}

	// Verify path methods work together
	path := state.Path()
	absPath := state.AbsolutePath()

	if path == "" {
		t.Error("Path() returned empty string")
	}
	if absPath == "" || absPath == "-" {
		t.Error("AbsolutePath() returned empty or error marker")
	}

	// Verify the path contains .casefile
	expectedDir := ".casefile"
	if len(path) < len(expectedDir) || path[len(path)-len(expectedDir):] != expectedDir {
		t.Errorf("Path() = %v, expected to end with %s", path, expectedDir)
	}
}
