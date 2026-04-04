package agents

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSupportedReturnsCopy(t *testing.T) {
	t.Parallel()

	got := Supported()
	if len(got) == 0 {
		t.Fatal("Supported returned no agents")
	}

	got[0].DisplayName = "changed"

	again := Supported()
	if again[0].DisplayName == "changed" {
		t.Fatal("Supported returned shared backing data")
	}
}

func TestLookupDisplayNamesAndResolveSkillsDir(t *testing.T) {
	t.Setenv("HOME", "/tmp/test-home")

	agent, ok := Lookup("codex")
	if !ok {
		t.Fatal("Lookup(codex) = not found")
	}
	if agent.DisplayName != "Codex" {
		t.Fatalf("unexpected agent display name: %q", agent.DisplayName)
	}

	projectDir, err := ResolveSkillsDir("cursor", false, "/repo")
	if err != nil {
		t.Fatalf("ResolveSkillsDir(project) error = %v", err)
	}
	if want := filepath.Join("/repo", ".cursor/skills"); projectDir != want {
		t.Fatalf("ResolveSkillsDir(project) = %q, want %q", projectDir, want)
	}

	globalDir, err := ResolveSkillsDir("claude", true, "/repo")
	if err != nil {
		t.Fatalf("ResolveSkillsDir(global) error = %v", err)
	}
	if want := filepath.Join("/tmp/test-home", ".claude/skills"); globalDir != want {
		t.Fatalf("ResolveSkillsDir(global) = %q, want %q", globalDir, want)
	}

	names := DisplayNames([]string{"codex", "unknown", "claude"})
	if want := []string{"Codex", "Claude Code"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("DisplayNames = %#v, want %#v", names, want)
	}
}

func TestResolveSkillsDirRejectsUnknownAgent(t *testing.T) {
	t.Parallel()

	if _, err := ResolveSkillsDir("unknown", false, "/repo"); err == nil {
		t.Fatal("expected error for unsupported agent")
	}
}
