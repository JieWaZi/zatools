package devwiki

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseResetScopesExpandsAllAndDeduplicates(t *testing.T) {
	t.Parallel()

	got, err := ParseResetScopes("wiki,all,raw,wiki")
	if err != nil {
		t.Fatalf("ParseResetScopes error = %v", err)
	}

	want := []string{"wiki", "raw", "log", "checkpoints"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseResetScopes = %#v, want %#v", got, want)
	}
}

func TestBuildResetPlanSkipsGitkeepAndTargetsExpectedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "features", ".gitkeep"), "")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "features", "note.md"), "# note\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "workflows", "flow.md"), "# flow\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "troubleshooting", "debug.md"), "# debug\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "sources", "legacy.md"), "# legacy source\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", "modules", "legacy.md"), "# legacy module\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "raw", "requirements", ".gitkeep"), "")
	mustWriteDevwikiFile(t, filepath.Join(root, "raw", "requirements", "spec.md"), "# spec\n")
	mustWriteDevwikiFile(t, filepath.Join(root, "wiki", ".checkpoints", "checkpoint.json"), "{}")

	plan, err := BuildResetPlan(root, []string{"wiki", "raw", "checkpoints"})
	if err != nil {
		t.Fatalf("BuildResetPlan error = %v", err)
	}

	if containsString(plan.Delete, filepath.Join(root, "wiki", "features", ".gitkeep")) {
		t.Fatal("plan should not delete .gitkeep files")
	}
	if !containsString(plan.Delete, filepath.Join(root, "wiki", "features", "note.md")) {
		t.Fatal("plan should delete wiki features")
	}
	if !containsString(plan.Delete, filepath.Join(root, "wiki", "workflows", "flow.md")) {
		t.Fatal("plan should delete wiki workflows")
	}
	if !containsString(plan.Delete, filepath.Join(root, "wiki", "troubleshooting", "debug.md")) {
		t.Fatal("plan should delete wiki troubleshooting")
	}
	if containsString(plan.Delete, filepath.Join(root, "wiki", "sources", "legacy.md")) {
		t.Fatal("plan should not target removed legacy wiki sources")
	}
	if containsString(plan.Delete, filepath.Join(root, "wiki", "modules", "legacy.md")) {
		t.Fatal("plan should not target removed legacy wiki modules")
	}
	if !containsString(plan.Delete, filepath.Join(root, "raw", "requirements", "spec.md")) {
		t.Fatal("plan should delete raw documents")
	}
	if !containsString(plan.Delete, filepath.Join(root, "wiki", ".checkpoints", "checkpoint.json")) {
		t.Fatal("plan should delete checkpoints")
	}
	if !containsString(plan.Reset, filepath.Join(root, "wiki", "index.md")) {
		t.Fatal("plan should reset wiki/index.md for wiki scope")
	}
}

func TestApplyResetPlanDeletesFilesAndRewritesTemplates(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	deletePath := filepath.Join(root, "wiki", "features", "note.md")
	indexPath := filepath.Join(root, "wiki", "index.md")
	logPath := filepath.Join(root, "wiki", "log.md")
	mustWriteDevwikiFile(t, deletePath, "# note\n")
	mustWriteDevwikiFile(t, indexPath, "old index\n")
	mustWriteDevwikiFile(t, logPath, "old log\n")

	result, err := ApplyResetPlan(ResetPlan{
		Root:   root,
		Scopes: []string{"wiki", "log"},
		Delete: []string{deletePath},
		Reset:  []string{indexPath, logPath},
	})
	if err != nil {
		t.Fatalf("ApplyResetPlan error = %v", err)
	}

	if result.DeletedCount != 1 || result.ResetCount != 2 {
		t.Fatalf("ApplyResetPlan result = %#v, want deleted=1 reset=2", result)
	}
	if _, err := os.Stat(deletePath); !os.IsNotExist(err) {
		t.Fatalf("deleted file still exists or wrong error: %v", err)
	}

	indexData := mustReadDevwikiFile(t, indexPath)
	if indexData != "# Wiki Index\n" {
		t.Fatalf("index content = %q", indexData)
	}
	logData := mustReadDevwikiFile(t, logPath)
	if logData != "# Wiki Log\n\n> Append-only chronological log.\n" {
		t.Fatalf("log content = %q", logData)
	}
}

func TestAppendLogWritesDatedEntry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteDevwikiFile(t, filepath.Join(root, "log.md"), "# Wiki Log\n\n")

	when := time.Date(2026, 4, 16, 8, 0, 0, 0, time.UTC)
	if err := AppendLog(root, "reset | scope=wiki", when); err != nil {
		t.Fatalf("AppendLog error = %v", err)
	}

	data := mustReadDevwikiFile(t, filepath.Join(root, "log.md"))
	if !strings.Contains(data, "## [2026-04-16] reset | scope=wiki\n") {
		t.Fatalf("log entry missing, got %q", data)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mustWriteDevwikiFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustReadDevwikiFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}
