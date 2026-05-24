package devwikiapp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"zatools/internal/devwiki/page"
)

// ReadOptions describes `zatools devwiki read` execution options.
type ReadOptions struct {
	Root   string
	Kind   string
	Slug   string
	View   string
	Format string
	Stdout io.Writer
}

func (s *Service) runRead(ctx context.Context, opts ReadOptions) error {
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

	rel, err := page.FindRootRelativePage(absRoot, kind, opts.Slug)
	if err != nil {
		return err
	}
	doc, err := page.Load(absRoot, rel)
	if err != nil {
		return err
	}
	if doc.Meta.Kind != "" && doc.Meta.Kind != kind {
		return fmt.Errorf("%s: frontmatter kind %q does not match requested kind %q", rel, doc.Meta.Kind, kind)
	}
	section, ok := doc.SectionByID(view)
	if !ok {
		return fmt.Errorf("%s: missing section %q", rel, view)
	}
	if view == "card" {
		if _, err := fmt.Fprintln(stdout, "---"); err != nil {
			return err
		}
		if _, err := stdout.Write(doc.RawMeta); err != nil {
			return err
		}
		if len(doc.RawMeta) == 0 || doc.RawMeta[len(doc.RawMeta)-1] != '\n' {
			if _, err := fmt.Fprintln(stdout); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(stdout, "---"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(stdout); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(stdout, section.Content)
	return err
}
