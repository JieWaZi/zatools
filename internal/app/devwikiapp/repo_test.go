package devwikiapp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/devwiki"
	"zatools/internal/skills"
)

func TestRepoAddInfoIncludesCodeReposAsJSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	var infoOut bytes.Buffer
	if err := service.RepoInfo(context.Background(), RepoInfoOptions{
		ProjectSlug: "huawei-zddi",
		Format:      "json",
		Stdout:      &infoOut,
	}); err != nil {
		t.Fatalf("RepoInfo() error = %v", err)
	}
	var info RepoInfo
	if err := json.Unmarshal(infoOut.Bytes(), &info); err != nil {
		t.Fatalf("Unmarshal repo info error = %v, output=%q", err, infoOut.String())
	}
	if info.ProjectSlug != "huawei-zddi" || info.ActiveSource != "local" || info.Sources.Local == nil || info.Sources.Local.Path != root {
		t.Fatalf("repo info = %#v", info)
	}
	want := []CodeRepoInfo{{Name: "zddiv3", Slug: "zddiv3", Path: codeRoot, Default: true}}
	if !reflect.DeepEqual(info.CodeRepos, want) {
		t.Fatalf("repo info code_repos = %#v, want %#v", info.CodeRepos, want)
	}
}

func TestRepoLinkWritesProjectNameIntoCodeRepoAgents(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "AGENTS.md"), "# DevWiki\n")
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(codeRoot, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "DevWiki project：`huawei-zddi`") || !strings.Contains(content, "--project huawei-zddi") {
		t.Fatalf("AGENTS.md missing project guidance:\n%s", content)
	}
	if !strings.Contains(content, "DevWiki 文档库根目录：`"+root+"`") || !strings.Contains(content, "必须先阅读 `"+filepath.Join(root, "AGENTS.md")+"`") {
		t.Fatalf("AGENTS.md missing local DevWiki root guidance:\n%s", content)
	}
}

func TestRepoLinkWritesCodeRepoAgentsForRemoteProjectConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		RemoteURL:   "http://devwiki.example.com:5697",
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}
	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(codeRoot, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "DevWiki project：`huawei-zddi`") || !strings.Contains(content, "--project huawei-zddi") {
		t.Fatalf("AGENTS.md missing project guidance:\n%s", content)
	}
	if strings.Contains(content, "DevWiki 文档库根目录") || strings.Contains(content, "必须先阅读") || strings.Contains(content, "config/search.yaml") || strings.Contains(content, "必须写入关联 DevWiki 文档库") {
		t.Fatalf("remote AGENTS.md should not include local DevWiki path guidance:\n%s", content)
	}
}

func TestRepoLinkInstallsCodeRepoSkills(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	for _, rel := range []string{
		".agents/skills/devwiki-code/SKILL.md",
		".agents/skills/devwiki-code-to-doc/SKILL.md",
		".agents/skills/devwiki-query/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(codeRoot, rel)); err != nil {
			t.Fatalf("missing installed code repo skill %s: %v", rel, err)
		}
	}
	for _, rel := range []string{
		".agents/skills/devwiki-ingest/SKILL.md",
		".agents/skills/devwiki-maintain/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(codeRoot, rel)); err == nil {
			t.Fatalf("repo link should not install doc-root skill %s", rel)
		}
	}
}

func TestRepoAddRemoteWritesRemoteSource(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	service := NewService()
	var addOut bytes.Buffer

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		RemoteURL:   "http://devwiki.example.com:5697",
		Stdout:      &addOut,
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}
	if !strings.Contains(addOut.String(), "已添加 DevWiki project `huawei-zddi`") ||
		!strings.Contains(addOut.String(), "remote: http://devwiki.example.com:5697") {
		t.Fatalf("RepoAdd(remote) output = %q", addOut.String())
	}

	var out bytes.Buffer
	if err := service.RepoInfo(context.Background(), RepoInfoOptions{
		ProjectSlug: "huawei-zddi",
		Stdout:      &out,
	}); err != nil {
		t.Fatalf("RepoInfo() error = %v", err)
	}
	var info RepoInfo
	if err := json.Unmarshal(out.Bytes(), &info); err != nil {
		t.Fatalf("Unmarshal repo info error = %v, output=%q", err, out.String())
	}
	if info.ActiveSource != "remote" || info.Sources.Remote == nil || info.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("repo info = %#v", info)
	}
	var raw map[string]any
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		t.Fatalf("Unmarshal raw repo info error = %v", err)
	}
	if _, ok := raw["source"]; ok {
		t.Fatalf("repo info should not contain duplicate source field: %s", out.String())
	}
}

