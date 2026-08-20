package command

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"casefile/internal/core"
)

func TestParseReportedIssues_ValidJSON(t *testing.T) {
	input := `[{"title":"t","description":"d","severity":"High","file":"a.go","line":10}]`

	issues, err := parseReportedIssues(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Title != "t" || issues[0].Severity != "High" || issues[0].Line != 10 {
		t.Errorf("unexpected issue: %+v", issues[0])
	}
}

func TestParseReportedIssues_FencedJSON(t *testing.T) {
	input := "```json\n[{\"title\":\"t\",\"description\":\"d\",\"severity\":\"Low\",\"file\":\"a.go\",\"line\":1}]\n```"

	issues, err := parseReportedIssues(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
}

func TestParseReportedIssues_EmptyArray(t *testing.T) {
	issues, err := parseReportedIssues("[]")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestParseReportedIssues_InvalidSeverity(t *testing.T) {
	input := `[{"title":"t","description":"d","severity":"Nope","file":"a.go","line":1}]`

	_, err := parseReportedIssues(input)
	if err == nil {
		t.Fatal("expected an error for an invalid severity")
	}
}

func TestParseReportedIssues_InvalidJSON(t *testing.T) {
	_, err := parseReportedIssues("not json")
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

// TestParseReportedIssues_ProsePreamble covers the "invalid character 'B'"
// failure: some models prefix their final answer with prose ("Based on my
// analysis: [...]") despite BuiltinIntent instructing them not to.
func TestParseReportedIssues_ProsePreamble(t *testing.T) {
	input := `Based on my analysis, here are the issues I found:
[{"title":"t","description":"d","severity":"Medium","file":"a.go","line":5}]`

	issues, err := parseReportedIssues(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Title != "t" {
		t.Errorf("unexpected issue: %+v", issues[0])
	}
}

func TestParseReportedIssues_ProseWithTrailingCommentary(t *testing.T) {
	input := `[{"title":"t","description":"d","severity":"Low","file":"a.go","line":1}]

Let me know if you'd like more detail.`

	issues, err := parseReportedIssues(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
}

func TestIssueFingerprint_Deterministic(t *testing.T) {
	a := issueFingerprint("file.go", 10, "title")
	b := issueFingerprint("file.go", 10, "title")

	if a != b {
		t.Errorf("expected deterministic fingerprint, got %q and %q", a, b)
	}
	if a == "" {
		t.Error("expected a non-empty fingerprint")
	}
}

func TestIssueFingerprint_DiffersOnInput(t *testing.T) {
	a := issueFingerprint("file.go", 10, "title")
	b := issueFingerprint("file.go", 11, "title")

	if a == b {
		t.Error("expected different fingerprints for different inputs")
	}
}

func TestWalkRepo_GitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}

	root := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	writeFile(t, filepath.Join(root, "tracked.go"), "package main\n")
	writeFile(t, filepath.Join(root, "ignored.log"), "noise\n")
	writeFile(t, filepath.Join(root, ".gitignore"), "*.log\n")

	run("add", "tracked.go", ".gitignore")
	run("commit", "-m", "init")

	writeFile(t, filepath.Join(root, "untracked.go"), "package main\n")

	files, err := walkRepo(root)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	sort.Strings(files)

	want := []string{".gitignore", "tracked.go", "untracked.go"}
	if len(files) != len(want) {
		t.Fatalf("expected %v, got %v", want, files)
	}
	for i, f := range want {
		if files[i] != f {
			t.Errorf("expected %v, got %v", want, files)
			break
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// newScanStubServer serves a two-step chat-completions exchange on every
// pair of requests: a tool call to casefile_search, then a final JSON-array
// answer reporting a single Critical issue in vuln.go.
func newScanStubServer(t *testing.T) *httptest.Server {
	t.Helper()

	toolCallResponse := `{"choices":[{"message":{"content":"","tool_calls":[
		{"id":"call_1","type":"function","function":{"name":"casefile_search","arguments":"{\"pattern\":\"SELECT\"}"}}
	]}}]}`
	finalResponse := `{"choices":[{"message":{"content":"[{\"title\":\"SQL Injection\",\"description\":\"string concat into query\",\"severity\":\"Critical\",\"file\":\"vuln.go\",\"line\":4}]"}}]}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if requestCount%2 == 0 {
			fmt.Fprint(w, toolCallResponse)
		} else {
			fmt.Fprint(w, finalResponse)
		}
		requestCount++
	}))
	t.Cleanup(server.Close)

	return server
}

// setUpScanRepo initializes a git repo containing a vulnerable file, a
// Casefile State pointed at baseURL, and chdirs into the repo root.
func setUpScanRepo(t *testing.T, baseURL string) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}

	root := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")

	writeFile(t, filepath.Join(root, "vuln.go"), "package main\n\nfunc run(input string) {\n\tquery := \"SELECT * FROM users WHERE name = '\" + input + \"'\"\n\t_ = query\n}\n")
	run("add", "vuln.go")
	run("commit", "-m", "init")

	if _, err := core.NewState(root); err != nil {
		t.Fatalf("create state: %v", err)
	}

	config := fmt.Sprintf("provider:\n    name: openai\n    model: gpt-test\n    base-url: %s\n    api-key: test-key\nintent: \"\"\n", baseURL)
	writeFile(t, filepath.Join(root, "casefile.config.yaml"), config)

	chdir(t, root)
}

func TestRunScan_EndToEnd_CreatesIssue(t *testing.T) {
	server := newScanStubServer(t)
	setUpScanRepo(t, server.URL)

	output := captureStdout(t, func() {
		if err := runScan(scanCmd, nil); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	if !strings.Contains(output, "1 new") {
		t.Errorf("expected summary to report 1 new issue, got %q", output)
	}

	state, err := core.LoadState()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	tx, err := state.DB().BeginTransaction()
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	var count int
	if err := tx.Get(&count, `SELECT COUNT(*) FROM T_Issue WHERE file = 'vuln.go' AND status = 'Open'`); err != nil {
		t.Fatalf("query issue count: %v", err)
	}
	_ = tx.Commit()

	if count != 1 {
		t.Errorf("expected 1 vuln.go issue in the database, got %d", count)
	}
}

func TestRunScan_DedupOnRerun(t *testing.T) {
	server := newScanStubServer(t)
	setUpScanRepo(t, server.URL)

	if err := runScan(scanCmd, nil); err != nil {
		t.Fatalf("expected no error on first scan, got %v", err)
	}

	output := captureStdout(t, func() {
		if err := runScan(scanCmd, nil); err != nil {
			t.Fatalf("expected no error on second scan, got %v", err)
		}
	})

	if !strings.Contains(output, "0 new") {
		t.Errorf("expected second scan to report 0 new issues, got %q", output)
	}

	state, err := core.LoadState()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	tx, err := state.DB().BeginTransaction()
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	var count int
	if err := tx.Get(&count, `SELECT COUNT(*) FROM T_Issue WHERE file = 'vuln.go' AND status = 'Open'`); err != nil {
		t.Fatalf("query issue count: %v", err)
	}
	_ = tx.Commit()

	if count != 1 {
		t.Errorf("expected still only 1 vuln.go issue after rerun, got %d", count)
	}
}

func TestRunScan_ProviderError(t *testing.T) {
	setUpScanRepo(t, "http://127.0.0.1:1")

	err := runScan(scanCmd, nil)
	if err == nil {
		t.Fatal("expected an error when the provider is unreachable")
	}
}
