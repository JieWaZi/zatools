package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Build parses, validates, and renders a DevWiki graph from a root directory.
func Build(root string) (Graph, []Issue, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Graph{}, nil, err
	}
	pages, issues, err := LoadPages(absRoot)
	if err != nil {
		return Graph{}, issues, err
	}
	if len(pages) == 0 {
		issues = append(issues, Issue{Level: IssueError, Message: "no capability, feature, or workflow pages found"})
		return Graph{}, issues, errors.New("no graph input pages found")
	}

	index := map[PageType]map[string]Page{
		PageTypeCapability: {},
		PageTypeFeature:    {},
		PageTypeWorkflow:   {},
	}
	for _, page := range pages {
		if _, exists := index[page.Type][page.Slug]; exists {
			issues = append(issues, Issue{
				Level:   IssueError,
				Path:    page.Path,
				Message: fmt.Sprintf("duplicate slug %q for %s", page.Slug, page.Type),
			})
			continue
		}
		index[page.Type][page.Slug] = page
	}
	if hasErrors(issues) {
		return Graph{}, issues, errors.New("graph validation failed")
	}

	graph := Graph{
		SchemaVersion: SchemaVersion,
		Project:       loadProject(absRoot),
		BuiltAt:       time.Now(),
		Documents:     map[string]Document{},
	}
	for _, page := range pages {
		node := pageNode(page)
		graph.Nodes = append(graph.Nodes, node)
		graph.Documents[node.ID] = Document{
			Type:    page.Type,
			Path:    page.Path,
			Title:   node.Title,
			Summary: node.Summary,
		}
		if page.Status == "" {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: "missing status"})
		}
		if page.Confidence == "" {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: "missing confidence"})
		}
	}

	edgeMap := map[string]*Edge{}
	for _, page := range pages {
		switch page.Type {
		case PageTypeCapability:
			issues = append(issues, buildCapabilityEdges(edgeMap, index, page)...)
		case PageTypeFeature:
			issues = append(issues, buildFeatureEdges(edgeMap, index, page)...)
		case PageTypeWorkflow:
			issues = append(issues, buildWorkflowEdges(edgeMap, index, page)...)
		}
	}
	if hasErrors(issues) {
		return Graph{}, issues, errors.New("graph validation failed")
	}

	for _, edge := range edgeMap {
		sort.Strings(edge.Sources)
		graph.Edges = append(graph.Edges, *edge)
	}
	sort.Slice(graph.Edges, func(i, j int) bool { return graph.Edges[i].ID < graph.Edges[j].ID })
	graph.Warnings = filterIssues(issues, IssueWarning)
	return graph, issues, nil
}

func buildCapabilityEdges(edgeMap map[string]*Edge, index map[PageType]map[string]Page, page Page) []Issue {
	var issues []Issue
	for _, feature := range page.Features {
		target, ok := index[PageTypeFeature][feature]
		if !ok {
			issues = append(issues, Issue{Level: IssueError, Path: page.Path, Message: fmt.Sprintf("missing feature %q", feature)})
			continue
		}
		addEdge(edgeMap, "contains", nodeID(PageTypeCapability, page.Slug), nodeID(PageTypeFeature, feature), "包含功能", page.Path)
		if !contains(target.Capabilities, page.Slug) {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("reverse relation missing: feature %q does not list capability %q", feature, page.Slug)})
		}
	}
	for _, related := range page.RelatedCapabilities {
		if _, ok := index[PageTypeCapability][related]; !ok {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("related capability %q does not exist", related)})
			continue
		}
		addUndirectedEdge(edgeMap, "related", nodeID(PageTypeCapability, page.Slug), nodeID(PageTypeCapability, related), "相关能力", page.Path)
	}
	return issues
}

