// Package tools provides the agent tool definitions available to Casefile.
package tools

import (
	"casefile/internal/core/tool"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Search returns the casefile_search tool definition: a repo-wide, grep-backed
// search letting an agent find where a pattern or symbol appears across the
// codebase.
func Search() *tool.Tool {
	return &tool.Tool{
		Name: "casefile_search",
		Description: "Repo-wide grep-backed search. Finds where a pattern or symbol appears " +
			"across the codebase, returning matches as file:line. This is the entry point for " +
			"\"where does X appear\" style questions, before narrowing in on a specific file.",
		Parameters: tool.Schema{
			Type: "object",
			Properties: map[string]tool.Property{
				"pattern": {
					Type:        "string",
					Description: "The pattern to search for.",
				},
				"path": {
					Type:        "string",
					Description: "Optional subdirectory to scope the search to, relative to the repo root. Defaults to the repo root.",
				},
			},
			Required: []string{"pattern"},
		},
		Handler: handleSearch,
	}
}

// handleSearch runs the search against the repo, scoped to the executor's root.
func handleSearch(ctx *tool.ExecutorContext, args tool.Arguments) (string, error) {
	pattern := args["pattern"]
	if pattern == "" {
		return "", errors.New("pattern must not be empty")
	}

	searchPath, err := ctx.ResolvePath(args["path"])
	if err != nil {
		return "", err
	}

	output, err := ctx.Run("grep", "-rn", "-I", "--exclude-dir=.git", "-e", pattern, searchPath)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "No matches found.", nil
		}
		return "", fmt.Errorf("search failed: %w", err)
	}

	results := formatMatches(string(output), ctx.Root())
	if results == "" {
		return "No matches found.", nil
	}
	return results, nil
}

// formatMatches rewrites raw "path:line:content" grep output into results
// relative to root, one per line.
func formatMatches(output string, root string) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	formatted := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		relPath, err := filepath.Rel(root, parts[0])
		if err != nil {
			relPath = parts[0]
		}
		formatted = append(formatted, fmt.Sprintf("%s:%s: %s", relPath, parts[1], parts[2]))
	}
	return strings.Join(formatted, "\n")
}
