package devwikiapp

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/app/skillapp"
	"zatools/internal/devwiki"
	"zatools/internal/qmd"
	"zatools/internal/skills"
)

func TestResolveTargetDirUsesCurrentWorkingDirectory(t *testing.T) {
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

	target, err := service.resolveTargetDir()
	if err != nil {
		t.Fatalf("resolveTargetDir error = %v", err)
	}

	want, err := filepath.Abs(child)
	if err != nil {
		t.Fatalf("Abs(child) error = %v", err)
	}
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
	if opts.Lang != "zh" {
		t.Fatalf("Lang = %q, want zh", opts.Lang)
	}
}

func TestNormalizeInitOptionsForcesZhLang(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	opts, err := service.normalizeInitOptions(InitOptions{
		ProjectName: "Sample",
		Agent:       "codex",
		Lang:        "en",
		CodeDirs:    []string{"."},
	})
	if err != nil {
		t.Fatalf("normalizeInitOptions error = %v", err)
	}
	if opts.Lang != "zh" {
		t.Fatalf("Lang = %q, want zh", opts.Lang)
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
		{Name: "devwiki-project-router", Description: "router"},
		{Name: "devwiki-ingest", Description: "ingest"},
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

	sourceDir := filepath.Join(root, "builtin-skills", "devwiki-project-router")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(sourceDir) error = %v", err)
	}
	mustWriteFileDevwikiApp(t, filepath.Join(sourceDir, "SKILL.md"), "---\nname: devwiki-project-router\ndescription: router\n---\n")

	err := service.installSelectedSkills(root, "codex", false, "zh", []skills.Skill{
		{Name: "devwiki-project-router", Description: "router", Dir: sourceDir, RelativeDir: "."},
	})
	if err != nil {
		t.Fatalf("installSelectedSkills error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "devwiki-project-router", "SKILL.md")); err != nil {
		t.Fatalf("missing installed skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".agents", "skills", "devwiki-project-router", "SKILL.md")); err == nil {
		t.Fatal("generated devwiki project should not receive project-scope installed skills")
	}

	lock, err := skills.LoadLock(filepath.Join(root, skills.LockFileName))
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.SkillAsset)["devwiki-project-router"]
	if entry.Source != "zatools/devwiki#zh" {
		t.Fatalf("Source = %q, want %q", entry.Source, "zatools/devwiki#zh")
	}
}

func TestInitCreatesProjectAndInstallsCodexSkillsIntoCurrentProjectRoot(t *testing.T) {
	docRoot := t.TempDir()
	codeRoot := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(docRoot),
		IsTTY:     false,
	})

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Sample Project",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{codeRoot},
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	for _, rel := range []string{
		"README.md",
		"AGENTS.md",
		"config/project.yaml",
		"wiki/index.md",
	} {
		if _, err := os.Stat(filepath.Join(docRoot, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(docRoot, "wiki/relations.yml")); err == nil {
		t.Fatal("wiki/relations.yml should not be generated")
	}
	if _, err := os.Stat(filepath.Join(docRoot, "CLAUDE.md")); err == nil {
		t.Fatal("codex init should not generate CLAUDE.md")
	}
	for _, rel := range []string{
		".agents/skills/devwiki-project-router/SKILL.md",
		".agents/skills/devwiki-ingest/SKILL.md",
		".agents/skills/devwiki-maintain/SKILL.md",
		".agents/skills/devwiki-code/SKILL.md",
		".agents/skills/devwiki-query/SKILL.md",
		".agents/skills/devwiki-code-to-doc/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(docRoot, rel)); err != nil {
			t.Fatalf("missing doc-root installed skill %s: %v", rel, err)
		}
	}
	gitignoreData, err := os.ReadFile(filepath.Join(docRoot, ".gitignore"))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	gitignore := string(gitignoreData)
	for _, want := range []string{".agents", ".cache", ".zatools-lock.json"} {
		if !strings.Contains(gitignore, want) {
			t.Fatalf(".gitignore missing %q:\n%s", want, gitignore)
		}
	}
	if _, err := os.Stat(filepath.Join(docRoot, ".zatools-lock.json")); err != nil {
		t.Fatalf("missing project lock file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(codeRoot, "AGENTS.md")); err == nil {
		t.Fatal("init should not create code repo AGENTS.md")
	}
	if _, err := os.Stat(filepath.Join(codeRoot, ".agents")); err == nil {
		t.Fatal("init should not install code repo skills")
	}

	lock, err := skills.LoadLock(filepath.Join(docRoot, skills.LockFileName))
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry := lock.Entries(skills.SkillAsset)["devwiki-project-router"]
	if entry.Source != "zatools/devwiki#zh" {
		t.Fatalf("Source = %q, want %q", entry.Source, "zatools/devwiki#zh")
	}
}

func TestInitCreatesProjectAndInstallsCursorSkillsIntoCurrentProjectRoot(t *testing.T) {
	docRoot := t.TempDir()
	codeRoot := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(docRoot),
		IsTTY:     false,
	})

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "Sample Project",
		Agent:       "cursor",
		Lang:        "zh",
		CodeDirs:    []string{codeRoot},
		Yes:         true,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(docRoot, "AGENTS.md")); err != nil {
		t.Fatalf("missing AGENTS.md for cursor runtime: %v", err)
	}
	if _, err := os.Stat(filepath.Join(docRoot, "CLAUDE.md")); err == nil {
		t.Fatal("cursor init should not generate CLAUDE.md")
	}
	if _, err := os.Stat(filepath.Join(docRoot, ".cursor", "skills", "devwiki-project-router", "SKILL.md")); err != nil {
		t.Fatalf("missing installed cursor-scope skill: %v", err)
	}
	gitignoreData, err := os.ReadFile(filepath.Join(docRoot, ".gitignore"))
	if err != nil {
		t.Fatalf("ReadFile(.gitignore) error = %v", err)
	}
	gitignore := string(gitignoreData)
	for _, want := range []string{".cursor", ".cache", ".zatools-lock.json"} {
		if !strings.Contains(gitignore, want) {
			t.Fatalf(".gitignore missing %q:\n%s", want, gitignore)
		}
	}
	if _, err := os.Stat(filepath.Join(codeRoot, "AGENTS.md")); err == nil {
		t.Fatal("cursor init should not create code repo AGENTS.md")
	}
}

