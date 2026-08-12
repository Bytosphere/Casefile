package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutorContext_NewExecutorContext_Success(t *testing.T) {
	ctx := context.Background()
	root := "/test/root"

	executorCtx := NewExecutorContext(ctx, root)

	if executorCtx == nil {
		t.Fatal("expected executor context to be non-nil")
	}
	if executorCtx.root != root {
		t.Errorf("expected root %q, got %q", root, executorCtx.root)
	}
}

func TestExecutorContext_ResolvePath_Success(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "executor-context-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	executorCtx := NewExecutorContext(context.Background(), tmpDir)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple subdirectory",
			input:    "subdir",
			expected: filepath.Join(tmpDir, "subdir"),
		},
		{
			name:     "nested subdirectory",
			input:    "a/b/c",
			expected: filepath.Join(tmpDir, "a", "b", "c"),
		},
		{
			name:     "file in directory",
			input:    "dir/file.txt",
			expected: filepath.Join(tmpDir, "dir", "file.txt"),
		},
		{
			name:     "current directory",
			input:    ".",
			expected: tmpDir,
		},
		{
			name:     "empty path",
			input:    "",
			expected: tmpDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executorCtx.ResolvePath(tt.input)

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExecutorContext_ResolvePath_Escaping(t *testing.T) {
	root := "/test/root"

	executorCtx := NewExecutorContext(context.Background(), root)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "parent directory escape",
			input: "../escape",
		},
		{
			name:  "sibling directory escape",
			input: "../sibling",
		},
		{
			name:  "double parent escape",
			input: "../../escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executorCtx.ResolvePath(tt.input)

			if err == nil {
				t.Error("expected error for path escaping root")
			}
			if result != "" {
				t.Errorf("expected empty result, got %q", result)
			}
		})
	}
}

func TestExecutorContext_Run_Success(t *testing.T) {
	executorCtx := NewExecutorContext(context.Background(), "/test/root")

	// Use a command that should succeed - echo with output
	result, err := executorCtx.Run("echo", "hello")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if string(result) == "" {
		t.Error("expected non-empty result")
	}
}

func TestExecutorContext_Run_CommandNotFound(t *testing.T) {
	executorCtx := NewExecutorContext(context.Background(), "/test/root")

	// Use a command that doesn't exist
	_, err := executorCtx.Run("nonexistent-command-12345")

	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestExecutorContext_Run_CommandError(t *testing.T) {
	executorCtx := NewExecutorContext(context.Background(), "/test/root")

	// Use a command that fails - exit 1
	_, err := executorCtx.Run("false")

	if err == nil {
		t.Error("expected error from failing command")
	}
}

// Test context interface methods

func TestExecutorContext_Deadline(t *testing.T) {
	// Test with context that has no deadline
	ctx := context.Background()
	executorCtx := NewExecutorContext(ctx, "/test/root")

	deadline, ok := executorCtx.Deadline()
	if ok {
		t.Error("expected ok to be false for context without deadline")
	}
	if !deadline.IsZero() {
		t.Error("expected zero deadline")
	}

	// Test with context that has a deadline
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	executorCtxWithDeadline := NewExecutorContext(ctxWithDeadline, "/test/root")
	deadline, ok = executorCtxWithDeadline.Deadline()

	if !ok {
		t.Error("expected ok to be true for context with deadline")
	}
	if deadline.IsZero() {
		t.Error("expected non-zero deadline")
	}
}

func TestExecutorContext_Done(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	executorCtx := NewExecutorContext(ctx, "/test/root")

	done := executorCtx.Done()
	if done == nil {
		t.Error("expected non-nil done channel")
	}

	// Cancel the context
	cancel()

	// The done channel should be closed now
	select {
	case <-done:
		// Expected - channel is closed
	default:
		t.Error("expected done channel to be closed after cancel")
	}
}

func TestExecutorContext_Err(t *testing.T) {
	// Test with context that has no error
	ctx := context.Background()
	executorCtx := NewExecutorContext(ctx, "/test/root")

	err := executorCtx.Err()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test with canceled context
	ctxCanceled, cancel := context.WithCancel(ctx)
	cancel()

	executorCtxCanceled := NewExecutorContext(ctxCanceled, "/test/root")
	err = executorCtxCanceled.Err()

	if err == nil {
		t.Error("expected error from canceled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestExecutorContext_Value(t *testing.T) {
	key := "test-key"
	value := "test-value"
	ctx := context.WithValue(context.Background(), key, value)

	executorCtx := NewExecutorContext(ctx, "/test/root")

	result := executorCtx.Value(key)
	if result != value {
		t.Errorf("expected %q, got %q", value, result)
	}

	// Test with non-existent key
	result = executorCtx.Value("non-existent-key")
	if result != nil {
		t.Errorf("expected nil for non-existent key, got %v", result)
	}
}
