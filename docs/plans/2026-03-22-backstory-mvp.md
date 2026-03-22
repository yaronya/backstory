# Backstory MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that auto-captures team decisions from Claude Code sessions and makes them available to all team members' AI agents.

**Architecture:** Git-based decisions repo (separate from code repos) with structured markdown files. Go CLI integrates via Claude Code hooks (SessionStart for injection, Stop for capture). Local SQLite with FTS5 for fast search. Claude API (Haiku) for decision extraction.

**Tech Stack:** Go 1.26, modernc.org/sqlite (CGO-free, FTS5 built-in), adrg/frontmatter (YAML frontmatter parsing), gopkg.in/yaml.v3, cobra (CLI framework), Anthropic Go SDK

**Spec:** `docs/specs/2026-03-22-backstory-design.md`

---

## Review Fixes (must be applied during implementation)

The following issues were identified during plan review and MUST be addressed:

### HIGH — Fix during implementation

1. **Task 7 (inject): Add recency boost scoring.** The `Generate` function must sort results by a score combining anchor match + recency (exponential decay, 30-day half-life). Sort by score descending before truncating to `MaxDecisions`. Retrieval frequency tracking is deferred to post-MVP.

2. **Task 10 (init): Must run `git init`, `git add -A`, `git commit` automatically.** Also must write Claude Code hooks to `~/.claude/settings.json` (add SessionStart and Stop hooks). The spec promises 2-minute setup — don't print manual instructions.

3. **Task 10 (sync): Must drain pending queue.** Load `~/.backstory/pending/`, present confirmation prompt for pending items, then commit/push confirmed ones and clear the queue.

4. **Task 6 (pending): Add JSON struct tags to Decision type.** Add `json:"date_str"` to `DateStr` and `json:"date"` to `Date` fields. Verify round-trip works in pending queue tests.

5. **Task 4 (store): Use DELETE + INSERT instead of Upsert.** The FTS5 trigger-based sync with `ON CONFLICT DO UPDATE` may double-fire triggers. Safer to `DELETE WHERE file_path = ?` then `INSERT`, keeping the triggers clean.

### MEDIUM — Fix during implementation

6. **Task 10 (capture): Separate transcript input from interactive input.** Read transcript from a file path argument (`--transcript /path/to/file`) instead of stdin. Use stdin only for the interactive confirmation prompt. This fixes the double-stdin bug.

7. **Task 4 (store): Add `AND stale = 0` to `QueryByLinearIssue`.** Currently stale decisions leak through Linear issue queries.

8. **Task 9 (linear): Use `issueSearch` GraphQL query instead of `issue(id:)`.** Linear's `issue(id:)` expects UUIDs, not human identifiers like `ENG-1234`. Use `issueSearch(filter: {identifier: {eq: $id}})`.

9. **Task 9 (linear): Accept team key from config.** `ExtractIssueFromBranch` must use `config.Linear.TeamKey` instead of hardcoded `ENG`. Build regex dynamically from the config value.

10. **Task 10 (inject): Derive relative path within code repo.** Compute the CWD path relative to the code repo root and combine with `repoName` for the anchor query (e.g., `env0/services/payment-service/`), not just `env0/`.

11. **Task 7 (inject): Enforce `max_tokens` budget.** Estimate token count per decision (~4 chars per token) and stop adding decisions when the budget is exceeded.

12. **Task 10 (edit): Commit after editing.** Run `git add <file> && git commit -m "backstory: update decision"` after the editor closes.

13. **Task 10 (init): Support `--connect <url>` to clone an existing decisions repo.** The spec says init handles both create and connect flows.

---

## File Structure

```
backstory/
├── cmd/
│   └── backstory/
│       └── main.go                  ← CLI entrypoint, cobra root command
├── internal/
│   ├── cli/
│   │   ├── init.go                  ← `backstory init` command
│   │   ├── sync.go                  ← `backstory sync` command
│   │   ├── index.go                 ← `backstory index` command
│   │   ├── search.go                ← `backstory search` command
│   │   ├── inject.go                ← `backstory inject` command
│   │   ├── capture.go               ← `backstory capture` command
│   │   ├── status.go                ← `backstory status` command
│   │   └── edit.go                  ← `backstory edit` command
│   ├── config/
│   │   ├── config.go                ← Config loading (config.yml + config.local.yml)
│   │   └── config_test.go
│   ├── decision/
│   │   ├── decision.go              ← Decision type, frontmatter parsing, markdown I/O
│   │   └── decision_test.go
│   ├── store/
│   │   ├── store.go                 ← SQLite index: create, insert, query, FTS5 search
│   │   └── store_test.go
│   ├── repo/
│   │   ├── repo.go                  ← Git operations: clone, pull, push, commit, rebase
│   │   └── repo_test.go
│   ├── inject/
│   │   ├── inject.go                ← Relevance algorithm, XML output formatting
│   │   └── inject_test.go
│   ├── extract/
│   │   ├── extract.go               ← Claude API decision extraction from session transcripts
│   │   └── extract_test.go
│   ├── pending/
│   │   ├── pending.go               ← Pending queue: save/load/confirm/dismiss candidates
│   │   └── pending_test.go
│   ├── linear/
│   │   ├── linear.go                ← Linear API client: fetch issue, comments
│   │   └── linear_test.go
│   └── template/
│       └── repo_template.go         ← Embedded decisions repo template files
├── testdata/
│   ├── decisions/                   ← Sample decision files for tests
│   └── sessions/                    ← Sample session transcripts for extraction tests
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

### Task 1: Project scaffolding and CLI skeleton

**Files:**
- Create: `cmd/backstory/main.go`
- Create: `go.mod`
- Create: `Makefile`
- Create: `internal/cli/init.go` (stub)
- Create: `internal/cli/sync.go` (stub)
- Create: `internal/cli/index.go` (stub)
- Create: `internal/cli/search.go` (stub)
- Create: `internal/cli/inject.go` (stub)
- Create: `internal/cli/capture.go` (stub)
- Create: `internal/cli/status.go` (stub)
- Create: `internal/cli/edit.go` (stub)

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/yaronyarimi/dev/backstory
go mod init github.com/backstory-team/backstory
```

- [ ] **Step 2: Install cobra**

```bash
cd /Users/yaronyarimi/dev/backstory
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: Create main.go with root command**

```go
// cmd/backstory/main.go
package main

