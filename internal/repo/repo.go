package repo

import (
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	Path string
}

func NormalizeGitURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.ReplaceAll(url, ":", "/")
	return strings.ToLower(url)
}

func MatchesRemote(actual, configured string) bool {
	return NormalizeGitURL(actual) == NormalizeGitURL(configured)
}

func Clone(url, dest string) (*Repo, error) {
	cmd := exec.Command("git", "clone", url, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone: %w: %s", err, out)
	}
	return &Repo{Path: dest}, nil
}

func Open(path string) *Repo {
	return &Repo{Path: path}
}

func (r *Repo) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) Pull() error {
	_, err := r.git("pull", "--ff-only")
	return err
}

func (r *Repo) CommitAll(msg string) error {
	if _, err := r.git("add", "-A"); err != nil {
		return err
	}
	_, err := r.git("commit", "-m", msg)
	return err
}

func (r *Repo) CommitAndPush(file, msg string) error {
	if _, err := r.git("add", file); err != nil {
		return err
	}
	if _, err := r.git("commit", "-m", msg); err != nil {
		return err
	}
	_, err := r.git("push")
	return err
}

func (r *Repo) PushWithRebase(maxRetries int) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		_, err := r.git("push")
		if err == nil {
			return nil
		}
		lastErr = err
		if i < maxRetries {
			if _, rebaseErr := r.git("pull", "--rebase"); rebaseErr != nil {
				return rebaseErr
			}
		}
	}
	return lastErr
}

func (r *Repo) GetRemoteURL() (string, error) {
	return r.git("remote", "get-url", "origin")
}

func (r *Repo) GetCurrentBranch() (string, error) {
	return r.git("rev-parse", "--abbrev-ref", "HEAD")
}

func (r *Repo) GetRepoRoot() (string, error) {
	return r.git("rev-parse", "--show-toplevel")
}
