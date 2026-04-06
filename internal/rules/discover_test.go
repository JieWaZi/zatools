package rules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverParsesRulesAndBuildsStableInstallNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	commonDir := filepath.Join(root, "common")
	cursorDir := filepath.Join(root, "cursor", "backend")
	if err := os.MkdirAll(commonDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(common) error = %v", err)
	}
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(cursor) error = %v", err)
	}

	meta := "common:\n  name: shared-rules\n  description: common docs\n"
	if err := os.WriteFile(filepath.Join(root, "RULE.yaml"), []byte(meta), 0o644); err != nil {
		t.Fatalf("WriteFile(RULE.yaml) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(commonDir, "engineering.md"), []byte("# Engineering\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(cursorDir, "style.mdc"), []byte("# Style\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(mdc) error = %v", err)
	}

	found, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover error = %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("Discover len = %d, want 2", len(found))
	}

	byName := map[string]Rule{}
	for _, item := range found {
		byName[item.Name] = item
	}

	commonRule := byName["shared-rules"]
	if commonRule.Name != "shared-rules" {
		t.Fatalf("commonRule.Name = %q", commonRule.Name)
	}
	if !reflect.DeepEqual(commonRule.DetectedAgents, []string{ClaudeTag}) {
		t.Fatalf("commonRule.DetectedAgents = %#v", commonRule.DetectedAgents)
	}
	if commonRule.Description != "common docs" {
		t.Fatalf("commonRule.Description = %q", commonRule.Description)
	}
	if commonRule.RelativeDir != "common" {
		t.Fatalf("commonRule.RelativeDir = %q", commonRule.RelativeDir)
	}
	if commonRule.InstallName != "common" {
		t.Fatalf("commonRule.InstallName = %q", commonRule.InstallName)
	}

	cursorRule := byName["cursor"]
	if cursorRule.Name != "cursor" {
		t.Fatalf("cursorRule.Name = %q", cursorRule.Name)
	}
	if !reflect.DeepEqual(cursorRule.DetectedAgents, []string{CursorTag}) {
		t.Fatalf("cursorRule.DetectedAgents = %#v", cursorRule.DetectedAgents)
	}
	if cursorRule.Description != "style.mdc" {
		t.Fatalf("cursorRule.Description = %q", cursorRule.Description)
	}
	if cursorRule.RelativeDir != "cursor" {
		t.Fatalf("cursorRule.RelativeDir = %q", cursorRule.RelativeDir)
	}
	if cursorRule.InstallName != "cursor" {
		t.Fatalf("cursorRule.InstallName = %q", cursorRule.InstallName)
	}
}

func TestDefaultRootAndFind(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if got := DefaultRoot(root); got != root {
		t.Fatalf("DefaultRoot(no rules) = %q, want %q", got, root)
	}

	rulesDir := filepath.Join(root, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(rules) error = %v", err)
	}
	if got := DefaultRoot(root); got != rulesDir {
		t.Fatalf("DefaultRoot(rules) = %q, want %q", got, rulesDir)
	}

	rule, ok := Find([]Rule{{Name: "alpha"}, {Name: "beta"}}, "beta")
	if !ok || rule.Name != "beta" {
		t.Fatalf("Find = %#v, %v", rule, ok)
	}
}

func TestDiscoverSkipsUnsupportedExtensions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "ignore.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	found, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover error = %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("Discover = %#v, want no rules", found)
	}
}