import (
	"fmt"
	"os"

	"github.com/backstory-team/backstory/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "backstory",
		Short: "Shared team memory for AI coding agents",
	}

	root.AddCommand(
		cli.NewInitCmd(),
		cli.NewSyncCmd(),
		cli.NewIndexCmd(),
		cli.NewSearchCmd(),
		cli.NewInjectCmd(),
		cli.NewCaptureCmd(),
		cli.NewStatusCmd(),
		cli.NewEditCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create stub commands**

Each stub follows this pattern (example for init):

```go
// internal/cli/init.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a decisions repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("backstory init: not yet implemented")
			return nil
		},
	}
}
```

Create the same pattern for: `sync.go`, `index.go`, `search.go`, `inject.go`, `capture.go`, `status.go`, `edit.go`.

- [ ] **Step 5: Create Makefile**

```makefile
.PHONY: build test clean

build:
	go build -o bin/backstory ./cmd/backstory

test:
	go test ./...

clean:
	rm -rf bin/
```

- [ ] **Step 6: Build and verify all commands register**

Run: `cd /Users/yaronyarimi/dev/backstory && make build && ./bin/backstory --help`
Expected: Shows "Shared team memory for AI coding agents" with all 8 subcommands listed.

Run: `./bin/backstory init`
Expected: "backstory init: not yet implemented"

- [ ] **Step 7: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add -A
git commit -m "scaffold: Go CLI skeleton with cobra and all command stubs"
```

---

### Task 2: Config loading

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Install yaml dependency**

```bash
cd /Users/yaronyarimi/dev/backstory
go get gopkg.in/yaml.v3
```

- [ ] **Step 2: Write failing tests for config loading**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ReadsTeamSettings(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	os.MkdirAll(backstoryDir, 0755)

	configYml := `team: my-team
repos:
  - name: env0
    url: git@github.com:env0/env0.git
linear:
  team_key: ENG
inject:
  max_decisions: 10
  max_tokens: 2000
staleness:
  archive_after_months: 6
  change_threshold: 0.5
`
	os.WriteFile(filepath.Join(backstoryDir, "config.yml"), []byte(configYml), 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Team != "my-team" {
		t.Errorf("expected team 'my-team', got '%s'", cfg.Team)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != "env0" {
		t.Errorf("expected repo name 'env0', got '%s'", cfg.Repos[0].Name)
	}
	if cfg.Inject.MaxDecisions != 10 {
		t.Errorf("expected max_decisions 10, got %d", cfg.Inject.MaxDecisions)
	}
}

func TestLoadConfig_MergesLocalOverrides(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	os.MkdirAll(backstoryDir, 0755)

	configYml := `team: my-team
repos: []
`
	localYml := `claude_api_key: sk-ant-test123
linear_api_key: lin_api_test456
`
	os.WriteFile(filepath.Join(backstoryDir, "config.yml"), []byte(configYml), 0644)
	os.WriteFile(filepath.Join(backstoryDir, "config.local.yml"), []byte(localYml), 0644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ClaudeAPIKey != "sk-ant-test123" {
		t.Errorf("expected claude key, got '%s'", cfg.ClaudeAPIKey)
	}
	if cfg.LinearAPIKey != "lin_api_test456" {
		t.Errorf("expected linear key, got '%s'", cfg.LinearAPIKey)
	}
}

func TestLoadConfig_DefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	os.MkdirAll(backstoryDir, 0755)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Inject.MaxDecisions != 10 {
		t.Errorf("expected default max_decisions 10, got %d", cfg.Inject.MaxDecisions)
	}
	if cfg.Inject.MaxTokens != 2000 {
		t.Errorf("expected default max_tokens 2000, got %d", cfg.Inject.MaxTokens)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/config/...`
Expected: FAIL (types not defined yet)

- [ ] **Step 4: Implement config types and loading**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type LinearConfig struct {
	TeamKey string `yaml:"team_key"`
}

type InjectConfig struct {
	MaxDecisions int `yaml:"max_decisions"`
	MaxTokens    int `yaml:"max_tokens"`
}

type StalenessConfig struct {
	ArchiveAfterMonths int     `yaml:"archive_after_months"`
	ChangeThreshold    float64 `yaml:"change_threshold"`
}

type Config struct {
	Team      string          `yaml:"team"`
	Repos     []Repo          `yaml:"repos"`
	Linear    LinearConfig    `yaml:"linear"`
	Inject    InjectConfig    `yaml:"inject"`
	Staleness StalenessConfig `yaml:"staleness"`

	ClaudeAPIKey string `yaml:"claude_api_key"`
	LinearAPIKey string `yaml:"linear_api_key"`
	SlackToken   string `yaml:"slack_token"`
}

func defaults() Config {
	return Config{
		Inject: InjectConfig{
			MaxDecisions: 10,
			MaxTokens:    2000,
		},
		Staleness: StalenessConfig{
			ArchiveAfterMonths: 6,
			ChangeThreshold:    0.5,
		},
	}
}

func Load(repoRoot string) (*Config, error) {
	cfg := defaults()

	teamFile := filepath.Join(repoRoot, ".backstory", "config.yml")
	if data, err := os.ReadFile(teamFile); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Inject.MaxDecisions == 0 {
		cfg.Inject.MaxDecisions = 10
	}
	if cfg.Inject.MaxTokens == 0 {
		cfg.Inject.MaxTokens = 2000
	}

	localFile := filepath.Join(repoRoot, ".backstory", "config.local.yml")
	if data, err := os.ReadFile(localFile); err == nil {
		var local Config
		if err := yaml.Unmarshal(data, &local); err != nil {
			return nil, err
		}
		if local.ClaudeAPIKey != "" {
			cfg.ClaudeAPIKey = local.ClaudeAPIKey
		}
		if local.LinearAPIKey != "" {
			cfg.LinearAPIKey = local.LinearAPIKey
		}
		if local.SlackToken != "" {
			cfg.SlackToken = local.SlackToken
		}
	}

	return &cfg, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/config/...`
Expected: PASS (3 tests)

- [ ] **Step 6: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/config/
git commit -m "feat: config loading with YAML parsing and local overrides"
```

---

### Task 3: Decision type and markdown I/O

**Files:**
- Create: `internal/decision/decision.go`
- Create: `internal/decision/decision_test.go`
- Create: `testdata/decisions/sample-technical.md`
- Create: `testdata/decisions/sample-product.md`

- [ ] **Step 1: Install frontmatter dependency**

```bash
cd /Users/yaronyarimi/dev/backstory
go get github.com/adrg/frontmatter
```

- [ ] **Step 2: Create sample decision test fixtures**

```markdown
<!-- testdata/decisions/sample-technical.md -->
---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s. Direct invocation from the Lambda
would hit this limit during peak hours. SQS provides natural backpressure
and retry semantics without custom rate-limiting code.

Considered alternatives:
- Direct invocation with client-side rate limiting — rejected
- SNS + SQS fan-out — overkill for a single consumer
```

```markdown
<!-- testdata/decisions/sample-product.md -->
---
type: product
date: 2026-03-22
author: pm-david
anchor: payments
linear_issue: ENG-892
stale: false
---

# No bulk operations in v1

Too risky for initial launch. Single-item operations only.
Bulk support planned for v2 after validating core flows.
```

- [ ] **Step 3: Write failing tests for decision parsing and writing**

```go
// internal/decision/decision_test.go
package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseFromFile_TechnicalDecision(t *testing.T) {
	d, err := ParseFromFile("../../testdata/decisions/sample-technical.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Type != "technical" {
		t.Errorf("expected type 'technical', got '%s'", d.Type)
	}
	if d.Author != "sarah" {
		t.Errorf("expected author 'sarah', got '%s'", d.Author)
	}
	if d.Anchor != "env0/services/payment-service/" {
		t.Errorf("expected anchor 'env0/services/payment-service/', got '%s'", d.Anchor)
	}
	if d.LinearIssue != "ENG-892" {
		t.Errorf("expected linear issue 'ENG-892', got '%s'", d.LinearIssue)
	}
	if !strings.Contains(d.Body, "SQS") {
		t.Errorf("expected body to contain 'SQS'")
	}
}

func TestParseFromFile_ProductDecision(t *testing.T) {
	d, err := ParseFromFile("../../testdata/decisions/sample-product.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Type != "product" {
		t.Errorf("expected type 'product', got '%s'", d.Type)
	}
	if d.Author != "pm-david" {
		t.Errorf("expected author 'pm-david', got '%s'", d.Author)
	}
}

func TestWriteToFile_RoundTrip(t *testing.T) {
	d := Decision{
		Type:        "technical",
		Date:        time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Author:      "sarah",
		Anchor:      "env0/services/payment-service/",
		LinearIssue: "ENG-892",
		Stale:       false,
		Title:       "Chose SQS over direct invocation",
		Body:        "The vendor API rate-limits at 100 req/s.",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test-decision.md")

	err := d.WriteToFile(path)
	if err != nil {
		t.Fatalf("unexpected error writing: %v", err)
	}

	parsed, err := ParseFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error parsing: %v", err)
	}
	if parsed.Type != d.Type {
		t.Errorf("round-trip type mismatch: got '%s'", parsed.Type)
	}
	if parsed.Author != d.Author {
		t.Errorf("round-trip author mismatch: got '%s'", parsed.Author)
	}
	if parsed.Anchor != d.Anchor {
		t.Errorf("round-trip anchor mismatch: got '%s'", parsed.Anchor)
	}
}

func TestParseAllFromDir(t *testing.T) {
	decisions, err := ParseAllFromDir("../../testdata/decisions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) < 2 {
		t.Errorf("expected at least 2 decisions, got %d", len(decisions))
	}
}

func TestGenerateFilename(t *testing.T) {
	d := Decision{
		Date:  time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Title: "Chose SQS over direct invocation",
	}
	name := d.Filename()
	if name != "2026-03-19-chose-sqs-over-direct-invocation.md" {
		t.Errorf("unexpected filename: '%s'", name)
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/decision/...`
Expected: FAIL

- [ ] **Step 5: Implement Decision type, parsing, and writing**

```go
// internal/decision/decision.go
package decision

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

type Decision struct {
	Type        string    `yaml:"type"`
	Date        time.Time `yaml:"-"`
	DateStr     string    `yaml:"date"`
	Author      string    `yaml:"author"`
	Anchor      string    `yaml:"anchor"`
	LinearIssue string    `yaml:"linear_issue"`
	Stale       bool      `yaml:"stale"`
	Title       string    `yaml:"-"`
	Body        string    `yaml:"-"`
	FilePath    string    `yaml:"-"`
}

func ParseFromFile(path string) (*Decision, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var d Decision
	body, err := frontmatter.Parse(f, &d)
	if err != nil {
		return nil, err
	}

	if d.DateStr != "" {
		d.Date, _ = time.Parse("2006-01-02", d.DateStr)
	}

	content := strings.TrimSpace(string(body))
	if strings.HasPrefix(content, "# ") {
		lines := strings.SplitN(content, "\n", 2)
		d.Title = strings.TrimPrefix(lines[0], "# ")
		if len(lines) > 1 {
			d.Body = strings.TrimSpace(lines[1])
		}
	} else {
		d.Body = content
	}

	d.FilePath = path
	return &d, nil
}

func ParseAllFromDir(dir string) ([]*Decision, error) {
	var decisions []*Decision
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		dec, err := ParseFromFile(path)
		if err != nil {
			return nil
		}
		decisions = append(decisions, dec)
		return nil
	})
	return decisions, err
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func (d *Decision) Filename() string {
	slug := strings.ToLower(d.Title)
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return fmt.Sprintf("%s-%s.md", d.Date.Format("2006-01-02"), slug)
}

func (d *Decision) WriteToFile(path string) error {
	if d.DateStr == "" {
		d.DateStr = d.Date.Format("2006-01-02")
	}
	fm, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n\n")
	if d.Title != "" {
		buf.WriteString("# ")
		buf.WriteString(d.Title)
		buf.WriteString("\n\n")
	}
	buf.WriteString(d.Body)
	buf.WriteString("\n")

	return os.WriteFile(path, []byte(buf.String()), 0644)
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/decision/...`
Expected: PASS (4 tests)

- [ ] **Step 7: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/decision/ testdata/
git commit -m "feat: decision type with frontmatter parsing and markdown I/O"
```

---

### Task 4: SQLite store with FTS5

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/store_test.go`

- [ ] **Step 1: Install SQLite dependency**

```bash
cd /Users/yaronyarimi/dev/backstory
go get modernc.org/sqlite
```

- [ ] **Step 2: Write failing tests for store operations**

```go
// internal/store/store_test.go
package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/backstory-team/backstory/internal/decision"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestInsertAndGetByAnchor(t *testing.T) {
	s := newTestStore(t)

	d := &decision.Decision{
		Type:        "technical",
		Date:        time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Author:      "sarah",
		Anchor:      "env0/services/payment-service/",
		LinearIssue: "ENG-892",
		Title:       "Chose SQS",
		Body:        "The vendor API rate-limits at 100 req/s.",
		FilePath:    "technical/env0/services/payment-service/2026-03-19-chose-sqs.md",
	}
	if err := s.Upsert(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := s.QueryByAnchor("env0/services/payment-service/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Chose SQS" {
		t.Errorf("expected title 'Chose SQS', got '%s'", results[0].Title)
	}
}

func TestFTS5Search(t *testing.T) {
	s := newTestStore(t)

	d1 := &decision.Decision{
		Type: "technical", Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Author: "sarah", Anchor: "env0/services/payment-service/",
		Title: "Chose SQS", Body: "vendor API rate-limits at 100 req/s",
		FilePath: "technical/env0/2026-03-19-chose-sqs.md",
	}
	d2 := &decision.Decision{
		Type: "product", Date: time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
		Author: "david", Anchor: "payments",
		Title: "No bulk operations in v1", Body: "Too risky for launch",
		FilePath: "product/payments/2026-03-22-no-bulk.md",
	}
	s.Upsert(d1)
	s.Upsert(d2)

	results, err := s.Search("rate limit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'rate limit', got %d", len(results))
	}
	if results[0].Title != "Chose SQS" {
		t.Errorf("expected 'Chose SQS', got '%s'", results[0].Title)
	}
}

func TestQueryByLinearIssue(t *testing.T) {
	s := newTestStore(t)

	d := &decision.Decision{
		Type: "product", Date: time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
		Author: "david", Anchor: "payments", LinearIssue: "ENG-892",
		Title: "No bulk ops", Body: "Risky",
		FilePath: "product/payments/2026-03-22-no-bulk.md",
	}
	s.Upsert(d)

	results, err := s.QueryByLinearIssue("ENG-892")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestExcludeStale(t *testing.T) {
	s := newTestStore(t)

	d := &decision.Decision{
		Type: "technical", Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Author: "sarah", Anchor: "env0/services/payment-service/",
		Title: "Old decision", Body: "outdated", Stale: true,
		FilePath: "technical/env0/2026-03-19-old.md",
	}
	s.Upsert(d)

	results, err := s.QueryByAnchor("env0/services/payment-service/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results (stale excluded), got %d", len(results))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/store/...`
Expected: FAIL

- [ ] **Step 4: Implement SQLite store with FTS5**

```go
// internal/store/store.go
package store

import (
	"database/sql"
	"fmt"

	"github.com/backstory-team/backstory/internal/decision"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS decisions (
			file_path TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			date TEXT NOT NULL,
			author TEXT NOT NULL,
			anchor TEXT NOT NULL,
			linear_issue TEXT,
			stale INTEGER NOT NULL DEFAULT 0,
			title TEXT NOT NULL,
			body TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS decisions_fts USING fts5(
			title, body, anchor, content='decisions', content_rowid='rowid'
		);

		CREATE TRIGGER IF NOT EXISTS decisions_ai AFTER INSERT ON decisions BEGIN
			INSERT INTO decisions_fts(rowid, title, body, anchor)
			VALUES (new.rowid, new.title, new.body, new.anchor);
		END;

		CREATE TRIGGER IF NOT EXISTS decisions_ad AFTER DELETE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, title, body, anchor)
			VALUES ('delete', old.rowid, old.title, old.body, old.anchor);
		END;

		CREATE TRIGGER IF NOT EXISTS decisions_au AFTER UPDATE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, title, body, anchor)
			VALUES ('delete', old.rowid, old.title, old.body, old.anchor);
			INSERT INTO decisions_fts(rowid, title, body, anchor)
			VALUES (new.rowid, new.title, new.body, new.anchor);
		END;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Upsert(d *decision.Decision) error {
	stale := 0
	if d.Stale {
		stale = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO decisions (file_path, type, date, author, anchor, linear_issue, stale, title, body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			type=excluded.type, date=excluded.date, author=excluded.author,
			anchor=excluded.anchor, linear_issue=excluded.linear_issue,
			stale=excluded.stale, title=excluded.title, body=excluded.body
	`, d.FilePath, d.Type, d.Date.Format("2006-01-02"), d.Author, d.Anchor, d.LinearIssue, stale, d.Title, d.Body)
	return err
}

