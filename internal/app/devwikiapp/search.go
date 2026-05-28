package devwikiapp

import (
	"context"
	"encoding/json"
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
	root := opts.Root
	if strings.TrimSpace(opts.Project) != "" {
		cfg, err := devwiki.LoadRepoConfig(opts.Project)
		if err != nil {
			return err
		}
		if cfg.Source.Type == devwiki.RepoSourceRemote {
			return s.searchRemote(ctx, cfg, opts)
		}
		root = cfg.Source.Path
	}
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
		results, err := retrieval.SearchIndexTable(absRoot, queries)
		if err != nil {
			return err
		}
		return encodeSearchJSON(stdout, results)
	case "glossary":
		results, err := retrieval.SearchGlossaryTable(absRoot, queries)
		if err != nil {
			return err
		}
		return encodeSearchJSON(stdout, results)
	case page.KindTopic, page.KindWorkflow:
	default:
		return fmt.Errorf("unsupported devwiki search kind %q", kind)
	}

	results, err := retrieval.SearchPages(ctx, absRoot, kind, queries)
	if err != nil {
		return err
	}
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
	return retrieval.NormalizeQueries(raw)
}
