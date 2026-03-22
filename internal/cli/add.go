package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/repo"
	"github.com/spf13/cobra"
)

func NewAddCmd() *cobra.Command {
	var decisionType string
	var anchor string
	var author string
	var linear string
	var title string

	cmd := &cobra.Command{
		Use:   "add [message...]",
		Short: "Manually add a decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := os.Getenv("BACKSTORY_REPO")
			if repoPath == "" {
				return fmt.Errorf("BACKSTORY_REPO not set")
			}
			if decisionType != "technical" && decisionType != "product" {
				return fmt.Errorf("--type must be \"technical\" or \"product\"")
			}
			if anchor == "" {
				return fmt.Errorf("--anchor is required")
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if author == "" {
				author = os.Getenv("USER")
			}

			var body string
			if len(args) > 0 {
				body = strings.Join(args, " ")
			} else {
				scanner := bufio.NewScanner(os.Stdin)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				body = strings.TrimSpace(strings.Join(lines, "\n"))
			}

			return runAdd(repoPath, decisionType, anchor, author, linear, title, body)
		},
	}

	cmd.Flags().StringVar(&decisionType, "type", "", "Decision type: \"technical\" or \"product\" (required)")
	cmd.Flags().StringVar(&anchor, "anchor", "", "Code path or feature area (required)")
	cmd.Flags().StringVar(&author, "author", "", "Author name (default: $USER)")
	cmd.Flags().StringVar(&linear, "linear", "", "Linear issue ID (optional)")
	cmd.Flags().StringVar(&title, "title", "", "Decision title (required)")

	return cmd
}

func runAdd(repoPath, decisionType, anchor, author, linear, title, body string) error {
	now := time.Now()
	d := &decision.Decision{
		Type:        decisionType,
		Date:        now,
		DateStr:     now.Format("2006-01-02"),
		Author:      author,
		Anchor:      anchor,
		LinearIssue: linear,
		Title:       title,
		Body:        body,
	}

	dir := filepath.Join(repoPath, decisionType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, d.Filename())
	if err := d.WriteToFile(filePath); err != nil {
		return fmt.Errorf("writing decision: %w", err)
	}

	fmt.Printf("Written: %s/%s\n", decisionType, d.Filename())

	r := repo.Open(repoPath)

	if err := r.CommitAll(fmt.Sprintf("backstory: add %s", d.Filename())); err != nil {
		return fmt.Errorf("committing: %w", err)
	}

	if err := r.PushWithRebase(3); err != nil {
		fmt.Printf("Warning: push failed (may not have remote): %v\n", err)
	}

	fmt.Println("Decision added and pushed.")
	return nil
}
