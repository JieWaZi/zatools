package ruleapp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/skills"
)

func TestServiceLifecycleCommandsAcrossAgents(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	sourceDir := filepath.Join(projectDir, "source-rules")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "RULE.yaml"), "common:\n  name: shared-rules\n  description: common docs\n")
	mdPath := filepath.Join(sourceDir, "rules", "common", "engineering.md")
	mdcPath := filepath.Join(sourceDir, "rules", "cursor", "backend", "style.mdc")
	mustWriteFile(t, mdPath, "---\ndescription: engineering guide\n---\n# v1\n")
	mustWriteFile(t, mdcPath, "---\ndescription: backend style\n---\n# v1\n")

	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), sourceDir, AddOptions{
			Yes:       true,
			RuleNames: []string{"shared-rules"},
			Agents:    []string{"cursor"},
		})
	}); err != nil {
		t.Fatalf("Service.Add(markdown->cursor) error = %v", err)
	}

	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), sourceDir, AddOptions{
			Yes:       true,
			RuleNames: []string{"cursor"},
			Agents:    []string{"claude"},
		})
	}); err != nil {
		t.Fatalf("Service.Add(mdc->claude) error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}

	markdownRule := lock.Entries(skills.RuleAsset)["shared-rules"]
	if markdownRule.DetectedAgent != "claude" {
		t.Fatalf("markdown rule detected agent = %q, want claude", markdownRule.DetectedAgent)
	}
	if markdownRule.SourceSubdir != "" {
		t.Fatalf("markdown rule source_subdir = %q, want empty", markdownRule.SourceSubdir)
	}
	if got := markdownRule.Agents; !reflect.DeepEqual(got, []string{"cursor"}) {
		t.Fatalf("markdown rule agents = %#v, want [cursor]", got)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".cursor", "rules", "common", "engineering.md")); err != nil {
		t.Fatalf("markdown rule not installed to cursor: %v", err)
	}

	cursorRule := lock.Entries(skills.RuleAsset)["cursor"]
	if cursorRule.DetectedAgent != "cursor" {
		t.Fatalf("cursor rule detected agent = %q, want cursor", cursorRule.DetectedAgent)
	}
	if cursorRule.SourceSubdir != "" {
		t.Fatalf("cursor rule source_subdir = %q, want empty", cursorRule.SourceSubdir)
	}
	if got := cursorRule.Agents; !reflect.DeepEqual(got, []string{"claude"}) {
		t.Fatalf("cursor rule agents = %#v, want [claude]", got)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude", "rules", "cursor", "backend", "style.mdc")); err != nil {
		t.Fatalf("cursor rule not installed to claude: %v", err)
	}

	if err := captureStdout(t, func() error {
		return service.Check(context.Background())
	}); err != nil {
		t.Fatalf("Service.Check error = %v", err)
	}

	mustWriteFile(t, mdPath, "---\ndescription: engineering guide\n---\n# v2\n")

	if err := captureStdout(t, func() error {
		return service.Update(context.Background())
	}); err != nil {
		t.Fatalf("Service.Update error = %v", err)
	}

	updatedLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(updated) error = %v", err)
	}
	if updatedLock.Entries(skills.RuleAsset)["shared-rules"].Hash == markdownRule.Hash {
		t.Fatal("expected markdown rule hash to change after update")
	}

	if err := captureStdout(t, func() error {
		return service.Remove(context.Background(), RemoveOptions{
			RuleNames: []string{"shared-rules", "cursor"},
		})
	}); err != nil {
		t.Fatalf("Service.Remove error = %v", err)
	}

	finalLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(final) error = %v", err)
	}
	if len(finalLock.Entries(skills.RuleAsset)) != 0 {
		t.Fatalf("final lock rules = %#v, want empty", finalLock.Entries(skills.RuleAsset))
	}
}

func TestServiceListOnlyDefaultAgentsAndUnsupportedAgent(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	sourceDir := filepath.Join(projectDir, "source-list-only")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "common", "team.md"), "---\ndescription: team\n---\n")

	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), sourceDir, AddOptions{ListOnly: true, Yes: true})
	}); err != nil {
		t.Fatalf("Service.Add(list-only) error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	if len(lock.Entries(skills.RuleAsset)) != 0 {
		t.Fatalf("list-only rules = %#v, want empty", lock.Entries(skills.RuleAsset))
	}

	if got := defaultAgents(); !reflect.DeepEqual(got, []string{"claude", "cursor"}) {
		t.Fatalf("defaultAgents = %#v", got)
	}

	if err := service.Add(context.Background(), sourceDir, AddOptions{Yes: true, Agents: []string{"codex"}}); err == nil {
		t.Fatal("expected unsupported agent error")
	}
}

