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
		Use:   "search [query]",
		Short: "Search decisions by query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			query := strings.Join(args, " ")
			return runSearch(repoPath, query)
		},
	}
}

func runSearch(repoPath, query string) error {
	dbPath := filepath.Join(repoPath, ".backstory", "index.db")
	s, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer s.Close()

	results, err := s.Search(query)
	if err != nil {
		return fmt.Errorf("searching: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No decisions found.")
		return nil
	}

	for _, d := range results {
		fmt.Printf("[%s] %s (%s by %s)\n", d.DateStr, d.Title, d.Type, d.Author)
		if d.Anchor != "" {
			fmt.Printf("  anchor: %s\n", d.Anchor)
		}
		if d.LinearIssue != "" {
			fmt.Printf("  linear: %s\n", d.LinearIssue)
		}
	}

	return nil
}
