package extract_test

import (
	"strings"
	"testing"

	"github.com/yaronya/backstory/internal/extract"
)

func TestParseExtractionResponse(t *testing.T) {
	input := `[
		{
			"title": "Use SQS for backpressure",
			"body": "Decided to use SQS instead of direct invocation to handle backpressure naturally.",
			"anchor": "services/payment-service/handler.ts",
			"type": "technical",
			"alternatives_considered": "Direct invocation with client-side rate limiting"
		},
		{
			"title": "Exponential backoff for Stripe retries",
			"body": "Added exponential backoff for Stripe webhook retries.",
			"anchor": "services/payment-service/handler.ts",
			"type": "technical",
			"alternatives_considered": ""
		}
	]`

	decisions, err := extract.ParseExtractionResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(decisions))
	}
	if decisions[0].Title != "Use SQS for backpressure" {
		t.Errorf("unexpected title: %q", decisions[0].Title)
	}
	if decisions[1].Title != "Exponential backoff for Stripe retries" {
		t.Errorf("unexpected title: %q", decisions[1].Title)
	}
	if decisions[0].Anchor != "services/payment-service/handler.ts" {
		t.Errorf("unexpected anchor: %q", decisions[0].Anchor)
	}
	if decisions[1].Anchor != "services/payment-service/handler.ts" {
		t.Errorf("unexpected anchor: %q", decisions[1].Anchor)
	}
	if !strings.Contains(decisions[0].Body, "Direct invocation with client-side rate limiting") {
		t.Errorf("alternatives not appended to body: %q", decisions[0].Body)
	}
}

func TestParseExtractionResponse_EmptyArray(t *testing.T) {
	decisions, err := extract.ParseExtractionResponse("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected 0 decisions, got %d", len(decisions))
	}
}

func TestParseExtractionResponse_InvalidJSON(t *testing.T) {
	_, err := extract.ParseExtractionResponse("not valid json {{{")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestBuildExtractionPrompt(t *testing.T) {
	transcript := "I decided to use Redis instead of Memcached for caching."
	repoName := "my-service"
	workDir := "/workspace/my-service"

	prompt := extract.BuildExtractionPrompt(transcript, repoName, workDir)

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !strings.Contains(prompt, transcript) {
		t.Error("prompt should contain the transcript")
	}
	if !strings.Contains(prompt, repoName) {
		t.Error("prompt should contain the repo name")
	}
	if !strings.Contains(prompt, workDir) {
		t.Error("prompt should contain the work dir")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("prompt should mention JSON")
	}
	if !strings.Contains(prompt, "decision") {
		t.Error("prompt should mention decision")
	}
}
