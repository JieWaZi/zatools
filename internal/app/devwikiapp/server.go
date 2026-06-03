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
)

// ServerOptions describes `zatools devwiki server` execution options.
type ServerOptions struct {
	Root    string
	Project string
	Host    string
	Port    int
	Stdout  io.Writer
}

func (s *Service) runServer(ctx context.Context, opts ServerOptions) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	root := opts.Root
	if opts.Project != "" {
		cfg, err := devwiki.LoadRepoConfig(opts.Project)
		if err != nil {
			return err
		}
		source, err := devwiki.ActiveRepoSource(cfg)
		if err != nil {
			return err
		}
		if source.Type != devwiki.RepoSourceLocal {
			return fmt.Errorf("devwiki server can only serve local project sources")
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
	url, err := devwikigraph.ServeAPI(ctx, devwikigraph.ServerOptions{Root: absRoot, Host: opts.Host, Port: opts.Port})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "DevWiki HTTP API serving at %s\n", url)
	<-ctx.Done()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}
