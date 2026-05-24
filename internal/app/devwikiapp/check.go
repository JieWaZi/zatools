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
	for _, rel := range files {
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

func collectDocumentCheckFiles(root string, paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{filepath.Join(root, "wiki")}
	}
	var files []string
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
				if isSectionedWikiDocument(rel) {
					files = append(files, rel)
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
			if isSectionedWikiDocument(rel) {
				files = append(files, rel)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func isSectionedWikiDocument(rel string) bool {
	rel = filepath.ToSlash(filepath.Clean(rel))
	return strings.HasPrefix(rel, "wiki/topics/") || strings.HasPrefix(rel, "wiki/workflows/")
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
