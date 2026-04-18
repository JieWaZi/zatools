package devwikiapp

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/app/skillapp"
	"zatools/internal/devwiki"
	"zatools/internal/skills"
)

func TestResolveTargetDirUsesDetectedProjectRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")
	child := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(child),
		IsTTY:     false,
	})
	service.qmdWarmup = func(context.Context, string) error { return nil }

	target, err := service.resolveTargetDir("Sample Project")
	if err != nil {
		t.Fatalf("resolveTargetDir error = %v", err)
	}

	want := filepath.Join(root, "devwiki-sample-project")
	if target != want {
		t.Fatalf("target = %q, want %q", target, want)
	}
}

func TestNormalizeInitOptionsDefaultsToProjectScope(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	opts, err := service.normalizeInitOptions(InitOptions{
		ProjectName: "Sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{"."},
	})
	if err != nil {
		t.Fatalf("normalizeInitOptions error = %v", err)
	}

	if opts.Global {
		t.Fatalf("Global = true, want false")
	}
	if len(opts.CodeDirs) != 1 || opts.CodeDirs[0] != root {
		t.Fatalf("CodeDirs = %#v, want [%q]", opts.CodeDirs, root)
	}
}

func TestNormalizeInitOptionsAcceptsCursorRuntime(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	opts, err := service.normalizeInitOptions(InitOptions{
		ProjectName: "Sample",
		Agent:       "cursor",
		Lang:        "zh",
		CodeDirs:    []string{"."},
	})
	if err != nil {
		t.Fatalf("normalizeInitOptions error = %v", err)
	}
	if opts.Agent != "cursor" {
		t.Fatalf("Agent = %q, want %q", opts.Agent, "cursor")
	}
}

func TestResolveSelectedSkillsDefaultsToAllInNonTTYMode(t *testing.T) {
	t.Parallel()

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(t.TempDir()),
		IsTTY:     false,
	})

	found := []skills.Skill{
		{Name: "devwiki-setup", Description: "setup"},
		{Name: "devwiki-init", Description: "init"},
	}

	selected, err := service.resolveSelectedSkills(found, InitOptions{})
	if err != nil {
		t.Fatalf("resolveSelectedSkills error = %v", err)
	}
	if len(selected) != len(found) {
		t.Fatalf("len(selected) = %d, want %d", len(selected), len(found))
	}
}

func TestInstallSelectedSkillsWritesIntoCurrentProjectRootByDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	target := filepath.Join(root, "devwiki-sample")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll(target) error = %v", err)
	}

	sourceDir := filepath.Join(root, "builtin-skills", "devwiki-setup")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(sourceDir) error = %v", err)
	}
	mustWriteFileDevwikiApp(t, filepath.Join(sourceDir, "SKILL.md"), "---\nname: devwiki-setup\ndescription: setup\n---\n")

	err := service.installSelectedSkills(root, "codex", false, "zh", []skills.Skill{
		{Name: "devwiki-setup", Description: "setup", Dir: sourceDir, RelativeDir: "."},
	})
	if err != nil {
		t.Fatalf("installSelectedSkills error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "devwiki-setup", "SKILL.md")); err != nil {
		t.Fatalf("missing installed skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".agents", "skills", "devwiki-setup", "SKILL.md")); err == nil {
		t.Fatal("generated devwiki project should not receive project-scope installed skills")
	}

	lock, err := skills.LoadLock(filepath.Join(root, skills.LockFileName))
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.SkillAsset)["devwiki-setup"]
	if entry.Source != "zatools/devwiki#zh" {
		t.Fatalf("Source = %q, want %q", entry.Source, "zatools/devwiki#zh")
	}
}

func TestInitCreatesProjectAndInstallsCodexSkillsIntoCurrentProjectRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")
	child := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll(child) error = %v", err)
	}

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(child),
		IsTTY:     false,
	})
	service.qmdWarmup = func(context.Context, string) error { return nil }

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Sample Project",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{root},
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	target := filepath.Join(root, "devwiki-sample-project")
	for _, rel := range []string{
		"README.md",
		"AGENTS.md",
		"config/project.yaml",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); err == nil {
		t.Fatal("codex init should not generate CLAUDE.md")
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "devwiki-setup", "SKILL.md")); err != nil {
		t.Fatalf("missing installed project-scope skill: %v", err)
	}
	rootAgentsData, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(root AGENTS.md) error = %v", err)
	}
	rootAgents := string(rootAgentsData)
	if !strings.Contains(rootAgents, "./devwiki-sample-project/AGENTS.md") {
		t.Fatalf("root AGENTS.md missing DevWiki runtime bridge:\n%s", rootAgents)
	}
	if _, err := os.Stat(filepath.Join(root, ".zatools-lock.json")); err != nil {
		t.Fatalf("missing project lock file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".agents")); err == nil {
		t.Fatal("generated devwiki project should not own project-scope .agents")
	}
	if _, err := os.Stat(filepath.Join(target, ".zatools-lock.json")); err == nil {
		t.Fatal("generated devwiki project should not own project-scope lock file")
	}

	lock, err := skills.LoadLock(filepath.Join(root, skills.LockFileName))
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.SkillAsset)["devwiki-setup"]
	if entry.Source != "zatools/devwiki#zh" {
		t.Fatalf("Source = %q, want %q", entry.Source, "zatools/devwiki#zh")
	}
}

