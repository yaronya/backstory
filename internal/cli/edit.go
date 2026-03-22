package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yaronya/backstory/internal/repo"
	"github.com/spf13/cobra"
)

func NewEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit [file]",
		Short: "Edit an existing decision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			return runEdit(repoPath, args[0])
		},
	}
}

func runEdit(repoPath, filePath string) error {
	fullPath := filePath
	if !filepath.IsAbs(filePath) {
		fullPath = filepath.Join(repoPath, filePath)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", fullPath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, fullPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	r := repo.Open(repoPath)

	relPath, err := filepath.Rel(repoPath, fullPath)
	if err != nil {
		relPath = fullPath
	}

	if commitErr := r.CommitAndPush(relPath, fmt.Sprintf("backstory: edit %s", relPath)); commitErr != nil {
		if commitAllErr := r.CommitAll(fmt.Sprintf("backstory: edit %s", relPath)); commitAllErr != nil {
			return fmt.Errorf("committing: %w", commitAllErr)
		}
		fmt.Println("Committed locally (push may require manual sync).")
		return nil
	}

	fmt.Println("Decision updated and pushed.")
	return nil
}
