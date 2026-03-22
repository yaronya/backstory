package cli

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/repo"
	"github.com/backstory-team/backstory/internal/template"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var pathFlag string
	var connectFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a decisions repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if connectFlag != "" {
				return runInitConnect(connectFlag, pathFlag)
			}
			return runInitNew(pathFlag)
		},
	}

	cmd.Flags().StringVar(&pathFlag, "path", "backstory-decisions", "Directory name for the decisions repo")
	cmd.Flags().StringVar(&connectFlag, "connect", "", "Clone an existing decisions repo from URL")

	return cmd
}

func runInitConnect(url, dest string) error {
	fmt.Printf("Cloning %s into %s...\n", url, dest)
	_, err := repo.Clone(url, dest)
	if err != nil {
		return fmt.Errorf("cloning repo: %w", err)
	}
	abs, _ := filepath.Abs(dest)
	fmt.Printf("Decisions repo cloned to %s\n", abs)
	printHookInstructions(abs)
	return nil
}

func runInitNew(dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("directory %s already exists", dest)
	}

	fmt.Printf("Creating decisions repo at %s...\n", dest)

	err := fs.WalkDir(template.RepoTemplate, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel("files", path)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, readErr := fs.ReadFile(template.RepoTemplate, path)
		if readErr != nil {
			return readErr
		}
		if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
			return mkErr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return fmt.Errorf("scaffolding repo: %w", err)
	}

	gitInit := exec.Command("git", "init", dest)
	if out, gitErr := gitInit.CombinedOutput(); gitErr != nil {
		return fmt.Errorf("git init: %w: %s", gitErr, out)
	}

	r := repo.Open(dest)
	if err := r.CommitAll("Initial backstory decisions repo"); err != nil {
		return fmt.Errorf("creating initial commit: %w", err)
	}

	abs, _ := filepath.Abs(dest)
	fmt.Printf("Decisions repo created at %s\n", abs)
	printHookInstructions(abs)
	return nil
}

func printHookInstructions(repoPath string) {
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Set BACKSTORY_REPO=%s in your shell profile\n", repoPath)
	fmt.Println("  2. Configure your Claude Code hook to run: backstory inject")
	fmt.Println("     Add a PreToolUse hook in your Claude Code settings with:")
	fmt.Println("       command: backstory inject")
	fmt.Println("       event: on_tool_call")
	fmt.Println("  3. Edit .backstory/config.yml with your team settings")
	fmt.Println("  4. Add API keys to .backstory/config.local.yml")
}
