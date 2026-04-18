package skillapp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

func TestInitSkillCreatesTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "My Skill")
	if err := initSkill(dir); err != nil {
		t.Fatalf("initSkill error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(SKILL.md) error = %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "name: my-skill") || !strings.Contains(text, "# my-skill") {
		t.Fatalf("template content = %q", text)
	}

	if err := initSkill(dir); err == nil {
		t.Fatal("expected initSkill to reject existing SKILL.md")
	}
}

func TestSelectSkillsResolveAgentsScopeAndRemoveNames(t *testing.T) {
	service := newTestService(t)
	found := []skills.Skill{
		{Name: "alpha", Description: "A"},
		{Name: "beta", Description: "B"},
	}

	selected, err := service.selectSkills(found, AddOptions{SkillNames: []string{"beta"}})
	if err != nil {
		t.Fatalf("selectSkills(named) error = %v", err)
	}
	if want := []skills.Skill{{Name: "beta", Description: "B"}}; !reflect.DeepEqual(selected, want) {
		t.Fatalf("selectSkills(named) = %#v, want %#v", selected, want)
	}

	selected, err = service.selectSkills(found, AddOptions{Yes: true})
	if err != nil {
		t.Fatalf("selectSkills(default) error = %v", err)
	}
	if !reflect.DeepEqual(selected, found) {
		t.Fatalf("selectSkills(default) = %#v, want %#v", selected, found)
	}

	if _, err := service.selectSkills(found, AddOptions{SkillNames: []string{"missing"}}); err == nil {
		t.Fatal("expected selectSkills to reject unknown names")
	}

	agentKeys, proceed, err := service.resolveAgents(AddOptions{Agents: []string{"claude-code", "codex", "codex"}})
	if err != nil {
		t.Fatalf("resolveAgents(explicit) error = %v", err)
	}
	if !proceed || !reflect.DeepEqual(agentKeys, []string{"claude", "codex"}) {
		t.Fatalf("resolveAgents(explicit) = %#v, proceed=%v", agentKeys, proceed)
	}

	agentKeys, global, proceed, err := service.resolveInstallTargets(AddOptions{Yes: true, Global: true, ScopeProvided: true})
	if err != nil {
		t.Fatalf("resolveInstallTargets error = %v", err)
	}
	if !proceed || !global || !reflect.DeepEqual(agentKeys, []string{"codex", "cursor", "claude"}) {
		t.Fatalf("resolveInstallTargets = %#v, global=%v, proceed=%v", agentKeys, global, proceed)
	}

	names, err := service.resolveRemoveNames([]skills.InstalledAsset{{Name: "b"}, {Name: "a"}}, RemoveOptions{All: true})
	if err != nil {
		t.Fatalf("resolveRemoveNames(all) error = %v", err)
	}
	if !reflect.DeepEqual(names, []string{"b", "a"}) {
		t.Fatalf("resolveRemoveNames(all) = %#v", names)
	}

	names, err = service.resolveRemoveNames(nil, RemoveOptions{SkillNames: []string{"x"}})
	if err != nil {
		t.Fatalf("resolveRemoveNames(explicit) error = %v", err)
	}
	if !reflect.DeepEqual(names, []string{"x"}) {
		t.Fatalf("resolveRemoveNames(explicit) = %#v", names)
	}

	names, err = service.resolveRemoveNames(nil, RemoveOptions{})
	if err != nil {
		t.Fatalf("resolveRemoveNames(non-tty) error = %v", err)
	}
	if names != nil {
		t.Fatalf("resolveRemoveNames(non-tty) = %#v, want nil", names)
	}
}

