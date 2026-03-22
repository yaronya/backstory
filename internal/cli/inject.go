package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		Short: "Inject relevant decisions into agent context",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInject()
		},
	}
}

func runInject() error {
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

	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	codeRepo := repo.Open(cwd)

	remoteURL, err := codeRepo.GetRemoteURL()
	if err != nil {
		return nil
	}

	repoName := ""
	for _, r := range cfg.Repos {
		if matchesRemote(remoteURL, r.URL) {
			repoName = r.Name
			break
		}
	}
	if repoName == "" {
		return nil
	}

	repoRoot, err := codeRepo.GetRepoRoot()
	if err != nil {
		return nil
	}

	relPath, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return nil
	}

	anchor := repoName
	if relPath != "." && relPath != "" {
		anchor = repoName + "/" + relPath
	}

	branch, err := codeRepo.GetCurrentBranch()
	if err != nil {
		return nil
	}

	linearIssue := ""
	if cfg.Linear.TeamKey != "" {
		linearIssue = linear.ExtractIssueFromBranch(branch, cfg.Linear.TeamKey)
	}

	engine := inject.New(s, cfg)
	output, err := engine.Generate(anchor, linearIssue)
	if err != nil {
		return nil
	}

	if output == "" {
		return nil
	}

	if linearIssue != "" && cfg.LinearAPIKey != "" {
		client := linear.NewClient(cfg.LinearAPIKey)
		issue, fetchErr := client.FetchIssue(context.Background(), linearIssue)
		if fetchErr == nil && issue != nil {
			output += "\n" + linear.FormatIssueXML(issue)
		}
	}

	fmt.Print(output)
	return nil
}

func matchesRemote(actual, configured string) bool {
	actual = normalizeGitURL(actual)
	configured = normalizeGitURL(configured)
	return actual == configured
}

func normalizeGitURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.ReplaceAll(url, ":", "/")
	return strings.ToLower(url)
}
