package decision

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/adrg/frontmatter"
)

type Decision struct {
	Type        string    `yaml:"type" json:"type"`
	Date        time.Time `yaml:"-" json:"date"`
	DateStr     string    `yaml:"date" json:"date_str"`
	Author      string    `yaml:"author" json:"author"`
	Anchor      string    `yaml:"anchor" json:"anchor"`
	LinearIssue string    `yaml:"linear_issue" json:"linear_issue"`
	Stale       bool      `yaml:"stale" json:"stale"`
	Title       string    `yaml:"-" json:"title"`
	Body        string    `yaml:"-" json:"body"`
	FilePath    string    `yaml:"-" json:"file_path"`
}

var multiHyphen = regexp.MustCompile(`-{2,}`)

func ParseFromFile(path string) (*Decision, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	d := &Decision{}
	body, err := frontmatter.Parse(strings.NewReader(string(data)), d)
	if err != nil {
		return nil, err
	}

	if d.DateStr != "" {
		parsed, err := time.Parse("2006-01-02", d.DateStr)
		if err != nil {
			return nil, fmt.Errorf("parsing date %q: %w", d.DateStr, err)
		}
		d.Date = parsed
	}

	content := strings.TrimSpace(string(body))
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			d.Title = strings.TrimPrefix(line, "# ")
			rest := strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
			d.Body = rest
			break
		}
	}

	d.FilePath = path
	return d, nil
}

func ParseAllFromDir(dir string) ([]*Decision, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var decisions []*Decision
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		d, err := ParseFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		decisions = append(decisions, d)
	}

	return decisions, nil
}

func (d *Decision) Filename() string {
	slug := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return '-'
	}, d.Title)

	slug = multiHyphen.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return d.DateStr + "-" + slug + ".md"
}

func (d *Decision) WriteToFile(path string) error {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("type: %s\n", d.Type))
	sb.WriteString(fmt.Sprintf("date: %s\n", d.DateStr))
	sb.WriteString(fmt.Sprintf("author: %s\n", d.Author))
	sb.WriteString(fmt.Sprintf("anchor: %s\n", d.Anchor))
	sb.WriteString(fmt.Sprintf("linear_issue: %s\n", d.LinearIssue))
	sb.WriteString(fmt.Sprintf("stale: %v\n", d.Stale))
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", d.Title))
	sb.WriteString(d.Body)

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}
