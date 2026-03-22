package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaronya/backstory/internal/config"
	"github.com/yaronya/backstory/internal/store"
	"github.com/spf13/cobra"
)

func NewStaleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stale",
		Short: "Mark decisions stale when their anchored path no longer exists",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			return runStale(repoPath)
		},
	}
}

func runStale(repoPath string) error {
	cfg, err := config.Load(repoPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	dbPath := filepath.Join(repoPath, ".backstory", "index.db")
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer s.Close()

	anchors, err := s.GetAnchors()
	if err != nil {
		return fmt.Errorf("getting anchors: %w", err)
	}

	markedStale := 0
	for _, anchor := range anchors {
		if anchorPathExists(anchor, cfg.Repos) {
			continue
		}
		if err := s.MarkStale(anchor); err != nil {
			return fmt.Errorf("marking stale for anchor %q: %w", anchor, err)
		}
		fmt.Printf("Marked stale: %s (path not found in any configured repo)\n", anchor)
		markedStale++
	}

	if markedStale == 0 {
		fmt.Println("No stale decisions found.")
	} else {
		fmt.Printf("Marked %d anchor(s) as stale.\n", markedStale)
	}
	return nil
}

func anchorPathExists(anchor string, repos []config.Repo) bool {
	repoName, subPath, _ := strings.Cut(anchor, "/")
	for _, r := range repos {
		if r.Name != repoName || r.LocalPath == "" {
			continue
		}
		checkPath := r.LocalPath
		if subPath != "" {
			checkPath = filepath.Join(r.LocalPath, subPath)
		}
		if _, err := os.Stat(checkPath); err == nil {
			return true
		}
	}
	return false
}
