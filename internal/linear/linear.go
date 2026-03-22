package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Comment struct {
	Body     string
	UserName string
}

type Issue struct {
	ID          string
	Identifier  string
	Title       string
	Description string
	Comments    []Comment
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type graphQLResponse struct {
	Data struct {
		Issues struct {
			Nodes []struct {
				ID          string `json:"id"`
				Identifier  string `json:"identifier"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Comments    struct {
					Nodes []struct {
						Body string `json:"body"`
						User struct {
							Name string `json:"name"`
						} `json:"user"`
					} `json:"nodes"`
				} `json:"comments"`
			} `json:"nodes"`
		} `json:"issues"`
	} `json:"data"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func ParseIssueResponse(data []byte) (*Issue, error) {
	var resp graphQLResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing graphql response: %w", err)
	}
	nodes := resp.Data.Issues.Nodes
	if len(nodes) == 0 {
		return nil, nil
	}
	n := nodes[0]
	comments := make([]Comment, len(n.Comments.Nodes))
	for i, c := range n.Comments.Nodes {
		comments[i] = Comment{Body: c.Body, UserName: c.User.Name}
	}
	return &Issue{
		ID:          n.ID,
		Identifier:  n.Identifier,
		Title:       n.Title,
		Description: n.Description,
		Comments:    comments,
	}, nil
}

func (c *Client) FetchIssue(ctx context.Context, identifier string) (*Issue, error) {
	query := `query($filter: IssueFilter) {
  issues(filter: $filter, first: 1) {
    nodes {
      id
      identifier
      title
      description
      comments { nodes { body user { name } } }
    }
  }
}`
	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"identifier": map[string]interface{}{
				"eq": identifier,
			},
		},
	}
	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linear API status %d: %s", resp.StatusCode, respBody)
	}

	issue, err := ParseIssueResponse(respBody)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s not found", identifier)
	}
	return issue, nil
}

func FormatIssueXML(issue *Issue) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<linear issue=%q>\n", issue.Identifier))
	sb.WriteString(fmt.Sprintf("Title: %s\n", html.EscapeString(issue.Title)))
	sb.WriteString(fmt.Sprintf("Description: %s\n", html.EscapeString(issue.Description)))
	if len(issue.Comments) > 0 {
		sb.WriteString("Discussion:\n")
		for _, c := range issue.Comments {
			sb.WriteString(fmt.Sprintf("%s: %s\n", html.EscapeString(c.UserName), html.EscapeString(c.Body)))
		}
	}
	sb.WriteString("</linear>")
	return sb.String()
}

func ExtractIssueFromBranch(branch, teamKey string) string {
	pattern := fmt.Sprintf(`(?i)(?:^|[^a-zA-Z])(%s)-(\d+)`, regexp.QuoteMeta(teamKey))
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(branch)
	if match == nil {
		return ""
	}
	return strings.ToUpper(match[1]) + "-" + match[2]
}