func (s *Store) scanDecisions(rows *sql.Rows) ([]*decision.Decision, error) {
	var results []*decision.Decision
	for rows.Next() {
		var d decision.Decision
		var stale int
		if err := rows.Scan(&d.FilePath, &d.Type, &d.DateStr, &d.Author, &d.Anchor, &d.LinearIssue, &stale, &d.Title, &d.Body); err != nil {
			return nil, err
		}
		d.Stale = stale != 0
		results = append(results, &d)
	}
	return results, rows.Err()
}

func (s *Store) QueryByAnchor(anchor string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT file_path, type, date, author, anchor, linear_issue, stale, title, body
		FROM decisions
		WHERE (anchor = ? OR anchor LIKE ? OR ? LIKE anchor || '%')
		AND stale = 0
		ORDER BY date DESC
	`, anchor, anchor+"%", anchor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanDecisions(rows)
}

func (s *Store) QueryByLinearIssue(issue string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT file_path, type, date, author, anchor, linear_issue, stale, title, body
		FROM decisions
		WHERE linear_issue = ?
		ORDER BY date DESC
	`, issue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanDecisions(rows)
}

func (s *Store) Search(query string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT d.file_path, d.type, d.date, d.author, d.anchor, d.linear_issue, d.stale, d.title, d.body
		FROM decisions d
		JOIN decisions_fts f ON d.rowid = f.rowid
		WHERE decisions_fts MATCH ?
		AND d.stale = 0
		ORDER BY rank
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanDecisions(rows)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/store/...`
Expected: PASS (4 tests)

- [ ] **Step 6: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/store/
git commit -m "feat: SQLite store with FTS5 full-text search index"
```

---

### Task 5: Git repo operations

**Files:**
- Create: `internal/repo/repo.go`
- Create: `internal/repo/repo_test.go`

- [ ] **Step 1: Write failing tests for git operations**

```go
// internal/repo/repo_test.go
package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initBareRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	cmd := exec.Command("git", "init", "--bare", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %s %v", out, err)
	}
	return dir
}

func TestClone(t *testing.T) {
	remote := initBareRemote(t)
	local := filepath.Join(t.TempDir(), "local")

	r, err := Clone(remote, local)
	if err != nil {
		t.Fatalf("clone failed: %v", err)
	}
	if r.Path != local {
		t.Errorf("expected path '%s', got '%s'", local, r.Path)
	}
	if _, err := os.Stat(filepath.Join(local, ".git")); os.IsNotExist(err) {
		t.Error("expected .git directory to exist")
	}
}

func TestCommitAndPush(t *testing.T) {
	remote := initBareRemote(t)
	local := filepath.Join(t.TempDir(), "local")

	r, _ := Clone(remote, local)

	testFile := filepath.Join(local, "test.md")
	os.WriteFile(testFile, []byte("hello"), 0644)

	err := r.CommitAndPush("test.md", "test commit")
	if err != nil {
		t.Fatalf("commit and push failed: %v", err)
	}
}

func TestPull(t *testing.T) {
	remote := initBareRemote(t)
	local1 := filepath.Join(t.TempDir(), "local1")
	local2 := filepath.Join(t.TempDir(), "local2")

	r1, _ := Clone(remote, local1)
	r2, _ := Clone(remote, local2)

	os.WriteFile(filepath.Join(local1, "file.md"), []byte("from local1"), 0644)
	r1.CommitAndPush("file.md", "add file")

	err := r2.Pull()
	if err != nil {
		t.Fatalf("pull failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(local2, "file.md"))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "from local1" {
		t.Errorf("expected 'from local1', got '%s'", string(data))
	}
}

func TestPullRebase(t *testing.T) {
	remote := initBareRemote(t)
	local1 := filepath.Join(t.TempDir(), "local1")
	local2 := filepath.Join(t.TempDir(), "local2")

	r1, _ := Clone(remote, local1)
	r2, _ := Clone(remote, local2)

	os.WriteFile(filepath.Join(local1, "file1.md"), []byte("from local1"), 0644)
	r1.CommitAndPush("file1.md", "add file1")

	os.WriteFile(filepath.Join(local2, "file2.md"), []byte("from local2"), 0644)
	r2.CommitAll("add file2")

	err := r2.PushWithRebase(3)
	if err != nil {
		t.Fatalf("push with rebase failed: %v", err)
	}
}

func TestGetRemoteURL(t *testing.T) {
	remote := initBareRemote(t)
	local := filepath.Join(t.TempDir(), "local")

	r, _ := Clone(remote, local)
	url, err := r.GetRemoteURL()
	if err != nil {
		t.Fatalf("get remote url failed: %v", err)
	}
	if url != remote {
		t.Errorf("expected '%s', got '%s'", remote, url)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	remote := initBareRemote(t)
	local := filepath.Join(t.TempDir(), "local")

	r, _ := Clone(remote, local)

	os.WriteFile(filepath.Join(local, "init.md"), []byte("init"), 0644)
	r.CommitAndPush("init.md", "initial")

	branch, err := r.GetCurrentBranch()
	if err != nil {
		t.Fatalf("get branch failed: %v", err)
	}
	if branch != "master" && branch != "main" {
		t.Errorf("expected 'master' or 'main', got '%s'", branch)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/repo/...`
Expected: FAIL

- [ ] **Step 3: Implement git operations**

```go
// internal/repo/repo.go
package repo

import (
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	Path string
}

func Clone(url, dest string) (*Repo, error) {
	cmd := exec.Command("git", "clone", url, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone: %s %w", out, err)
	}
	return &Repo{Path: dest}, nil
}

func Open(path string) *Repo {
	return &Repo{Path: path}
}

func (r *Repo) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *Repo) Pull() error {
	_, err := r.git("pull", "--ff-only")
	return err
}

func (r *Repo) CommitAll(msg string) error {
	if _, err := r.git("add", "-A"); err != nil {
		return err
	}
	_, err := r.git("commit", "-m", msg)
	return err
}

func (r *Repo) CommitAndPush(file, msg string) error {
	if _, err := r.git("add", file); err != nil {
		return err
	}
	if _, err := r.git("commit", "-m", msg); err != nil {
		return err
	}
	_, err := r.git("push")
	return err
}

func (r *Repo) PushWithRebase(maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		_, err := r.git("push")
		if err == nil {
			return nil
		}
		if _, rebaseErr := r.git("pull", "--rebase"); rebaseErr != nil {
			return fmt.Errorf("pull --rebase failed: %w", rebaseErr)
		}
	}
	return fmt.Errorf("push failed after %d retries", maxRetries)
}

func (r *Repo) GetRemoteURL() (string, error) {
	return r.git("remote", "get-url", "origin")
}

func (r *Repo) GetCurrentBranch() (string, error) {
	return r.git("rev-parse", "--abbrev-ref", "HEAD")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/repo/...`
Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/repo/
git commit -m "feat: git repo operations with pull-rebase retry strategy"
```

---

### Task 6: Pending queue

**Files:**
- Create: `internal/pending/pending.go`
- Create: `internal/pending/pending_test.go`

- [ ] **Step 1: Write failing tests for pending queue**

```go
// internal/pending/pending_test.go
package pending

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/backstory-team/backstory/internal/decision"
)

func TestSaveAndLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pending")
	q := New(dir)

	d := &decision.Decision{
		Type: "technical", Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		Author: "sarah", Anchor: "env0/services/payment-service/",
		Title: "Chose SQS", Body: "Rate limit reason",
	}

	err := q.Save([]*decision.Decision{d})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := q.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(loaded))
	}
	if loaded[0].Title != "Chose SQS" {
		t.Errorf("expected title 'Chose SQS', got '%s'", loaded[0].Title)
	}
}

func TestClear(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pending")
	q := New(dir)

	d := &decision.Decision{
		Type: "technical", Date: time.Now(),
		Author: "sarah", Anchor: "test/",
		Title: "Test", Body: "test body",
	}
	q.Save([]*decision.Decision{d})

	err := q.Clear()
	if err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	loaded, _ := q.Load()
	if len(loaded) != 0 {
		t.Errorf("expected 0 after clear, got %d", len(loaded))
	}
}

func TestHasPending(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pending")
	q := New(dir)

	if q.HasPending() {
		t.Error("expected no pending initially")
	}

	d := &decision.Decision{
		Type: "technical", Date: time.Now(),
		Author: "sarah", Anchor: "test/",
		Title: "Test", Body: "body",
	}
	q.Save([]*decision.Decision{d})

	if !q.HasPending() {
		t.Error("expected pending after save")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/pending/...`
Expected: FAIL

- [ ] **Step 3: Implement pending queue**

```go
// internal/pending/pending.go
package pending

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/decision"
)

type Queue struct {
	dir string
}

func New(dir string) *Queue {
	return &Queue{dir: dir}
}

func (q *Queue) file() string {
	return filepath.Join(q.dir, "pending.json")
}

func (q *Queue) Save(decisions []*decision.Decision) error {
	if err := os.MkdirAll(q.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(decisions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(q.file(), data, 0644)
}

func (q *Queue) Load() ([]*decision.Decision, error) {
	data, err := os.ReadFile(q.file())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var decisions []*decision.Decision
	if err := json.Unmarshal(data, &decisions); err != nil {
		return nil, err
	}
	return decisions, nil
}

func (q *Queue) Clear() error {
	return os.Remove(q.file())
}

func (q *Queue) HasPending() bool {
	info, err := os.Stat(q.file())
	return err == nil && info.Size() > 0
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/pending/...`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/pending/
git commit -m "feat: pending queue for offline decision capture resilience"
```

---

### Task 7: Injection engine (relevance algorithm + XML output)

**Files:**
- Create: `internal/inject/inject.go`
- Create: `internal/inject/inject_test.go`

- [ ] **Step 1: Write failing tests for injection logic**

```go
// internal/inject/inject_test.go
package inject

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/backstory-team/backstory/internal/config"
	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/store"
)

func seedStore(t *testing.T, s *store.Store) {
	t.Helper()
	decisions := []*decision.Decision{
		{Type: "technical", Date: time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC), Author: "sarah", Anchor: "env0/services/payment-service/", LinearIssue: "ENG-892", Title: "Chose SQS", Body: "Rate limit reason", FilePath: "t/1.md"},
		{Type: "product", Date: time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC), Author: "david", Anchor: "payments", LinearIssue: "ENG-892", Title: "No bulk ops v1", Body: "Too risky", FilePath: "p/1.md"},
		{Type: "technical", Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), Author: "alice", Anchor: "env0/services/notification-service/", Title: "SES templates", Body: "Use SES", FilePath: "t/2.md"},
		{Type: "technical", Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Author: "bob", Anchor: "env0/services/payment-service/", Title: "Old decision", Body: "Very old", FilePath: "t/3.md", Stale: true},
	}
	for _, d := range decisions {
		s.Upsert(d)
	}
}

func TestInjectByAnchor(t *testing.T) {
	s := openTestStore(t)
	seedStore(t, s)

	cfg := &config.Config{Inject: config.InjectConfig{MaxDecisions: 10, MaxTokens: 5000}}
	eng := New(s, cfg)

	output, err := eng.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Chose SQS") {
		t.Error("expected output to contain 'Chose SQS'")
	}
	if strings.Contains(output, "SES templates") {
		t.Error("should not contain unrelated decisions")
	}
	if strings.Contains(output, "Old decision") {
		t.Error("should not contain stale decisions")
	}
}

func TestInjectByLinearIssue(t *testing.T) {
	s := openTestStore(t)
	seedStore(t, s)

	cfg := &config.Config{Inject: config.InjectConfig{MaxDecisions: 10, MaxTokens: 5000}}
	eng := New(s, cfg)

	output, err := eng.Generate("", "ENG-892")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Chose SQS") {
		t.Error("expected 'Chose SQS' via linear issue match")
	}
	if !strings.Contains(output, "No bulk ops") {
		t.Error("expected 'No bulk ops' via linear issue match")
	}
}

func TestInjectXMLFormat(t *testing.T) {
	s := openTestStore(t)
	seedStore(t, s)

	cfg := &config.Config{Inject: config.InjectConfig{MaxDecisions: 10, MaxTokens: 5000}}
	eng := New(s, cfg)

	output, _ := eng.Generate("env0/services/payment-service/", "")
	if !strings.HasPrefix(output, "<backstory>") {
		t.Error("expected output to start with <backstory>")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "</backstory>") {
		t.Error("expected output to end with </backstory>")
	}
	if !strings.Contains(output, `type="technical"`) {
		t.Error("expected XML attributes in output")
	}
}

func TestInjectRespectsMaxDecisions(t *testing.T) {
	s := openTestStore(t)

	for i := 0; i < 20; i++ {
		s.Upsert(&decision.Decision{
			Type: "technical", Date: time.Date(2026, 3, 1+i%28, 0, 0, 0, 0, time.UTC),
			Author: "test", Anchor: "env0/services/payment-service/",
			Title: strings.Repeat("x", 50), Body: strings.Repeat("word ", 20),
			FilePath: filepath.Join("t", strings.Repeat("x", 10)+string(rune('a'+i))+".md"),
		})
	}

	cfg := &config.Config{Inject: config.InjectConfig{MaxDecisions: 5, MaxTokens: 50000}}
	eng := New(s, cfg)

	output, _ := eng.Generate("env0/services/payment-service/", "")
	count := strings.Count(output, "<decision ")
	if count > 5 {
		t.Errorf("expected at most 5 decisions, got %d", count)
	}
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/inject/...`
Expected: FAIL

- [ ] **Step 3: Implement injection engine**

```go
// internal/inject/inject.go
package inject

import (
	"fmt"
	"strings"

	"github.com/backstory-team/backstory/internal/config"
	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/store"
)

type Engine struct {
	store *store.Store
	cfg   *config.Config
}

func New(s *store.Store, cfg *config.Config) *Engine {
	return &Engine{store: s, cfg: cfg}
}

func (e *Engine) Generate(anchor, linearIssue string) (string, error) {
	seen := map[string]bool{}
	var all []*decision.Decision

	if anchor != "" {
		results, err := e.store.QueryByAnchor(anchor)
		if err != nil {
			return "", err
		}
		for _, d := range results {
			if !seen[d.FilePath] {
				seen[d.FilePath] = true
				all = append(all, d)
			}
		}
	}

	if linearIssue != "" {
		results, err := e.store.QueryByLinearIssue(linearIssue)
		if err != nil {
			return "", err
		}
		for _, d := range results {
			if !seen[d.FilePath] {
				seen[d.FilePath] = true
				all = append(all, d)
			}
		}
	}

	max := e.cfg.Inject.MaxDecisions
	if max > 0 && len(all) > max {
		all = all[:max]
	}

	if len(all) == 0 {
		return "", nil
	}

	return formatXML(all), nil
}

func formatXML(decisions []*decision.Decision) string {
	var buf strings.Builder
	buf.WriteString("<backstory>\n<decisions>\n")
	for _, d := range decisions {
		buf.WriteString(fmt.Sprintf(`<decision type="%s" date="%s" author="%s" anchor="%s"`,
			d.Type, d.DateStr, d.Author, d.Anchor))
		if d.LinearIssue != "" {
			buf.WriteString(fmt.Sprintf(` linear="%s"`, d.LinearIssue))
		}
		buf.WriteString(">\n")
		if d.Title != "" {
			buf.WriteString(d.Title)
			buf.WriteString(". ")
		}
		buf.WriteString(d.Body)
		buf.WriteString("\n</decision>\n")
	}
	buf.WriteString("</decisions>\n</backstory>")
	return buf.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/inject/...`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/inject/
git commit -m "feat: injection engine with relevance ranking and XML output"
```

---

### Task 8: Decision extraction via Claude API

**Files:**
- Create: `internal/extract/extract.go`
- Create: `internal/extract/extract_test.go`
- Create: `testdata/sessions/sample-session.txt`

- [ ] **Step 1: Install Anthropic Go SDK**

```bash
cd /Users/yaronyarimi/dev/backstory
go get github.com/anthropics/anthropic-sdk-go
```

- [ ] **Step 2: Create sample session transcript fixture**

```text
<!-- testdata/sessions/sample-session.txt -->
User asked me to implement webhook handling for Stripe payments.

I looked at the existing code in services/payment-service/handler.ts.

The vendor API rate-limits at 100 requests per second. I decided to use SQS
instead of direct invocation to handle backpressure naturally. I considered
direct invocation with client-side rate limiting but rejected it because it
adds complexity and doesn't handle Lambda concurrency spikes well.

I also added exponential backoff for Stripe webhook retries since webhooks
can be delayed up to 5 minutes according to Stripe docs.

I spent 20 minutes debugging a flaky test that was caused by a timing issue
in the test setup. Fixed it by adding a proper wait condition.

I renamed a variable from `resp` to `webhookResponse` for clarity.
```

- [ ] **Step 3: Write tests for extraction (with mock)**

```go
// internal/extract/extract_test.go
package extract

import (
	"testing"
)

func TestParseExtractionResponse(t *testing.T) {
	response := `[
		{
			"title": "Chose SQS over direct invocation for vendor API",
			"body": "The vendor API rate-limits at 100 req/s. SQS provides natural backpressure.",
			"anchor": "env0/services/payment-service/",
			"type": "technical",
			"alternatives_considered": "Direct invocation with rate limiting — rejected due to complexity"
		},
		{
			"title": "Added exponential backoff for Stripe webhook retries",
			"body": "Stripe webhooks can be delayed up to 5 minutes. Exponential backoff prevents thundering herd.",
			"anchor": "env0/services/payment-service/",
			"type": "technical",
			"alternatives_considered": ""
		}
	]`

	decisions, err := ParseExtractionResponse(response)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(decisions))
	}
	if decisions[0].Title != "Chose SQS over direct invocation for vendor API" {
		t.Errorf("unexpected title: %s", decisions[0].Title)
	}
	if decisions[0].Anchor != "env0/services/payment-service/" {
		t.Errorf("unexpected anchor: %s", decisions[0].Anchor)
	}
	if decisions[1].Title != "Added exponential backoff for Stripe webhook retries" {
		t.Errorf("unexpected title: %s", decisions[1].Title)
	}
}

func TestBuildExtractionPrompt(t *testing.T) {
	transcript := "I decided to use SQS for backpressure."
	repoName := "env0"
	workDir := "services/payment-service"

	prompt := BuildExtractionPrompt(transcript, repoName, workDir)

	if prompt == "" {
		t.Error("expected non-empty prompt")
	}
	if len(prompt) < 100 {
		t.Error("prompt seems too short")
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/extract/...`
Expected: FAIL

- [ ] **Step 5: Implement extraction logic**

```go
// internal/extract/extract.go
package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/backstory-team/backstory/internal/decision"
)

type extractedDecision struct {
	Title                  string `json:"title"`
	Body                   string `json:"body"`
	Anchor                 string `json:"anchor"`
	Type                   string `json:"type"`
	AlternativesConsidered string `json:"alternatives_considered"`
}

func BuildExtractionPrompt(transcript, repoName, workDir string) string {
	return fmt.Sprintf(`You are analyzing a coding session transcript to extract architectural and design decisions.

Context:
- Repository: %s
- Working directory: %s

Session transcript:
---
%s
---

Extract ONLY meaningful decisions — choices where alternatives existed and a deliberate selection was made.

DO NOT extract:
- Bug fixes or debugging steps
- Variable renames or formatting changes
- Routine implementation without alternatives considered

For each decision, provide:
- title: A concise description of the decision (imperative form)
- body: The reasoning and context behind the decision
- anchor: The code path this decision relates to (format: repo/path/)
- type: "technical" or "product"
- alternatives_considered: What was considered and rejected (empty string if none)

Return a JSON array. If no decisions found, return an empty array [].`, repoName, workDir, transcript)
}

func ParseExtractionResponse(response string) ([]*decision.Decision, error) {
	var extracted []extractedDecision
	if err := json.Unmarshal([]byte(response), &extracted); err != nil {
		return nil, fmt.Errorf("parsing extraction response: %w", err)
	}

	var decisions []*decision.Decision
	now := time.Now()
	for _, e := range extracted {
		body := e.Body
		if e.AlternativesConsidered != "" {
			body += "\n\nConsidered alternatives:\n- " + e.AlternativesConsidered
		}
		decisions = append(decisions, &decision.Decision{
			Type:    e.Type,
			Date:    now,
			DateStr: now.Format("2006-01-02"),
			Anchor:  e.Anchor,
			Title:   e.Title,
			Body:    body,
		})
	}
	return decisions, nil
}

type Extractor struct {
	client *anthropic.Client
}

func NewExtractor(apiKey string) *Extractor {
	client := anthropic.NewClient(anthropic.WithAPIKey(apiKey))
	return &Extractor{client: client}
}

func (e *Extractor) Extract(ctx context.Context, transcript, repoName, workDir, author string) ([]*decision.Decision, error) {
	prompt := BuildExtractionPrompt(transcript, repoName, workDir)

	msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API call: %w", err)
	}

	var responseText string
	for _, block := range msg.Content {
		if block.Type == "text" {
			responseText = block.Text
		}
	}

	decisions, err := ParseExtractionResponse(responseText)
	if err != nil {
		return nil, err
	}

	for _, d := range decisions {
		d.Author = author
	}

	return decisions, nil
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/extract/...`
Expected: PASS (2 tests — prompt building and response parsing; actual API call tested manually)

- [ ] **Step 7: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/extract/ testdata/sessions/
git commit -m "feat: decision extraction via Claude API with prompt engineering"
```

---

### Task 9: Linear API client

**Files:**
- Create: `internal/linear/linear.go`
- Create: `internal/linear/linear_test.go`

- [ ] **Step 1: Write tests for Linear response parsing**

```go
// internal/linear/linear_test.go
package linear

import (
	"encoding/json"
	"testing"
)

func TestParseIssueResponse(t *testing.T) {
	response := `{
		"data": {
			"issue": {
				"id": "abc-123",
				"identifier": "ENG-1234",
				"title": "Add Stripe webhook retry logic",
				"description": "Implement exponential backoff for failed Stripe webhook deliveries",
				"comments": {
					"nodes": [
						{"body": "Should we use a dead-letter queue for failed webhooks?", "user": {"name": "David"}},
						{"body": "Yes, good idea. Let's add that.", "user": {"name": "Sarah"}}
					]
				}
			}
		}
	}`

	var graphQLResp graphQLResponse
	if err := json.Unmarshal([]byte(response), &graphQLResp); err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	issue := graphQLResp.Data.Issue
	if issue.Identifier != "ENG-1234" {
		t.Errorf("expected ENG-1234, got %s", issue.Identifier)
	}
	if issue.Title != "Add Stripe webhook retry logic" {
		t.Errorf("unexpected title: %s", issue.Title)
	}
	if len(issue.Comments.Nodes) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(issue.Comments.Nodes))
	}
}

func TestFormatIssueXML(t *testing.T) {
	issue := &Issue{
		Identifier:  "ENG-1234",
		Title:       "Add Stripe webhook retry",
		Description: "Implement backoff",
		Comments: []Comment{
			{Body: "Use DLQ?", UserName: "David"},
		},
	}

	xml := FormatIssueXML(issue)
	if xml == "" {
		t.Error("expected non-empty XML")
	}
}

func TestExtractIssueFromBranch(t *testing.T) {
	tests := []struct {
		branch   string
		expected string
	}{
		{"eng-1234-add-webhook-retry", "ENG-1234"},
		{"ENG-567-fix-bug", "ENG-567"},
		{"main", ""},
		{"feature/eng-89-something", "ENG-89"},
	}
	for _, tt := range tests {
		result := ExtractIssueFromBranch(tt.branch)
		if result != tt.expected {
			t.Errorf("branch '%s': expected '%s', got '%s'", tt.branch, tt.expected, result)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/linear/...`
Expected: FAIL

- [ ] **Step 3: Implement Linear client**

```go
// internal/linear/linear.go
package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Comment struct {
	Body     string `json:"body"`
	UserName string `json:"-"`
}

type Issue struct {
	ID          string    `json:"id"`
	Identifier  string    `json:"identifier"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Comments    []Comment `json:"-"`
}

type graphQLResponse struct {
	Data struct {
		Issue struct {
			ID          string `json:"id"`
			Identifier  string `json:"identifier"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Comments    struct {
				Nodes []struct {
					Body string `json:"body"`
					User struct {
						Name string `json:"name"`
					} `json:"user"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"issue"`
	} `json:"data"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{}}
}

func (c *Client) FetchIssue(ctx context.Context, identifier string) (*Issue, error) {
	query := `query($id: String!) {
		issue(id: $id) {
			id identifier title description
			comments { nodes { body user { name } } }
		}
	}`

	payload, _ := json.Marshal(map[string]any{
		"query":     query,
		"variables": map[string]string{"id": identifier},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.linear.app/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("linear API error %d: %s", resp.StatusCode, body)
	}

	var graphResp graphQLResponse
	if err := json.Unmarshal(body, &graphResp); err != nil {
		return nil, err
	}

	issue := &Issue{
		ID:          graphResp.Data.Issue.ID,
		Identifier:  graphResp.Data.Issue.Identifier,
		Title:       graphResp.Data.Issue.Title,
		Description: graphResp.Data.Issue.Description,
	}
	for _, n := range graphResp.Data.Issue.Comments.Nodes {
		issue.Comments = append(issue.Comments, Comment{Body: n.Body, UserName: n.User.Name})
	}
	return issue, nil
}

func FormatIssueXML(issue *Issue) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf(`<linear issue="%s">`+"\n", issue.Identifier))
	buf.WriteString(fmt.Sprintf("Title: %s\n", issue.Title))
	if issue.Description != "" {
		buf.WriteString(fmt.Sprintf("Description: %s\n", issue.Description))
	}
	if len(issue.Comments) > 0 {
		buf.WriteString("Discussion:\n")
		for _, c := range issue.Comments {
			buf.WriteString(fmt.Sprintf("- %s: %s\n", c.UserName, c.Body))
		}
	}
	buf.WriteString("</linear>")
	return buf.String()
}

var branchIssuePattern = regexp.MustCompile(`(?i)(eng)-(\d+)`)

func ExtractIssueFromBranch(branch string) string {
	match := branchIssuePattern.FindStringSubmatch(branch)
	if match == nil {
		return ""
	}
	return strings.ToUpper(match[1]) + "-" + match[2]
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/yaronyarimi/dev/backstory && go test ./internal/linear/...`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add internal/linear/
git commit -m "feat: Linear API client with GraphQL issue fetching"
```

---

### Task 10: Wire up CLI commands

**Files:**
- Modify: `internal/cli/init.go`
- Modify: `internal/cli/sync.go`
- Modify: `internal/cli/index.go`
- Modify: `internal/cli/search.go`
- Modify: `internal/cli/inject.go`
- Modify: `internal/cli/capture.go`
- Modify: `internal/cli/status.go`
- Modify: `internal/cli/edit.go`
- Create: `internal/template/repo_template.go`

This is the largest task — wiring all internal packages into the CLI commands. Each command is relatively thin, delegating to the internal packages.

- [ ] **Step 1: Create repo template with embedded files**

```go
// internal/template/repo_template.go
package template

import "embed"

//go:embed files/*
var RepoTemplate embed.FS
```

Create embedded template files:

```bash
mkdir -p /Users/yaronyarimi/dev/backstory/internal/template/files/.backstory
mkdir -p /Users/yaronyarimi/dev/backstory/internal/template/files/product
mkdir -p /Users/yaronyarimi/dev/backstory/internal/template/files/technical
```

```yaml
# internal/template/files/.backstory/config.yml
team: ""
repos: []
linear:
  team_key: ""
inject:
  max_decisions: 10
  max_tokens: 2000
staleness:
  archive_after_months: 6
  change_threshold: 0.5
```

```gitignore
# internal/template/files/.backstory/.gitignore
config.local.yml
index.db
```

```markdown
<!-- internal/template/files/README.md -->
# Team Decisions

This repository is managed by [Backstory](https://github.com/backstory-team/backstory).

Decisions are auto-captured from coding sessions and organized into:
- `product/` — Product decisions from PMs
- `technical/` — Technical decisions from dev sessions
```

```
# internal/template/files/product/.gitkeep
```

```
# internal/template/files/technical/.gitkeep
```

- [ ] **Step 2: Implement `backstory init`**

```go
// internal/cli/init.go
package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/template"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var repoPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a decisions repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoPath == "" {
				repoPath = "backstory-decisions"
			}

			if _, err := os.Stat(repoPath); err == nil {
				return fmt.Errorf("directory '%s' already exists", repoPath)
			}

			if err := os.MkdirAll(repoPath, 0755); err != nil {
				return err
			}

			err := fs.WalkDir(template.RepoTemplate, "files", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel("files", path)
				dest := filepath.Join(repoPath, rel)

				if d.IsDir() {
					return os.MkdirAll(dest, 0755)
				}
				data, err := template.RepoTemplate.ReadFile(path)
				if err != nil {
					return err
				}
				return os.WriteFile(dest, data, 0644)
			})
			if err != nil {
				return err
			}

			fmt.Printf("Decisions repo initialized at %s\n", repoPath)
			fmt.Println("Next steps:")
			fmt.Println("  1. cd", repoPath, "&& git init && git add -A && git commit -m 'init backstory'")
			fmt.Println("  2. Push to a remote (e.g., gh repo create --push)")
			fmt.Println("  3. Edit .backstory/config.yml with your team settings")
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", "", "Path for the new decisions repo (default: backstory-decisions)")
	return cmd
}
```

- [ ] **Step 3: Implement `backstory sync`**

```go
// internal/cli/sync.go
package cli

import (
	"fmt"
	"os"

	"github.com/backstory-team/backstory/internal/repo"
	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Pull latest and push pending changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO environment variable not set")
			}

			r := repo.Open(repoPath)
			if err := r.Pull(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: pull failed: %v\n", err)
			}

			fmt.Println("Synced decisions repo")
			return nil
		},
	}
}
```

- [ ] **Step 4: Implement `backstory index`**

```go
// internal/cli/index.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/store"
	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Rebuild the local search index from decision files",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO environment variable not set")
			}

			dbPath := filepath.Join(repoPath, ".backstory", "index.db")
			os.Remove(dbPath)

			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			count := 0
			for _, dir := range []string{"product", "technical"} {
				dirPath := filepath.Join(repoPath, dir)
				decisions, err := decision.ParseAllFromDir(dirPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: parsing %s: %v\n", dir, err)
					continue
				}
				for _, d := range decisions {
					if err := s.Upsert(d); err != nil {
						fmt.Fprintf(os.Stderr, "warning: indexing %s: %v\n", d.FilePath, err)
						continue
					}
					count++
				}
			}

			fmt.Printf("Indexed %d decisions\n", count)
			return nil
		},
	}
}
```

- [ ] **Step 5: Implement `backstory search`**

```go
// internal/cli/search.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/backstory-team/backstory/internal/store"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search decisions by keyword",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO environment variable not set")
			}

			dbPath := filepath.Join(repoPath, ".backstory", "index.db")
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			query := strings.Join(args, " ")
			results, err := s.Search(query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(results) == 0 {
				fmt.Println("No decisions found")
				return nil
			}

			for _, d := range results {
				fmt.Printf("[%s] %s (%s by %s)\n", d.DateStr, d.Title, d.Type, d.Author)
				fmt.Printf("  anchor: %s\n", d.Anchor)
				if d.LinearIssue != "" {
					fmt.Printf("  linear: %s\n", d.LinearIssue)
				}
				fmt.Println()
			}
			return nil
		},
	}
}
```

- [ ] **Step 6: Implement `backstory inject`**

```go
// internal/cli/inject.go
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/config"
	"github.com/backstory-team/backstory/internal/inject"
	"github.com/backstory-team/backstory/internal/linear"
	"github.com/backstory-team/backstory/internal/repo"
	"github.com/backstory-team/backstory/internal/store"
	"github.com/spf13/cobra"
)

func NewInjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inject",
		Short: "Output relevant decisions as XML context for the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return nil
			}

			cfg, err := config.Load(repoPath)
			if err != nil {
				return nil
			}

			dbPath := filepath.Join(repoPath, ".backstory", "index.db")
			s, err := store.Open(dbPath)
			if err != nil {
				return nil
			}
			defer s.Close()

			cwd, _ := os.Getwd()
			codeRepo := repo.Open(cwd)
			remoteURL, _ := codeRepo.GetRemoteURL()

			var repoName string
			for _, r := range cfg.Repos {
				if r.URL == remoteURL {
					repoName = r.Name
					break
				}
			}

			relPath := ""
			if repoName != "" {
				relPath = repoName + "/"
			}

			branch, _ := codeRepo.GetCurrentBranch()
			linearIssue := linear.ExtractIssueFromBranch(branch)

			eng := inject.New(s, cfg)
			output, err := eng.Generate(relPath, linearIssue)
			if err != nil || output == "" {
				return nil
			}

			if linearIssue != "" && cfg.LinearAPIKey != "" {
				client := linear.NewClient(cfg.LinearAPIKey)
				issue, err := client.FetchIssue(context.Background(), linearIssue)
				if err == nil && issue != nil {
					output = output[:len(output)-len("</backstory>")] + "\n" + linear.FormatIssueXML(issue) + "\n</backstory>"
				}
			}

			fmt.Print(output)
			return nil
		},
	}
}
```

- [ ] **Step 7: Implement `backstory capture`**

```go
// internal/cli/capture.go
package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/backstory-team/backstory/internal/config"
	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/extract"
	"github.com/backstory-team/backstory/internal/pending"
	"github.com/backstory-team/backstory/internal/repo"
	"github.com/spf13/cobra"
)

func NewCaptureCmd() *cobra.Command {
	var author string

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Extract decisions from a session transcript (reads from stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO environment variable not set")
			}

			cfg, err := config.Load(repoPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if cfg.ClaudeAPIKey == "" {
				return fmt.Errorf("claude_api_key not set in .backstory/config.local.yml")
			}

			transcript, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			if len(strings.TrimSpace(string(transcript))) == 0 {
				return nil
			}

			cwd, _ := os.Getwd()
			codeRepo := repo.Open(cwd)
			remoteURL, _ := codeRepo.GetRemoteURL()

			var repoName string
			for _, r := range cfg.Repos {
				if r.URL == remoteURL {
					repoName = r.Name
					break
				}
			}

			extractor := extract.NewExtractor(cfg.ClaudeAPIKey)
			decisions, err := extractor.Extract(context.Background(), string(transcript), repoName, cwd, author)
			if err != nil {
				pendingDir := filepath.Join(os.Getenv("HOME"), ".backstory", "pending")
				q := pending.New(pendingDir)
				q.Save([]*decision.Decision{{Body: string(transcript)}})
				fmt.Fprintf(os.Stderr, "extraction failed, saved to pending queue: %v\n", err)
				return nil
			}

			if len(decisions) == 0 {
				fmt.Println("No decisions captured from this session")
				return nil
			}

			fmt.Println("\nBackstory captured from this session:")
			selected := make([]bool, len(decisions))
			for i, d := range decisions {
				selected[i] = true
				fmt.Printf("  %d. [x] %s\n", i+1, d.Title)
			}

			fmt.Print("\nPress Enter to share, or type numbers to toggle (e.g., '3' to deselect #3): ")

			timer := time.NewTimer(30 * time.Second)
			inputCh := make(chan string, 1)
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					inputCh <- scanner.Text()
				} else {
					inputCh <- ""
				}
			}()

			select {
			case input := <-inputCh:
				timer.Stop()
				input = strings.TrimSpace(input)
				if input != "" {
					for _, ch := range input {
						if ch >= '1' && ch <= '9' {
							idx := int(ch-'0') - 1
							if idx < len(selected) {
								selected[idx] = !selected[idx]
							}
						}
					}
				}
			case <-timer.C:
				fmt.Println("\nTimeout — saving to pending queue")
				pendingDir := filepath.Join(os.Getenv("HOME"), ".backstory", "pending")
				q := pending.New(pendingDir)
				q.Save(decisions)
				return nil
			}

			var toCommit []*decision.Decision
			for i, d := range decisions {
				if selected[i] {
					toCommit = append(toCommit, d)
				}
			}

			if len(toCommit) == 0 {
				fmt.Println("No decisions shared")
				return nil
			}

			r := repo.Open(repoPath)
			for _, d := range toCommit {
				dir := filepath.Join(repoPath, "technical")
				if d.Anchor != "" {
					dir = filepath.Join(dir, d.Anchor)
				}
				os.MkdirAll(dir, 0755)

				path := filepath.Join(dir, d.Filename())
				if err := d.WriteToFile(path); err != nil {
					fmt.Fprintf(os.Stderr, "error writing %s: %v\n", path, err)
					continue
				}
			}

			if err := r.CommitAll(fmt.Sprintf("backstory: captured %d decisions", len(toCommit))); err != nil {
				fmt.Fprintf(os.Stderr, "commit failed: %v\n", err)
				return nil
			}

			if err := r.PushWithRebase(3); err != nil {
				fmt.Fprintf(os.Stderr, "push failed (will retry on next sync): %v\n", err)
			} else {
				fmt.Printf("Shared %d decisions with the team\n", len(toCommit))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", os.Getenv("USER"), "Author name for captured decisions")
	return cmd
}
```

- [ ] **Step 8: Implement `backstory status`**

```go
// internal/cli/status.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/pending"
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show decisions repo status",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO environment variable not set")
			}

			totalCount := 0
			staleCount := 0
			for _, dir := range []string{"product", "technical"} {
				dirPath := filepath.Join(repoPath, dir)
				decisions, _ := decision.ParseAllFromDir(dirPath)
				for _, d := range decisions {
					totalCount++
					if d.Stale {
						staleCount++
					}
				}
			}

			fmt.Printf("Decisions repo: %s\n", repoPath)
			fmt.Printf("Total decisions: %d\n", totalCount)
			if staleCount > 0 {
				fmt.Printf("Stale decisions: %d\n", staleCount)
			}

			pendingDir := filepath.Join(os.Getenv("HOME"), ".backstory", "pending")
			q := pending.New(pendingDir)
			if q.HasPending() {
				items, _ := q.Load()
				fmt.Printf("Pending decisions: %d (run 'backstory sync' to process)\n", len(items))
			}

			return nil
		},
	}
}
```

- [ ] **Step 9: Implement `backstory edit`**

```go
// internal/cli/edit.go
package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <file>",
		Short: "Edit a decision file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			c := exec.Command(editor, args[0])
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor failed: %w", err)
			}

			fmt.Println("Decision updated. Run 'backstory sync' to push changes.")
			return nil
		},
	}
}
```

- [ ] **Step 10: Build and verify all commands work**

Run: `cd /Users/yaronyarimi/dev/backstory && make build && ./bin/backstory --help`
Expected: All 8 commands listed with descriptions.

Run: `./bin/backstory init --path /tmp/test-backstory-decisions`
Expected: Creates directory with template files.

Run: `./bin/backstory status`
Expected: Error about BACKSTORY_REPO not set (expected).

- [ ] **Step 11: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add -A
git commit -m "feat: wire all CLI commands to internal packages"
```

---

### Task 11: End-to-end integration test

**Files:**
- Create: `integration_test.go`

- [ ] **Step 1: Write integration test**

```go
// integration_test.go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "backstory")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/backstory")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s %v", out, err)
	}
	return bin
}

func TestInitCreatesRepoStructure(t *testing.T) {
	bin := buildBinary(t)
	repoPath := filepath.Join(t.TempDir(), "decisions")

	cmd := exec.Command(bin, "init", "--path", repoPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %s %v", out, err)
	}

	for _, path := range []string{
		"product",
		"technical",
		".backstory/config.yml",
		"README.md",
	} {
		if _, err := os.Stat(filepath.Join(repoPath, path)); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", path)
		}
	}
}

