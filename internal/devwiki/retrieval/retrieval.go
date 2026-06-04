package retrieval

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zatools/internal/devwiki/page"
	"zatools/internal/qmd"

	"gopkg.in/yaml.v3"
)

const searchRRFK = 60.0

// SearchResult is one compact DevWiki topic/workflow search hit.
type SearchResult struct {
	File  string `json:"file"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Score string `json:"score"`
}

// IndexSearchResult is one compact wiki/index.md search hit.
type IndexSearchResult struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

// GlossarySearchResult is one compact wiki/glossary.md search hit.
type GlossarySearchResult struct {
	Glossary    string `json:"glossary"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

// NormalizeQueries trims empty query terms while preserving user order.
func NormalizeQueries(raw []string) []string {
	queries := make([]string, 0, len(raw))
	for _, query := range raw {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		queries = append(queries, query)
	}
	return queries
}

// ReadText reads a topic/workflow section using the same text contract as the CLI.
func ReadText(root string, kind string, slug string, view string) (string, error) {
	if kind != page.KindTopic && kind != page.KindWorkflow {
		return "", fmt.Errorf("unsupported devwiki read kind %q", kind)
	}
	if view == "" {
		view = "card"
	}
	switch view {
	case "card", "core", "explain":
	default:
		return "", fmt.Errorf("unsupported %s view %q", kind, view)
	}

	rel, err := page.FindRootRelativePage(root, kind, slug)
	if err != nil {
		return "", err
	}
	doc, err := page.Load(root, rel)
	if err != nil {
		return "", err
	}
	if doc.Meta.Kind != "" && doc.Meta.Kind != kind {
		return "", fmt.Errorf("%s: frontmatter kind %q does not match requested kind %q", rel, doc.Meta.Kind, kind)
	}
	section, ok := doc.SectionByID(view)
	if !ok {
		return "", fmt.Errorf("%s: missing section %q", rel, view)
	}
	if view != "card" {
		return section.Content + "\n", nil
	}
	meta, err := MarshalCardMeta(doc.Meta)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.Write(meta)
	if len(meta) == 0 || meta[len(meta)-1] != '\n' {
		builder.WriteByte('\n')
	}
	builder.WriteString("---\n\n")
	builder.WriteString(section.Content)
	builder.WriteByte('\n')
	return builder.String(), nil
}

type cardMeta struct {
	Title      string `yaml:"title"`
	Status     string `yaml:"status"`
	Summary    string `yaml:"summary"`
	Confidence string `yaml:"confidence"`
}

// MarshalCardMeta serializes the filtered metadata shown in card views.
func MarshalCardMeta(meta page.Meta) ([]byte, error) {
	return yaml.Marshal(cardMeta{
		Title:      meta.Title,
		Status:     meta.Status,
		Summary:    meta.Summary,
		Confidence: meta.Confidence,
	})
}

// SearchPages searches topic/workflow pages through qmd and fuses multi-query results.
func SearchPages(ctx context.Context, root string, kind string, queries []string) ([]SearchResult, error) {
	switch kind {
	case page.KindTopic, page.KindWorkflow:
	default:
		return nil, fmt.Errorf("unsupported devwiki search kind %q", kind)
	}
	resultSets := make([][]SearchResult, len(queries))
	for i, query := range queries {
		var searchOut bytes.Buffer
		if err := qmd.RunCommandInDir(ctx, root, []string{"search", query}, qmd.Models{}, &searchOut, os.Stderr); err != nil {
			return nil, err
		}
		resultSets[i] = ParseQMDSearchOutput(searchOut.String(), kind)
	}
	results := FuseSearchResults(resultSets)
	if results == nil {
		results = []SearchResult{}
	}
	FillSearchResultSlugs(root, kind, results)
	return results, nil
}

// FuseSearchResults combines multiple qmd result sets with reciprocal rank fusion.
func FuseSearchResults(resultSets [][]SearchResult) []SearchResult {
	if len(resultSets) == 0 {
		return nil
	}
	if len(resultSets) == 1 {
		return resultSets[0]
	}

	type fusedResult struct {
		result SearchResult
		score  float64
		order  int
	}
	fused := make(map[string]*fusedResult)
	order := 0
	for _, results := range resultSets {
		for rank, result := range results {
			if result.File == "" {
				continue
			}
			item, ok := fused[result.File]
			if !ok {
				item = &fusedResult{result: result, order: order}
				fused[result.File] = item
				order++
			}
			item.score += 1 / (searchRRFK + float64(rank+1))
		}
	}

	items := make([]fusedResult, 0, len(fused))
	var maxScore float64
	for _, item := range fused {
		items = append(items, *item)
		if item.score > maxScore {
			maxScore = item.score
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score == items[j].score {
			return items[i].order < items[j].order
		}
		return items[i].score > items[j].score
	})

	results := make([]SearchResult, 0, len(items))
	for _, item := range items {
		result := item.result
		result.Score = formatFusedSearchScore(item.score, maxScore)
		results = append(results, result)
	}
	return results
}

func formatFusedSearchScore(score float64, maxScore float64) string {
	if maxScore <= 0 {
		return "0%"
	}
	percent := int((score/maxScore)*100 + 0.5)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return fmt.Sprintf("%d%%", percent)
}

// ParseQMDSearchOutput keeps qmd search hits for the requested topic/workflow kind.
func ParseQMDSearchOutput(output string, kind string) []SearchResult {
	dir := "topics"
	if kind == page.KindWorkflow {
		dir = "workflows"
	}

	var results []SearchResult
	var currentFile string
	var currentTitle string
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			currentFile = ""
			currentTitle = ""
			continue
		}
		if file, ok := ParseQMDSearchFileLine(line, dir); ok {
			currentFile = file
			currentTitle = ""
			continue
		}
		if currentFile == "" {
			continue
		}
		if title, ok := parseQMDSearchTitleLine(line); ok {
			currentTitle = title
			continue
		}
		score, ok := parseQMDSearchScoreLine(line)
		if !ok {
			continue
		}
		results = append(results, SearchResult{File: currentFile, Title: currentTitle, Score: score})
		currentFile = ""
		currentTitle = ""
	}
	return results
}

