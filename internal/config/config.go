package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Repo struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	LocalPath string `yaml:"local_path"`
}

type LinearConfig struct {
	TeamKey string `yaml:"team_key"`
}

type InjectConfig struct {
	MaxDecisions int `yaml:"max_decisions"`
	MaxTokens    int `yaml:"max_tokens"`
}

type ExtractConfig struct {
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
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
	Extract   ExtractConfig   `yaml:"extract"`
	Staleness StalenessConfig `yaml:"staleness"`

	ClaudeAPIKey string `yaml:"claude_api_key"`
	LinearAPIKey string `yaml:"linear_api_key"`
	SlackToken   string `yaml:"slack_token"`
}

func applyDefaults(cfg *Config) {
	if cfg.Inject.MaxDecisions == 0 {
		cfg.Inject.MaxDecisions = 10
	}
	if cfg.Inject.MaxTokens == 0 {
		cfg.Inject.MaxTokens = 2000
	}
	if cfg.Extract.Model == "" {
		cfg.Extract.Model = "claude-haiku-4-5-20251001"
	}
	if cfg.Extract.MaxTokens == 0 {
		cfg.Extract.MaxTokens = 4096
	}
	if cfg.Staleness.ArchiveAfterMonths == 0 {
		cfg.Staleness.ArchiveAfterMonths = 6
	}
	if cfg.Staleness.ChangeThreshold == 0 {
		cfg.Staleness.ChangeThreshold = 0.5
	}
}

func loadYAMLFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

func Load(repoRoot string) (*Config, error) {
	cfg := &Config{}
	backstoryDir := filepath.Join(repoRoot, ".backstory")

	mainPath := filepath.Join(backstoryDir, "config.yml")
	if err := loadYAMLFile(mainPath, cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	applyDefaults(cfg)

	localPath := filepath.Join(backstoryDir, "config.local.yml")
	local := &Config{}
	if err := loadYAMLFile(localPath, local); err != nil && !errors.Is(err, os.ErrNotExist) {
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

	return cfg, nil
}
