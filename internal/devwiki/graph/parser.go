package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"zatools/internal/devwiki/page"
)

type frontmatter struct {
	Title            string   `yaml:"title"`
	Slug             string   `yaml:"slug"`
	Kind             string   `yaml:"kind"`
	Status           string   `yaml:"status"`
	Summary          string   `yaml:"summary"`
	Workflows        []string `yaml:"workflows"`
	Topics           []string `yaml:"topics"`
	RelatedTopics    []string `yaml:"related_topics"`
	RelatedWorkflows []string `yaml:"related_workflows"`
	Confidence       string   `yaml:"confidence"`
	SearchTerms      []string `yaml:"search_terms"`
}

// LoadPages parses all topic and workflow pages from a DevWiki root.
func LoadPages(root string) ([]Page, []Issue, error) {
	var pages []Page
	var issues []Issue
	for _, spec := range []struct {
		dir string
		typ PageType
	}{
		{"wiki/topics", PageTypeTopic},
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
	slug := page.NormalizeReference(fm.Slug)
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
	if strings.TrimSpace(fm.Kind) != "" && PageType(strings.TrimSpace(fm.Kind)) != typ {
		issues = append(issues, Issue{Level: IssueWarning, Path: rel, Message: fmt.Sprintf("kind %q does not match directory type %q", fm.Kind, typ)})
	}
	return Page{
		Type:             typ,
		Path:             rel,
		Slug:             slug,
		Title:            title,
		Summary:          strings.TrimSpace(fm.Summary),
		Status:           strings.TrimSpace(fm.Status),
		Confidence:       strings.TrimSpace(fm.Confidence),
		SearchTerms:      trimStrings(fm.SearchTerms),
		Workflows:        page.NormalizeReferences(fm.Workflows),
		Topics:           page.NormalizeReferences(fm.Topics),
		RelatedTopics:    page.NormalizeReferences(fm.RelatedTopics),
		RelatedWorkflows: page.NormalizeReferences(fm.RelatedWorkflows),
	}, issues, nil
}

func extractFrontmatter(data []byte) ([]byte, bool) {
	raw, _, ok := page.ExtractFrontmatter(data)
	return raw, ok
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
