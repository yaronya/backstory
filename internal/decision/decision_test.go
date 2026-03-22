package decision_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/backstory-team/backstory/internal/decision"
)

func TestParseFromFile_TechnicalDecision(t *testing.T) {
	d, err := decision.ParseFromFile("../../testdata/decisions/sample-technical.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.Type != "technical" {
		t.Errorf("Type: got %q, want %q", d.Type, "technical")
	}
	expectedDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	if !d.Date.Equal(expectedDate) {
		t.Errorf("Date: got %v, want %v", d.Date, expectedDate)
	}
	if d.Author != "sarah" {
		t.Errorf("Author: got %q, want %q", d.Author, "sarah")
	}
	if d.Anchor != "env0/services/payment-service/" {
		t.Errorf("Anchor: got %q, want %q", d.Anchor, "env0/services/payment-service/")
	}
	if d.LinearIssue != "ENG-892" {
		t.Errorf("LinearIssue: got %q, want %q", d.LinearIssue, "ENG-892")
	}
	if d.Stale != false {
		t.Errorf("Stale: got %v, want false", d.Stale)
	}
	if d.Title != "Chose SQS over direct invocation for vendor API" {
		t.Errorf("Title: got %q, want %q", d.Title, "Chose SQS over direct invocation for vendor API")
	}
	if d.FilePath != "../../testdata/decisions/sample-technical.md" {
		t.Errorf("FilePath: got %q, want %q", d.FilePath, "../../testdata/decisions/sample-technical.md")
	}
}

func TestParseFromFile_ProductDecision(t *testing.T) {
	d, err := decision.ParseFromFile("../../testdata/decisions/sample-product.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.Type != "product" {
		t.Errorf("Type: got %q, want %q", d.Type, "product")
	}
	expectedDate := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	if !d.Date.Equal(expectedDate) {
		t.Errorf("Date: got %v, want %v", d.Date, expectedDate)
	}
	if d.Author != "pm-david" {
		t.Errorf("Author: got %q, want %q", d.Author, "pm-david")
	}
	if d.Anchor != "payments" {
		t.Errorf("Anchor: got %q, want %q", d.Anchor, "payments")
	}
	if d.Title != "No bulk operations in v1" {
		t.Errorf("Title: got %q, want %q", d.Title, "No bulk operations in v1")
	}
}

func TestWriteToFile_RoundTrip(t *testing.T) {
	d := &decision.Decision{
		Type:        "technical",
		Date:        time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		DateStr:     "2026-01-15",
		Author:      "alice",
		Anchor:      "services/auth/",
		LinearIssue: "ENG-100",
		Stale:       false,
		Title:       "Use JWT over session cookies",
		Body:        "Stateless auth simplifies horizontal scaling.\n",
	}

	dir := t.TempDir()
	path := filepath.Join(dir, d.Filename())

	if err := d.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	parsed, err := decision.ParseFromFile(path)
	if err != nil {
		t.Fatalf("ParseFromFile error: %v", err)
	}

	if parsed.Type != d.Type {
		t.Errorf("Type: got %q, want %q", parsed.Type, d.Type)
	}
	if !parsed.Date.Equal(d.Date) {
		t.Errorf("Date: got %v, want %v", parsed.Date, d.Date)
	}
	if parsed.Author != d.Author {
		t.Errorf("Author: got %q, want %q", parsed.Author, d.Author)
	}
	if parsed.Anchor != d.Anchor {
		t.Errorf("Anchor: got %q, want %q", parsed.Anchor, d.Anchor)
	}
	if parsed.LinearIssue != d.LinearIssue {
		t.Errorf("LinearIssue: got %q, want %q", parsed.LinearIssue, d.LinearIssue)
	}
	if parsed.Stale != d.Stale {
		t.Errorf("Stale: got %v, want %v", parsed.Stale, d.Stale)
	}
	if parsed.Title != d.Title {
		t.Errorf("Title: got %q, want %q", parsed.Title, d.Title)
	}
}

func TestParseAllFromDir(t *testing.T) {
	decisions, err := decision.ParseAllFromDir("../../testdata/decisions/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(decisions) < 2 {
		t.Errorf("expected at least 2 decisions, got %d", len(decisions))
	}
}

func TestParseAllFromDir_NonExistentDir(t *testing.T) {
	_, err := decision.ParseAllFromDir("../../testdata/nonexistent/")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestGenerateFilename(t *testing.T) {
	d := &decision.Decision{
		Date:    time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
		DateStr: "2026-03-19",
		Title:   "Chose SQS over direct invocation",
	}

	got := d.Filename()
	want := "2026-03-19-chose-sqs-over-direct-invocation.md"

	if got != want {
		t.Errorf("Filename: got %q, want %q", got, want)
	}
}

func TestGenerateFilename_SpecialChars(t *testing.T) {
	d := &decision.Decision{
		Date:    time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		DateStr: "2026-01-05",
		Title:   "Use JWT (v2) over session/cookies!",
	}

	got := d.Filename()

	if got == "" {
		t.Error("expected non-empty filename")
	}

	info, err := os.Stat(got)
	_ = info
	_ = err
}
