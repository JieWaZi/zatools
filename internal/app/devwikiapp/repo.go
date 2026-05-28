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
	"zatools/internal/skills"
	"zatools/internal/ui"
)

// RepoAddOptions describes `zatools devwiki repo add` options.
type RepoAddOptions struct {
	ProjectSlug string
	RemoteURL   string
	LocalPath   string
	Stdout      io.Writer
}

// RepoLinkOptions describes `zatools devwiki repo link` options.
type RepoLinkOptions struct {
	ProjectSlug string
	RepoSlug    string
	Path        string
	Stdout      io.Writer
}

// RepoInfoOptions describes `zatools devwiki repo info` options.
type RepoInfoOptions struct {
	ProjectSlug string
	Format      string
	Stdout      io.Writer
}

// RepoInfo is the JSON shape emitted by `devwiki repo info`.
type RepoInfo struct {
	ProjectSlug string             `json:"project_slug"`
	ProjectName string             `json:"project_name"`
	Language    string             `json:"language"`
	Source      devwiki.RepoSource `json:"source"`
	CodeRepos   []CodeRepoInfo     `json:"code_repos"`
}

// CodeRepoInfo is the JSON shape emitted for bound code repositories.
type CodeRepoInfo struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Path    string `json:"path"`
	Default bool   `json:"default"`
}

// RepoAdd creates or updates a per-user DevWiki source configuration.
func (s *Service) RepoAdd(ctx context.Context, opts RepoAddOptions) error {
	return s.runRepoAdd(ctx, opts)
}

// RepoLink binds a local code repository path to a DevWiki project.
func (s *Service) RepoLink(ctx context.Context, opts RepoLinkOptions) error {
	return s.runRepoLink(ctx, opts)
}

// RepoInfo prints the user-level DevWiki source configuration as JSON.
func (s *Service) RepoInfo(ctx context.Context, opts RepoInfoOptions) error {
	return s.runRepoInfo(ctx, opts)
}

func (s *Service) runRepoAdd(ctx context.Context, opts RepoAddOptions) error {
	_ = ctx
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	project := strings.TrimSpace(opts.ProjectSlug)
	if project == "" {
		return fmt.Errorf("devwiki repo add project slug is required")
	}
	remoteURL := strings.TrimSpace(opts.RemoteURL)
	localPath := strings.TrimSpace(opts.LocalPath)
	if (remoteURL == "") == (localPath == "") {
		return fmt.Errorf("devwiki repo add requires exactly one local path or --remote url")
	}

	cfg := devwiki.RepoConfig{
		ProjectName: project,
		ProjectSlug: project,
		Language:    ui.DefaultLang,
	}
	if existing, err := devwiki.LoadRepoConfig(project); err == nil {
		cfg = existing
		cfg.ProjectName = project
		cfg.ProjectSlug = project
		if strings.TrimSpace(cfg.Language) == "" {
			cfg.Language = ui.DefaultLang
		}
	}
	sourceText := ""
	if remoteURL != "" {
		cfg.Source = devwiki.RepoSource{Type: devwiki.RepoSourceRemote, URL: remoteURL}
		sourceText = "remote: " + remoteURL
	} else {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return err
		}
		cfg.Source = devwiki.RepoSource{Type: devwiki.RepoSourceLocal, Path: absPath}
		sourceText = "local: " + absPath
	}
	if err := devwiki.SaveRepoConfig(cfg); err != nil {
		return err
	}
	configPath, err := devwiki.RepoConfigPath(cfg.ProjectSlug)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, ui.Messages().DevwikiRepoAddSuccessFmt, cfg.ProjectSlug, sourceText, configPath)
	return err
}