func TestCheckInstalledSkillsAndInstallForAgents(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	currentDir := filepath.Join(projectDir, "current-skill")
	mustWriteFileApp(t, filepath.Join(currentDir, "SKILL.md"), "---\nname: current\ndescription: current\n---\n")
	currentHash, err := skills.HashDir(currentDir)
	if err != nil {
		t.Fatalf("HashDir(currentDir) error = %v", err)
	}

	outdatedDir := filepath.Join(projectDir, "outdated-skill")
	mustWriteFileApp(t, filepath.Join(outdatedDir, "SKILL.md"), "---\nname: outdated\ndescription: outdated\n---\n")
	mustWriteFileApp(t, filepath.Join(outdatedDir, "data.txt"), "new")

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}

	lock := skills.LockFile{
		Skills: map[string]skills.InstalledAsset{
			"broken": {
				Name:   "broken",
				Source: "://bad",
			},
			"current": {
				Name:   "current",
				Source: currentDir,
				Hash:   currentHash,
			},
			"outdated": {
				Name:   "outdated",
				Source: outdatedDir,
				Hash:   "stale-hash",
			},
		},
	}
	if err := skills.SaveLock(lockPath, lock); err != nil {
		t.Fatalf("SaveLock error = %v", err)
	}

	results, err := service.checkInstalledSkills(context.Background(), false)
	if err != nil {
		t.Fatalf("checkInstalledSkills error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("checkInstalledSkills len = %d, want 3", len(results))
	}

	statusByName := map[string]string{}
	for _, result := range results {
		statusByName[result.Asset.Name] = result.Status
	}
	if statusByName["broken"] != "invalid-source" || statusByName["current"] != "current" || statusByName["outdated"] != "outdated" {
		t.Fatalf("checkInstalledSkills statuses = %#v", statusByName)
	}

	installSourceDir := filepath.Join(projectDir, "install-src")
	mustWriteFileApp(t, filepath.Join(installSourceDir, "SKILL.md"), "---\nname: Demo Skill\ndescription: demo\n---\n")
	mustWriteFileApp(t, filepath.Join(installSourceDir, "notes.txt"), "demo")

	entry, err := service.installForAgents(
		skills.Source{Original: installSourceDir, Type: "local"},
		skills.Skill{Name: "Demo Skill", Description: "demo", Dir: installSourceDir, RelativeDir: "."},
		[]string{"codex", "cursor"},
		false,
	)
	if err != nil {
		t.Fatalf("installForAgents error = %v", err)
	}
	if len(entry.AgentPaths) != 2 || entry.Path == "" {
		t.Fatalf("installForAgents entry = %#v", entry)
	}
	for _, path := range entry.AgentPaths {
		if _, err := os.Stat(filepath.Join(path, "notes.txt")); err != nil {
			t.Fatalf("installed agent path missing file in %q: %v", path, err)
		}
	}
}

func TestFormattingAndPathHelpers(t *testing.T) {
	service := newTestService(t)
	t.Setenv("HOME", filepath.Join(service.runtime.Workspace.ProjectDir(), "home"))

	summary := common.FormatSourceSummary(skills.Source{
		Type:     "github",
		RepoURL:  "https://github.com/a/b.git",
		Ref:      "main",
		Subpath:  "skills/demo",
		LocalDir: "/tmp/ignored",
	})
	if !strings.Contains(summary, "https://github.com/a/b.git") || !strings.Contains(summary, "skills/demo") {
		t.Fatalf("formatSourceSummary = %q", summary)
	}

	got := common.ShortenPath(filepath.Join(service.runtime.Workspace.ProjectDir(), "dir", "file"), service.runtime.Workspace.ProjectDir())
	if got != "dir/file" {
		t.Fatalf("shortenPath(project) = %q", got)
	}

	if rel, ok := common.RelativeToRoot(filepath.Join(service.runtime.Workspace.ProjectDir(), "a"), service.runtime.Workspace.ProjectDir(), "."); !ok || rel != "a" {
		t.Fatalf("relativeToRoot = %q, %v", rel, ok)
	}
	if _, ok := common.RelativeToRoot("/tmp", "/var", "."); ok {
		t.Fatal("relativeToRoot should reject path outside root")
	}

	normalized, err := common.NormalizeAgentKeys([]string{"cursor", "claude-code", "cursor", "codex"}, skills.SkillAsset, ui.Messages().UnsupportedAgentFmt)
	if err != nil {
		t.Fatalf("normalizeAgents error = %v", err)
	}
	if want := []string{"claude", "codex", "cursor"}; !reflect.DeepEqual(normalized, want) {
		t.Fatalf("normalizeAgents = %#v, want %#v", normalized, want)
	}
	if _, err := common.NormalizeAgentKeys([]string{"bad"}, skills.SkillAsset, ui.Messages().UnsupportedAgentFmt); err == nil {
		t.Fatal("expected normalizeAgents to reject unsupported agent")
	}

	dirs, err := common.ResolveAgentDirectories(skills.SkillAsset, []string{"codex", "cursor"}, false, service.runtime.Workspace.ProjectDir())
	if err != nil {
		t.Fatalf("resolveAgentDirectories error = %v", err)
	}
	if len(dirs) != 2 {
		t.Fatalf("resolveAgentDirectories = %#v", dirs)
	}

	targets, err := service.describeScopeTargets([]string{"codex", "cursor"}, false)
	if err != nil {
		t.Fatalf("describeScopeTargets error = %v", err)
	}
	if !strings.Contains(targets, ".agents/skills") || !strings.Contains(targets, ".cursor/skills") {
		t.Fatalf("describeScopeTargets = %q", targets)
	}

	lines := service.buildInstallSummary(
		skills.Source{Type: "local", LocalDir: "/tmp/source"},
		[]skills.Skill{{Name: "alpha"}, {Name: "beta"}},
		[]string{"codex", "cursor"},
		false,
	)
	if len(lines) != 5 || !strings.Contains(lines[4], "alpha, beta") {
		t.Fatalf("buildInstallSummary = %#v", lines)
	}
}

