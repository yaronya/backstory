package inject

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/yaronya/backstory/internal/config"
	"github.com/yaronya/backstory/internal/decision"
	"github.com/yaronya/backstory/internal/store"
)

type Engine struct {
	store *store.Store
	cfg   *config.Config
}

func New(s *store.Store, cfg *config.Config) *Engine {
	return &Engine{store: s, cfg: cfg}
}

func (e *Engine) Generate(anchor, linearIssue string) (string, error) {
	seen := map[string]*decision.Decision{}

	if anchor != "" {
		results, err := e.store.QueryByAnchor(anchor)
		if err != nil {
			return "", fmt.Errorf("querying by anchor: %w", err)
		}
		for _, d := range results {
			seen[d.FilePath] = d
		}
	}

	if linearIssue != "" {
		results, err := e.store.QueryByLinearIssue(linearIssue)
		if err != nil {
			return "", fmt.Errorf("querying by linear issue: %w", err)
		}
		for _, d := range results {
			seen[d.FilePath] = d
		}
	}

	if len(seen) == 0 {
		return "", nil
	}

	type scored struct {
		d     *decision.Decision
		score float64
	}

	now := time.Now()
	candidates := make([]scored, 0, len(seen))
	for _, d := range seen {
		days := now.Sub(d.Date).Hours() / 24.0
		score := math.Exp(-0.693 * days / 30.0)
		candidates = append(candidates, scored{d: d, score: score})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	max := e.cfg.Inject.MaxDecisions
	if max > 0 && len(candidates) > max {
		candidates = candidates[:max]
	}

	tokenBudget := e.cfg.Inject.MaxTokens
	var selected []*decision.Decision
	usedTokens := 0
	for _, c := range candidates {
		content := c.d.Title + ". " + c.d.Body
		tokens := len(content) / 4
		if tokenBudget > 0 && usedTokens+tokens > tokenBudget {
			break
		}
		usedTokens += tokens
		selected = append(selected, c.d)
	}

	return formatXML(selected), nil
}

func formatXML(decisions []*decision.Decision) string {
	var sb strings.Builder
	sb.WriteString("<backstory>\n")
	sb.WriteString("<decisions>\n")
	for _, d := range decisions {
		sb.WriteString("<decision")
		sb.WriteString(fmt.Sprintf(" type=%q", d.Type))
		sb.WriteString(fmt.Sprintf(" date=%q", d.DateStr))
		sb.WriteString(fmt.Sprintf(" author=%q", d.Author))
		sb.WriteString(fmt.Sprintf(" anchor=%q", d.Anchor))
		if d.LinearIssue != "" {
			sb.WriteString(fmt.Sprintf(" linear=%q", d.LinearIssue))
		}
		sb.WriteString(">\n")
		content := d.Title + ". " + d.Body
		sb.WriteString(content)
		sb.WriteString("\n</decision>\n")
	}
	sb.WriteString("</decisions>\n")
	sb.WriteString("</backstory>")
	return sb.String()
}
