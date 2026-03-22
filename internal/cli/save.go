package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yaronya/backstory/internal/extract"
	"github.com/yaronya/backstory/internal/repo"
)

func NewSaveCmd() *cobra.Command {
	var author string

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save pre-extracted decisions from JSON (reads stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			if author == "" {
				author = os.Getenv("USER")
			}
			return runSave(repoPath, author)
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "Author name (default: $USER)")

	return cmd
}

func runSave(repoPath, author string) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading JSON from stdin: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("no input provided on stdin")
	}

	decisions, err := extract.ParseExtractionResponse(string(data))
	if err != nil {
		return fmt.Errorf("parsing decisions JSON: %w", err)
	}

	if len(decisions) == 0 {
		fmt.Println("No decisions to save.")
		return nil
	}

	for _, d := range decisions {
		d.Author = author
	}

	for _, d := range decisions {
		dir := filepath.Join(repoPath, d.Type)
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return fmt.Errorf("creating directory %s: %w", dir, mkErr)
		}
		filePath := filepath.Join(dir, d.Filename())
		if writeErr := d.WriteToFile(filePath); writeErr != nil {
			return fmt.Errorf("writing decision: %w", writeErr)
		}
		fmt.Printf("Saved: %s/%s\n", d.Type, d.Filename())
	}

	r := repo.Open(repoPath)
	if commitErr := r.CommitAll("backstory: save decisions"); commitErr != nil {
		return fmt.Errorf("committing: %w", commitErr)
	}

	if pushErr := r.PushWithRebase(3); pushErr != nil {
		fmt.Printf("Warning: push failed (may not have remote): %v\n", pushErr)
	}

	fmt.Printf("Saved %d decision(s).\n", len(decisions))
	return nil
}
