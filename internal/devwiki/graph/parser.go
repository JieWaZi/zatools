package graph

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Title               string   `yaml:"title"`
	Slug                string   `yaml:"slug"`
	Status              string   `yaml:"status"`
	Summary             string   `yaml:"summary"`
	Features            []string `yaml:"features"`
	Capabilities        []string `yaml:"capabilities"`
	Workflow            any      `yaml:"workflow"`
	RelatedCapabilities []string `yaml:"related_capabilities"`
	RelatedFeatures     []string `yaml:"related_features"`
	RelatedWorkflows    []string `yaml:"related_workflows"`
	Confidence          string   `yaml:"confidence"`
	SearchTerms         []string `yaml:"search_terms"`
}

// LoadPages parses all capability, feature, and workflow pages from a DevWiki root.
func LoadPages(root string) ([]Page, []Issue, error) {
	var pages []Page
	var issues []Issue
	for _, spec := range []struct {
		dir string
		typ PageType
	}{
		{"wiki/capabilities", PageTypeCapability},
		{"wiki/features", PageTypeFeature},
		{"wiki/workflows", PageTypeWorkflow},
	} {
		loaded, pageIssues, err := loadPagesInDir(root, spec.dir, spec.typ)
		if err != nil {
			return nil, nil, err
		}
		pages = append(pages, loaded...)
		issues = append(issues, pageIssues...)
	}
	sort.Slice(pages, func(i, j int) bool {
		if pages[i].Type != pages[j].Type {
			return pages[i].Type < pages[j].Type
		}
		return pages[i].Slug < pages[j].Slug
	})
	return pages, issues, nil
}

func loadPagesInDir(root string, relDir string, typ PageType) ([]Page, []Issue, error) {
	dir := filepath.Join(root, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var pages []Page
	var issues []Issue
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(relDir, entry.Name()))
		page, pageIssues, err := parsePage(root, rel, typ)
		if err != nil {
			return nil, nil, err
		}
		pages = append(pages, page)
		issues = append(issues, pageIssues...)
	}
	return pages, issues, nil
}

func parsePage(root string, rel string, typ PageType) (Page, []Issue, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return Page{}, nil, err
	}
	raw, ok := extractFrontmatter(data)
	if !ok {
		return Page{}, []Issue{{Level: IssueWarning, Path: rel, Message: "missing YAML frontmatter"}}, nil
	}
	var fm frontmatter
	if err := yaml.Unmarshal(raw, &fm); err != nil {
		return Page{}, nil, fmt.Errorf("%s: parse frontmatter: %w", rel, err)
	}
	issues := make([]Issue, 0, 3)
	slug := normalizeReference(fm.Slug)
	if slug == "" {
		slug = strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
		issues = append(issues, Issue{Level: IssueWarning, Path: rel, Message: "missing slug; inferred from filename"})
	}
	title := strings.TrimSpace(fm.Title)
	if title == "" {
		title = slug
		issues = append(issues, Issue{Level: IssueWarning, Path: rel, Message: "missing title; using slug"})
	}
	if strings.TrimSpace(fm.Summary) == "" {
		issues = append(issues, Issue{Level: IssueWarning, Path: rel, Message: "missing summary"})
	}
	workflows, err := normalizeWorkflowField(fm.Workflow)
	if err != nil {
		return Page{}, nil, fmt.Errorf("%s: %w", rel, err)
	}
	return Page{
		Type:                typ,
		Path:                rel,
		Slug:                slug,
		Title:               title,
		Summary:             strings.TrimSpace(fm.Summary),
		Status:              strings.TrimSpace(fm.Status),
		Confidence:          strings.TrimSpace(fm.Confidence),
		SearchTerms:         trimStrings(fm.SearchTerms),
		Features:            normalizeReferences(fm.Features),
		Capabilities:        normalizeReferences(fm.Capabilities),
		Workflows:           workflows,
		RelatedCapabilities: normalizeReferences(fm.RelatedCapabilities),
		RelatedFeatures:     normalizeReferences(fm.RelatedFeatures),
		RelatedWorkflows:    normalizeReferences(fm.RelatedWorkflows),
	}, issues, nil
}

func extractFrontmatter(data []byte) ([]byte, bool) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, false
	}
	rest := data[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return nil, false
	}
	return rest[:end], true
}

func normalizeWorkflowField(value any) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		ref := normalizeReference(v)
		if ref == "" {
			return nil, nil
		}
		return []string{ref}, nil
	case []any:
		values := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("workflow entries must be strings")
			}
			if ref := normalizeReference(s); ref != "" {
				values = append(values, ref)
			}
		}
		return values, nil
	default:
		return nil, fmt.Errorf("workflow must be a string or list")
	}
}

func normalizeReferences(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		ref := normalizeReference(value)
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

func normalizeReference(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[[")
	value = strings.TrimSuffix(value, "]]")
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".md")
	value = filepath.ToSlash(value)
	if strings.Contains(value, "/") {
		value = pathBase(value)
	}
	return strings.TrimSpace(value)
}

func pathBase(value string) string {
	parts := strings.Split(value, "/")
	return parts[len(parts)-1]
}

func trimStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
