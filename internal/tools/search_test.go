package tools

import (
	"casefile/internal/core/tool"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func TestSearch_Definition(t *testing.T) {
	def := Search()

	if def.Name != "casefile_search" {
		t.Errorf("expected name %q, got %q", "casefile_search", def.Name)
	}
	if def.Handler == nil {
		t.Fatal("expected handler to be set")
	}
	required := def.Parameters.Required
	if len(required) != 1 || required[0] != "pattern" {
		t.Errorf("expected required [pattern], got %v", required)
	}
}

func TestHandleSearch_Matches(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n\nfunc helloWorld() {}\n")
	writeFile(t, root, "sub/other.go", "package sub\n\n// helloWorld is used here too\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleSearch(ctx, tool.Arguments{"pattern": "helloWorld"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == "No matches found." {
		t.Fatal("expected matches, got no-matches result")
	}
	if !strings.Contains(result, "main.go:3:") {
		t.Errorf("expected result to reference main.go:3, got %q", result)
	}
	if !strings.Contains(result, "sub/other.go:3:") {
		t.Errorf("expected result to reference sub/other.go:3, got %q", result)
	}
}

func TestHandleSearch_ScopedToPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a/target.go", "package a\n\nfunc scopedMatch() {}\n")
	writeFile(t, root, "b/target.go", "package b\n\nfunc scopedMatch() {}\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleSearch(ctx, tool.Arguments{"pattern": "scopedMatch", "path": "a"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "a/target.go") {
		t.Errorf("expected result to reference a/target.go, got %q", result)
	}
	if strings.Contains(result, "b/target.go") {
		t.Errorf("expected result to not reference b/target.go, got %q", result)
	}
}

func TestHandleSearch_NoMatches(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleSearch(ctx, tool.Arguments{"pattern": "doesNotExistAnywhere"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "No matches found." {
		t.Errorf("expected no-matches result, got %q", result)
	}
}

func TestHandleSearch_EmptyPattern(t *testing.T) {
	root := t.TempDir()
	ctx := tool.NewExecutorContext(context.Background(), root)

	_, err := handleSearch(ctx, tool.Arguments{"pattern": ""})

	if err == nil {
		t.Error("expected error for empty pattern")
	}
}

func TestHandleSearch_PathEscape(t *testing.T) {
	root := t.TempDir()
	ctx := tool.NewExecutorContext(context.Background(), root)

	_, err := handleSearch(ctx, tool.Arguments{"pattern": "foo", "path": "../escape"})

	if err == nil {
		t.Error("expected error for path escaping root")
	}
}
