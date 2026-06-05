package devwikiapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"zatools/internal/devwiki"
	"zatools/internal/ui"
)

// SkillRefsOptions describes `devwiki skill refs` command options.
type SkillRefsOptions struct {
	// Stdout receives command output.
	Stdout io.Writer
}

// SkillRefsCheck checks duplicated DevWiki skill references without modifying files.
func (s *Service) SkillRefsCheck(ctx context.Context, opts SkillRefsOptions) error {
	_ = ctx
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	copy := ui.Messages()
	repoRoot, err := devwiki.FindZatoolsRepoRoot(s.runtime.Workspace.CWD)
	if err != nil {
		return err
	}
	issues, err := devwiki.CheckDevwikiReferenceGroups(repoRoot)
	if err != nil {
		return err
	}
	if len(issues) == 0 {
		_, err = fmt.Fprintln(stdout, copy.DevwikiSkillRefsOK)
		return err
	}
	_, _ = fmt.Fprintln(stdout, devwiki.FormatReferenceGroupIssues(issues))
	return errors.New(copy.DevwikiSkillRefsCheckFailed)
}

// SkillRefsFix syncs duplicated DevWiki skill references from their canonical files.
func (s *Service) SkillRefsFix(ctx context.Context, opts SkillRefsOptions) error {
	_ = ctx
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	copy := ui.Messages()
	repoRoot, err := devwiki.FindZatoolsRepoRoot(s.runtime.Workspace.CWD)
	if err != nil {
		return err
	}
	updated, err := devwiki.FixDevwikiReferenceGroups(repoRoot)
	if err != nil {
		return err
	}
	if len(updated) == 0 {
		_, err = fmt.Fprintln(stdout, copy.DevwikiSkillRefsNoChanges)
		return err
	}
	for _, file := range updated {
		if _, err := fmt.Fprintf(stdout, copy.DevwikiSkillRefsFixedFmt, file); err != nil {
			return err
		}
	}
	return nil
}
