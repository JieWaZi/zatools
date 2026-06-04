package devwikiapp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"zatools/internal/devwiki"
	"zatools/internal/devwiki/page"
	"zatools/internal/devwiki/retrieval"
)

// SearchOptions describes `zatools devwiki search` execution options.
type SearchOptions struct {
	Root       string
	Project    string
	Kind       string
	Query      string
	QueryTerms []string
	Stdout     io.Writer
}

// GlossaryKeywordsOptions describes `zatools devwiki glossary keywords`.
type GlossaryKeywordsOptions struct {
	Root    string
	Project string
	Stdout  io.Writer
}

// SearchResult is one compact DevWiki search hit.
type SearchResult = retrieval.SearchResult

// IndexSearchResult is one compact wiki/index.md search hit.
type IndexSearchResult = retrieval.IndexSearchResult

// GlossarySearchResult is one compact wiki/glossary.md search hit.
type GlossarySearchResult = retrieval.GlossarySearchResult

func (s *Service) runSearch(ctx context.Context, opts SearchOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	opts.Stdout = stdout
	absRoot, remote, err := s.resolveDevwikiSource(opts.Root, opts.Project)
	if err != nil {
		return err
	}
	if remote != nil {
		return s.searchRemote(ctx, *remote, opts)
	}
	kind := strings.TrimSpace(opts.Kind)
	queries := normalizeSearchQueries(opts)
	if len(queries) == 0 {
		return fmt.Errorf("devwiki search query cannot be empty")
	}
	switch kind {
	case "index":
		results, err := retrieval.SearchIndexTable(absRoot, queries)
		if err != nil {
			return err
		}
		return writeIndexSearchTable(stdout, results)
	case "glossary":
		results, err := retrieval.SearchGlossaryTable(absRoot, queries)
		if err != nil {
			return err
		}
		return writeGlossarySearchTable(stdout, results)
	case page.KindTopic, page.KindWorkflow:
	default:
		return fmt.Errorf("unsupported devwiki search kind %q", kind)
	}

	results, err := retrieval.SearchPages(ctx, absRoot, kind, queries)
	if err != nil {
		return err
	}
	return writePageSearchTable(stdout, results)
}

// GlossaryKeywords writes unique glossary terms, one per line.
func (s *Service) GlossaryKeywords(ctx context.Context, opts GlossaryKeywordsOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	opts.Stdout = stdout
	absRoot, remote, err := s.resolveDevwikiSource(opts.Root, opts.Project)
	if err != nil {
		return err
	}
	if remote != nil {
		return s.glossaryKeywordsRemote(ctx, *remote, opts)
	}
	keywords, err := retrieval.GlossaryKeywords(absRoot)
	if err != nil {
		return err
	}
	for _, keyword := range keywords {
		if _, err := fmt.Fprintln(stdout, keyword); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resolveDevwikiSource(root string, project string) (string, *devwiki.RepoSource, error) {
	if strings.TrimSpace(project) != "" {
		cfg, err := devwiki.LoadRepoConfig(project)
		if err != nil {
			return "", nil, err
		}
		source, err := devwiki.ActiveRepoSource(cfg)
		if err != nil {
			return "", nil, err
		}
		if source.Type == devwiki.RepoSourceRemote {
			return "", &source, nil
		}
		root = source.Path
	}
	if root == "" {
		root = s.runtime.Workspace.CWD
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, err
	}
	return absRoot, nil, nil
}

func writeIndexSearchTable(stdout io.Writer, results []IndexSearchResult) error {
	if _, err := fmt.Fprintln(stdout, "|type|description|slug|"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(stdout, "|%s|%s|%s|\n",
			pipeCell(result.Type),
			pipeCell(result.Description),
			pipeCell(result.Slug),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeGlossarySearchTable(stdout io.Writer, results []GlossarySearchResult) error {
	if _, err := fmt.Fprintln(stdout, "|glossary|type|description|slug|"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(stdout, "|%s|%s|%s|%s|\n",
			pipeCell(result.Glossary),
			pipeCell(result.Type),
			pipeCell(result.Description),
			pipeCell(result.Slug),
		); err != nil {
			return err
		}
	}
	return nil
}

func writePageSearchTable(stdout io.Writer, results []SearchResult) error {
	if _, err := fmt.Fprintln(stdout, "|file|slug|title|score|"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(stdout, "|%s|%s|%s|%s|\n",
			pipeCell(result.File),
			pipeCell(result.Slug),
			pipeCell(result.Title),
			pipeCell(result.Score),
		); err != nil {
			return err
		}
	}
	return nil
}

func pipeCell(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return strings.TrimSpace(value)
}

func normalizeSearchQueries(opts SearchOptions) []string {
	raw := opts.QueryTerms
	if len(raw) == 0 {
		raw = []string{opts.Query}
	}
	return retrieval.NormalizeQueries(raw)
}
