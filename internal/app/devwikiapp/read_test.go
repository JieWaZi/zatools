package devwikiapp

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/skills"
)

func TestReadTopicCardIncludesFilteredMetadataAndCardSection(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/topics/vip.md", `---
title: "VIP"
slug: "vip"
kind: topic
status: draft
summary: "VIP topic"
formatter: "markdown"
aliases: ["vip-service"]
confidence: high
---
# VIP

<!-- devwiki:section id=card -->
## 导航卡
card body
<!-- /devwiki:section -->

<!-- devwiki:section id=core -->
## 核心内容
core body
<!-- /devwiki:section -->

<!-- devwiki:section id=explain -->
## 详细说明
explain body
<!-- /devwiki:section -->
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	if err := service.Read(context.Background(), ReadOptions{
		Root:   root,
		Kind:   "topic",
		Slug:   "vip",
		View:   "card",
		Format: "text",
		Stdout: &out,
	}); err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `title: VIP`) ||
		!strings.Contains(got, `status: draft`) ||
		!strings.Contains(got, `summary: VIP topic`) ||
		!strings.Contains(got, `confidence: high`) ||
		!strings.Contains(got, "card body") {
		t.Fatalf("card output = %q", got)
	}
	if strings.Contains(got, `slug: "vip"`) || strings.Contains(got, "kind: topic") || strings.Contains(got, `formatter: "markdown"`) || strings.Contains(got, "aliases:") {
		t.Fatalf("card output should filter extra frontmatter: %q", got)
	}
	if strings.Contains(got, "core body") {
		t.Fatalf("card output should not include core: %q", got)
	}
}

func TestReadWorkflowCoreOmitsFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/workflows/vip-runtime.md", `---
title: "VIP Runtime"
slug: "vip-runtime"
kind: workflow
summary: "VIP runtime"
topics: ["vip"]
---
# VIP Runtime

<!-- devwiki:section id=card -->
## 工程卡
card body
<!-- /devwiki:section -->

<!-- devwiki:section id=core -->
## 代码定位
core body
<!-- /devwiki:section -->

<!-- devwiki:section id=explain -->
## 实现说明
explain body
<!-- /devwiki:section -->
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	if err := service.Read(context.Background(), ReadOptions{
		Root:   root,
		Kind:   "workflow",
		Slug:   "vip-runtime",
		View:   "core",
		Format: "text",
		Stdout: &out,
	}); err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	got := out.String()
	if strings.Contains(got, "---") || strings.Contains(got, "card body") || !strings.Contains(got, "core body") {
		t.Fatalf("core output = %q", got)
	}
}

func TestReadRejectsUnsupportedFormat(t *testing.T) {
	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Read(context.Background(), ReadOptions{
		Root:   root,
		Kind:   "topic",
		Slug:   "vip",
		View:   "card",
		Format: "json",
		Stdout: &out,
	})
	if err == nil || !strings.Contains(err.Error(), "only text is supported") {
		t.Fatalf("Read() error = %v, want unsupported format", err)
	}
}

func TestReadUsesLocalProjectConfigWhenProjectProvided(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/topics/vip.md", `---
title: "VIP"
slug: "vip"
kind: topic
summary: "VIP topic"
---
# VIP

<!-- devwiki:section id=card -->
card body
<!-- /devwiki:section -->
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(t.TempDir())})
	var out bytes.Buffer

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd() error = %v", err)
	}
	if err := service.Read(context.Background(), ReadOptions{
		Project: "sample",
		Kind:    "topic",
		Slug:    "vip",
		View:    "card",
		Format:  "text",
		Stdout:  &out,
	}); err != nil {
		t.Fatalf("Read(project) error = %v", err)
	}
	if !strings.Contains(out.String(), "card body") {
		t.Fatalf("Read(project) output = %q", out.String())
	}
}

func TestReadUsesRemoteProjectConfigWhenProjectProvided(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/devwiki/read" || r.Method != http.MethodPost {
			t.Fatalf("unexpected remote request %s %s", r.Method, r.URL.Path)
		}
		user, password, ok := r.BasicAuth()
		if !ok || user != "devwiki" || password != "T19xwxc3n2I38F1A" {
			t.Fatalf("unexpected basic auth user=%q password=%q ok=%v", user, password, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"remote card body"}`))
	}))
	defer server.Close()

	service := NewService()
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		RemoteURL:   server.URL,
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}
	var out bytes.Buffer
	if err := service.Read(context.Background(), ReadOptions{
		Project: "sample",
		Kind:    "topic",
		Slug:    "vip",
		View:    "card",
		Format:  "text",
		Stdout:  &out,
	}); err != nil {
		t.Fatalf("Read(remote project) error = %v", err)
	}
	if strings.TrimSpace(out.String()) != "remote card body" {
		t.Fatalf("Read(remote project) output = %q", out.String())
	}
}

func writeDevwikiReadFixture(t *testing.T, root string, rel string, content string) {
	t.Helper()
	mustWriteFileDevwikiApp(t, filepath.Join(root, filepath.FromSlash(rel)), content)
}
