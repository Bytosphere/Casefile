package tools

import (
	"casefile/internal/core/tool"
	"context"
	"strings"
	"testing"
)

func TestFile_Definition(t *testing.T) {
	def := File()

	if def.Name != "casefile_file" {
		t.Errorf("expected name %q, got %q", "casefile_file", def.Name)
	}
	if def.Handler == nil {
		t.Fatal("expected handler to be set")
	}
	required := def.Parameters.Required
	if len(required) != 1 || required[0] != "path" {
		t.Errorf("expected required [path], got %v", required)
	}
}

func TestHandleFile_FullContents(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n\nfunc main() {}\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleFile(ctx, tool.Arguments{"path": "main.go"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	expected := "1: package main\n2: \n3: func main() {}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandleFile_WithPattern(t *testing.T) {
	root := t.TempDir()
	lines := make([]string, 0, 20)
	for i := 1; i <= 20; i++ {
		lines = append(lines, "line")
	}
	lines[9] = "target"
	writeFile(t, root, "big.txt", strings.Join(lines, "\n")+"\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleFile(ctx, tool.Arguments{"path": "big.txt", "pattern": "target"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "10: target") {
		t.Errorf("expected result to reference matching line 10, got %q", result)
	}
	resultLines := strings.Split(result, "\n")
	if len(resultLines) != 2*contextLines+1 {
		t.Errorf("expected %d lines of context, got %d: %q", 2*contextLines+1, len(resultLines), result)
	}
}

func TestHandleFile_PatternNoMatches(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	result, err := handleFile(ctx, tool.Arguments{"path": "main.go", "pattern": "doesNotExistAnywhere"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "No matches found." {
		t.Errorf("expected no-matches result, got %q", result)
	}
}

func TestHandleFile_EmptyPath(t *testing.T) {
	root := t.TempDir()
	ctx := tool.NewExecutorContext(context.Background(), root)

	_, err := handleFile(ctx, tool.Arguments{"path": ""})

	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestHandleFile_PathEscape(t *testing.T) {
	root := t.TempDir()
	ctx := tool.NewExecutorContext(context.Background(), root)

	_, err := handleFile(ctx, tool.Arguments{"path": "../escape.go"})

	if err == nil {
		t.Error("expected error for path escaping root")
	}
}

func TestHandleFile_NonExistentFile(t *testing.T) {
	root := t.TempDir()
	ctx := tool.NewExecutorContext(context.Background(), root)

	_, err := handleFile(ctx, tool.Arguments{"path": "missing.go"})

	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestHandleFile_Directory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "sub/file.go", "package sub\n")

	ctx := tool.NewExecutorContext(context.Background(), root)
	_, err := handleFile(ctx, tool.Arguments{"path": "sub"})

	if err == nil {
		t.Error("expected error when path is a directory")
	}
}