// FillSearchResultSlugs loads each result page and fills its frontmatter slug.
func FillSearchResultSlugs(root string, kind string, results []SearchResult) {
	dir := "topics"
	if kind == page.KindWorkflow {
		dir = "workflows"
	}
	for index := range results {
		rel := filepath.ToSlash(filepath.Join("wiki", dir, results[index].File))
		doc, err := page.Load(root, rel)
		if err != nil {
			continue
		}
		results[index].Slug = doc.Meta.Slug
	}
}

// ParseQMDSearchFileLine extracts a topic/workflow markdown filename from qmd output.
func ParseQMDSearchFileLine(line string, dir string) (string, bool) {
	firstField := strings.Fields(line)
	if len(firstField) == 0 {
		return "", false
	}
	path := firstField[0]
	if index := strings.LastIndex(path, ":"); index >= 0 && isQMDLineNumberSuffix(path[index+1:]) {
		path = path[:index]
	}
	path = strings.Trim(filepath.ToSlash(path), "/")
	prefix := dir + "/"
	if strings.HasPrefix(path, prefix) {
		return filepath.Base(path), true
	}
	marker := "/" + prefix
	index := strings.Index(path, marker)
	if index < 0 {
		return "", false
	}
	return filepath.Base(path[index+len(marker):]), true
}

func parseQMDSearchScoreLine(line string) (string, bool) {
	label, value, ok := strings.Cut(line, ":")
	if !ok || !strings.EqualFold(strings.TrimSpace(label), "Score") {
		return "", false
	}
	score := strings.TrimSpace(value)
	return score, score != ""
}

func parseQMDSearchTitleLine(line string) (string, bool) {
	label, value, ok := strings.Cut(line, ":")
	if !ok || !strings.EqualFold(strings.TrimSpace(label), "Title") {
		return "", false
	}
	title := strings.TrimSpace(value)
	return title, title != ""
}