func TestInitCreatesProjectAndInstallsCursorSkillsIntoCurrentProjectRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")
	child := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll(child) error = %v", err)
	}

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(child),
		IsTTY:     false,
	})
	service.qmdWarmup = func(context.Context, string) error { return nil }

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Sample Project",
		Agent:       "cursor",
		Lang:        "zh",
		CodeDirs:    []string{root},
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	target := filepath.Join(root, "devwiki-sample-project")
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); err != nil {
		t.Fatalf("missing AGENTS.md for cursor runtime: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); err == nil {
		t.Fatal("cursor init should not generate CLAUDE.md")
	}
	if _, err := os.Stat(filepath.Join(root, ".cursor", "skills", "devwiki-setup", "SKILL.md")); err != nil {
		t.Fatalf("missing installed cursor-scope skill: %v", err)
	}
	rootAgentsData, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(root AGENTS.md) error = %v", err)
	}
	if !strings.Contains(string(rootAgentsData), "./devwiki-sample-project/AGENTS.md") {
		t.Fatalf("root AGENTS.md missing cursor DevWiki runtime bridge:\n%s", string(rootAgentsData))
	}
}

func TestInitDoesNotRequireFinalConfirmationPrompt(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")
	child := filepath.Join(root, "nested", "work")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("MkdirAll(child) error = %v", err)
	}

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(child),
		IsTTY:     false,
	})
	service.qmdWarmup = func(context.Context, string) error { return nil }

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "No Prompt Project",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{root},
		Yes:         false,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	target := filepath.Join(root, "devwiki-no-prompt-project")
	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Fatalf("missing README.md after init without final confirmation: %v", err)
	}
}

func TestInitWarmsQMDModelsForGeneratedProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	var warmedRoot string
	service.qmdWarmup = func(_ context.Context, projectRoot string) error {
		warmedRoot = projectRoot
		return nil
	}

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Warmup Project",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{root},
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	want := filepath.Join(root, "devwiki-warmup-project")
	if warmedRoot != want {
		t.Fatalf("warmedRoot = %q, want %q", warmedRoot, want)
	}
}

func TestInitReturnsWarmupErrorButKeepsGeneratedProject(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	service.qmdWarmup = func(_ context.Context, projectRoot string) error {
		return errors.New("warmup failed")
	}

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Warmup Failure",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{root},
		Yes:         true,
	})
	if err == nil {
		t.Fatal("Init should fail when qmd warmup fails")
	}
	if !strings.Contains(err.Error(), "warmup failed") {
		t.Fatalf("Init error = %v", err)
	}

	target := filepath.Join(root, "devwiki-warmup-failure")
	if _, statErr := os.Stat(filepath.Join(target, "README.md")); statErr != nil {
		t.Fatalf("generated project should remain on warmup failure: %v", statErr)
	}
}

func TestUpdateMigratesLegacyDevwikiSourcesAndOnlyRefreshesDevwikiSkills(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	builtinRoot, cleanup, err := devwiki.ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	found, err := skills.Discover(builtinRoot)
	if err != nil {
		t.Fatalf("Discover builtin skills error = %v", err)
	}
	var selected []skills.Skill
	for _, skill := range found {
		if skill.Name == "devwiki-setup" {
			selected = append(selected, skill)
			break
		}
	}
	if len(selected) != 1 {
		t.Fatalf("selected builtin skills = %#v, want devwiki-setup", selected)
	}

	if err := service.installSelectedSkills(root, "codex", false, "zh", selected); err != nil {
		t.Fatalf("installSelectedSkills error = %v", err)
	}

	skillService := skillapp.NewServiceWithRuntime(service.runtime)
	localSource := filepath.Join(root, "custom-skill")
	mustWriteFileDevwikiApp(t, filepath.Join(localSource, "SKILL.md"), "---\nname: custom\ndescription: custom\n---\n")
	mustWriteFileDevwikiApp(t, filepath.Join(localSource, "content.txt"), "v1")
	if err := captureDevwikiStdout(t, func() error {
		return skillService.Add(context.Background(), localSource, skillapp.AddOptions{
			Yes:           true,
			ScopeProvided: true,
			Agents:        []string{"codex"},
		})
	}); err != nil {
		t.Fatalf("skillService.Add error = %v", err)
	}

	lockPath := filepath.Join(root, skills.LockFileName)
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	devEntry := lock.Entries(skills.SkillAsset)["devwiki-setup"]
	devEntry.Source = filepath.Join(root, "tmp", "legacy-devwiki-setup")
	devEntry.Hash = "stale-devwiki"
	lock.Entries(skills.SkillAsset)["devwiki-setup"] = devEntry

	customEntry := lock.Entries(skills.SkillAsset)["custom"]
	customEntry.Hash = "stale-custom"
	lock.Entries(skills.SkillAsset)["custom"] = customEntry
	if err := skills.SaveLock(lockPath, lock); err != nil {
		t.Fatalf("SaveLock error = %v", err)
	}

	if err := captureDevwikiStdout(t, func() error {
		return service.Update(context.Background())
	}); err != nil {
		t.Fatalf("Service.Update error = %v", err)
	}

	updatedLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(updated) error = %v", err)
	}
	updatedDevEntry := updatedLock.Entries(skills.SkillAsset)["devwiki-setup"]
	if updatedDevEntry.Source != "zatools/devwiki#zh" {
		t.Fatalf("updated devwiki source = %q, want %q", updatedDevEntry.Source, "zatools/devwiki#zh")
	}
	if updatedDevEntry.Hash == "stale-devwiki" {
		t.Fatal("expected devwiki hash to refresh")
	}
	if updatedLock.Entries(skills.SkillAsset)["custom"].Hash != "stale-custom" {
		t.Fatal("expected non-devwiki skill to remain untouched by devwiki update")
	}
}

func mustWriteFileDevwikiApp(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func captureDevwikiStdout(t *testing.T, fn func() error) error {
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
