package repo_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaronya/backstory/internal/repo"
)

func initBareRemote(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "remote.git")
	cmd := exec.Command("git", "init", "--bare", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %s %v", out, err)
	}
	return dir
}

func initAndSeedRemote(t *testing.T) string {
	t.Helper()
	remote := initBareRemote(t)
	tmp := t.TempDir()
	seed, err := repo.Clone(remote, filepath.Join(tmp, "seed"))
	if err != nil {
		t.Fatalf("seed clone failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seed.Path, "init.txt"), []byte("init"), 0644); err != nil {
		t.Fatalf("write init.txt: %v", err)
	}
	configureGitUser(t, seed.Path)
	if err := seed.CommitAndPush("init.txt", "initial commit"); err != nil {
		t.Fatalf("seed CommitAndPush: %v", err)
	}
	return remote
}

func configureGitUser(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %s %v", args, out, err)
		}
	}
}

func TestClone(t *testing.T) {
	remote := initBareRemote(t)
	dest := filepath.Join(t.TempDir(), "cloned")
	r, err := repo.Clone(remote, dest)
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.Path, ".git")); os.IsNotExist(err) {
		t.Fatal(".git directory does not exist after clone")
	}
}

func TestCommitAndPush(t *testing.T) {
	remote := initBareRemote(t)
	tmp := t.TempDir()
	r, err := repo.Clone(remote, filepath.Join(tmp, "work"))
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}
	configureGitUser(t, r.Path)

	if err := os.WriteFile(filepath.Join(r.Path, "hello.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := r.CommitAndPush("hello.txt", "add hello"); err != nil {
		t.Fatalf("CommitAndPush failed: %v", err)
	}
}

func TestPull(t *testing.T) {
	remote := initAndSeedRemote(t)
	tmp := t.TempDir()

	r1, err := repo.Clone(remote, filepath.Join(tmp, "clone1"))
	if err != nil {
		t.Fatalf("clone1 failed: %v", err)
	}
	configureGitUser(t, r1.Path)

	r2, err := repo.Clone(remote, filepath.Join(tmp, "clone2"))
	if err != nil {
		t.Fatalf("clone2 failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(r1.Path, "new.txt"), []byte("from-r1"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := r1.CommitAndPush("new.txt", "add new.txt"); err != nil {
		t.Fatalf("CommitAndPush: %v", err)
	}

	if err := r2.Pull(); err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(r2.Path, "new.txt"))
	if err != nil {
		t.Fatalf("read new.txt in r2: %v", err)
	}
	if string(data) != "from-r1" {
		t.Fatalf("expected 'from-r1', got %q", string(data))
	}
}

func TestPullRebase(t *testing.T) {
	remote := initAndSeedRemote(t)
	tmp := t.TempDir()

	r1, err := repo.Clone(remote, filepath.Join(tmp, "clone1"))
	if err != nil {
		t.Fatalf("clone1 failed: %v", err)
	}
	configureGitUser(t, r1.Path)

	r2, err := repo.Clone(remote, filepath.Join(tmp, "clone2"))
	if err != nil {
		t.Fatalf("clone2 failed: %v", err)
	}
	configureGitUser(t, r2.Path)

	if err := os.WriteFile(filepath.Join(r1.Path, "from-r1.txt"), []byte("r1"), 0644); err != nil {
		t.Fatalf("write r1 file: %v", err)
	}
	if err := r1.CommitAndPush("from-r1.txt", "push from r1"); err != nil {
		t.Fatalf("r1 CommitAndPush: %v", err)
	}

	if err := os.WriteFile(filepath.Join(r2.Path, "from-r2.txt"), []byte("r2"), 0644); err != nil {
		t.Fatalf("write r2 file: %v", err)
	}
	if err := r2.CommitAll("push from r2"); err != nil {
		t.Fatalf("r2 CommitAll: %v", err)
	}

	if err := r2.PushWithRebase(3); err != nil {
		t.Fatalf("PushWithRebase failed: %v", err)
	}
}

func TestGetRemoteURL(t *testing.T) {
	remote := initBareRemote(t)
	r, err := repo.Clone(remote, filepath.Join(t.TempDir(), "cloned"))
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	url, err := r.GetRemoteURL()
	if err != nil {
		t.Fatalf("GetRemoteURL failed: %v", err)
	}
	if url != remote {
		t.Fatalf("expected remote URL %q, got %q", remote, url)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	remote := initAndSeedRemote(t)
	r, err := repo.Clone(remote, filepath.Join(t.TempDir(), "cloned"))
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	branch, err := r.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	branch = strings.TrimSpace(branch)
	if branch != "master" && branch != "main" {
		t.Fatalf("expected 'master' or 'main', got %q", branch)
	}
}