func isQMDLineNumberSuffix(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// SearchIndexTable searches wiki/index.md with the CLI table contract.
func SearchIndexTable(root string, queries []string) ([]IndexSearchResult, error) {
	data, err := os.ReadFile(filepath.Join(root, "wiki", "index.md"))
	if err != nil {
		return nil, err
	}
	rows := ParseMarkdownTableRows(string(data))
	results := make([]IndexSearchResult, 0, len(rows))
	for _, row := range rows {
		result := IndexSearchResult{
			Type:        row["type"],
			Description: row["description"],
			Slug:        row["slug"],
		}
		if result.Type == "" || result.Description == "" || result.Slug == "" {
			continue
		}
		if SearchTableRowMatches(queries, result.Type, result.Description, result.Slug) {
			results = append(results, result)
		}
	}
	if results == nil {
		results = []IndexSearchResult{}
	}
	return results, nil
}

// SearchGlossaryTable searches wiki/glossary.md with the CLI table contract.
func SearchGlossaryTable(root string, queries []string) ([]GlossarySearchResult, error) {
	data, err := os.ReadFile(filepath.Join(root, "wiki", "glossary.md"))
	if err != nil {
		return nil, err
	}
	rows := ParseMarkdownTableRows(string(data))
	results := make([]GlossarySearchResult, 0, len(rows))
	for _, row := range rows {
		result := GlossarySearchResult{
			Glossary:    row["glossary"],
			Type:        row["type"],
			Description: row["description"],
			Slug:        row["slug"],
		}
		if result.Glossary == "" || result.Type == "" || result.Description == "" || result.Slug == "" {
			continue
		}
		if SearchTableRowMatches(queries, result.Glossary, result.Type, result.Description, result.Slug) {
			results = append(results, result)
		}
	}
	if results == nil {
		results = []GlossarySearchResult{}
	}
	return results, nil
}

// GlossaryKeywords returns unique glossary terms in table order.
func GlossaryKeywords(root string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "wiki", "glossary.md"))
	if err != nil {
		return nil, err
	}
	rows := ParseMarkdownTableRows(string(data))
	keywords := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		keyword := row["glossary"]
		if keyword == "" || row["type"] == "" || row["description"] == "" || row["slug"] == "" {
			continue
		}
		if _, ok := seen[keyword]; ok {
			continue
		}
		seen[keyword] = struct{}{}
		keywords = append(keywords, keyword)
	}
	if keywords == nil {
		keywords = []string{}
	}
	return keywords, nil
}

// ParseMarkdownTableRows parses pipe table rows into lowercase-header maps.
func ParseMarkdownTableRows(text string) []map[string]string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var headers []string
	var rows []map[string]string
	for _, line := range lines {
		cells, ok := ParseMarkdownTableLine(line)
		if !ok {
			headers = nil
			continue
		}
		if len(headers) == 0 {
			headers = make([]string, len(cells))
			for i, cell := range cells {
				headers[i] = strings.ToLower(strings.TrimSpace(cell))
			}
			continue
		}
		if IsMarkdownTableSeparator(cells) {
			continue
		}
		if len(cells) != len(headers) {
			continue
		}
		row := make(map[string]string, len(headers))
		for i, header := range headers {
			row[header] = strings.TrimSpace(cells[i])
		}
		rows = append(rows, row)
	}
	return rows
}

// ParseMarkdownTableLine parses one markdown pipe table row.
func ParseMarkdownTableLine(line string) ([]string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
		return nil, false
	}
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}
	return cells, len(cells) > 0
}

// IsMarkdownTableSeparator reports whether a parsed table row is a separator row.
func IsMarkdownTableSeparator(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			return false
		}
		for _, char := range cell {
			if char != '-' && char != ':' {
				return false
			}
		}
	}
	return true
}

// SearchTableRowMatches returns true when any query appears in the row values.
func SearchTableRowMatches(queries []string, values ...string) bool {
	haystack := strings.ToLower(strings.Join(values, "\n"))
	for _, query := range queries {
		query = strings.ToLower(strings.TrimSpace(query))
		if query == "" {
			continue
		}
		if strings.Contains(haystack, query) {
			return true
		}
	}
	return false
}