func buildFeatureEdges(edgeMap map[string]*Edge, index map[PageType]map[string]Page, page Page) []Issue {
	var issues []Issue
	for _, capability := range page.Capabilities {
		target, ok := index[PageTypeCapability][capability]
		if !ok {
			issues = append(issues, Issue{Level: IssueError, Path: page.Path, Message: fmt.Sprintf("missing capability %q", capability)})
			continue
		}
		addEdge(edgeMap, "contains", nodeID(PageTypeCapability, capability), nodeID(PageTypeFeature, page.Slug), "包含功能", page.Path)
		if !contains(target.Features, page.Slug) {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("reverse relation missing: capability %q does not list feature %q", capability, page.Slug)})
		}
	}
	for _, workflow := range page.Workflows {
		target, ok := index[PageTypeWorkflow][workflow]
		if !ok {
			issues = append(issues, Issue{Level: IssueError, Path: page.Path, Message: fmt.Sprintf("missing workflow %q", workflow)})
			continue
		}
		addEdge(edgeMap, "implemented_by", nodeID(PageTypeFeature, page.Slug), nodeID(PageTypeWorkflow, workflow), "实现流程", page.Path)
		if !contains(target.Features, page.Slug) {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("reverse relation missing: workflow %q does not list feature %q", workflow, page.Slug)})
		}
	}
	for _, related := range page.RelatedFeatures {
		if _, ok := index[PageTypeFeature][related]; !ok {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("related feature %q does not exist", related)})
			continue
		}
		addUndirectedEdge(edgeMap, "related", nodeID(PageTypeFeature, page.Slug), nodeID(PageTypeFeature, related), "相关功能", page.Path)
	}
	return issues
}

func buildWorkflowEdges(edgeMap map[string]*Edge, index map[PageType]map[string]Page, page Page) []Issue {
	var issues []Issue
	for _, feature := range page.Features {
		target, ok := index[PageTypeFeature][feature]
		if !ok {
			issues = append(issues, Issue{Level: IssueError, Path: page.Path, Message: fmt.Sprintf("missing feature %q", feature)})
			continue
		}
		addEdge(edgeMap, "implemented_by", nodeID(PageTypeFeature, feature), nodeID(PageTypeWorkflow, page.Slug), "实现流程", page.Path)
		if !contains(target.Workflows, page.Slug) {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("reverse relation missing: feature %q does not list workflow %q", feature, page.Slug)})
		}
	}
	for _, related := range page.RelatedWorkflows {
		if _, ok := index[PageTypeWorkflow][related]; !ok {
			issues = append(issues, Issue{Level: IssueWarning, Path: page.Path, Message: fmt.Sprintf("related workflow %q does not exist", related)})
			continue
		}
		addUndirectedEdge(edgeMap, "related", nodeID(PageTypeWorkflow, page.Slug), nodeID(PageTypeWorkflow, related), "相关流程", page.Path)
	}
	return issues
}

func pageNode(page Page) Node {
	return Node{
		ID:          nodeID(page.Type, page.Slug),
		Type:        page.Type,
		Slug:        page.Slug,
		Title:       page.Title,
		Summary:     page.Summary,
		Status:      page.Status,
		Confidence:  page.Confidence,
		Path:        page.Path,
		SearchTerms: page.SearchTerms,
	}
}

func nodeID(typ PageType, slug string) string {
	return string(typ) + ":" + slug
}

func addEdge(edges map[string]*Edge, typ string, source string, target string, label string, path string) {
	id := source + "->" + target
	edge, ok := edges[id]
	if !ok {
		edge = &Edge{ID: id, Type: typ, Source: source, Target: target, Label: label}
		edges[id] = edge
	}
	if !contains(edge.Sources, path) {
		edge.Sources = append(edge.Sources, path)
	}
}

func addUndirectedEdge(edges map[string]*Edge, typ string, a string, b string, label string, path string) {
	source, target := a, b
	if source > target {
		source, target = target, source
	}
	addEdge(edges, typ, source, target, label, path)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasErrors(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Level == IssueError {
			return true
		}
	}
	return false
}

func filterIssues(issues []Issue, level IssueLevel) []Issue {
	out := make([]Issue, 0)
	for _, issue := range issues {
		if issue.Level == level {
			out = append(out, issue)
		}
	}
	return out
}

func loadProject(root string) Project {
	project := Project{Name: filepath.Base(root), Slug: filepath.Base(root), Root: root}
	data, err := os.ReadFile(filepath.Join(root, "config", "project.yaml"))
	if err != nil {
		return project
	}
	var parsed struct {
		ProjectName string `yaml:"project_name"`
		ProjectSlug string `yaml:"project_slug"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return project
	}
	if strings.TrimSpace(parsed.ProjectName) != "" {
		project.Name = strings.TrimSpace(parsed.ProjectName)
	}
	if strings.TrimSpace(parsed.ProjectSlug) != "" {
		project.Slug = strings.TrimSpace(parsed.ProjectSlug)
	}
	return project
}
