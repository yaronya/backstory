package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "backstory")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/backstory")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s %v", out, err)
	}
	return bin
}

func runCmd(t *testing.T, bin string, env []string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running %v: %v", args, err)
		}
	}
	return string(out), exitCode
}

func TestInitCreatesRepoStructure(t *testing.T) {
	bin := buildBinary(t)
	dir := filepath.Join(t.TempDir(), "decisions")

	cmd := exec.Command(bin, "init", "--path", dir)
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %s %v", out, err)
	}

	for _, rel := range []string{
		"product",
		"technical",
		filepath.Join(".backstory", "config.yml"),
		"README.md",
		".git",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}
}

func TestInitConnect(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()

	bare := filepath.Join(tmp, "bare.git")
	if out, err := exec.Command("git", "init", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %s %v", out, err)
	}

	source := filepath.Join(tmp, "source")
	sourceCmd := func(name string, args ...string) {
		t.Helper()
		c := exec.Command(name, args...)
		c.Dir = source
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("command %s %v failed: %s %v", name, args, out, err)
		}
	}

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	sourceCmd("git", "init")
	sourceCmd("git", "remote", "add", "origin", bare)

	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, ".backstory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".backstory", "config.yml"), []byte("team: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceCmd("git", "add", ".")
	sourceCmd("git", "commit", "-m", "init")
	sourceCmd("git", "push", "origin", "HEAD:master")

	cloneDir := filepath.Join(tmp, "clone")
	cmd := exec.Command(bin, "init", "--connect", bare, "--path", cloneDir)
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init --connect failed: %s %v", out, err)
	}

	for _, rel := range []string{"README.md", ".git"} {
		if _, err := os.Stat(filepath.Join(cloneDir, rel)); err != nil {
			t.Errorf("expected %s to exist in clone: %v", rel, err)
		}
	}
}

func TestIndexAndSearch(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "decisions")

	initCmd := exec.Command(bin, "init", "--path", repoDir)
	initCmd.Env = os.Environ()
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %s %v", out, err)
	}

	decisionContent := `---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s. Direct invocation from the Lambda
would hit this limit during peak hours. SQS provides natural backpressure
and retry semantics without custom rate-limiting code.
`
	decisionPath := filepath.Join(repoDir, "technical", "2026-03-19-chose-sqs.md")
	if err := os.WriteFile(decisionPath, []byte(decisionContent), 0o644); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(), "BACKSTORY_REPO="+repoDir)

	out, code := runCmd(t, bin, env, "index")
	if code != 0 {
		t.Fatalf("index failed (exit %d): %s", code, out)
	}

	out, code = runCmd(t, bin, env, "search", "SQS")
	if code != 0 {
		t.Fatalf("search failed (exit %d): %s", code, out)
	}
	if !strings.Contains(out, "Chose SQS") {
		t.Errorf("expected search output to contain 'Chose SQS', got: %s", out)
	}
}

func TestStatus(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "decisions")

	initCmd := exec.Command(bin, "init", "--path", repoDir)
	initCmd.Env = os.Environ()
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %s %v", out, err)
	}

	decisionContent := `---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s.
`
	decisionPath := filepath.Join(repoDir, "technical", "2026-03-19-chose-sqs.md")
	if err := os.WriteFile(decisionPath, []byte(decisionContent), 0o644); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(), "BACKSTORY_REPO="+repoDir)

	out, code := runCmd(t, bin, env, "status")
	if code != 0 {
		t.Fatalf("status failed (exit %d): %s", code, out)
	}
	if !strings.Contains(out, "Total decisions: 1") {
		t.Errorf("expected 'Total decisions: 1', got: %s", out)
	}
}

func TestInjectDoesNotCrash(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "decisions")

	initCmd := exec.Command(bin, "init", "--path", repoDir)
	initCmd.Env = os.Environ()
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %s %v", out, err)
	}

	decisionContent := `---
type: technical
date: 2026-03-19
author: sarah
anchor: env0/services/payment-service/
linear_issue: ENG-892
stale: false
---

# Chose SQS over direct invocation for vendor API

The vendor API rate-limits at 100 req/s.
`
	decisionPath := filepath.Join(repoDir, "technical", "2026-03-19-chose-sqs.md")
	if err := os.WriteFile(decisionPath, []byte(decisionContent), 0o644); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(), "BACKSTORY_REPO="+repoDir)
	indexOut, code := runCmd(t, bin, env, "index")
	if code != 0 {
		t.Fatalf("index failed (exit %d): %s", code, indexOut)
	}

	_, code = runCmd(t, bin, env, "inject")
	if code != 0 {
		t.Errorf("inject should not crash, but exited with code %d", code)
	}
}
