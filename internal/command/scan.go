package command

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"casefile/internal/agent"
	"casefile/internal/core"
	"casefile/internal/core/tool"
	"casefile/internal/model"
	"casefile/internal/provider/openai"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans the codebase and returns Issues",
	Args:  cobra.NoArgs,
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

// reportedIssue is the shape the model is instructed to report Issues in,
// per agent.BuiltinIntent.
type reportedIssue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

func runScan(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	slog.Info("scan: loading state")
	state, err := core.LoadState()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	slog.Info("scan: walking repo", "root", root)
	files, err := walkRepo(root)
	if err != nil {
		return fmt.Errorf("walk repo: %w", err)
	}
	slog.Info("scan: repo walked", "files", len(files))

	providerConfig := state.Config().Provider
	p := openai.New(providerConfig)
	executor := tool.NewExecutor(root)
	loop := agent.New(p, state.Tools(), executor)

	userMessage := fmt.Sprintf("Repository files (%d total):\n%s\n\nBegin your audit.",
		len(files), strings.Join(files, "\n"))

	slog.Info("scan: starting agent loop", "provider", providerConfig.Name, "model", providerConfig.Model)
	result, err := loop.Run(ctx, agent.BuiltinIntent, userMessage)
	if err != nil {
		return fmt.Errorf("run agent loop: %w", err)
	}
	slog.Info("scan: agent loop finished", "response_length", len(result))

	issues, err := parseReportedIssues(result)
	if err != nil {
		return fmt.Errorf("parse reported issues: %w", err)
	}
	slog.Info("scan: issues reported by model", "count", len(issues))

	slog.Info("scan: beginning transaction")
	db := state.DB()
	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	var newCount, dupCount int
	bySeverity := make(map[string]int)

	for i, issue := range issues {
		bySeverity[issue.Severity]++

		var count int
		err = tx.Get(&count, `
			SELECT COUNT(*)
			FROM T_Issue
			WHERE file = $1 AND line = $2 AND title = $3 AND status = 'Open'
		`, issue.File, issue.Line, issue.Title)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("check duplicate issue: %w", err)
		}

		if count > 0 {
			slog.Debug("scan: skipping duplicate issue", "index", i, "title", issue.Title, "file", issue.File, "line", issue.Line)
			dupCount++
			continue
		}

		fingerprint := issueFingerprint(issue.File, issue.Line, issue.Title)

		_, err = tx.Exec(`
			INSERT INTO T_Issue (title, description, severity, file, line, status, fingerprint)
			VALUES ($1, $2, $3, $4, $5, 'Open', $6)
		`, issue.Title, issue.Description, issue.Severity, issue.File, issue.Line, fingerprint)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert issue: %w", err)
		}

		slog.Debug("scan: inserted new issue", "index", i, "title", issue.Title, "file", issue.File, "line", issue.Line, "severity", issue.Severity)
		newCount++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	slog.Info("scan: transaction committed", "new", newCount, "duplicate", dupCount)

	printScanSummary(len(issues), newCount, dupCount, bySeverity)

	return nil
}

// printScanSummary prints a short human-readable summary of a scan.
func printScanSummary(total, newCount, dupCount int, bySeverity map[string]int) {
	fmt.Printf("Scan complete: %d issue(s) found, %d new, %d already tracked.\n", total, newCount, dupCount)
	for _, severity := range []model.Severity{
		model.SeverityCritical,
		model.SeverityHigh,
		model.SeverityMedium,
		model.SeverityLow,
	} {
		if count := bySeverity[string(severity)]; count > 0 {
			fmt.Printf("  %s: %d\n", severity, count)
		}
	}
}

// issueFingerprint deterministically derives a placeholder fingerprint from
// an issue's file, line, and title. Real fingerprinting is a later
// milestone; this only satisfies T_Issue's NOT NULL constraint.
func issueFingerprint(file string, line int, title string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", file, line, title)))
	return hex.EncodeToString(sum[:])
}

// parseReportedIssues unmarshals the model's final answer into a list of
// reportedIssue, defensively stripping an optional ```json fence and any
// prose the model added around the array despite instructions not to, and
// validates each severity against model.Severity's four values.
func parseReportedIssues(result string) ([]reportedIssue, error) {
	trimmed := strings.TrimSpace(result)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	jsonSlice, err := extractJSONArray(trimmed)
	if err != nil {
		slog.Warn("scan: model response did not contain a JSON array", "error", err, "raw", trimmed)
		return nil, fmt.Errorf("unmarshal issues: %w", err)
	}

	var issues []reportedIssue
	if err := json.Unmarshal([]byte(jsonSlice), &issues); err != nil {
		slog.Warn("scan: failed to unmarshal extracted JSON array", "error", err, "extracted", jsonSlice)
		return nil, fmt.Errorf("unmarshal issues: %w", err)
	}

	validSeverities := map[string]bool{
		string(model.SeverityLow):      true,
		string(model.SeverityMedium):   true,
		string(model.SeverityHigh):     true,
		string(model.SeverityCritical): true,
	}

	for i, issue := range issues {
		if !validSeverities[issue.Severity] {
			return nil, fmt.Errorf("issue %d: invalid severity %q", i, issue.Severity)
		}
	}

	slog.Debug("scan: parsed reported issues", "count", len(issues))

	return issues, nil
}

// extractJSONArray returns the substring of s spanning its first '[' and
// matching last ']', tolerating any prose the model wrapped the JSON array
// in (e.g. "Based on my analysis: [...]"). It returns an error if s
// contains no '[' or ']' at all.
func extractJSONArray(s string) (string, error) {
	start := strings.IndexByte(s, '[')
	end := strings.LastIndexByte(s, ']')
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("no JSON array found in model response")
	}
	return s[start : end+1], nil
}

// walkRepo returns the repo-relative file list for root. It shells out to
// `git ls-files` (respecting .gitignore) when git is available, falling
// back to a plain filepath.WalkDir (skipping only .git/) otherwise.
func walkRepo(root string) ([]string, error) {
	if files, err := walkRepoGit(root); err == nil {
		return files, nil
	}
	return walkRepoFS(root)
}

// walkRepoGit lists files via `git ls-files --cached --others
// --exclude-standard`, NUL-delimited so filenames with spaces or newlines
// are handled correctly.
func walkRepoGit(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	cmd.Dir = root

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	entries := strings.Split(strings.TrimRight(out.String(), "\x00"), "\x00")
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry != "" {
			files = append(files, entry)
		}
	}
	return files, nil
}

// walkRepoFS walks root directly, skipping only .git/, and returns
// repo-relative paths.
func walkRepoFS(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
