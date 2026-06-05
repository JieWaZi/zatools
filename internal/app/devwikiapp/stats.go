package devwikiapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"zatools/internal/devwiki/stats"
)

// StatsKeywordsOptions describes `zatools devwiki stats keywords`.
type StatsKeywordsOptions struct {
	Root   string
	Stdout io.Writer
}

// StatsKeywords aggregates search terms from queries-*.jsonl into keywords.json.
func (s *Service) StatsKeywords(ctx context.Context, opts StatsKeywordsOptions) error {
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

	result, err := stats.UpdateKeywords(stats.UpdateKeywordsOptions{
		Root: absRoot,
		Now:  time.Now(),
	})
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout); err != nil {
		return err
	}
	return nil
}