func TestInitDoesNotRequireFinalConfirmationPrompt(t *testing.T) {
	docRoot := t.TempDir()
	codeRoot := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(docRoot),
		IsTTY:     false,
	})

	err := service.Init(context.Background(), InitOptions{
		ProjectName: "No Prompt Project",
		Agent:       "codex",
		Lang:        "zh",
		CodeDirs:    []string{codeRoot},
		Yes:         false,
	})
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(docRoot, "README.md")); err != nil {
		t.Fatalf("missing README.md after init without final confirmation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(codeRoot, "AGENTS.md")); err == nil {
		t.Fatal("init should not create code repo AGENTS.md")
	}
}

func TestInitDoesNotWarmQMDModelsAndPrintsManualDownloadHint(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "go.mod"), "module example\n")

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	var output string
	err := captureDevwikiStdoutText(t, func() error {
		runErr := service.Init(context.Background(), InitOptions{
			ProjectName: "Warmup Project",
			Agent:       "codex",
			Lang:        "zh",
			CodeDirs:    []string{root},
			Yes:         true,
		})
		return runErr
	}, &output)
	if err != nil {
		t.Fatalf("Init error = %v", err)
	}

	if !strings.Contains(output, "zatools qmd download --root .") {
		t.Fatalf("Init output missing manual qmd download hint:\n%s", output)
	}
}

func TestUpdateRefreshesDevwikiSourcesAndLeavesOtherSkillsUntouched(t *testing.T) {
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
		if skill.Name == "devwiki-ingest" {
			selected = append(selected, skill)
			break
		}
	}
	if len(selected) != 1 {
		t.Fatalf("selected builtin skills = %#v, want devwiki-ingest", selected)
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
	devEntry := lock.Entries(skills.SkillAsset)["devwiki-ingest"]
	devEntry.Source = filepath.Join(root, "tmp", "cached-devwiki-ingest")
	devEntry.Hash = "stale-devwiki"
	lock.Entries(skills.SkillAsset)["devwiki-ingest"] = devEntry

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
	updatedDevEntry := updatedLock.Entries(skills.SkillAsset)["devwiki-ingest"]
	if updatedDevEntry.Source != "zatools/devwiki#zh" {
		t.Fatalf("updated devwiki source = %q, want %q", updatedDevEntry.Source, "zatools/devwiki#zh")
	}
	if updatedDevEntry.Hash == "stale-devwiki" {
		t.Fatal("expected devwiki hash to refresh")
	}
	ingestData, err := os.ReadFile(filepath.Join(root, ".agents", "skills", "devwiki-ingest", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(updated ingest) error = %v", err)
	}
	if !strings.Contains(string(ingestData), "加载 `devwiki-topic`") || !strings.Contains(string(ingestData), "加载 `devwiki-workflow`") {
		t.Fatalf("updated devwiki-ingest missing page-writing skill dispatch guidance:\n%s", string(ingestData))
	}
	if updatedLock.Entries(skills.SkillAsset)["custom"].Hash != "stale-custom" {
		t.Fatal("expected non-devwiki skill to remain untouched by devwiki update")
	}
}

func TestUpdateInstallsMissingBuiltinSkillsInDevwikiRoot(t *testing.T) {
	root := t.TempDir()
	codeRoot := t.TempDir()
	if err := devwiki.GenerateProject(root, devwiki.ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []devwiki.CodeRepo{
			{Name: "code", Slug: "code", Path: codeRoot, Default: true},
		},
	}); err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

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
	selected := make([]skills.Skill, 0, len(found)-1)
	for _, skill := range found {
		if skill.Name == "devwiki-maintain" {
			continue
		}
		selected = append(selected, skill)
	}
	if len(selected) != len(found)-1 {
		t.Fatalf("selected skills = %d, found = %d; expected only maintain to be omitted", len(selected), len(found))
	}
	if err := service.installSelectedSkills(root, "codex", false, "zh", selected); err != nil {
		t.Fatalf("installSelectedSkills error = %v", err)
	}

	lockPath := filepath.Join(root, skills.LockFileName)
	beforeLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(before) error = %v", err)
	}
	if _, ok := beforeLock.Entries(skills.SkillAsset)["devwiki-maintain"]; ok {
		t.Fatal("test setup should not have devwiki-maintain installed")
	}

	ctx := qmd.WithCommandRunner(context.Background(), devwikiQMDHelperRunner(t, ""))
	if err := captureDevwikiStdout(t, func() error {
		return service.Update(ctx)
	}); err != nil {
		t.Fatalf("Service.Update error = %v", err)
	}

	afterLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(after) error = %v", err)
	}
	entry, ok := afterLock.Entries(skills.SkillAsset)["devwiki-maintain"]
	if !ok {
		t.Fatalf("expected devwiki update to install missing devwiki-maintain, got %#v", afterLock.Entries(skills.SkillAsset))
	}
	if entry.Source != "zatools/devwiki#zh" {
		t.Fatalf("maintain Source = %q, want %q", entry.Source, "zatools/devwiki#zh")
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "devwiki-maintain", "SKILL.md")); err != nil {
		t.Fatalf("missing installed devwiki-maintain skill: %v", err)
	}
}

