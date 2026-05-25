package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"zatools/internal/devwiki/page"
)

func TestLoadPagesParsesCurrentTemplateFields(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/vip-failover.md", `---
title: "VIP 接管"
slug: "vip-failover"
kind: topic
status: active
summary: "VIP topic"
workflows:
  - "workflow-vip-failover"
related_topics:
  - "[[ha-state]]"
confidence: medium
---
# VIP 接管
`)
	writeGraphFile(t, root, "wiki/topics/ha-state.md", `---
title: "高可用状态"
slug: "ha-state"
kind: topic
status: active
summary: "HA state"
---
# 高可用状态
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-vip-failover.md", `---
title: "VIP 接管实现"
slug: "workflow-vip-failover"
kind: workflow
status: active
summary: "VIP workflow"
topics:
  - "vip-failover"
related_workflows:
  - "workflow-ha-state"
confidence: medium
---
# VIP 接管实现
`)

	pages, issues, err := LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages() error = %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("LoadPages() issues = %#v, want none", issues)
	}
	if len(pages) != 3 {
		t.Fatalf("len(pages) = %d, want 3", len(pages))
	}

	topic := findPage(t, pages, PageTypeTopic, "vip-failover")
	if topic.Title != "VIP 接管" || topic.Summary != "VIP topic" {
		t.Fatalf("topic title/summary = %q/%q", topic.Title, topic.Summary)
	}
	if !reflect.DeepEqual(topic.RelatedTopics, []string{"ha-state"}) {
		t.Fatalf("topic related_topics = %#v, want ha-state", topic.RelatedTopics)
	}
	if !reflect.DeepEqual(topic.Workflows, []string{"workflow-vip-failover"}) {
		t.Fatalf("topic workflows = %#v", topic.Workflows)
	}
	workflow := findPage(t, pages, PageTypeWorkflow, "workflow-vip-failover")
	if !reflect.DeepEqual(workflow.Topics, []string{"vip-failover"}) {
		t.Fatalf("workflow topics = %#v", workflow.Topics)
	}
}

func TestLoadPagesFallsBackToFilenameSlugAndWarns(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/no-slug.md", `---
title: "No Slug"
summary: "missing slug"
---
# No Slug
`)

	pages, issues, err := LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages() error = %v", err)
	}
	page := findPage(t, pages, PageTypeTopic, "no-slug")
	if page.Slug != "no-slug" {
		t.Fatalf("Slug = %q, want no-slug", page.Slug)
	}
	if !hasIssue(issues, IssueWarning, "missing slug") {
		t.Fatalf("issues = %#v, want missing slug warning", issues)
	}
}

func TestLoadPagesRejectsInvalidYAML(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/broken.md", "---\ntitle: [broken\n---\n# Broken\n")

	_, _, err := LoadPages(root)
	if err == nil {
		t.Fatal("LoadPages() error = nil, want YAML error")
	}
}

func TestNormalizeReference(t *testing.T) {
	tests := map[string]string{
		"vip-failover":                "vip-failover",
		"[[vip-failover]]":            "vip-failover",
		"wiki/topics/vip-failover.md": "vip-failover",
		" topic-vip ":                 "topic-vip",
	}
	for input, want := range tests {
		if got := page.NormalizeReference(input); got != want {
			t.Fatalf("NormalizeReference(%q) = %q, want %q", input, got, want)
		}
	}
}

func writeGraphFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func findPage(t *testing.T, pages []Page, typ PageType, slug string) Page {
	t.Helper()
	for _, page := range pages {
		if page.Type == typ && page.Slug == slug {
			return page
		}
	}
	t.Fatalf("missing page %s:%s in %#v", typ, slug, pages)
	return Page{}
}

func hasIssue(issues []Issue, level IssueLevel, contains string) bool {
	for _, issue := range issues {
		if issue.Level == level && strings.Contains(issue.Message, contains) {
			return true
		}
	}
	return false
}
