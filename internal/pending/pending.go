package pending

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/backstory-team/backstory/internal/decision"
)

type Queue struct {
	dir string
}

func New(dir string) *Queue {
	return &Queue{dir: dir}
}

func (q *Queue) filePath() string {
	return filepath.Join(q.dir, "pending.json")
}

func (q *Queue) Save(decisions []*decision.Decision) error {
	if err := os.MkdirAll(q.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(decisions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(q.filePath(), data, 0o644)
}

func (q *Queue) Load() ([]*decision.Decision, error) {
	data, err := os.ReadFile(q.filePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var decisions []*decision.Decision
	if err := json.Unmarshal(data, &decisions); err != nil {
		return nil, err
	}
	return decisions, nil
}

func (q *Queue) Clear() error {
	err := os.Remove(q.filePath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (q *Queue) HasPending() bool {
	info, err := os.Stat(q.filePath())
	return err == nil && info.Size() > 0
}
