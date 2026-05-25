// Package page parses DevWiki Markdown pages.
package page

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// KindTopic identifies a topic page.
	KindTopic = "topic"
	// KindWorkflow identifies a workflow page.
	KindWorkflow = "workflow"
	// KindTroubleshooting identifies a troubleshooting page.
	KindTroubleshooting = "troubleshooting"
)

var sectionStartPattern = regexp.MustCompile(`^<!--\s*devwiki:section\s+id=([A-Za-z0-9_-]+)\s*-->\s*$`)

// Section is one marked DevWiki page section.
type Section struct {
	ID        string
	Title     string
	Content   string
	StartLine int
	EndLine   int
}

// Meta is the normalized YAML frontmatter used by topic/workflow pages.
type Meta struct {
	Title            string   `yaml:"title"`
	Slug             string   `yaml:"slug"`
	Kind             string   `yaml:"kind"`
	Status           string   `yaml:"status"`
	Summary          string   `yaml:"summary"`
	Aliases          []string `yaml:"aliases"`
	Workflows        []string `yaml:"workflows"`
	Topics           []string `yaml:"topics"`
	RelatedTopics    []string `yaml:"related_topics"`
	RelatedWorkflows []string `yaml:"related_workflows"`
	Troubleshooting  []string `yaml:"troubleshooting"`
	Confidence       string   `yaml:"confidence"`
	LastVerifiedAt   string   `yaml:"last_verified_at"`
}

// Document is a parsed DevWiki Markdown page.
type Document struct {
	Path     string
	Meta     Meta
	RawMeta  []byte
	Sections []Section
}

// Parse parses one Markdown page.
func Parse(rel string, data []byte) (Document, error) {
	rawMeta, body, ok := ExtractFrontmatter(data)
	if !ok {
		return Document{}, fmt.Errorf("%s: missing YAML frontmatter", rel)
	}
	var meta Meta
	if err := yaml.Unmarshal(rawMeta, &meta); err != nil {
		return Document{}, fmt.Errorf("%s: parse frontmatter: %w", rel, err)
	}
	sections, err := ParseSections(body)
	if err != nil {
		return Document{}, fmt.Errorf("%s: %w", rel, err)
	}
	return Document{
		Path:     rel,
		Meta:     normalizeMeta(meta),
		RawMeta:  rawMeta,
		Sections: sections,
	}, nil
}

// Load reads and parses one Markdown page from root-relative path rel.
func Load(root string, rel string) (Document, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return Document{}, err
	}
	return Parse(rel, data)
}

// ExtractFrontmatter returns YAML frontmatter and Markdown body.
func ExtractFrontmatter(data []byte) ([]byte, []byte, bool) {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, data, false
	}
	rest := data[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return nil, data, false
	}
	bodyStart := end + len("\n---")
	if len(rest) > bodyStart && rest[bodyStart] == '\r' {
		bodyStart++
	}
	if len(rest) > bodyStart && rest[bodyStart] == '\n' {
		bodyStart++
	}
	return rest[:end], rest[bodyStart:], true
}