func TestRepoAddStoresLocalAndRemoteSourcesAndRepoUseSwitchesActiveSource(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		RemoteURL:   "http://devwiki.example.com:5697",
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}

	var remoteOut bytes.Buffer
	if err := service.RepoInfo(context.Background(), RepoInfoOptions{
		ProjectSlug: "huawei-zddi",
		Stdout:      &remoteOut,
	}); err != nil {
		t.Fatalf("RepoInfo(remote active) error = %v", err)
	}
	var remoteInfo RepoInfo
	if err := json.Unmarshal(remoteOut.Bytes(), &remoteInfo); err != nil {
		t.Fatalf("Unmarshal repo info error = %v, output=%q", err, remoteOut.String())
	}
	if remoteInfo.ActiveSource != devwiki.RepoSourceRemote || remoteInfo.Sources.Local == nil || remoteInfo.Sources.Remote == nil {
		t.Fatalf("remote active info = %#v", remoteInfo)
	}
	if remoteInfo.Sources.Local.Path != root || remoteInfo.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("sources = %#v", remoteInfo.Sources)
	}

	if err := service.RepoUse(context.Background(), RepoUseOptions{
		ProjectSlug: "huawei-zddi",
		SourceType:  devwiki.RepoSourceLocal,
	}); err != nil {
		t.Fatalf("RepoUse(local) error = %v", err)
	}

	var localOut bytes.Buffer
	if err := service.RepoInfo(context.Background(), RepoInfoOptions{
		ProjectSlug: "huawei-zddi",
		Stdout:      &localOut,
	}); err != nil {
		t.Fatalf("RepoInfo(local active) error = %v", err)
	}
	var localInfo RepoInfo
	if err := json.Unmarshal(localOut.Bytes(), &localInfo); err != nil {
		t.Fatalf("Unmarshal repo info error = %v, output=%q", err, localOut.String())
	}
	if localInfo.ActiveSource != devwiki.RepoSourceLocal || localInfo.Sources.Local == nil || localInfo.Sources.Local.Path != root {
		t.Fatalf("local active info = %#v", localInfo)
	}
	if localInfo.Sources.Remote == nil || localInfo.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("remote source was not preserved: %#v", localInfo.Sources.Remote)
	}
}

func TestRepoAddLocalWritesReadableSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	service := NewService()
	var out bytes.Buffer

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
		Stdout:      &out,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}

	if !strings.Contains(out.String(), "已添加 DevWiki project `huawei-zddi`") ||
		!strings.Contains(out.String(), "local: "+root) {
		t.Fatalf("RepoAdd(local) output = %q", out.String())
	}
}

func TestRepoLinkWritesReadableSuccess(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	var out bytes.Buffer

	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
		Stdout:      &out,
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	for _, want := range []string{
		"已绑定代码仓 `zddiv3` 到 DevWiki project `huawei-zddi`",
		"repo: " + codeRoot,
		"installed skills: devwiki-code, devwiki-code-to-doc, devwiki-query",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("RepoLink() output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRepoInitRequiresTTY(t *testing.T) {
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(t.TempDir()),
		IsTTY:     false,
	})

	err := service.RepoInit(context.Background(), RepoInitOptions{})

	if err == nil || !strings.Contains(err.Error(), "requires an interactive terminal") {
		t.Fatalf("RepoInit() error = %v, want interactive terminal error", err)
	}
}

