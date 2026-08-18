package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecutorContext represents a running executor on a Tool. It provides various
// operations to facilitate tool calls.
type ExecutorContext struct {
	root string
	ctx  context.Context
}

// NewExecutorContext creates a new instance from the current given context.
func NewExecutorContext(ctx context.Context, root string) *ExecutorContext {
	return &ExecutorContext{root: root, ctx: ctx}
}

// Root returns the project root directory this context is bound to.
func (ex *ExecutorContext) Root() string {
	return ex.root
}

// ResolvePath enforces the path boundaries to root. It returns the resolved path upon
// success.
func (ex *ExecutorContext) ResolvePath(path string) (string, error) {
	absolutePath := filepath.Join(ex.root, path)
	absolutePath = filepath.Clean(absolutePath)
	relativePath, err := filepath.Rel(ex.root, absolutePath)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return "", fmt.Errorf("path escapes project root: %s", path)
	}
	return absolutePath, nil
}

// Run runs a shell binary with arguments.
func (ex *ExecutorContext) Run(bin string, args ...string) ([]byte, error) {
	cmd := exec.Command(bin, args...)

	// Attach an output for the command.
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (ex *ExecutorContext) Deadline() (time.Time, bool) {
	return ex.ctx.Deadline()
}

func (ex *ExecutorContext) Done() <-chan struct{} {
	return ex.ctx.Done()
}

func (ex *ExecutorContext) Err() error {
	return ex.ctx.Err()
}

func (ex *ExecutorContext) Value(key any) any {
	return ex.ctx.Value(key)
}
