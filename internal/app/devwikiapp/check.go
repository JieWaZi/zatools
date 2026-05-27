package devwikiapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	devwikigraph "zatools/internal/devwiki/graph"
	devwikipage "zatools/internal/devwiki/page"
)

// CheckOptions describes `zatools devwiki check` execution options.
type CheckOptions struct {
	Root   string
	Types  []string
	Paths  []string
	Stdout io.Writer
}

type documentCheckFile struct {
	Rel  string
	Kind string
}

func (s *Service) runCheck(ctx context.Context, opts CheckOptions) error {
	_ = ctx
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
	types, err := normalizeCheckTypes(opts.Types)
	if err != nil {
		return err
	}
	for _, typ := range types {
		switch typ {
		case "document":
			if err := runDocumentCheck(stdout, absRoot, opts.Paths); err != nil {
				return err
			}
		case "graph":
			if err := runGraphCheck(stdout, absRoot); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported devwiki check type %q", typ)
		}
	}
	return nil
}

func normalizeCheckTypes(input []string) ([]string, error) {
	if len(input) == 0 {
		return []string{"document", "graph"}, nil
	}
	var out []string
	for _, value := range input {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			switch part {
			case "document", "graph":
				out = append(out, part)
			default:
				return nil, fmt.Errorf("unsupported devwiki check type %q", part)
			}
		}
	}
	if len(out) == 0 {
		return nil, errors.New("devwiki check type is required")
	}
	return out, nil
}

func runDocumentCheck(stdout io.Writer, root string, paths []string) error {
	files, err := collectDocumentCheckFiles(root, paths)
	if err != nil {
		return err
	}
	var issues []string
	for _, file := range files {
		rel := file.Rel
		if file.Kind != "sectioned" {
			issues = append(issues, checkSupportFile(root, rel)...)
			continue
		}
		doc, err := devwikipage.Load(root, rel)
		if err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", rel, err))
			continue
		}
		for _, id := range []string{"card", "core", "explain"} {
			if _, ok := doc.SectionByID(id); !ok {
				issues = append(issues, fmt.Sprintf("%s: missing required section %q", rel, id))
			}
		}
	}
	for _, issue := range issues {
		fmt.Fprintf(stdout, "error %s\n", issue)
	}
	if len(issues) > 0 {
		return fmt.Errorf("document check failed with %d issue(s)", len(issues))
	}
	fmt.Fprintf(stdout, "DevWiki document check passed (%d file(s))\n", len(files))
	return nil
}

func collectDocumentCheckFiles(root string, paths []string) ([]documentCheckFile, error) {
	if len(paths) == 0 {
		paths = []string{filepath.Join(root, "wiki")}
	}
	var files []documentCheckFile
	for _, input := range paths {
		abs := input
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(root, input)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if filepath.Ext(abs) == ".md" {
				rel, err := filepath.Rel(root, abs)
				if err != nil {
					return nil, err
				}
				rel = filepath.ToSlash(rel)
				if kind := documentCheckKind(rel); kind != "" {
					files = append(files, documentCheckFile{Rel: rel, Kind: kind})
				}
			}
			continue
		}
		err = filepath.WalkDir(abs, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if kind := documentCheckKind(rel); kind != "" {
				files = append(files, documentCheckFile{Rel: rel, Kind: kind})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func documentCheckKind(rel string) string {
	rel = filepath.ToSlash(filepath.Clean(rel))
	if strings.HasPrefix(rel, "wiki/topics/") || strings.HasPrefix(rel, "wiki/workflows/") {
		return "sectioned"
	}
	switch rel {
	case "wiki/index.md", "wiki/glossary.md", "wiki/log.md":
		return "support"
	default:
		return ""
	}
}

func checkSupportFile(root string, rel string) []string {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", rel, err)}
	}
	text := string(data)
	switch rel {
	case "wiki/index.md":
		return checkRequiredTable(rel, text, []string{"type", "description", "slug"})
	case "wiki/glossary.md":
		return checkRequiredTable(rel, text, []string{"glossary", "type", "description", "slug"})
	case "wiki/log.md":
		if !strings.HasPrefix(strings.TrimLeft(text, "\ufeff\r\n\t "), "# Wiki Log") {
			return []string{fmt.Sprintf("%s: missing required title %q", rel, "# Wiki Log")}
		}
	}
	return nil
}

func checkRequiredTable(rel string, text string, required []string) []string {
	var issues []string
	headers, rows, ok := firstMarkdownTable(text)
	if !ok {
		return []string{fmt.Sprintf("%s: missing required table", rel)}
	}
	if len(headers) != len(required) {
		issues = append(issues, fmt.Sprintf("%s: table must have columns %s", rel, strings.Join(required, ", ")))
	} else {
		for index, header := range headers {
			if header != required[index] {
				issues = append(issues, fmt.Sprintf("%s: table column %d must be %q", rel, index+1, required[index]))
			}
		}
	}
	for rowIndex, row := range rows {
		if len(row) != len(headers) {
			issues = append(issues, fmt.Sprintf("%s: table row %d has %d cell(s), want %d", rel, rowIndex+1, len(row), len(headers)))
			continue
		}
		for cellIndex, cell := range row {
			if strings.TrimSpace(cell) == "" {
				name := fmt.Sprintf("column %d", cellIndex+1)
				if cellIndex < len(headers) {
					name = headers[cellIndex]
				}
				issues = append(issues, fmt.Sprintf("%s: table row %d has empty %s", rel, rowIndex+1, name))
			}
		}
	}
	return issues
}

func firstMarkdownTable(text string) ([]string, [][]string, bool) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for index := 0; index < len(lines); index++ {
		headers, ok := parseMarkdownTableLine(lines[index])
		if !ok || index+1 >= len(lines) {
			continue
		}
		separator, ok := parseMarkdownTableLine(lines[index+1])
		if !ok || !isMarkdownTableSeparator(separator) {
			continue
		}
		for i := range headers {
			headers[i] = strings.ToLower(strings.TrimSpace(headers[i]))
		}
		var rows [][]string
		for rowIndex := index + 2; rowIndex < len(lines); rowIndex++ {
			row, ok := parseMarkdownTableLine(lines[rowIndex])
			if !ok {
				break
			}
			rows = append(rows, row)
		}
		return headers, rows, true
	}
	return nil, nil, false
}

func runGraphCheck(stdout io.Writer, root string) error {
	_, issues, err := devwikigraph.Build(root)
	printGraphIssues(stdout, issues)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "DevWiki graph check passed")
	return nil
}
