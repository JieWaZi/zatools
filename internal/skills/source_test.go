package skills

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSourceRecognizesBuiltinDevwikiLibrary(t *testing.T) {
	t.Setenv("ZATOOLS_LANG", "en")

	source, err := ParseSource("zatools/devwiki")
	if err != nil {
		t.Fatalf("ParseSource returned error: %v", err)
	}

	if source.Type != "builtin" {
		t.Fatalf("Type = %q, want %q", source.Type, "builtin")
	}
	if source.Builtin != "devwiki" {
		t.Fatalf("Builtin = %q, want %q", source.Builtin, "devwiki")
	}
	if source.Ref != "en" {
		t.Fatalf("Ref = %q, want %q", source.Ref, "en")
	}
	if source.Original != "zatools/devwiki#en" {
		t.Fatalf("Original = %q, want %q", source.Original, "zatools/devwiki#en")
	}
}

func TestResolveSourceBuiltinDevwikiExtractsEmbeddedSkills(t *testing.T) {
	t.Parallel()

	source := NewBuiltinSource("devwiki", "zh")
	resolved, err := ResolveSource(t.Context(), source)
	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	defer func() {
		if cleanupErr := resolved.Cleanup(); cleanupErr != nil {
			t.Fatalf("cleanup resolved source: %v", cleanupErr)
		}
	}()

	searchRoot, err := resolved.SearchRoot()
	if err != nil {
		t.Fatalf("SearchRoot returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(searchRoot, "qmd-sync", "SKILL.md")); err != nil {
		t.Fatalf("missing builtin qmd-sync skill: %v", err)
	}
}

func TestParseSourceRejectsUnsafeGitHubSubpath(t *testing.T) {
	t.Parallel()

	_, err := ParseSource("owner/repo/../evil")
	if err == nil {
		t.Fatal("expected ParseSource to reject path traversal subpath")
	}
	if !strings.Contains(err.Error(), "unsafe subpath") {
		t.Fatalf("expected unsafe subpath error, got %v", err)
	}
}

func TestSearchRootRejectsUnsafeSubpath(t *testing.T) {
	t.Parallel()

	resolved := ResolvedSource{
		Source:  Source{Subpath: "../escape"},
		RootDir: filepath.Join(t.TempDir(), "repo"),
	}

	_, err := resolved.SearchRoot()
	if err == nil {
		t.Fatal("expected SearchRoot to reject path traversal subpath")
	}
	if !strings.Contains(err.Error(), "unsafe subpath") {
		t.Fatalf("expected unsafe subpath error, got %v", err)
	}
}

func TestResolveSourceChecksOutCommitRef(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	runGitOrFail(t, "", "init", repoDir)
	runGitOrFail(t, repoDir, "config", "user.name", "Test User")
	runGitOrFail(t, repoDir, "config", "user.email", "test@example.com")

	skillDir := filepath.Join(repoDir, "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo\ndescription: first\n---\n"), 0o644); err != nil {
		t.Fatalf("write first skill file: %v", err)
	}
	runGitOrFail(t, repoDir, "add", ".")
	runGitOrFail(t, repoDir, "commit", "-m", "first")
	firstCommit := strings.TrimSpace(runGitOutputOrFail(t, repoDir, "rev-parse", "HEAD"))

	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: demo\ndescription: second\n---\n"), 0o644); err != nil {
		t.Fatalf("write second skill file: %v", err)
	}
	runGitOrFail(t, repoDir, "add", ".")
	runGitOrFail(t, repoDir, "commit", "-m", "second")

	resolved, err := ResolveSource(t.Context(), Source{
		Type:    "git",
		RepoURL: repoDir,
		Ref:     firstCommit,
	})
	if err != nil {
		t.Fatalf("ResolveSource returned error: %v", err)
	}
	defer func() {
		if cleanupErr := resolved.Cleanup(); cleanupErr != nil {
			t.Fatalf("cleanup resolved source: %v", cleanupErr)
		}
	}()

	searchRoot, err := resolved.SearchRoot()
	if err != nil {
		t.Fatalf("SearchRoot returned error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(searchRoot, "demo", "SKILL.md"))
	if err != nil {
		t.Fatalf("read resolved skill file: %v", err)
	}
	if !strings.Contains(string(got), "description: first") {
		t.Fatalf("expected first commit content, got:\n%s", string(got))
	}
}

func runGitOrFail(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func runGitOutputOrFail(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}