// ParseSections extracts non-nested devwiki:section blocks from Markdown body.
func ParseSections(body []byte) ([]Section, error) {
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	lines := strings.SplitAfter(text, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	var sections []Section
	seen := map[string]struct{}{}
	var current *Section
	var content strings.Builder
	for index, rawLine := range lines {
		lineNumber := index + 1
		trimmed := strings.TrimSpace(strings.TrimSuffix(rawLine, "\n"))
		if matches := sectionStartPattern.FindStringSubmatch(trimmed); matches != nil {
			if current != nil {
				return nil, fmt.Errorf("nested section %q at line %d", matches[1], lineNumber)
			}
			id := matches[1]
			if _, ok := seen[id]; ok {
				return nil, fmt.Errorf("duplicate section id %q", id)
			}
			seen[id] = struct{}{}
			current = &Section{ID: id, StartLine: lineNumber}
			content.Reset()
			continue
		}
		if trimmed == "<!-- /devwiki:section -->" {
			if current == nil {
				return nil, fmt.Errorf("unexpected section end at line %d", lineNumber)
			}
			current.EndLine = lineNumber
			current.Content = strings.TrimRight(content.String(), "\n")
			if strings.TrimSpace(current.Content) == "" {
				return nil, fmt.Errorf("empty section %q", current.ID)
			}
			current.Title = firstMarkdownHeading(current.Content)
			sections = append(sections, *current)
			current = nil
			content.Reset()
			continue
		}
		if current != nil {
			content.WriteString(rawLine)
		}
	}
	if current != nil {
		return nil, fmt.Errorf("section %q missing end marker", current.ID)
	}
	return sections, nil
}

// FindRootRelativePage finds kind/slug in the canonical topic/workflow directory.
func FindRootRelativePage(root string, kind string, slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", errors.New("slug is required")
	}
	dir, err := DirForKind(kind)
	if err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(dir, slug+".md"))
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
		return rel, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
	if os.IsNotExist(err) {
		return "", fmt.Errorf("%s %q not found", kind, slug)
	}
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		candidateRel := filepath.ToSlash(filepath.Join(dir, entry.Name()))
		doc, err := Load(root, candidateRel)
		if err != nil {
			return "", err
		}
		if doc.Meta.Slug == slug {
			return candidateRel, nil
		}
	}
	return "", fmt.Errorf("%s %q not found", kind, slug)
}

// DirForKind returns the root-relative wiki directory for a page kind.
func DirForKind(kind string) (string, error) {
	switch kind {
	case KindTopic:
		return "wiki/topics", nil
	case KindWorkflow:
		return "wiki/workflows", nil
	case KindTroubleshooting:
		return "wiki/troubleshooting", nil
	default:
		return "", fmt.Errorf("unsupported page kind %q", kind)
	}
}

// SectionByID returns a section by id.
func (d Document) SectionByID(id string) (Section, bool) {
	for _, section := range d.Sections {
		if section.ID == id {
			return section, true
		}
	}
	return Section{}, false
}

// NormalizeReference normalizes wikilinks, paths, and direct slugs to a slug.
func NormalizeReference(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[[")
	value = strings.TrimSuffix(value, "]]")
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".md")
	value = filepath.ToSlash(value)
	if strings.Contains(value, "/") {
		parts := strings.Split(value, "/")
		value = parts[len(parts)-1]
	}
	return strings.TrimSpace(value)
}

// NormalizeReferences normalizes, de-duplicates, and sorts references in input order.
func NormalizeReferences(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		ref := NormalizeReference(value)
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

// ListMarkdownFiles lists Markdown files under root-relative dirs.
func ListMarkdownFiles(root string, dirs ...string) ([]string, error) {
	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				files = append(files, filepath.ToSlash(filepath.Join(dir, entry.Name())))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func normalizeMeta(meta Meta) Meta {
	meta.Title = strings.TrimSpace(meta.Title)
	meta.Slug = NormalizeReference(meta.Slug)
	meta.Kind = strings.TrimSpace(meta.Kind)
	meta.Status = strings.TrimSpace(meta.Status)
	meta.Summary = strings.TrimSpace(meta.Summary)
	meta.Aliases = trimStrings(meta.Aliases)
	meta.Workflows = NormalizeReferences(meta.Workflows)
	meta.Topics = NormalizeReferences(meta.Topics)
	meta.RelatedTopics = NormalizeReferences(meta.RelatedTopics)
	meta.RelatedWorkflows = NormalizeReferences(meta.RelatedWorkflows)
	meta.Troubleshooting = NormalizeReferences(meta.Troubleshooting)
	meta.Confidence = strings.TrimSpace(meta.Confidence)
	meta.LastVerifiedAt = strings.TrimSpace(meta.LastVerifiedAt)
	return meta
}

func firstMarkdownHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return ""
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
