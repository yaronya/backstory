package inject_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yaronya/backstory/internal/config"
	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/inject"
	"github.com/yaronya/backstory/internal/store"
)

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func makeDecision(filePath, decType, dateStr, anchor, linearIssue, title, body string, stale bool) *decision.Decision {
	parsed, _ := time.Parse("2006-01-02", dateStr)
	return &decision.Decision{
		Type:        decType,
		Date:        parsed,
		DateStr:     dateStr,
		Author:      "sarah",
		Anchor:      anchor,
		LinearIssue: linearIssue,
		Stale:       stale,
		Title:       title,
		Body:        body,
		FilePath:    filePath,
	}
}

func seedDecisions(t *testing.T, s *store.Store) {
	t.Helper()
	decisions := []*decision.Decision{
		makeDecision("/docs/sqs.md", "technical", "2026-03-19", "env0/services/payment-service/", "ENG-892", "Chose SQS", "The vendor API rate-limits at 100 req/s.", false),
		makeDecision("/docs/bulk.md", "product", "2026-03-22", "payments", "ENG-892", "No bulk ops v1", "No bulk operations in v1.", false),
		makeDecision("/docs/ses.md", "technical", "2026-03-20", "env0/services/notification-service/", "", "SES templates", "Use SES for email templates.", false),
		makeDecision("/docs/stale-sqs.md", "technical", "2026-01-01", "env0/services/payment-service/", "", "Old payment decision", "This is stale.", true),
	}
	for _, d := range decisions {
		if err := s.Upsert(d); err != nil {
			t.Fatalf("failed to seed decision %q: %v", d.FilePath, err)
		}
	}
}

func defaultConfig() *config.Config {
	return &config.Config{
		Inject: config.InjectConfig{
			MaxDecisions: 10,
			MaxTokens:    2000,
		},
	}
}

func TestInjectByAnchor(t *testing.T) {
	s := openTestStore(t)
	seedDecisions(t, s)

	engine := inject.New(s, defaultConfig())
	output, err := engine.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(output, "Chose SQS") {
		t.Errorf("expected output to contain %q", "Chose SQS")
	}
	if strings.Contains(output, "SES templates") {
		t.Errorf("expected output to NOT contain %q", "SES templates")
	}
	if strings.Contains(output, "Old payment decision") {
		t.Errorf("expected output to NOT contain stale decision")
	}
}

func TestInjectByLinearIssue(t *testing.T) {
	s := openTestStore(t)
	seedDecisions(t, s)

	engine := inject.New(s, defaultConfig())
	output, err := engine.Generate("", "ENG-892")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(output, "Chose SQS") {
		t.Errorf("expected output to contain %q", "Chose SQS")
	}
	if !strings.Contains(output, "No bulk ops v1") {
		t.Errorf("expected output to contain %q", "No bulk ops v1")
	}
}

func TestInjectXMLFormat(t *testing.T) {
	s := openTestStore(t)
	seedDecisions(t, s)

	engine := inject.New(s, defaultConfig())
	output, err := engine.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.HasPrefix(output, "<backstory>") {
		t.Errorf("expected output to start with <backstory>, got: %q", output[:min(len(output), 50)])
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "</backstory>") {
		t.Errorf("expected output to end with </backstory>")
	}
	if !strings.Contains(output, `type="technical"`) {
		t.Errorf("expected output to contain type=\"technical\"")
	}
}

func TestInjectRespectsMaxDecisions(t *testing.T) {
	s := openTestStore(t)

	for i := 0; i < 20; i++ {
		d := makeDecision(
			filepath.Join("/docs", "decision-"+string(rune('a'+i))+".md"),
			"technical",
			"2026-03-01",
			"env0/services/payment-service/",
			"",
			"Decision "+string(rune('A'+i)),
			"Body of decision.",
			false,
		)
		if err := s.Upsert(d); err != nil {
			t.Fatalf("Upsert failed: %v", err)
		}
	}

	cfg := &config.Config{
		Inject: config.InjectConfig{
			MaxDecisions: 5,
			MaxTokens:    100000,
		},
	}

	engine := inject.New(s, cfg)
	output, err := engine.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	count := strings.Count(output, "<decision ")
	if count > 5 {
		t.Errorf("expected at most 5 <decision tags, got %d", count)
	}
}

func TestInjectRespectsMaxTokens(t *testing.T) {
	s := openTestStore(t)

	for i := 0; i < 10; i++ {
		body := strings.Repeat("This is a long body of content. ", 20)
		d := makeDecision(
			filepath.Join("/docs", "long-"+string(rune('a'+i))+".md"),
			"technical",
			"2026-03-01",
			"env0/services/payment-service/",
			"",
			"Long Decision "+string(rune('A'+i)),
			body,
			false,
		)
		if err := s.Upsert(d); err != nil {
			t.Fatalf("Upsert failed: %v", err)
		}
	}

	cfg := &config.Config{
		Inject: config.InjectConfig{
			MaxDecisions: 100,
			MaxTokens:    200,
		},
	}

	engine := inject.New(s, cfg)
	output, err := engine.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	count := strings.Count(output, "<decision ")
	if count >= 10 {
		t.Errorf("expected token budget to limit decisions below 10, got %d <decision tags", count)
	}
}

func TestRecencyBoost(t *testing.T) {
	s := openTestStore(t)

	old := makeDecision("/docs/old.md", "technical", "2025-01-01", "env0/services/payment-service/", "", "Old decision", "This is the old one.", false)
	recent := makeDecision("/docs/new.md", "technical", "2026-03-22", "env0/services/payment-service/", "", "Recent decision", "This is the new one.", false)

	if err := s.Upsert(old); err != nil {
		t.Fatalf("Upsert old failed: %v", err)
	}
	if err := s.Upsert(recent); err != nil {
		t.Fatalf("Upsert recent failed: %v", err)
	}

	cfg := &config.Config{
		Inject: config.InjectConfig{
			MaxDecisions: 1,
			MaxTokens:    100000,
		},
	}

	engine := inject.New(s, cfg)
	output, err := engine.Generate("env0/services/payment-service/", "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(output, "Recent decision") {
		t.Errorf("expected newer decision to be selected with MaxDecisions=1")
	}
	if strings.Contains(output, "Old decision") {
		t.Errorf("expected older decision to be excluded with MaxDecisions=1")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
