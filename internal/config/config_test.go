package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ReadsTeamSettings(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	if err := os.MkdirAll(backstoryDir, 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `team: my-team
repos:
  - name: env0
    url: git@github.com:env0/env0.git
linear:
  team_key: ENG
inject:
  max_decisions: 15
  max_tokens: 3000
staleness:
  archive_after_months: 12
  change_threshold: 0.7
`
	if err := os.WriteFile(filepath.Join(backstoryDir, "config.yml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Team != "my-team" {
		t.Errorf("Team = %q, want %q", cfg.Team, "my-team")
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("len(Repos) = %d, want 1", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != "env0" {
		t.Errorf("Repos[0].Name = %q, want %q", cfg.Repos[0].Name, "env0")
	}
	if cfg.Repos[0].URL != "git@github.com:env0/env0.git" {
		t.Errorf("Repos[0].URL = %q, want %q", cfg.Repos[0].URL, "git@github.com:env0/env0.git")
	}
	if cfg.Linear.TeamKey != "ENG" {
		t.Errorf("Linear.TeamKey = %q, want %q", cfg.Linear.TeamKey, "ENG")
	}
	if cfg.Inject.MaxDecisions != 15 {
		t.Errorf("Inject.MaxDecisions = %d, want 15", cfg.Inject.MaxDecisions)
	}
	if cfg.Inject.MaxTokens != 3000 {
		t.Errorf("Inject.MaxTokens = %d, want 3000", cfg.Inject.MaxTokens)
	}
	if cfg.Staleness.ArchiveAfterMonths != 12 {
		t.Errorf("Staleness.ArchiveAfterMonths = %d, want 12", cfg.Staleness.ArchiveAfterMonths)
	}
	if cfg.Staleness.ChangeThreshold != 0.7 {
		t.Errorf("Staleness.ChangeThreshold = %f, want 0.7", cfg.Staleness.ChangeThreshold)
	}
}

func TestLoadConfig_MergesLocalOverrides(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	if err := os.MkdirAll(backstoryDir, 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `team: my-team
repos:
  - name: env0
    url: git@github.com:env0/env0.git
linear:
  team_key: ENG
`
	if err := os.WriteFile(filepath.Join(backstoryDir, "config.yml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	localYAML := `claude_api_key: sk-ant-abc123
linear_api_key: lin_api_xyz789
slack_token: xoxb-token
`
	if err := os.WriteFile(filepath.Join(backstoryDir, "config.local.yml"), []byte(localYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ClaudeAPIKey != "sk-ant-abc123" {
		t.Errorf("ClaudeAPIKey = %q, want %q", cfg.ClaudeAPIKey, "sk-ant-abc123")
	}
	if cfg.LinearAPIKey != "lin_api_xyz789" {
		t.Errorf("LinearAPIKey = %q, want %q", cfg.LinearAPIKey, "lin_api_xyz789")
	}
	if cfg.SlackToken != "xoxb-token" {
		t.Errorf("SlackToken = %q, want %q", cfg.SlackToken, "xoxb-token")
	}
	if cfg.Team != "my-team" {
		t.Errorf("Team = %q, want %q", cfg.Team, "my-team")
	}
}

func TestLoadConfig_DefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	backstoryDir := filepath.Join(dir, ".backstory")
	if err := os.MkdirAll(backstoryDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Inject.MaxDecisions != 10 {
		t.Errorf("Inject.MaxDecisions = %d, want 10", cfg.Inject.MaxDecisions)
	}
	if cfg.Inject.MaxTokens != 2000 {
		t.Errorf("Inject.MaxTokens = %d, want 2000", cfg.Inject.MaxTokens)
	}
	if cfg.Staleness.ArchiveAfterMonths != 6 {
		t.Errorf("Staleness.ArchiveAfterMonths = %d, want 6", cfg.Staleness.ArchiveAfterMonths)
	}
	if cfg.Staleness.ChangeThreshold != 0.5 {
		t.Errorf("Staleness.ChangeThreshold = %f, want 0.5", cfg.Staleness.ChangeThreshold)
	}
}
