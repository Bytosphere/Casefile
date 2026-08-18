package tools

import (
	"casefile/internal/core/tool"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// contextLines is the number of lines of surrounding context returned on
// either side of a match when a pattern is given.
const contextLines = 3

// contextLineRegex matches grep's "<line><sep><content>" output, where sep is
// ':' for a matching line or '-' for a context line.
var contextLineRegex = regexp.MustCompile(`^(\d+)[:-](.*)$`)

// File returns the casefile_file tool definition: a targeted file-read
// letting an agent pull the contents of a specific file, optionally focused
// around a pattern.
func File() *tool.Tool {
	return &tool.Tool{
		Name: "casefile_file",
		Description: "Reads the contents of a specific file, with line numbers. If a pattern is " +
			"given, returns only the context surrounding its matches instead of the whole file, " +
			"saving tokens. The natural follow-up to casefile_search once a relevant file has " +
			"been identified.",
		Parameters: tool.Schema{
			Type: "object",
			Properties: map[string]tool.Property{
				"path": {
					Type:        "string",
					Description: "Path to the file to read, relative to the repo root.",
				},
				"pattern": {
					Type:        "string",
					Description: "Optional pattern to focus the output on. When given, only the surrounding context of matches is returned.",
				},
			},
			Required: []string{"path"},
		},
		Handler: handleFile,
	}
}

// handleFile reads the requested file, scoped to the executor's root.
func handleFile(ctx *tool.ExecutorContext, args tool.Arguments) (string, error) {
	path := args["path"]
	if path == "" {
		return "", errors.New("path must not be empty")
	}

	resolvedPath, err := ctx.ResolvePath(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a file", path)
	}

	pattern := args["pattern"]
	if pattern == "" {
		return readFile(resolvedPath)
	}
	return readFileContext(ctx, resolvedPath, pattern)
}

// readFile returns the full contents of the file at path, with line numbers.
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	numbered := make([]string, len(lines))
	for i, line := range lines {
		numbered[i] = fmt.Sprintf("%d: %s", i+1, line)
	}
	return strings.Join(numbered, "\n"), nil
}

// readFileContext returns the lines surrounding matches of pattern in the
// file at path, with line numbers.
func readFileContext(ctx *tool.ExecutorContext, path, pattern string) (string, error) {
	output, err := ctx.Run("grep", "-n", "-C", strconv.Itoa(contextLines), "-e", pattern, path)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "No matches found.", nil
		}
		return "", fmt.Errorf("failed to search file: %w", err)
	}

	formatted := formatContext(string(output))
	if formatted == "" {
		return "No matches found.", nil
	}
	return formatted, nil
}

// formatContext rewrites raw grep "-C" output into a consistent "line: content"
// form, preserving "--" separators between non-adjacent match groups.
func formatContext(output string) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	formatted := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		if line == "--" {
			formatted = append(formatted, line)
			continue
		}
		match := contextLineRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		formatted = append(formatted, fmt.Sprintf("%s: %s", match[1], match[2]))
	}
	return strings.Join(formatted, "\n")
}
