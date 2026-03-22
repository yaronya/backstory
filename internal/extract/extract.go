package extract

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yaronya/backstory/internal/decision"
)

type extractedDecision struct {
	Title                  string `json:"title"`
	Body                   string `json:"body"`
	Anchor                 string `json:"anchor"`
	Type                   string `json:"type"`
	AlternativesConsidered string `json:"alternatives_considered"`
}

type Extractor struct {
	client    *anthropic.Client
	model     string
	maxTokens int
}

func BuildExtractionPrompt(transcript, repoName, workDir string) string {
	return fmt.Sprintf(`You are analyzing a coding session transcript to extract meaningful architectural and technical decisions.

Repository: %s
Working directory: %s

Extract ONLY decisions that represent genuine choices where alternatives existed and a deliberate choice was made. Examples include:
- Choosing a specific technology, library, or pattern over alternatives
- Making architectural trade-offs
- Selecting an approach after weighing options

Do NOT extract:
- Bug fixes or debugging sessions
- Variable or function renames
- Code formatting or style changes
- Routine implementation of stated requirements with no alternatives considered

Return a JSON array of decision objects. Each object must have these fields:
- title: short descriptive title of the decision
- body: explanation of the decision and reasoning
- anchor: the most relevant file path or identifier mentioned in context (empty string if none)
- type: either "technical" or "product"
- alternatives_considered: description of alternatives that were rejected (empty string if none mentioned)

Return an empty array [] if no meaningful decisions are found.

Transcript:
%s`, repoName, workDir, transcript)
}

func ParseExtractionResponse(response string) ([]*decision.Decision, error) {
	var extracted []extractedDecision
	if err := json.Unmarshal([]byte(response), &extracted); err != nil {
		return nil, fmt.Errorf("parsing extraction response: %w", err)
	}

	now := time.Now()
	decisions := make([]*decision.Decision, 0, len(extracted))
	for _, e := range extracted {
		body := e.Body
		if e.AlternativesConsidered != "" {
			body = body + "\n\nAlternatives considered: " + e.AlternativesConsidered
		}
		decisions = append(decisions, &decision.Decision{
			Title:   e.Title,
			Body:    strings.TrimSpace(body),
			Anchor:  e.Anchor,
			Type:    e.Type,
			Date:    now,
			DateStr: now.Format("2006-01-02"),
		})
	}
	return decisions, nil
}

func NewExtractor(apiKey, model string, maxTokens int) *Extractor {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Extractor{client: &client, model: model, maxTokens: maxTokens}
}

func (e *Extractor) Extract(ctx context.Context, transcript, repoName, workDir, author string) ([]*decision.Decision, error) {
	prompt := BuildExtractionPrompt(transcript, repoName, workDir)

	msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(e.model),
		MaxTokens: int64(e.maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling Claude API: %w", err)
	}

	var responseText string
	for _, block := range msg.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}
	if responseText == "" {
		return nil, fmt.Errorf("no text in Claude API response")
	}

	text := responseText

	decisions, err := ParseExtractionResponse(text)
	if err != nil {
		return nil, err
	}

	for _, d := range decisions {
		d.Author = author
	}
	return decisions, nil
}
