package tool

import (
	"testing"
)

func TestRegistry_NewRegistry_Success(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("expected registry to be non-nil")
	}
	if registry.tools == nil {
		t.Error("expected tools map to be non-nil")
	}
}

func TestRegistry_Register_Success(t *testing.T) {
	registry := NewRegistry()

	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "test result", nil
		},
	}

	registry.Register(tool)

	if len(registry.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(registry.tools))
	}
}

func TestRegistry_Register_MultipleTools(t *testing.T) {
	registry := NewRegistry()

	tool1 := &Tool{
		Name:        "tool-one",
		Description: "First tool",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "result 1", nil
		},
	}

	tool2 := &Tool{
		Name:        "tool-two",
		Description: "Second tool",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "result 2", nil
		},
	}

	registry.Register(tool1)
	registry.Register(tool2)

	if len(registry.tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(registry.tools))
	}

	// Verify both tools are in the registry
	if _, ok := registry.tools["tool-one"]; !ok {
		t.Error("expected tool-one to be registered")
	}
	if _, ok := registry.tools["tool-two"]; !ok {
		t.Error("expected tool-two to be registered")
	}
}

func TestRegistry_Get_ExistingTool(t *testing.T) {
	registry := NewRegistry()

	tool := &Tool{
		Name:        "existing-tool",
		Description: "An existing tool",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "result", nil
		},
	}

	registry.Register(tool)

	got, ok := registry.Get("existing-tool")

	if !ok {
		t.Fatal("expected to find existing-tool")
	}
	if got == nil {
		t.Fatal("expected tool to be non-nil")
	}
	if got.Name != "existing-tool" {
		t.Errorf("expected name 'existing-tool', got %q", got.Name)
	}
}

func TestRegistry_Get_NonExistingTool(t *testing.T) {
	registry := NewRegistry()

	got, ok := registry.Get("non-existing-tool")

	if ok {
		t.Error("expected not to find non-existing-tool")
	}
	if got != nil {
		t.Error("expected tool to be nil for non-existing tool")
	}
}

func TestRegistry_Get_AfterOverwrite(t *testing.T) {
	registry := NewRegistry()

	tool1 := &Tool{
		Name:        "overwrite-tool",
		Description: "First version",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "result 1", nil
		},
	}

	tool2 := &Tool{
		Name:        "overwrite-tool",
		Description: "Second version",
		Parameters:  Schema{},
		Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
			return "result 2", nil
		},
	}

	registry.Register(tool1)
	registry.Register(tool2)

	// Should only have one tool (the second one overwrote the first)
	if len(registry.tools) != 1 {
		t.Errorf("expected 1 tool after overwrite, got %d", len(registry.tools))
	}

	got, ok := registry.Get("overwrite-tool")
	if !ok {
		t.Fatal("expected to find overwrite-tool")
	}
	if got.Description != "Second version" {
		t.Errorf("expected 'Second version', got %q", got.Description)
	}
}

func TestRegistry_All_Empty(t *testing.T) {
	registry := NewRegistry()

	all := registry.All()

	if len(all) != 0 {
		t.Errorf("expected 0 tools, got %d", len(all))
	}
}

func TestRegistry_All_ReturnsAllRegistered(t *testing.T) {
	registry := NewRegistry()

	tool1 := &Tool{Name: "tool-one", Description: "First tool"}
	tool2 := &Tool{Name: "tool-two", Description: "Second tool"}

	registry.Register(tool1)
	registry.Register(tool2)

	all := registry.All()

	if len(all) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(all))
	}

	names := map[string]bool{}
	for _, t := range all {
		names[t.Name] = true
	}
	if !names["tool-one"] || !names["tool-two"] {
		t.Errorf("expected both tool-one and tool-two in All(), got %v", names)
	}
}

func TestRegistry_Integration(t *testing.T) {
	registry := NewRegistry()

	// Verify initial state
	if len(registry.tools) != 0 {
		t.Errorf("expected empty registry, got %d tools", len(registry.tools))
	}

	// Register multiple tools
	tools := []*Tool{
		{
			Name:        "integration-tool-1",
			Description: "First integration tool",
			Parameters:  Schema{},
			Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
				return "result 1", nil
			},
		},
		{
			Name:        "integration-tool-2",
			Description: "Second integration tool",
			Parameters:  Schema{},
			Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
				return "result 2", nil
			},
		},
		{
			Name:        "integration-tool-3",
			Description: "Third integration tool",
			Parameters:  Schema{},
			Handler: func(ctx *ExecutorContext, args Arguments) (string, error) {
				return "result 3", nil
			},
		},
	}

	for _, tool := range tools {
		registry.Register(tool)
	}

	// Verify all tools are registered
	if len(registry.tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(registry.tools))
	}

	// Retrieve and verify each tool
	for _, expected := range tools {
		got, ok := registry.Get(expected.Name)
		if !ok {
			t.Errorf("expected to find %s", expected.Name)
		}
		if got.Name != expected.Name {
			t.Errorf("expected name %q, got %q", expected.Name, got.Name)
		}
		if got.Description != expected.Description {
			t.Errorf("expected description %q, got %q", expected.Description, got.Description)
		}
	}

	// Verify non-existing tool returns false
	_, ok := registry.Get("does-not-exist")
	if ok {
		t.Error("expected not to find non-existing tool")
	}
}
