package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaronya/backstory/internal/config"
	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/extract"
	"github.com/yaronya/backstory/internal/pending"
	"github.com/yaronya/backstory/internal/repo"
	"github.com/spf13/cobra"
)

func NewCaptureCmd() *cobra.Command {
	var author string

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a new decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			if author == "" {
				author = os.Getenv("USER")
			}
			return runCapture(repoPath, author)
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "Author name (default: $USER)")

	return cmd
}

func runCapture(repoPath, author string) error {
	cfg, err := config.Load(repoPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.ClaudeAPIKey == "" {
		return fmt.Errorf("claude_api_key not set in config")
	}

	transcriptBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading transcript from stdin: %w", err)
	}
	transcript := transcriptBytes

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	codeRepo := repo.Open(cwd)
	remoteURL, _ := codeRepo.GetRemoteURL()

	repoName := ""
	for _, r := range cfg.Repos {
		if repo.MatchesRemote(remoteURL, r.URL) {
			repoName = r.Name
			break
		}
	}
	if repoName == "" {
		repoName = filepath.Base(cwd)
	}

	repoRoot, rootErr := codeRepo.GetRepoRoot()
	workDir := cwd
	if rootErr == nil {
		if rel, relErr := filepath.Rel(repoRoot, cwd); relErr == nil {
			workDir = rel
		}
	}

	fmt.Println("Extracting decisions from transcript...")
	extractor := extract.NewExtractor(cfg.ClaudeAPIKey, cfg.Extract.Model, cfg.Extract.MaxTokens)
	decisions, err := extractor.Extract(context.Background(), string(transcript), repoName, workDir, author)
	if err != nil {
		fmt.Printf("Extraction failed: %v\n", err)
		fmt.Println("Saving raw transcript to pending queue...")
		return saveToPending(repoPath, string(transcript), author)
	}

	if len(decisions) == 0 {
		fmt.Println("No decisions found in transcript.")
		return nil
	}

	fmt.Printf("\nFound %d decision(s):\n", len(decisions))
	confirmed := make([]bool, len(decisions))
	for i, d := range decisions {
		confirmed[i] = true
		fmt.Printf("  [x] %d. [%s] %s\n", i+1, d.Type, d.Title)
	}

	var input string
	tty, ttyErr := openTTY()
	if ttyErr == nil {
		defer tty.Close()
		fmt.Print("\nToggle numbers to deselect, or press Enter to confirm: ")

		inputCh := make(chan string, 1)
		go func() {
			reader := bufio.NewReader(tty)
			line, _ := reader.ReadString('\n')
			inputCh <- strings.TrimSpace(line)
		}()

		select {
		case input = <-inputCh:
		case <-time.After(30 * time.Second):
			fmt.Println("\nTimeout waiting for input. Saving to pending queue...")
			homeDir, _ := os.UserHomeDir()
			q := pending.New(filepath.Join(homeDir, ".backstory", "pending"))
			return q.Save(decisions)
		}
	}

	if input != "" {
		for _, token := range strings.Fields(input) {
			var idx int
			if _, parseErr := fmt.Sscanf(token, "%d", &idx); parseErr == nil {
				if idx >= 1 && idx <= len(confirmed) {
					confirmed[idx-1] = !confirmed[idx-1]
				}
			}
		}
	}

	var selected []*decision.Decision
	for i, d := range decisions {
		if confirmed[i] {
			selected = append(selected, d)
		}
	}

	if len(selected) == 0 {
		fmt.Println("No decisions selected.")
		return nil
	}

	r := repo.Open(repoPath)

	for _, d := range selected {
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

	if commitErr := r.CommitAll("backstory: capture decisions"); commitErr != nil {
		return fmt.Errorf("committing: %w", commitErr)
	}

	if pushErr := r.PushWithRebase(3); pushErr != nil {
		fmt.Printf("Warning: push failed (may not have remote): %v\n", pushErr)
	}

	fmt.Printf("Captured %d decision(s).\n", len(selected))
	return nil
}

func saveToPending(repoPath, transcript, author string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}
	q := pending.New(filepath.Join(homeDir, ".backstory", "pending"))

	d := &decision.Decision{
		Title:   "Raw transcript (extraction failed)",
		Body:    transcript,
		Type:    decision.TypeTechnical,
		Author:  author,
		DateStr: time.Now().Format("2006-01-02"),
		Date:    time.Now(),
	}
	return q.Save([]*decision.Decision{d})
}
