package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadPagesParsesCurrentTemplateFields(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/capabilities/ha.md", `---
title: "高可用能力"
slug: "ha"
status: active
summary: "HA capability"
features:
  - "[[vip-failover]]"
related_capabilities:
  - "dns-availability"
confidence: high
search_terms:
  - "HA"
---
# 高可用能力
`)
	writeGraphFile(t, root, "wiki/features/vip-failover.md", `---
title: "VIP 接管"
slug: "vip-failover"
status: active
summary: "VIP failover"
capabilities:
  - "wiki/capabilities/ha.md"
workflow: "workflow-vip-failover"
related_features:
  - "[[brain-split]]"
confidence: medium
search_terms:
  - "vip"
---
# VIP 接管
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-vip-failover.md", `---
title: "VIP 接管实现"
slug: "workflow-vip-failover"
status: active
summary: "VIP workflow"
features:
  - "vip-failover"
related_workflows:
  - "workflow-ha-state"
confidence: medium
search_terms:
  - "keepalived"
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

	feature := findPage(t, pages, PageTypeFeature, "vip-failover")
	if feature.Title != "VIP 接管" || feature.Summary != "VIP failover" {
		t.Fatalf("feature title/summary = %q/%q", feature.Title, feature.Summary)
	}
	if !reflect.DeepEqual(feature.Capabilities, []string{"ha"}) {
		t.Fatalf("feature capabilities = %#v, want ha", feature.Capabilities)
	}
	if !reflect.DeepEqual(feature.Workflows, []string{"workflow-vip-failover"}) {
		t.Fatalf("feature workflows = %#v", feature.Workflows)
	}
}

func TestLoadPagesFallsBackToFilenameSlugAndWarns(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/features/no-slug.md", `---
title: "No Slug"
summary: "missing slug"
---
# No Slug
`)

	pages, issues, err := LoadPages(root)
	if err != nil {
		t.Fatalf("LoadPages() error = %v", err)
	}
	page := findPage(t, pages, PageTypeFeature, "no-slug")
	if page.Slug != "no-slug" {
		t.Fatalf("Slug = %q, want no-slug", page.Slug)
	}
	if !hasIssue(issues, IssueWarning, "missing slug") {
		t.Fatalf("issues = %#v, want missing slug warning", issues)
	}
}

func TestLoadPagesRejectsInvalidYAML(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/features/broken.md", "---\ntitle: [broken\n---\n# Broken\n")

	_, _, err := LoadPages(root)
	if err == nil {
		t.Fatal("LoadPages() error = nil, want YAML error")
	}
}

func TestNormalizeReference(t *testing.T) {
	tests := map[string]string{
		"vip-failover":                  "vip-failover",
		"[[vip-failover]]":              "vip-failover",
		"wiki/features/vip-failover.md": "vip-failover",
		" feature-vip ":                 "feature-vip",
	}
	for input, want := range tests {
		if got := normalizeReference(input); got != want {
			t.Fatalf("normalizeReference(%q) = %q, want %q", input, got, want)
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
