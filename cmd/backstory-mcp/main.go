package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/repo"
	"github.com/yaronya/backstory/internal/store"
)

func main() {
	s := server.NewMCPServer(
		"backstory",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(mcp.NewTool("backstory_search",
		mcp.WithDescription("Search team decisions by keyword"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
	), handleSearch)

	s.AddTool(mcp.NewTool("backstory_add",
		mcp.WithDescription("Add a new team decision"),
		mcp.WithString("type", mcp.Required(), mcp.Description("Decision type: technical or product")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Decision title")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Decision body/reasoning")),
		mcp.WithString("anchor", mcp.Required(), mcp.Description("Code path or feature area")),
		mcp.WithString("linear_issue", mcp.Description("Linear issue ID (optional)")),
	), handleAdd)

	s.AddTool(mcp.NewTool("backstory_status",
		mcp.WithDescription("Show decisions repo status - total decisions, stale count"),
	), handleStatus)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "backstory-mcp: %v\n", err)
		os.Exit(1)
	}
}

func repoPath() (string, error) {
	p := os.Getenv("BACKSTORY_REPO")
	if p == "" {
		return "", fmt.Errorf("BACKSTORY_REPO not set")
	}
	return p, nil
}

func openStore(repoPath string) (*store.Store, error) {
	dbPath := filepath.Join(repoPath, ".backstory", "index.db")
	return store.Open(dbPath)
}

func handleSearch(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	rp, err := repoPath()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	s, err := openStore(rp)
	if err != nil {
		return mcp.NewToolResultError("failed to open index: " + err.Error()), nil
	}
	defer s.Close()

	results, err := s.Search(query)
	if err != nil {
		return mcp.NewToolResultError("search failed: " + err.Error()), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No decisions found"), nil
	}

	var output string
	for _, d := range results {
		output += fmt.Sprintf("[%s] %s (%s by %s)\n  anchor: %s\n", d.DateStr, d.Title, d.Type, d.Author, d.Anchor)
		if d.LinearIssue != "" {
			output += fmt.Sprintf("  linear: %s\n", d.LinearIssue)
		}
		output += "\n"
	}
	return mcp.NewToolResultText(output), nil
}

func handleAdd(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	decType := request.GetString("type", "")
	title := request.GetString("title", "")
	body := request.GetString("body", "")
	anchor := request.GetString("anchor", "")
	linearIssue := request.GetString("linear_issue", "")

	if decType != decision.TypeTechnical && decType != decision.TypeProduct {
		return mcp.NewToolResultError("type must be 'technical' or 'product'"), nil
	}
	if title == "" {
		return mcp.NewToolResultError("title is required"), nil
	}
	if body == "" {
		return mcp.NewToolResultError("body is required"), nil
	}
	if anchor == "" {
		return mcp.NewToolResultError("anchor is required"), nil
	}

	rp, err := repoPath()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	now := time.Now()
	author := os.Getenv("USER")
	if author == "" {
		author = "unknown"
	}

	d := &decision.Decision{
		Type:        decType,
		DateStr:     now.Format("2006-01-02"),
		Date:        now,
		Author:      author,
		Anchor:      anchor,
		LinearIssue: linearIssue,
		Title:       title,
		Body:        body,
	}

	dir := filepath.Join(rp, decType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mcp.NewToolResultError("failed to create directory: " + err.Error()), nil
	}

	filePath := filepath.Join(dir, d.Filename())
	if err := d.WriteToFile(filePath); err != nil {
		return mcp.NewToolResultError("failed to write decision: " + err.Error()), nil
	}

	r := repo.Open(rp)
	commitMsg := fmt.Sprintf("decision: add %s", d.Filename())
	if err := r.CommitAll(commitMsg); err != nil {
		return mcp.NewToolResultError("failed to commit: " + err.Error()), nil
	}

	if err := r.PushWithRebase(3); err != nil {
		return mcp.NewToolResultError("failed to push: " + err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Decision added: %s", filePath)), nil
}

func handleStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rp, err := repoPath()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var total, stale int
	for _, decType := range []string{decision.TypeTechnical, decision.TypeProduct} {
		dir := filepath.Join(rp, decType)
		decisions, err := decision.ParseAllFromDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return mcp.NewToolResultError("failed to read decisions: " + err.Error()), nil
		}
		for _, d := range decisions {
			total++
			if d.Stale {
				stale++
			}
		}
	}

	output := fmt.Sprintf("repo: %s\ntotal decisions: %d\nstale: %d\n", rp, total, stale)
	return mcp.NewToolResultText(output), nil
}
