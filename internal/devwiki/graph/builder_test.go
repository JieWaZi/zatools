package graph

import "testing"

func TestBuildGraphCreatesNormalizedEdges(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/capabilities/ha.md", `---
title: "HA"
slug: "ha"
status: active
summary: "HA"
features: ["vip"]
related_capabilities: ["dns"]
---
# HA
`)
	writeGraphFile(t, root, "wiki/capabilities/dns.md", `---
title: "DNS"
slug: "dns"
status: active
summary: "DNS"
related_capabilities: ["ha"]
---
# DNS
`)
	writeGraphFile(t, root, "wiki/features/vip.md", `---
title: "VIP"
slug: "vip"
status: active
summary: "VIP"
capabilities: ["ha"]
workflow: "workflow-vip"
related_features: ["brain-split"]
---
# VIP
`)
	writeGraphFile(t, root, "wiki/features/brain-split.md", `---
title: "Brain Split"
slug: "brain-split"
status: active
summary: "Brain Split"
---
# Brain Split
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-vip.md", `---
title: "VIP Workflow"
slug: "workflow-vip"
status: active
summary: "VIP Workflow"
features: ["vip"]
---
# VIP Workflow
`)

	graph, issues, err := Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if hasIssue(issues, IssueError, "") {
		t.Fatalf("issues contain error: %#v", issues)
	}
	assertEdge(t, graph, "contains", "capability:ha", "feature:vip")
	assertEdge(t, graph, "implemented_by", "feature:vip", "workflow:workflow-vip")
	assertEdge(t, graph, "related", "capability:dns", "capability:ha")
	assertEdge(t, graph, "related", "feature:brain-split", "feature:vip")
}

func TestBuildReportsDuplicateSlugAsError(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Sample\nproject_slug: sample\n")
	writeGraphFile(t, root, "wiki/features/a.md", "---\ntitle: A\nslug: same\nsummary: A\n---\n")
	writeGraphFile(t, root, "wiki/features/b.md", "---\ntitle: B\nslug: same\nsummary: B\n---\n")

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
	writeGraphFile(t, root, "wiki/features/vip.md", "---\ntitle: VIP\nslug: vip\nsummary: VIP\nworkflow: missing\n---\n")

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
	writeGraphFile(t, root, "wiki/capabilities/ha.md", "---\ntitle: HA\nslug: ha\nsummary: HA\nfeatures: [vip]\n---\n")
	writeGraphFile(t, root, "wiki/features/vip.md", "---\ntitle: VIP\nslug: vip\nsummary: VIP\n---\n")

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