func TestCollectionAndRemovalHelpers(t *testing.T) {
	service := newTestService(t)
	lock := skills.LockFile{
		Skills: map[string]skills.InstalledAsset{
			"b": {Name: "b"},
			"a": {Name: "a"},
		},
	}
	sorted := common.SortedInstalledAssets(lock, skills.SkillAsset)
	if got := []string{sorted[0].Name, sorted[1].Name}; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("sortedInstalledSkills names = %#v", got)
	}

	found, ok := findDiscoveredSkill([]skills.Skill{{Name: "alpha"}, {Name: "beta"}}, "beta")
	if !ok || found.Name != "beta" {
		t.Fatalf("findDiscoveredSkill = %#v, %v", found, ok)
	}

	entry := skills.InstalledAsset{Name: "demo", Agents: []string{"cursor", "codex"}}
	roots, err := service.resolveInstalledPathRoots(entry, false)
	if err != nil {
		t.Fatalf("resolveInstalledPathRoots error = %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("resolveInstalledPathRoots = %#v", roots)
	}

	allowedRoot := filepath.Join(t.TempDir(), "allowed")
	target := filepath.Join(allowedRoot, "demo")
	mustWriteFileApp(t, filepath.Join(target, "file.txt"), "data")

	if err := common.ValidateInstalledPath(target, []string{allowedRoot}); err != nil {
		t.Fatalf("validateInstalledPath(valid) error = %v", err)
	}
	if err := common.RemoveInstalledPath(target, []string{allowedRoot}); err != nil {
		t.Fatalf("removeInstalledPath error = %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed, stat err = %v", target, err)
	}

	if err := common.ValidateInstalledPath(allowedRoot, []string{allowedRoot}); err == nil {
		t.Fatal("expected validateInstalledPath to reject root removal")
	}
	if err := common.ValidateInstalledPath(filepath.Join(t.TempDir(), "other"), []string{allowedRoot}); err == nil {
		t.Fatal("expected validateInstalledPath to reject unexpected path")
	}
	if err := common.RemoveInstalledPath("", []string{allowedRoot}); err != nil {
		t.Fatalf("removeInstalledPath(empty) error = %v", err)
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

func mustWriteFileApp(t *testing.T, path string, content string) {
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

func TestServiceInitAndStatusHelpers(t *testing.T) {
	service := newTestService(t)
	if service.Runtime().Workspace == nil {
		t.Fatal("Runtime() returned nil workspace")
	}

	dir := filepath.Join(t.TempDir(), "service-init")
	if err := service.Init(context.Background(), dir); err != nil {
		t.Fatalf("Service.Init error = %v", err)
	}

	if got := ui.StatusText("current"); got == "" {
		t.Fatal("ui.StatusText(current) returned empty string")
	}
}

func TestServiceLifecycleCommands(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	sourceDir := filepath.Join(projectDir, "source-skill")
	mustWriteFileApp(t, filepath.Join(sourceDir, "SKILL.md"), "---\nname: demo\ndescription: demo skill\n---\n")
	mustWriteFileApp(t, filepath.Join(sourceDir, "content.txt"), "v1")

	addOpts := AddOptions{
		Yes:           true,
		ScopeProvided: true,
		Agents:        []string{"codex"},
	}
	if err := captureStdout(t, func() error {
		return service.Add(context.Background(), sourceDir, addOpts)
	}); err != nil {
		t.Fatalf("Service.Add error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry, ok := lock.Entries(skills.SkillAsset)["demo"]
	if !ok {
		t.Fatalf("lock skills = %#v, want demo entry", lock.Entries(skills.SkillAsset))
	}

	if err := captureStdout(t, func() error {
		return service.List(context.Background(), false)
	}); err != nil {
		t.Fatalf("Service.List error = %v", err)
	}

	if err := captureStdout(t, func() error {
		return service.Check(context.Background(), false)
	}); err != nil {
		t.Fatalf("Service.Check(current) error = %v", err)
	}

	mustWriteFileApp(t, filepath.Join(sourceDir, "content.txt"), "v2")

	if err := captureStdout(t, func() error {
		return service.Update(context.Background(), false)
	}); err != nil {
		t.Fatalf("Service.Update error = %v", err)
	}

	updatedLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(updated) error = %v", err)
	}
	if updatedLock.Entries(skills.SkillAsset)["demo"].Hash == entry.Hash {
		t.Fatal("expected Service.Update to refresh installed hash")
	}

	if err := captureStdout(t, func() error {
		return service.Remove(context.Background(), RemoveOptions{
			Global:     false,
			SkillNames: []string{"demo"},
		})
	}); err != nil {
		t.Fatalf("Service.Remove error = %v", err)
	}

	finalLock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock(final) error = %v", err)
	}
	if len(finalLock.Entries(skills.SkillAsset)) != 0 {
		t.Fatalf("final lock skills = %#v, want empty", finalLock.Entries(skills.SkillAsset))
	}
	if _, err := os.Stat(entry.Path); !os.IsNotExist(err) {
		t.Fatalf("expected installed path to be removed, stat err = %v", err)
	}
}

func TestServiceAddBuiltinDevwikiLibrary(t *testing.T) {
	service := newTestService(t)
	t.Setenv("ZATOOLS_LANG", "en")

	err := captureStdout(t, func() error {
		return service.Add(context.Background(), "zatools/devwiki", AddOptions{
			Yes:           true,
			ScopeProvided: true,
			Agents:        []string{"codex"},
		})
	})
	if err != nil {
		t.Fatalf("Service.Add(builtin devwiki) error = %v", err)
	}

	lockPath, err := service.runtime.Workspace.LockFilePath(false)
	if err != nil {
		t.Fatalf("LockFilePath error = %v", err)
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock error = %v", err)
	}
	entry, ok := lock.Entries(skills.SkillAsset)["devwiki-setup"]
	if !ok {
		t.Fatalf("expected devwiki-setup in lock, got %#v", lock.Entries(skills.SkillAsset))
	}
	if entry.Source != "zatools/devwiki#en" {
		t.Fatalf("Source = %q, want %q", entry.Source, "zatools/devwiki#en")
	}
	if !strings.HasSuffix(entry.SourceSubdir, "setup") {
		t.Fatalf("SourceSubdir = %q, want suffix %q", entry.SourceSubdir, "setup")
	}
}

func TestServiceEmptyFlowsAndListOnly(t *testing.T) {
	service := newTestService(t)
	projectDir := service.runtime.Workspace.ProjectDir()

	sourceDir := filepath.Join(projectDir, "source-list-only")
	mustWriteFileApp(t, filepath.Join(sourceDir, "SKILL.md"), "---\nname: listonly\ndescription: demo skill\n---\n")

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
	if len(lock.Entries(skills.SkillAsset)) != 0 {
		t.Fatalf("list-only lock skills = %#v, want empty", lock.Entries(skills.SkillAsset))
	}

	for name, fn := range map[string]func() error{
		"list":   func() error { return service.List(context.Background(), false) },
		"remove": func() error { return service.Remove(context.Background(), RemoveOptions{}) },
		"check":  func() error { return service.Check(context.Background(), false) },
		"update": func() error { return service.Update(context.Background(), false) },
	} {
		if err := captureStdout(t, fn); err != nil {
			t.Fatalf("Service.%s empty flow error = %v", name, err)
		}
	}
}

func TestAdditionalHelperBranches(t *testing.T) {
	service := newTestService(t)

	if got, err := service.selectSkills([]skills.Skill{{Name: "one"}}, AddOptions{SkillNames: []string{"*"}}); err != nil || len(got) != 1 {
		t.Fatalf("selectSkills(*) = %#v, err=%v", got, err)
	}

	agentKeys, proceed, err := service.resolveAgents(AddOptions{})
	if err != nil || !proceed || !reflect.DeepEqual(agentKeys, []string{"codex", "cursor", "claude"}) {
		t.Fatalf("resolveAgents(default) = %#v proceed=%v err=%v", agentKeys, proceed, err)
	}

	global, proceed, err := service.resolveScope(AddOptions{ScopeProvided: true, Global: true}, []string{"codex"})
	if err != nil || !proceed || !global {
		t.Fatalf("resolveScope(scope provided) = global=%v proceed=%v err=%v", global, proceed, err)
	}

	names, err := service.resolveRemoveNames(nil, RemoveOptions{Yes: true})
	if err != nil || names != nil {
		t.Fatalf("resolveRemoveNames(yes) = %#v err=%v", names, err)
	}

	if ok, err := common.ConfirmInstall(true, ui.Messages().PromptInstallNow); err != nil || !ok {
		t.Fatalf("confirmInstall(true) = ok=%v err=%v", ok, err)
	}

	t.Setenv("HOME", filepath.Join(service.runtime.Workspace.ProjectDir(), "home"))
	homePath := filepath.Join(os.Getenv("HOME"), "bin", "zatools")
	if got := common.ShortenPath(homePath, "/unrelated"); !strings.HasPrefix(got, "~/") {
		t.Fatalf("shortenPath(home) = %q", got)
	}
}

func TestNewServiceUsesDetectedRuntime(t *testing.T) {
	service := NewService()
	if service == nil || service.Runtime().Workspace == nil {
		t.Fatal("NewService returned incomplete runtime")
	}
}