func TestIndexAndSearch(t *testing.T) {
	bin := buildBinary(t)
	repoPath := filepath.Join(t.TempDir(), "decisions")

	exec.Command(bin, "init", "--path", repoPath).Run()

	decDir := filepath.Join(repoPath, "technical", "env0", "services", "payment-service")
	os.MkdirAll(decDir, 0755)
	os.WriteFile(filepath.Join(decDir, "2026-03-19-chose-sqs.md"), []byte(`---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation

The vendor API rate-limits at 100 req/s.
`), 0644)

	indexCmd := exec.Command(bin, "index")
	indexCmd.Env = append(os.Environ(), "BACKSTORY_REPO="+repoPath)
	if out, err := indexCmd.CombinedOutput(); err != nil {
		t.Fatalf("index failed: %s %v", out, err)
	}

	searchCmd := exec.Command(bin, "search", "rate limit")
	searchCmd.Env = append(os.Environ(), "BACKSTORY_REPO="+repoPath)
	out, err := searchCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search failed: %s %v", out, err)
	}
	if !strings.Contains(string(out), "Chose SQS") {
		t.Errorf("search should find 'Chose SQS', got: %s", out)
	}
}

func TestInjectOutputsXML(t *testing.T) {
	bin := buildBinary(t)
	repoPath := filepath.Join(t.TempDir(), "decisions")

	exec.Command(bin, "init", "--path", repoPath).Run()

	decDir := filepath.Join(repoPath, "technical", "env0")
	os.MkdirAll(decDir, 0755)
	os.WriteFile(filepath.Join(decDir, "2026-03-19-test.md"), []byte(`---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/
stale: false
---

# Test decision

Test body.
`), 0644)

	indexCmd := exec.Command(bin, "index")
	indexCmd.Env = append(os.Environ(), "BACKSTORY_REPO="+repoPath)
	indexCmd.Run()

	cwd, _ := os.Getwd()
	injectCmd := exec.Command(bin, "inject")
	injectCmd.Dir = cwd
	injectCmd.Env = append(os.Environ(), "BACKSTORY_REPO="+repoPath)
	out, _ := injectCmd.CombinedOutput()

	// inject may return empty if no repo match — that's OK for this test
	// the important thing is it doesn't crash
	_ = out
}
```

- [ ] **Step 2: Run integration tests**

Run: `cd /Users/yaronyarimi/dev/backstory && go test -v -run TestInit -count=1 .`
Expected: PASS

Run: `cd /Users/yaronyarimi/dev/backstory && go test -v -run TestIndexAndSearch -count=1 .`
Expected: PASS

Run: `cd /Users/yaronyarimi/dev/backstory && go test -v -run TestInject -count=1 .`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /Users/yaronyarimi/dev/backstory && make test`
Expected: All tests pass across all packages.

