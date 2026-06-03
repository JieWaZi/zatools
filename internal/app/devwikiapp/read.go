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

// ReadOptions describes `zatools devwiki read` execution options.
type ReadOptions struct {
	Root    string
	Project string
	Kind    string
	Slug    string
	View    string
	Format  string
	Stdout  io.Writer
}

func (s *Service) runRead(ctx context.Context, opts ReadOptions) error {
	_ = ctx
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
		source, err := devwiki.ActiveRepoSource(cfg)
		if err != nil {
			return err
		}
		if source.Type == devwiki.RepoSourceRemote {
			return s.readRemote(ctx, source, opts)
		}
		root = source.Path
	}
	if root == "" {
		root = s.runtime.Workspace.CWD
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	format := strings.TrimSpace(opts.Format)
	if format == "" {
		format = "text"
	}
	if format != "text" {
		return fmt.Errorf("unsupported devwiki read format %q; only text is supported", format)
	}
	kind := strings.TrimSpace(opts.Kind)
	if kind != page.KindTopic && kind != page.KindWorkflow {
		return fmt.Errorf("unsupported devwiki read kind %q", kind)
	}
	view := strings.TrimSpace(opts.View)
	if view == "" {
		view = "card"
	}
	switch view {
	case "card", "core", "explain":
	default:
		return fmt.Errorf("unsupported %s view %q", kind, view)
	}

	text, err := retrieval.ReadText(absRoot, kind, opts.Slug, view)
	if err != nil {
		return err
	}
	_, err = stdout.Write([]byte(text))
	return err
}
