package graph

import "testing"

func TestBuildGraphCreatesNormalizedEdges(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/topics/vip.md", `---
title: "VIP"
slug: "vip"
kind: topic
status: active
summary: "VIP"
workflows: ["workflow-vip"]
related_topics: ["dns"]
---
# VIP
`)
	writeGraphFile(t, root, "wiki/topics/dns.md", `---
title: "DNS"
slug: "dns"
kind: topic
status: active
summary: "DNS"
related_topics: ["vip"]
---
# DNS
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-vip.md", `---
title: "VIP Workflow"
slug: "workflow-vip"
kind: workflow
status: active
summary: "VIP Workflow"
topics: ["vip"]
related_workflows: ["workflow-ha"]
---
# VIP Workflow
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-ha.md", `---
title: "HA Workflow"
slug: "workflow-ha"
kind: workflow
status: active
summary: "HA Workflow"
---
# HA Workflow
`)

	graph, issues, err := Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if hasIssue(issues, IssueError, "") {
		t.Fatalf("issues contain error: %#v", issues)
	}
	assertEdge(t, graph, "implemented_by", "topic:vip", "workflow:workflow-vip")
	assertEdge(t, graph, "related", "topic:dns", "topic:vip")
	assertEdge(t, graph, "related", "workflow:workflow-ha", "workflow:workflow-vip")
}

func TestBuildReportsDuplicateSlugAsError(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/topics/a.md", "---\ntitle: A\nslug: same\nsummary: A\n---\n")
	writeGraphFile(t, root, "wiki/topics/b.md", "---\ntitle: B\nslug: same\nsummary: B\n---\n")

	_, issues, err := Build(root)
	if err == nil {
		t.Fatal("Build() error = nil, want duplicate slug error")
	}
	if !hasIssue(issues, IssueError, "duplicate slug") {
		t.Fatalf("issues = %#v, want duplicate slug error", issues)
	}
}

func TestBuildReportsMissingMainRelationAsError(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/topics/vip.md", "---\ntitle: VIP\nslug: vip\nsummary: VIP\nworkflows: [missing]\n---\n")

	_, issues, err := Build(root)
	if err == nil {
		t.Fatal("Build() error = nil, want missing workflow error")
	}
	if !hasIssue(issues, IssueError, "missing workflow") {
		t.Fatalf("issues = %#v, want missing workflow error", issues)
	}
}

func TestBuildWarnsForMissingReverseRelation(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/topics/vip.md", "---\ntitle: VIP\nslug: vip\nsummary: VIP\nworkflows: [workflow-vip]\n---\n")
	writeGraphFile(t, root, "wiki/workflows/workflow-vip.md", "---\ntitle: VIP Workflow\nslug: workflow-vip\nsummary: VIP Workflow\n---\n")

	_, issues, err := Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !hasIssue(issues, IssueWarning, "reverse relation") {
		t.Fatalf("issues = %#v, want reverse relation warning", issues)
	}
}

func assertEdge(t *testing.T, graph Graph, typ string, source string, target string) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.Type == typ && edge.Source == source && edge.Target == target {
			return
		}
	}
	t.Fatalf("missing edge %s %s -> %s in %#v", typ, source, target, graph.Edges)
}
