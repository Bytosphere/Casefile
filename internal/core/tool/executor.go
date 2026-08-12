package tool

import (
	"context"
	"errors"
)

// Executor is the execution layer that executes a given Tool.
type Executor struct {
	// root is the project directory of the codebase.
	root string
}

func NewExecutor(root string) *Executor {
	return &Executor{
		root: root,
	}
}

// Run executes the given tool and returns the result.
func (e *Executor) Run(ctx context.Context, tool *Tool, args Arguments) (string, error) {
	if err := e.validateRequiredArguments(tool, args); err != nil {
		return "", errors.New("no path property provided")
	}

	// Start a new context.
	executorCtx := NewExecutorContext(ctx, e.root)

	return tool.Handler(executorCtx, args)
}

// validateRequiredArguments ensures all required arguments are available to be passed
// to the tool.
func (e *Executor) validateRequiredArguments(tool *Tool, args Arguments) error {
	requiredArgs := tool.Parameters.Required
	for _, name := range requiredArgs {
		if _, ok := args[name]; !ok {
			return errors.New("required arguments not satisfied")
		}
	}
	return nil
}
