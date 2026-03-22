package pending_test

import (
	"os"
	"testing"
	"time"

	"github.com/backstory-team/backstory/internal/decision"
	"github.com/backstory-team/backstory/internal/pending"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	q := pending.New(dir)

	d := &decision.Decision{Title: "use postgres"}
	if err := q.Save([]*decision.Decision{d}); err != nil {
		t.Fatal(err)
	}

	decisions, err := q.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}
	if decisions[0].Title != "use postgres" {
		t.Errorf("expected title %q, got %q", "use postgres", decisions[0].Title)
	}
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	q := pending.New(dir)

	if err := q.Save([]*decision.Decision{{Title: "use redis"}}); err != nil {
		t.Fatal(err)
	}
	if err := q.Clear(); err != nil {
		t.Fatal(err)
	}

	decisions, err := q.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected 0 decisions after clear, got %d", len(decisions))
	}
}

func TestHasPending(t *testing.T) {
	dir := t.TempDir()
	q := pending.New(dir)

	if q.HasPending() {
		t.Error("expected HasPending to be false before any save")
	}

	if err := q.Save([]*decision.Decision{{Title: "use kafka"}}); err != nil {
		t.Fatal(err)
	}

	if !q.HasPending() {
		t.Error("expected HasPending to be true after save")
	}
}

func TestRoundTripPreservesAllFields(t *testing.T) {
	dir := t.TempDir()
	q := pending.New(dir)

	date := time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC)
	original := &decision.Decision{
		Type:        "architecture",
		Date:        date,
		DateStr:     "2024-05-15",
		Author:      "alice",
		Anchor:      "some-anchor",
		LinearIssue: "ENG-999",
		Stale:       true,
		Title:       "use event sourcing",
		Body:        "we decided to use event sourcing because...",
		FilePath:    "/path/to/file.md",
	}

	if err := q.Save([]*decision.Decision{original}); err != nil {
		t.Fatal(err)
	}

	decisions, err := q.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	got := decisions[0]

	if got.Type != original.Type {
		t.Errorf("Type: expected %q, got %q", original.Type, got.Type)
	}
	if !got.Date.Equal(original.Date) {
		t.Errorf("Date: expected %v, got %v", original.Date, got.Date)
	}
	if got.DateStr != original.DateStr {
		t.Errorf("DateStr: expected %q, got %q", original.DateStr, got.DateStr)
	}
	if got.Author != original.Author {
		t.Errorf("Author: expected %q, got %q", original.Author, got.Author)
	}
	if got.Anchor != original.Anchor {
		t.Errorf("Anchor: expected %q, got %q", original.Anchor, got.Anchor)
	}
	if got.LinearIssue != original.LinearIssue {
		t.Errorf("LinearIssue: expected %q, got %q", original.LinearIssue, got.LinearIssue)
	}
	if got.Stale != original.Stale {
		t.Errorf("Stale: expected %v, got %v", original.Stale, got.Stale)
	}
	if got.Title != original.Title {
		t.Errorf("Title: expected %q, got %q", original.Title, got.Title)
	}
	if got.Body != original.Body {
		t.Errorf("Body: expected %q, got %q", original.Body, got.Body)
	}
	if got.FilePath != original.FilePath {
		t.Errorf("FilePath: expected %q, got %q", original.FilePath, got.FilePath)
	}

	_ = os.Remove(dir)
}
