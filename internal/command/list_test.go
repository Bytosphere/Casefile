package command

import (
	"strings"
	"testing"

	"casefile/internal/core"
)

// setUpListState initializes a fresh State (which seeds T_Issue via the
// V1_Init migration) rooted at a temp dir and chdirs into it, matching what
// core.LoadState() expects to find.
func setUpListState(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	if _, err := core.NewState(tmpDir); err != nil {
		t.Fatalf("create state: %v", err)
	}

	chdir(t, tmpDir)
}

// runListWithFlags sets listCmd's severity/status/file flags (defaulting to
// empty when unset) and runs runList, returning captured stdout.
func runListWithFlags(t *testing.T, severity, status, file string) string {
	t.Helper()

	for name, value := range map[string]string{
		"severity": severity,
		"status":   status,
		"file":     file,
	} {
		if err := listCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}

	return captureStdout(t, func() {
		if err := runList(listCmd, nil); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

// countIssueLines counts the "#NNNN" header lines in list output.
func countIssueLines(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "#") {
			count++
		}
	}
	return count
}

func TestRunList_NoFilters_ReturnsAllSeededIssues(t *testing.T) {
	setUpListState(t)

	output := runListWithFlags(t, "", "", "")

	if got := countIssueLines(output); got != 8 {
		t.Errorf("expected 8 seeded issues, got %d\noutput:\n%s", got, output)
	}
}

func TestRunList_FilterBySeverity(t *testing.T) {
	setUpListState(t)

	output := runListWithFlags(t, "Critical", "", "")

	if got := countIssueLines(output); got != 2 {
		t.Errorf("expected 2 Critical issues, got %d\noutput:\n%s", got, output)
	}
	if !strings.Contains(output, "SQL Injection Risk") || !strings.Contains(output, "Hardcoded Credential") {
		t.Errorf("expected the two Critical seeded issues, got:\n%s", output)
	}
}

func TestRunList_FilterByStatus(t *testing.T) {
	setUpListState(t)

	output := runListWithFlags(t, "", "Closed", "")

	if got := countIssueLines(output); got != 2 {
		t.Errorf("expected 2 Closed issues, got %d\noutput:\n%s", got, output)
	}
}

func TestRunList_FilterByFile(t *testing.T) {
	setUpListState(t)

	output := runListWithFlags(t, "", "", "handler")

	if got := countIssueLines(output); got != 2 {
		t.Errorf("expected 2 issues under internal/handler, got %d\noutput:\n%s", got, output)
	}
}

func TestRunList_NoMatches(t *testing.T) {
	setUpListState(t)

	output := runListWithFlags(t, "Low", "Closed", "")

	if got := countIssueLines(output); got != 0 {
		t.Errorf("expected 0 issues, got %d\noutput:\n%s", got, output)
	}
}
