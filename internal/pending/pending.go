package pending

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yaronya/backstory/internal/decision"
)

type Queue struct {
	dir string
}

func New(dir string) *Queue {
	return &Queue{dir: dir}
}

func (q *Queue) file() string {
	return filepath.Join(q.dir, "pending.json")
}

func (q *Queue) lock() (*os.File, error) {
	lockPath := q.file() + ".lock"
	if err := os.MkdirAll(q.dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("could not acquire lock: %w", err)
		}
	}
	return f, nil
}

func (q *Queue) unlock(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func (q *Queue) save(decisions []*decision.Decision) error {
	if err := os.MkdirAll(q.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(decisions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(q.file(), data, 0o644)
}

func (q *Queue) Append(decisions []*decision.Decision) error {
	lf, err := q.lock()
	if err != nil {
		return err
	}
	defer q.unlock(lf)

	existing, _ := q.Load()
	existing = append(existing, decisions...)
	return q.save(existing)
}

func (q *Queue) Save(decisions []*decision.Decision) error {
	lf, err := q.lock()
	if err != nil {
		return err
	}
	defer q.unlock(lf)

	return q.save(decisions)
}

func (q *Queue) Load() ([]*decision.Decision, error) {
	data, err := os.ReadFile(q.file())
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
	lf, err := q.lock()
	if err != nil {
		return err
	}
	defer q.unlock(lf)

	err = os.Remove(q.file())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (q *Queue) HasPending() bool {
	info, err := os.Stat(q.file())
	return err == nil && info.Size() > 0
}