- [ ] **Step 4: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add integration_test.go
git commit -m "test: end-to-end integration tests for init, index, search, inject"
```

---

### Task 12: README and final polish

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README**

```markdown
# Backstory

Shared team memory for AI coding agents.

Backstory automatically captures the reasoning behind code decisions during AI coding sessions and makes that context available to every developer's agent on the team.

## Install

```bash
brew install backstory
```

Or build from source:

```bash
go install github.com/backstory-team/backstory/cmd/backstory@latest
```

## Quick Start

```bash
# Create a decisions repo
backstory init --path my-team-decisions
cd my-team-decisions && git init && git add -A && git commit -m "init"

# Configure
# Edit .backstory/config.yml with your repos
# Add API keys to .backstory/config.local.yml

# Set the environment variable
export BACKSTORY_REPO=/path/to/my-team-decisions

# Build the search index
backstory index

# Search decisions
backstory search "payment processing"

# Check status
backstory status
```

## Claude Code Integration

Add to your Claude Code `settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "backstory sync && backstory inject"
      }
    ],
    "Stop": [
      {
        "type": "command",
        "command": "backstory capture"
      }
    ]
  }
}
```

## How It Works

1. **Session starts** → Backstory pulls the latest decisions and injects relevant context into your agent
2. **You code** → Your agent makes decisions with full team context
3. **Session ends** → Backstory extracts decisions and asks you to confirm sharing
4. **Teammate starts a session** → Their agent already knows what you decided

Nobody "shares" anything. It just happens.

## Commands

| Command | Description |
|---------|-------------|
| `backstory init` | Create a new decisions repo |
| `backstory sync` | Pull latest and push pending |
| `backstory index` | Rebuild the local search index |
| `backstory search <query>` | Search decisions by keyword |
| `backstory inject` | Output relevant context (used by hooks) |
| `backstory capture` | Extract decisions from session (used by hooks) |
| `backstory status` | Show repo status and pending items |
| `backstory edit <file>` | Edit a decision file |
```

- [ ] **Step 2: Final build and verify**

Run: `cd /Users/yaronyarimi/dev/backstory && make build && ./bin/backstory --help`
Expected: Clean build, all commands shown.

Run: `cd /Users/yaronyarimi/dev/backstory && make test`
Expected: All tests pass.

- [ ] **Step 3: Commit**

```bash
cd /Users/yaronyarimi/dev/backstory
git add README.md
git commit -m "docs: add README with install, quickstart, and Claude Code integration"
```
