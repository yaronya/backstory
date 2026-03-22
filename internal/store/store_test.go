package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func makeDecision(filePath, anchor, linearIssue, title, body string, stale bool) *decision.Decision {
	return &decision.Decision{
		Type:        "adr",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		DateStr:     "2024-01-15",
		Author:      "alice",
		Anchor:      anchor,
		LinearIssue: linearIssue,
		Stale:       stale,
		Title:       title,
		Body:        body,
		FilePath:    filePath,
	}
}

func TestInsertAndGetByAnchor(t *testing.T) {
	s := newTestStore(t)

	d := makeDecision("/docs/auth.md", "auth/jwt", "", "Use JWT for auth", "We decided to use JWT tokens.", false)
	if err := s.Upsert(d); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	results, err := s.QueryByAnchor("auth/jwt")
	if err != nil {
		t.Fatalf("QueryByAnchor failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Use JWT for auth" {
		t.Errorf("expected title %q, got %q", "Use JWT for auth", results[0].Title)
	}
}

func TestFTS5Search(t *testing.T) {
	s := newTestStore(t)

	d1 := makeDecision("/docs/rate-limit.md", "api/rate-limit", "", "Rate Limit Policy", "We decided to enforce rate limiting on all public endpoints.", false)
	d2 := makeDecision("/docs/caching.md", "api/caching", "", "Caching Strategy", "We use Redis for distributed caching to improve performance.", false)

	if err := s.Upsert(d1); err != nil {
		t.Fatalf("Upsert d1 failed: %v", err)
	}
	if err := s.Upsert(d2); err != nil {
		t.Fatalf("Upsert d2 failed: %v", err)
	}

	results, err := s.Search("rate limit")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].FilePath != "/docs/rate-limit.md" {
		t.Errorf("expected file path %q, got %q", "/docs/rate-limit.md", results[0].FilePath)
	}
}

func TestQueryByLinearIssue(t *testing.T) {
	s := newTestStore(t)

	d := makeDecision("/docs/eng892.md", "billing/subscription", "ENG-892", "Subscription Model Change", "Switching to monthly billing.", false)
	if err := s.Upsert(d); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	results, err := s.QueryByLinearIssue("ENG-892")
	if err != nil {
		t.Fatalf("QueryByLinearIssue failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].LinearIssue != "ENG-892" {
		t.Errorf("expected linear issue %q, got %q", "ENG-892", results[0].LinearIssue)
	}
}

func TestExcludeStale(t *testing.T) {
	s := newTestStore(t)

	d := makeDecision("/docs/old-auth.md", "auth/basic", "", "Use Basic Auth", "We decided to use basic auth (now stale).", true)
	if err := s.Upsert(d); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	results, err := s.QueryByAnchor("auth/basic")
	if err != nil {
		t.Fatalf("QueryByAnchor failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for stale decision, got %d", len(results))
	}
}

func TestExcludeStaleFromLinearQuery(t *testing.T) {
	s := newTestStore(t)

	d := makeDecision("/docs/stale-eng100.md", "infra/old", "ENG-100", "Old Infra Decision", "This is now stale.", true)
	if err := s.Upsert(d); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	results, err := s.QueryByLinearIssue("ENG-100")
	if err != nil {
		t.Fatalf("QueryByLinearIssue failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for stale decision, got %d", len(results))
	}
}

func TestUpsertOverwrites(t *testing.T) {
	s := newTestStore(t)

	d := makeDecision("/docs/db.md", "db/postgres", "", "Use Postgres", "Original body.", false)
	if err := s.Upsert(d); err != nil {
		t.Fatalf("first Upsert failed: %v", err)
	}

	d2 := makeDecision("/docs/db.md", "db/postgres", "", "Use PostgreSQL (updated)", "Updated body with more detail.", false)
	if err := s.Upsert(d2); err != nil {
		t.Fatalf("second Upsert failed: %v", err)
	}

	results, err := s.QueryByAnchor("db/postgres")
	if err != nil {
		t.Fatalf("QueryByAnchor failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result after upsert, got %d", len(results))
	}
	if results[0].Title != "Use PostgreSQL (updated)" {
		t.Errorf("expected updated title, got %q", results[0].Title)
	}
}
