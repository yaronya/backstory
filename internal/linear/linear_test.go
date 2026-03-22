package linear_test

import (
	"strings"
	"testing"

	"github.com/yaronya/backstory/internal/linear"
)

func TestParseIssueResponse(t *testing.T) {
	jsonResponse := `{
		"data": {
			"issues": {
				"nodes": [
					{
						"id": "abc-123-uuid",
						"identifier": "ENG-1234",
						"title": "Add webhook retry logic",
						"description": "We need exponential backoff for webhook retries.",
						"comments": {
							"nodes": [
								{
									"body": "Looks good to me",
									"user": { "name": "Alice" }
								},
								{
									"body": "Please add tests",
									"user": { "name": "Bob" }
								}
							]
						}
					}
				]
			}
		}
	}`

	issue, err := linear.ParseIssueResponse([]byte(jsonResponse))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue == nil {
		t.Fatal("expected non-nil issue")
	}
	if issue.ID != "abc-123-uuid" {
		t.Errorf("unexpected ID: %q", issue.ID)
	}
	if issue.Identifier != "ENG-1234" {
		t.Errorf("unexpected identifier: %q", issue.Identifier)
	}
	if issue.Title != "Add webhook retry logic" {
		t.Errorf("unexpected title: %q", issue.Title)
	}
	if issue.Description != "We need exponential backoff for webhook retries." {
		t.Errorf("unexpected description: %q", issue.Description)
	}
	if len(issue.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(issue.Comments))
	}
	if issue.Comments[0].Body != "Looks good to me" {
		t.Errorf("unexpected comment body: %q", issue.Comments[0].Body)
	}
	if issue.Comments[0].UserName != "Alice" {
		t.Errorf("unexpected comment user: %q", issue.Comments[0].UserName)
	}
	if issue.Comments[1].Body != "Please add tests" {
		t.Errorf("unexpected comment body: %q", issue.Comments[1].Body)
	}
	if issue.Comments[1].UserName != "Bob" {
		t.Errorf("unexpected comment user: %q", issue.Comments[1].UserName)
	}
}

func TestParseIssueResponse_Empty(t *testing.T) {
	jsonResponse := `{"data": {"issues": {"nodes": []}}}`

	issue, err := linear.ParseIssueResponse([]byte(jsonResponse))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue != nil {
		t.Errorf("expected nil issue for empty nodes, got %+v", issue)
	}
}

func TestFormatIssueXML(t *testing.T) {
	issue := &linear.Issue{
		ID:          "uuid-001",
		Identifier:  "ENG-1234",
		Title:       "Add webhook retry logic",
		Description: "Implement exponential backoff.",
		Comments: []linear.Comment{
			{Body: "LGTM", UserName: "Alice"},
			{Body: "Ship it", UserName: "Bob"},
		},
	}

	xml := linear.FormatIssueXML(issue)

	if xml == "" {
		t.Fatal("expected non-empty XML output")
	}
	if !strings.Contains(xml, `<linear issue="ENG-1234">`) {
		t.Errorf("missing opening tag with identifier: %q", xml)
	}
	if !strings.Contains(xml, "Add webhook retry logic") {
		t.Errorf("missing title: %q", xml)
	}
	if !strings.Contains(xml, "Implement exponential backoff.") {
		t.Errorf("missing description: %q", xml)
	}
	if !strings.Contains(xml, "LGTM") {
		t.Errorf("missing comment body: %q", xml)
	}
	if !strings.Contains(xml, "Alice") {
		t.Errorf("missing comment user: %q", xml)
	}
	if !strings.Contains(xml, "Ship it") {
		t.Errorf("missing second comment: %q", xml)
	}
	if !strings.Contains(xml, "</linear>") {
		t.Errorf("missing closing tag: %q", xml)
	}
}

func TestExtractIssueFromBranch(t *testing.T) {
	tests := []struct {
		branch   string
		teamKey  string
		expected string
	}{
		{"eng-1234-add-webhook-retry", "ENG", "ENG-1234"},
		{"ENG-567-fix-bug", "ENG", "ENG-567"},
		{"main", "ENG", ""},
		{"feature/eng-89-something", "ENG", "ENG-89"},
		{"plat-42-new-feature", "PLAT", "PLAT-42"},
		{"fe-100-component", "FE", "FE-100"},
	}

	for _, tc := range tests {
		result := linear.ExtractIssueFromBranch(tc.branch, tc.teamKey)
		if result != tc.expected {
			t.Errorf("ExtractIssueFromBranch(%q, %q) = %q, want %q", tc.branch, tc.teamKey, result, tc.expected)
		}
	}
}
