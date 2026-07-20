# Roadmap

## Version 0.1.0 — Project Initialization & Storage

### Command: `casefile init`

Sets up everything a project needs before it can be scanned.

- Creates the `.casefile/` state directory
- Bootstraps the database and default config in one pass

### Storage: SQLite Schema & Migrations

The persistence layer everything else in Casefile is built on.

- Initial schema covering Issues and their metadata.
- A migration mechanism so the schema can evolve across later versions
  without breaking existing state.

### Configuration

The baseline config file `init` writes out.

- Defines where state lives, default provider, default intent.
- Structured so later versions (providers, intents) extend it rather than
  redefine it.

---

## Version 0.2.0 — Agent Provider

Establishes how Casefile talks to a model, even though only one built-in
provider exists at this stage. Getting the abstraction right now means later
provider work is additive rather than a rewrite.

### Provider: Foundation

A defined interface between Casefile's orchestration logic and whatever
backend actually generates findings.

- Request/response contract agents are built against.
- Kept intentionally provider-agnostic, even with a single implementation
  behind it.

### Provider: Built-in Default

The one concrete implementation of that interface for now.

- Hardcoded to a single provider/model.
- Wired in as the only option — multi-provider selection comes later, once
  real usage has exercised this interface.

---

## Version 0.3.0 — First Scan

### Command: `casefile scan`

Runs the provider from 0.2.0 against a single built-in intent to produce
Issues, using the storage layer from 0.1.0.

- Walks the repo.
- Writes Issues to the database.

### Schema: Issue Model

The fields that make up an Issue, since every later command builds on it.

- Sequential ID, title, description, severity, file, line, status, created date.
- Placeholder field for a code-context fingerprint (built out in 0.6.0).

---

## Version 0.4.0 — Issue Listing & Filtering

### Command: `casefile list`

The primary way to see what a scan found.

- Sorted by severity by default.

### Flags: List Filtering

Filtering lives in the same version as `list` itself rather than being
bolted on later.

- Combinable in a single invocation.

#### Filters

- `--severity`
- `--status`
- `--file`.

### Output: Terminal Formatting

A readability pass on `list` output.

- Aligned columns.
- Color by severity.

---

## Version 0.5.0 — Export

### Command: `casefile export`

Produces a clean Markdown report from stored Issues.

- Grouped by severity.
- Written to a single output file.

### Output: Markdown Report Formatting

The actual shape of the exported document.

- Consistent heading structure per severity group.
- Designed to be readable as-is, not just a dump of the `list` output.

---

## Version 0.6.0 — Issue Lifecycle

### Command: `casefile close`

Marks Issues resolved.

- Accepts a sequence and/or range of numbers (`close 1 3 5-7`).
- Optional note/reason stored on the Issue.

### Behavior: Lifecycle Visibility

Keeps closed Issues out of the way without hiding them entirely.

- Excluded from default `list` output.
- Visible via `list --status closed`.

---

## Version 0.7.0 — Fingerprints

A dedicated milestone to design and build the mechanism that lets Casefile
tell whether an Issue still applies after code changes. This is the
foundation `update` (next milestone) depends on, so it's scoped on its own
rather than bundled in.

### Design: Fingerprint Strategy

What identifies an Issue beyond file/line, since line numbers drift.

- Candidates: hash of surrounding code context, AST-node-level fingerprint,
  or a combination.
- Needs a tolerance model — how much surrounding code can change before a
  fingerprint is considered "broken" vs. "still matches".

### Storage: Fingerprint Persistence & Comparison

Persists fingerprints and builds the diff logic later commands will call.

- Fingerprint stored per Issue at scan time.
- Comparison routine: same / drifted-but-matches / no-longer-found.

---

## Version 0.8.0 — Verified Re-evaluation

The part that makes Casefile more than a one-shot linter, built on the
fingerprinting work above.

### Command: `casefile update`

Re-scans and checks each open Issue against its stored fingerprint.

- Auto-closes Issues that are verifiably fixed.
- Flags Issues where context changed too much to verify automatically.

### Flag: `--dry-run`

Preview mode for `update`.

- Shows what would change without committing to it.

---

## Version 0.9.0 — Intents

Intents get their own version rather than riding along with scan or provider
work, since they need a real configuration mechanism of their own.

### Concept: Intents

Named strategies describing what a scan should look for and how it weighs
severity.

- Ship built-in intents: security, performance, style/smells.

### Config: Intent Configuration

The mechanism that lets users define intents beyond the built-ins.

- New section of `.casefile/config` dedicated to intent definitions.
- Selected via `scan --intent <name>` (repeatable, for multiple intents per
  scan).

---

## Version 0.10.0 — Multi-Provider Support

Builds directly on the provider abstraction from 0.2.0 to support more than
the single built-in provider.

### Architecture: Additional Providers

Implements the 0.2.0 interface for more backends.

- At least one additional provider alongside the built-in default.

### Config: Provider Selection

Chooses provider and model outside of code.

- Set via config, overridable per-scan.

---

## Version 0.11.0 — QA Test-Case Ingestion

The second half of Casefile's stated scope: using existing test cases to
focus a scan.

### Feature: Test-Case Ingestion

Accepts a Markdown file describing test cases for a feature.

- Exact entry point (a `scan` flag vs. a dedicated subcommand) decided during
  implementation.

### Behavior: Focused Scanning

Uses ingested test cases to steer the agent's attention.

- Findings tagged so exports can distinguish QA-driven results from general
  scans.

---

## Version 0.12.0 — Polish & Output Formats

### Flag: `--json`

Machine-readable output.

- Available on both `list` and `export`.

### Feature: Progress Reporting

Feedback during long-running scans.

- Spinner/progress bar instead of a silent wait.

### Feature: Shell Completions

Standard CLI ergonomics.

- Bash/zsh/fish completions via Cobra's generator.

### Feature: Config Validation

Catches problems before a command runs.

- Actionable error messages on malformed config.
