package devwikiapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"zatools/internal/devwiki"
	devwikigraph "zatools/internal/devwiki/graph"
	"zatools/internal/ui"
)

// GraphOptions describes `zatools devwiki graph` execution options.
type GraphOptions struct {
	Root    string
	Project string
	Host    string
	Port    int
	NoOpen  bool
	Force   bool
	Check   bool
	NoServe bool
	Stdout  io.Writer
	Stderr  io.Writer
}

func (s *Service) runGraph(ctx context.Context, opts GraphOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	root := opts.Root
	if opts.Project != "" {
		cfg, err := devwiki.LoadRepoConfig(opts.Project)
		if err != nil {
			return err
		}
		if cfg.Source.Type != devwiki.RepoSourceLocal {
			return fmt.Errorf("devwiki graph can only serve local project sources")
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
	if opts.Check {
		_, issues, err := devwikigraph.Build(absRoot)
		printGraphIssues(stdout, issues)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, "DevWiki graph check passed")
		return nil
	}

	outDir := filepath.Join(absRoot, ".devwiki", "graph")
	current, err := devwikigraph.BuildManifest(absRoot)
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(outDir, "manifest.json")
	rebuilt := opts.Force
	if !rebuilt {
		previous, err := devwikigraph.ReadManifest(manifestPath)
		if err != nil || !previous.IsFresh(current) {
			rebuilt = true
		}
	}
	if rebuilt {
		graph, issues, err := devwikigraph.Build(absRoot)
		printGraphIssues(stdout, issues)
		if err != nil {
			return err
		}
		if err := devwikigraph.WriteOutputs(outDir, graph, current); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "DevWiki graph rebuilt: %s\n", outDir)
	} else {
		fmt.Fprintf(stdout, "DevWiki graph cache is fresh: %s\n", outDir)
	}
	if opts.NoServe {
		return nil
	}
	url, err := devwikigraph.Serve(ctx, devwikigraph.ServerOptions{Dir: outDir, Root: absRoot, Host: opts.Host, Port: opts.Port})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "DevWiki graph serving at %s\n", url)
	if !opts.NoOpen {
		if err := devwikigraph.OpenBrowser(url); err != nil {
			fmt.Fprintf(stderr, ui.Messages().DevwikiGraphOpenFailedFmt, ui.Yellow, ui.Reset, err, url)
		}
	}
	<-ctx.Done()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}

func printGraphIssues(w io.Writer, issues []devwikigraph.Issue) {
	for _, issue := range issues {
		path := issue.Path
		if path == "" {
			path = "-"
		}
		fmt.Fprintf(w, "%s %s: %s\n", issue.Level, path, issue.Message)
	}
}
