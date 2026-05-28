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
	if info.ProjectSlug != "huawei-zddi" || info.Source.Type != "local" || info.Source.Path != root {
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
	} {
		if _, err := os.Stat(filepath.Join(codeRoot, rel)); err != nil {
			t.Fatalf("missing installed code repo skill %s: %v", rel, err)
		}
	}
	for _, rel := range []string{
		".agents/skills/devwiki-query/SKILL.md",
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
	if info.Source.Type != "remote" || info.Source.URL != "http://devwiki.example.com:5697" {
		t.Fatalf("repo info source = %#v", info.Source)
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
		"installed skills: devwiki-code, devwiki-code-to-doc",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("RepoLink() output missing %q:\n%s", want, out.String())
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
