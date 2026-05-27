package devwikiapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zatools/internal/devwiki/page"
	"zatools/internal/qmd"

	"golang.org/x/sync/errgroup"
)

const searchRRFK = 60.0

// SearchOptions describes `zatools devwiki search` execution options.
type SearchOptions struct {
	Root       string
	Kind       string
	Query      string
	QueryTerms []string
	Stdout     io.Writer
}

// SearchResult is one compact DevWiki search hit.
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

func (s *Service) runSearch(ctx context.Context, opts SearchOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	root := opts.Root
	if root == "" {
		root = s.runtime.Workspace.CWD
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	kind := strings.TrimSpace(opts.Kind)
	queries := normalizeSearchQueries(opts)
	if len(queries) == 0 {
		return fmt.Errorf("devwiki search query cannot be empty")
	}
	switch kind {
	case "index":
		results, err := searchIndexTable(absRoot, queries)
		if err != nil {
			return err
		}
		return encodeSearchJSON(stdout, results)
	case "glossary":
		results, err := searchGlossaryTable(absRoot, queries)
		if err != nil {
			return err
		}
		return encodeSearchJSON(stdout, results)
	case page.KindTopic, page.KindWorkflow:
	default:
		return fmt.Errorf("unsupported devwiki search kind %q", kind)
	}

	resultSets := make([][]SearchResult, len(queries))
	group, groupCtx := errgroup.WithContext(ctx)
	for i, query := range queries {
		i, query := i, query
		group.Go(func() error {
			var searchOut bytes.Buffer
			if err := qmd.RunCommandInDir(groupCtx, absRoot, []string{"search", query}, qmd.Models{}, &searchOut, os.Stderr); err != nil {
				return err
			}
			resultSets[i] = parseQMDSearchOutput(searchOut.String(), kind)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	results := fuseSearchResults(resultSets)
	if results == nil {
		results = []SearchResult{}
	}
	fillSearchResultSlugs(absRoot, kind, results)
	return encodeSearchJSON(stdout, results)
}

func encodeSearchJSON(stdout io.Writer, value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func normalizeSearchQueries(opts SearchOptions) []string {
	raw := opts.QueryTerms
	if len(raw) == 0 {
		raw = []string{opts.Query}
	}
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

func fuseSearchResults(resultSets [][]SearchResult) []SearchResult {
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

func parseQMDSearchOutput(output string, kind string) []SearchResult {
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
		if file, ok := parseQMDSearchFileLine(line, dir); ok {
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

func fillSearchResultSlugs(root string, kind string, results []SearchResult) {
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

func parseQMDSearchFileLine(line string, dir string) (string, bool) {
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

func searchIndexTable(root string, queries []string) ([]IndexSearchResult, error) {
	data, err := os.ReadFile(filepath.Join(root, "wiki", "index.md"))
	if err != nil {
		return nil, err
	}
	rows := parseMarkdownTableRows(string(data))
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
		if searchTableRowMatches(queries, result.Type, result.Description, result.Slug) {
			results = append(results, result)
		}
	}
	if results == nil {
		results = []IndexSearchResult{}
	}
	return results, nil
}

func searchGlossaryTable(root string, queries []string) ([]GlossarySearchResult, error) {
	data, err := os.ReadFile(filepath.Join(root, "wiki", "glossary.md"))
	if err != nil {
		return nil, err
	}
	rows := parseMarkdownTableRows(string(data))
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
		if searchTableRowMatches(queries, result.Glossary, result.Type, result.Description, result.Slug) {
			results = append(results, result)
		}
	}
	if results == nil {
		results = []GlossarySearchResult{}
	}
	return results, nil
}

func parseMarkdownTableRows(text string) []map[string]string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var headers []string
	var rows []map[string]string
	for _, line := range lines {
		cells, ok := parseMarkdownTableLine(line)
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
		if isMarkdownTableSeparator(cells) {
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

func parseMarkdownTableLine(line string) ([]string, bool) {
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

func isMarkdownTableSeparator(cells []string) bool {
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

func searchTableRowMatches(queries []string, values ...string) bool {
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