func (s *Service) runRepoLink(ctx context.Context, opts RepoLinkOptions) error {
	_ = ctx
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	project := strings.TrimSpace(opts.ProjectSlug)
	repoSlug := strings.TrimSpace(opts.RepoSlug)
	repoPath := strings.TrimSpace(opts.Path)
	if project == "" || repoSlug == "" || repoPath == "" {
		return fmt.Errorf("devwiki repo link requires project, repo slug, and path")
	}
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("code repo path is not a directory: %s", absPath)
	}

	cfg, err := devwiki.LoadRepoConfig(project)
	if err != nil {
		return err
	}
	repo := devwiki.CodeRepo{
		Name:    repoSlug,
		Slug:    repoSlug,
		Path:    absPath,
		Default: len(cfg.CodeRepos) == 0,
	}
	updated := false
	for i, existing := range cfg.CodeRepos {
		if existing.Slug == repoSlug {
			repo.Default = existing.Default
			cfg.CodeRepos[i] = repo
			updated = true
			break
		}
	}
	if !updated {
		cfg.CodeRepos = append(cfg.CodeRepos, repo)
	}
	if !hasDefaultCodeRepo(cfg.CodeRepos) && len(cfg.CodeRepos) > 0 {
		cfg.CodeRepos[0].Default = true
	}
	if err := devwiki.SaveRepoConfig(cfg); err != nil {
		return err
	}
	devwikiRoot := ""
	if cfg.Source.Type == devwiki.RepoSourceLocal {
		devwikiRoot = cfg.Source.Path
	}
	if err := devwiki.EnsureCodeRepoDevwikiLink(absPath, devwikiRoot, cfg.ProjectSlug, "codex", cfg.Language); err != nil {
		return err
	}
	if err := s.installCodeRepoDevwikiSkills(absPath, cfg.Language); err != nil {
		return err
	}
	configPath, err := devwiki.RepoConfigPath(cfg.ProjectSlug)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, ui.Messages().DevwikiRepoLinkSuccessFmt, repoSlug, cfg.ProjectSlug, absPath, configPath)
	return err
}

func (s *Service) runRepoInfo(ctx context.Context, opts RepoInfoOptions) error {
	_ = ctx
	if err := requireJSONFormat(opts.Format); err != nil {
		return err
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	if strings.TrimSpace(opts.ProjectSlug) == "" {
		projects, err := devwiki.ListRepoProjectSlugs()
		if err != nil {
			return err
		}
		return encodeIndentedJSON(stdout, projects)
	}
	cfg, err := devwiki.LoadRepoConfig(opts.ProjectSlug)
	if err != nil {
		return err
	}
	repos := codeRepoInfos(cfg.CodeRepos)
	return encodeIndentedJSON(stdout, RepoInfo{
		ProjectSlug: cfg.ProjectSlug,
		ProjectName: cfg.ProjectName,
		Language:    cfg.Language,
		Source:      cfg.Source,
		CodeRepos:   repos,
	})
}

func codeRepoInfos(codeRepos []devwiki.CodeRepo) []CodeRepoInfo {
	infos := make([]CodeRepoInfo, 0, len(codeRepos))
	for _, repo := range codeRepos {
		infos = append(infos, CodeRepoInfo{
			Name:    repo.Name,
			Slug:    repo.Slug,
			Path:    repo.Path,
			Default: repo.Default,
		})
	}
	return infos
}

func hasDefaultCodeRepo(repos []devwiki.CodeRepo) bool {
	for _, repo := range repos {
		if repo.Default {
			return true
		}
	}
	return false
}

func (s *Service) installCodeRepoDevwikiSkills(codeRoot string, lang string) error {
	skillsRoot, cleanup, err := devwiki.ExtractBuiltinSkills(lang)
	if err != nil {
		return err
	}
	defer cleanup()

	found, err := skills.Discover(skillsRoot)
	if err != nil {
		return err
	}
	selected := selectDevwikiSkillsByName(found, "devwiki-code", "devwiki-code-to-doc")
	if len(selected) != 2 {
		return fmt.Errorf("missing built-in code repo DevWiki skills")
	}
	if err := s.installSelectedSkills(codeRoot, "codex", false, lang, selected); err != nil {
		return err
	}
	return ensureProjectInstallGitignore(codeRoot, "codex")
}

func selectDevwikiSkillsByName(found []skills.Skill, names ...string) []skills.Skill {
	wanted := make(map[string]struct{}, len(names))
	for _, name := range names {
		wanted[name] = struct{}{}
	}
	selected := make([]skills.Skill, 0, len(names))
	for _, skill := range found {
		if _, ok := wanted[skill.Name]; ok {
			selected = append(selected, skill)
		}
	}
	return selected
}

func requireJSONFormat(format string) error {
	format = strings.TrimSpace(format)
	if format == "" || format == "json" {
		return nil
	}
	return fmt.Errorf("unsupported devwiki repo format %q; v1 supports json only", format)
}

func encodeIndentedJSON(stdout io.Writer, value any) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
