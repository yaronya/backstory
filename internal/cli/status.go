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
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			return runStatus(repoPath)
		},
	}
}

func runStatus(repoPath string) error {
	fmt.Printf("Decisions repo: %s\n", repoPath)

	total := 0
	stale := 0

	for _, dir := range []string{"product", "technical"} {
		dirPath := filepath.Join(repoPath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}
		decisions, err := decision.ParseAllFromDir(dirPath)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", dir, err)
		}
		for _, d := range decisions {
			total++
			if d.Stale {
				stale++
			}
		}
	}

	fmt.Printf("Total decisions: %d\n", total)
	fmt.Printf("Stale decisions: %d\n", stale)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}
	q := pending.New(filepath.Join(homeDir, ".backstory", "pending"))
	if q.HasPending() {
		items, loadErr := q.Load()
		if loadErr != nil {
			return fmt.Errorf("loading pending queue: %w", loadErr)
		}
		fmt.Printf("Pending decisions: %d\n", len(items))
	} else {
		fmt.Println("Pending decisions: 0")
	}

	return nil
}
