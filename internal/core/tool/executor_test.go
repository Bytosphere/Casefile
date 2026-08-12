package tool

import (
	"context"
	"errors"
	"testing"
)

func TestExecutor_NewExecutor_Success(t *testing.T) {
	root := "/test/root"

	executor := NewExecutor(root)

	if executor == nil {
		t.Fatal("expected executor to be non-nil")
	}
	if executor.root != root {
		t.Errorf("expected root %q, got %q", root, executor.root)
	}
}

func TestExecutor_Run_Success(t *testing.T) {
	registry := NewRegistry()

	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: Schema{
			Required: []string{"path"},
		},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "test result", nil
		},
	}

	registry.Register(tool)

	executor := NewExecutor("/test/root")
	ctx := context.Background()
	args := Arguments{"path": "/some/path"}

	result, err := executor.Run(ctx, tool, args)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "test result" {
		t.Errorf("expected %q, got %q", "test result", result)
	}
}

func TestExecutor_Run_MissingRequiredArgument(t *testing.T) {
	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: Schema{
			Required: []string{"path", "mode"},
		},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "test result", nil
		},
	}

	executor := NewExecutor("/test/root")
	ctx := context.Background()
	args := Arguments{"path": "/some/path"} // missing "mode"

	result, err := executor.Run(ctx, tool, args)

	if err == nil {
		t.Error("expected error for missing required argument")
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestExecutor_Run_ToolHandlerError(t *testing.T) {
	tool := &Tool{
		Name:        "error-tool",
		Description: "A tool that returns error",
		Parameters: Schema{
			Required: []string{"path"},
		},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "", errors.New("handler error")
		},
	}

	executor := NewExecutor("/test/root")
	ctx := context.Background()
	args := Arguments{"path": "/some/path"}

	result, err := executor.Run(ctx, tool, args)

	if err == nil {
		t.Error("expected error from handler")
	}
	if err.Error() != "handler error" {
		t.Errorf("expected error message 'handler error', got %q", err.Error())
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestExecutor_validateRequiredArguments_Success(t *testing.T) {
	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: Schema{
			Required: []string{"path", "mode"},
		},
	}

	executor := NewExecutor("/test/root")
	args := Arguments{
		"path": "/some/path",
		"mode": "read",
	}

	err := executor.validateRequiredArguments(tool, args)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExecutor_validateRequiredArguments_Missing(t *testing.T) {
	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: Schema{
			Required: []string{"path", "mode", "flag"},
		},
	}

	executor := NewExecutor("/test/root")
	args := Arguments{
		"path": "/some/path",
		// missing "mode" and "flag"
	}

	err := executor.validateRequiredArguments(tool, args)

	if err == nil {
		t.Error("expected error for missing required arguments")
	}
}