func TestServiceRootRuleOnlyInstallsAndHashesSelectedFiles(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	sourceDir := filepath.Join(projectDir, "source-root-rules")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "RULE.yaml"), "name: repo-rules\ndescription: root package\n")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "team.md"), "# Team v1\n")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "cursor.mdc"), "# Cursor v1\n")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "nested", "child.md"), "# Child v1\n")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "README.txt"), "ignore me\n")

	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), sourceDir, AddOptions{
			Yes:       true,
			RuleNames: []string{"repo-rules"},
			Agents:    []string{"claude"},
		})
	}); err != nil {
		t.Fatalf("Service.Add(root-rules) error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.RuleAsset)["repo-rules"]
	if !reflect.DeepEqual(entry.SourceFiles, []string{"cursor.mdc", "team.md"}) {
		t.Fatalf("root rule source_files = %#v, want [cursor.mdc team.md]", entry.SourceFiles)
	}

	installPath := filepath.Join(projectDir, ".claude", "rules", "root-rules")
	if _, err := os.Stat(filepath.Join(installPath, "team.md")); err != nil {
		t.Fatalf("root markdown rule missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "cursor.mdc")); err != nil {
		t.Fatalf("root cursor rule missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "nested", "child.md")); !os.IsNotExist(err) {
		t.Fatalf("nested child unexpectedly installed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "README.txt")); !os.IsNotExist(err) {
		t.Fatalf("unrelated file unexpectedly installed, stat err = %v", err)
	}

	mustWriteFile(t, filepath.Join(sourceDir, "rules", "nested", "child.md"), "# Child v2\n")
	results, err := service.checkInstalledRules(context.Background())
	if err != nil {
		t.Fatalf("checkInstalledRules error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("checkInstalledRules len = %d, want 1", len(results))
	}
	if results[0].Status != "current" {
		t.Fatalf("root rule status after unrelated child change = %q, want current", results[0].Status)
	}
}

func TestServiceRelativeLocalRuleSourceRemainsStableAcrossDirectories(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()
	workspaceRoot := filepath.Dir(projectDir)
	sourceDir := filepath.Join(workspaceRoot, "shared-rules")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "common", "team.md"), "# Team v1\n")

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir(projectDir) error = %v", err)
	}

	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), "../shared-rules", AddOptions{
			Yes:       true,
			RuleNames: []string{"common"},
			Agents:    []string{"claude"},
		})
	}); err != nil {
		t.Fatalf("Service.Add(relative local source) error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.RuleAsset)["common"]
	storedSource, err := filepath.EvalSymlinks(entry.Source)
	if err != nil {
		t.Fatalf("EvalSymlinks(stored source) error = %v", err)
	}
	wantSource, err := filepath.EvalSymlinks(sourceDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(sourceDir) error = %v", err)
	}
	if storedSource != wantSource {
		t.Fatalf("stored source = %q, want %q", storedSource, wantSource)
	}

	subdir := filepath.Join(projectDir, "nested", "work")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("MkdirAll(subdir) error = %v", err)
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("Chdir(subdir) error = %v", err)
	}

	if err := captureStdout(t, func() error {
		return service.Check(context.Background())
	}); err != nil {
		t.Fatalf("Service.Check(from different dir) error = %v", err)
	}

	mustWriteFile(t, filepath.Join(sourceDir, "rules", "common", "team.md"), "# Team v2\n")
	if err := captureStdout(t, func() error {
		return service.Update(context.Background())
	}); err != nil {
		t.Fatalf("Service.Update(from different dir) error = %v", err)
	}

	updatedLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(updated) error = %v", err)
	}
	if updatedLock.Entries(skills.RuleAsset)["common"].Hash == entry.Hash {
		t.Fatal("expected updated hash after modifying relative local rule source")
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	projectDir := t.TempDir()
	t.Setenv("HOME", filepath.Join(projectDir, "home"))
	return NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(projectDir),
		IsTTY:     false,
	})
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func captureStdout(t *testing.T, fn func() error) error {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe error = %v", err)
	}

	os.Stdout = writer
	runErr := fn()
	_ = writer.Close()
	os.Stdout = original
	_, _ = io.ReadAll(reader)
	_ = reader.Close()
	return runErr
}