func TestInstallRepoInitDocSkillsInstallsAllDevwikiSkillsForAgents(t *testing.T) {
	docRoot := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(docRoot),
		IsTTY:     false,
	})

	if err := service.installRepoInitDocSkills(docRoot, []string{"codex", "cursor"}, "zh"); err != nil {
		t.Fatalf("installRepoInitDocSkills() error = %v", err)
	}

	for _, rel := range []string{
		".agents/skills/devwiki-ingest/SKILL.md",
		".agents/skills/devwiki-topic/SKILL.md",
		".agents/skills/devwiki-workflow/SKILL.md",
		".agents/skills/devwiki-maintain/SKILL.md",
		".agents/skills/devwiki-code/SKILL.md",
		".agents/skills/devwiki-code-to-doc/SKILL.md",
		".agents/skills/devwiki-query/SKILL.md",
		".cursor/skills/devwiki-query/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(docRoot, rel)); err != nil {
			t.Fatalf("missing doc skill %s: %v", rel, err)
		}
	}
}

func TestRepoAddAndDocSkillInstallSkipsRemoteDocRoot(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	docRoot := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{
		Workspace: skills.NewWorkspace(docRoot),
		IsTTY:     false,
	})

	cfg, err := service.applyRepoInitSource(context.Background(), RepoInitSource{
		ProjectSlug: "remote-project",
		SourceType:  devwiki.RepoSourceRemote,
		RemoteURL:   "http://devwiki.example.com:5697",
		Agents:      []string{"codex", "cursor"},
	})
	if err != nil {
		t.Fatalf("applyRepoInitSource(remote) error = %v", err)
	}

	if cfg.ActiveSource != devwiki.RepoSourceRemote || cfg.Sources.Remote == nil || cfg.Sources.Remote.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("config = %#v, want remote source", cfg)
	}
	for _, rel := range []string{
		".agents/skills/devwiki-query/SKILL.md",
		".cursor/skills/devwiki-query/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(docRoot, rel)); err == nil {
			t.Fatalf("remote source should not install doc skill %s", rel)
		}
	}
}

func TestRepoLinkInstallsCodeRepoSkillsForAgentsOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	docRoot := t.TempDir()
	codeRoot := filepath.Join(t.TempDir(), "zddiv3")
	mustWriteFileDevwikiApp(t, filepath.Join(codeRoot, "go.mod"), "module example\n")
	service := NewService()

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "huawei-zddi",
		LocalPath:   docRoot,
	}); err != nil {
		t.Fatalf("RepoAdd(local) error = %v", err)
	}
	if err := service.RepoLink(context.Background(), RepoLinkOptions{
		ProjectSlug: "huawei-zddi",
		RepoSlug:    "zddiv3",
		Path:        codeRoot,
		Agents:      []string{"codex", "cursor"},
	}); err != nil {
		t.Fatalf("RepoLink() error = %v", err)
	}

	for _, rel := range []string{
		".agents/skills/devwiki-code/SKILL.md",
		".agents/skills/devwiki-code-to-doc/SKILL.md",
		".agents/skills/devwiki-query/SKILL.md",
		".cursor/skills/devwiki-code/SKILL.md",
		".cursor/skills/devwiki-code-to-doc/SKILL.md",
		".cursor/skills/devwiki-query/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(codeRoot, rel)); err != nil {
			t.Fatalf("missing code repo skill %s: %v", rel, err)
		}
	}
	for _, rel := range []string{
		".agents/skills/devwiki-ingest/SKILL.md",
		".agents/skills/devwiki-maintain/SKILL.md",
		".cursor/skills/devwiki-ingest/SKILL.md",
		".cursor/skills/devwiki-maintain/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(codeRoot, rel)); err == nil {
			t.Fatalf("code repo should not install doc-root skill %s", rel)
		}
	}
}

func TestRepoInfoWithoutProjectPrintsOnlyProjectNames(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	service := NewService()

	for _, project := range []string{"zeta", "alpha"} {
		if err := service.RepoAdd(context.Background(), RepoAddOptions{
			ProjectSlug: project,
			RemoteURL:   "http://devwiki.example.com:5697/" + project,
		}); err != nil {
			t.Fatalf("RepoAdd(%s) error = %v", project, err)
		}
	}

	var out bytes.Buffer
	if err := service.RepoInfo(context.Background(), RepoInfoOptions{
		Stdout: &out,
	}); err != nil {
		t.Fatalf("RepoInfo() error = %v", err)
	}
	var got []string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal repo info list error = %v, output=%q", err, out.String())
	}
	want := []string{"alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("repo info projects = %#v, want %#v", got, want)
	}
}