func TestUpdateRefreshesQMDIndexAfterSkillUpdate(t *testing.T) {
	root := t.TempDir()
	codeRoot := t.TempDir()
	if err := devwiki.GenerateProject(root, devwiki.ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []devwiki.CodeRepo{
			{Name: "code", Slug: "code", Path: codeRoot, Default: true},
		},
	}); err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

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
	if err := service.installSelectedSkills(root, "codex", false, "zh", found); err != nil {
		t.Fatalf("installSelectedSkills error = %v", err)
	}

	ctx := qmd.WithCommandRunner(context.Background(), devwikiQMDHelperRunner(t, ""))
	var output string
	if err := captureDevwikiStdoutText(t, func() error {
		return service.Update(ctx)
	}, &output); err != nil {
		t.Fatalf("Service.Update error = %v", err)
	}

	for _, want := range []string{
		"argv=qmd collection add " + filepath.Join(root, "wiki") + " --name devwiki-sample-wiki",
		"argv=qmd update",
		"argv=qmd embed",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("update output missing %q:\n%s", want, output)
		}
	}
}

func TestUpdateReportsQMDRefreshFailureWithoutFailing(t *testing.T) {
	root := t.TempDir()
	codeRoot := t.TempDir()
	if err := devwiki.GenerateProject(root, devwiki.ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []devwiki.CodeRepo{
			{Name: "code", Slug: "code", Path: codeRoot, Default: true},
		},
	}); err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(root),
		IsTTY:     false,
	})

	ctx := qmd.WithCommandRunner(context.Background(), devwikiQMDHelperRunner(t, "update"))
	var output string
	if err := captureDevwikiStdoutText(t, func() error {
		return service.Update(ctx)
	}, &output); err != nil {
		t.Fatalf("Service.Update should not fail when qmd update fails: %v", err)
	}
	if !strings.Contains(output, "qmd update") || !strings.Contains(output, "失败") {
		t.Fatalf("expected qmd failure warning in output:\n%s", output)
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

func captureDevwikiStdoutText(t *testing.T, fn func() error, output *string) error {
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
	data, _ := io.ReadAll(reader)
	_ = reader.Close()
	*output = string(data)
	return runErr
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}

func devwikiQMDHelperRunner(t *testing.T, failSubcommand string) func(context.Context, string, ...string) *exec.Cmd {
	t.Helper()

	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		commandArgs := []string{"-test.run=TestDevwikiQMDHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_DEVWIKI_QMD_HELPER_PROCESS=1",
			"GO_WANT_DEVWIKI_QMD_FAIL="+failSubcommand,
		)
		return cmd
	}
}

func TestDevwikiQMDHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_DEVWIKI_QMD_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	if fail := os.Getenv("GO_WANT_DEVWIKI_QMD_FAIL"); fail != "" && len(args) > 1 && args[1] == fail {
		_, _ = os.Stderr.WriteString("qmd failed\n")
		os.Exit(1)
	}
	_, _ = os.Stdout.WriteString("argv=" + strings.Join(args, " ") + "\n")
	os.Exit(0)
}
