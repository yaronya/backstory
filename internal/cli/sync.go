package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/pending"
	"github.com/backstory-team/backstory/internal/repo"
	"github.com/backstory-team/backstory/internal/store"
	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync decisions repo with remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			return runSync(repoPath)
		},
	}
}

func runSync(repoPath string) error {
	r := repo.Open(repoPath)

	fmt.Println("Pulling latest changes...")
	if err := r.Pull(); err != nil {
		fmt.Printf("Warning: pull failed (may not have remote): %v\n", err)
	}

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

		if len(items) > 0 {
			fmt.Printf("\n%d pending decision(s):\n", len(items))
			for i, d := range items {
				fmt.Printf("  %d. [%s] %s (by %s)\n", i+1, d.Type, d.Title, d.Author)
			}
			fmt.Print("\nCommit and push these decisions? [Y/n] ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer == "" || answer == "y" || answer == "yes" {
				for _, d := range items {
					dir := filepath.Join(repoPath, d.Type)
					if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
						return fmt.Errorf("creating directory %s: %w", dir, mkErr)
					}
					filePath := filepath.Join(dir, d.Filename())
					if writeErr := d.WriteToFile(filePath); writeErr != nil {
						return fmt.Errorf("writing decision: %w", writeErr)
					}
					fmt.Printf("  Written: %s/%s\n", d.Type, d.Filename())
				}

				if commitErr := r.CommitAll("backstory: add pending decisions"); commitErr != nil {
					return fmt.Errorf("committing: %w", commitErr)
				}

				if pushErr := r.PushWithRebase(3); pushErr != nil {
					fmt.Printf("Warning: push failed: %v\n", pushErr)
				}

				if clearErr := q.Clear(); clearErr != nil {
					return fmt.Errorf("clearing pending queue: %w", clearErr)
				}
				fmt.Println("Pending decisions committed and pushed.")
			} else {
				fmt.Println("Skipped pending decisions.")
			}
		}
	}

	fmt.Println("Rebuilding index...")
	if err := rebuildIndex(repoPath); err != nil {
		return fmt.Errorf("rebuilding index: %w", err)
	}

	fmt.Println("Sync complete.")
	return nil
}

func rebuildIndex(repoPath string) error {
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
		if _, statErr := os.Stat(dirPath); os.IsNotExist(statErr) {
			continue
		}
		decisions, parseErr := decision.ParseAllFromDir(dirPath)
		if parseErr != nil {
			return fmt.Errorf("parsing %s: %w", dir, parseErr)
		}
		for _, d := range decisions {
			if upsertErr := s.Upsert(d); upsertErr != nil {
				return fmt.Errorf("indexing %s: %w", d.FilePath, upsertErr)
			}
			count++
		}
	}

	fmt.Printf("Indexed %d decisions.\n", count)
	return nil
}
