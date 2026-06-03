package devwikiapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	common "zatools/internal/app/common"
	"zatools/internal/devwiki"
	"zatools/internal/platform/agents"
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
	Agents      []string
	Stdout      io.Writer
}

// RepoUseOptions describes `zatools devwiki repo use` options.
type RepoUseOptions struct {
	ProjectSlug string
	SourceType  string
	Stdout      io.Writer
}

// RepoInitOptions describes `zatools devwiki repo init` options.
type RepoInitOptions struct{}

// RepoInitSource contains collected source answers for repo init.
type RepoInitSource struct {
	ProjectSlug string
	SourceType  string
	LocalPath   string
	RemoteURL   string
	Agents      []string
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
	ProjectSlug  string              `json:"project_slug"`
	ProjectName  string              `json:"project_name"`
	Language     string              `json:"language"`
	ActiveSource string              `json:"active_source"`
	Sources      devwiki.RepoSources `json:"sources,omitempty"`
	CodeRepos    []CodeRepoInfo      `json:"code_repos"`
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

// RepoUse switches the active source for a configured DevWiki project.
func (s *Service) RepoUse(ctx context.Context, opts RepoUseOptions) error {
	return s.runRepoUse(ctx, opts)
}

// RepoInit interactively creates a DevWiki repo config and optional code links.
func (s *Service) RepoInit(ctx context.Context, opts RepoInitOptions) error {
	return s.runRepoInit(ctx, opts)
}

// RepoInfo prints the user-level DevWiki source configuration as JSON.
func (s *Service) RepoInfo(ctx context.Context, opts RepoInfoOptions) error {
	return s.runRepoInfo(ctx, opts)
}

func (s *Service) runRepoInit(ctx context.Context, opts RepoInitOptions) error {
	_ = opts
	copy := ui.Messages()
	if !s.runtime.IsTTY {
		return errors.New(copy.DevwikiRepoInitTTYRequired)
	}

	project, err := promptLine(copy.PromptDevwikiProjectName, "")
	if err != nil {
		return err
	}
	project = strings.TrimSpace(project)
	if project == "" {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	selectedAgents, err := selectRepoInitAgents()
	if err != nil {
		return err
	}
	if len(selectedAgents) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	sourceType, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
		Message: copy.PromptDevwikiRepoSource,
		Items: []ui.Option{
			{Value: devwiki.RepoSourceLocal, Label: copy.DevwikiRepoSourceLocalLabel},
			{Value: devwiki.RepoSourceRemote, Label: copy.DevwikiRepoSourceRemoteLabel},
		},
	})
	if err != nil {
		return err
	}
	if cancelled {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	source := RepoInitSource{
		ProjectSlug: project,
		SourceType:  sourceType,
		Agents:      selectedAgents,
	}
	switch sourceType {
	case devwiki.RepoSourceLocal:
		value, err := promptLine(copy.PromptDevwikiRepoLocalPath, s.runtime.Workspace.CWD)
		if err != nil {
			return err
		}
		source.LocalPath = value
	case devwiki.RepoSourceRemote:
		value, err := promptLine(copy.PromptDevwikiRepoRemoteURL, "")
		if err != nil {
			return err
		}
		source.RemoteURL = value
	}
	if _, err := s.applyRepoInitSource(ctx, source); err != nil {
		return err
	}

	linkedAny := false
	for {
		prompt := copy.PromptDevwikiRepoLinkCode
		items := []ui.Option{
			{Value: "link", Label: copy.DevwikiRepoLinkCodeLabel},
			{Value: "continue", Label: copy.DevwikiRepoContinueLabel},
		}
		if linkedAny {
			prompt = copy.PromptDevwikiRepoLinkMore
			items = []ui.Option{
				{Value: "link", Label: copy.DevwikiRepoLinkAnotherLabel},
				{Value: "finish", Label: copy.DevwikiRepoFinishLabel},
			}
		}
		choice, cancelled, err := ui.SelectOne(ui.SelectOneOptions{Message: prompt, Items: items})
		if err != nil {
			return err
		}
		if cancelled || choice == "continue" || choice == "finish" {
			return nil
		}
		repoSlug, err := promptLine(copy.PromptDevwikiRepoCodeName, "")
		if err != nil {
			return err
		}
		repoPath, err := promptLine(copy.PromptDevwikiRepoCodePath, s.runtime.Workspace.CWD)
		if err != nil {
			return err
		}
		if err := s.RepoLink(ctx, RepoLinkOptions{
			ProjectSlug: project,
			RepoSlug:    repoSlug,
			Path:        repoPath,
			Agents:      selectedAgents,
		}); err != nil {
			return err
		}
		linkedAny = true
	}
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
		source := devwiki.RepoSource{Type: devwiki.RepoSourceRemote, URL: remoteURL}
		cfg.ActiveSource = devwiki.RepoSourceRemote
		cfg.Sources.Remote = &source
		sourceText = "remote: " + remoteURL
	} else {
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return err
		}
		source := devwiki.RepoSource{Type: devwiki.RepoSourceLocal, Path: absPath}
		cfg.ActiveSource = devwiki.RepoSourceLocal
		cfg.Sources.Local = &source
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

func (s *Service) runRepoUse(ctx context.Context, opts RepoUseOptions) error {
	_ = ctx
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	project := strings.TrimSpace(opts.ProjectSlug)
	sourceType := strings.TrimSpace(opts.SourceType)
	if project == "" || sourceType == "" {
		return fmt.Errorf("devwiki repo use requires project and source type")
	}
	cfg, err := devwiki.LoadRepoConfig(project)
	if err != nil {
		return err
	}
	switch sourceType {
	case devwiki.RepoSourceLocal:
		if cfg.Sources.Local == nil {
			return fmt.Errorf("local devwiki source is not configured for project %q", project)
		}
	case devwiki.RepoSourceRemote:
		if cfg.Sources.Remote == nil {
			return fmt.Errorf("remote devwiki source is not configured for project %q", project)
		}
	default:
		return fmt.Errorf("unsupported devwiki source type %q", sourceType)
	}
	cfg.ActiveSource = sourceType
	if err := devwiki.SaveRepoConfig(cfg); err != nil {
		return err
	}
	configPath, err := devwiki.RepoConfigPath(cfg.ProjectSlug)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, ui.Messages().DevwikiRepoUseSuccessFmt, cfg.ProjectSlug, sourceType, configPath)
	return err
}

func (s *Service) applyRepoInitSource(ctx context.Context, opts RepoInitSource) (devwiki.RepoConfig, error) {
	addOpts := RepoAddOptions{
		ProjectSlug: opts.ProjectSlug,
		Stdout:      opts.Stdout,
	}
	switch opts.SourceType {
	case devwiki.RepoSourceLocal:
		addOpts.LocalPath = opts.LocalPath
	case devwiki.RepoSourceRemote:
		addOpts.RemoteURL = opts.RemoteURL
	default:
		return devwiki.RepoConfig{}, fmt.Errorf("unsupported devwiki source type %q", opts.SourceType)
	}
	if err := s.RepoAdd(ctx, addOpts); err != nil {
		return devwiki.RepoConfig{}, err
	}
	cfg, err := devwiki.LoadRepoConfig(opts.ProjectSlug)
	if err != nil {
		return devwiki.RepoConfig{}, err
	}
	activeSource, err := devwiki.ActiveRepoSource(cfg)
	if err != nil {
		return devwiki.RepoConfig{}, err
	}
	if activeSource.Type == devwiki.RepoSourceLocal {
		agentKeys, err := normalizeRepoAgents(opts.Agents)
		if err != nil {
			return devwiki.RepoConfig{}, err
		}
		if err := s.installRepoInitDocSkills(activeSource.Path, agentKeys, cfg.Language); err != nil {
			return devwiki.RepoConfig{}, err
		}
	}
	return cfg, nil
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
	activeSource, err := devwiki.ActiveRepoSource(cfg)
	if err != nil {
		return err
	}
	if activeSource.Type == devwiki.RepoSourceLocal {
		devwikiRoot = activeSource.Path
	}
	if err := devwiki.EnsureCodeRepoDevwikiLink(absPath, devwikiRoot, cfg.ProjectSlug, "codex", cfg.Language); err != nil {
		return err
	}
	agentKeys, err := normalizeRepoAgents(opts.Agents)
	if err != nil {
		return err
	}
	if err := s.installCodeRepoDevwikiSkills(absPath, agentKeys, cfg.Language); err != nil {
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
		ProjectSlug:  cfg.ProjectSlug,
		ProjectName:  cfg.ProjectName,
		Language:     cfg.Language,
		ActiveSource: cfg.ActiveSource,
		Sources:      cfg.Sources,
		CodeRepos:    repos,
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

func (s *Service) installRepoInitDocSkills(docRoot string, agentKeys []string, lang string) error {
	return s.installBuiltinDevwikiSkillsForAgents(docRoot, agentKeys, lang, nil)
}

func (s *Service) installCodeRepoDevwikiSkills(codeRoot string, agentKeys []string, lang string) error {
	return s.installBuiltinDevwikiSkillsForAgents(codeRoot, agentKeys, lang, []string{
		"devwiki-code",
		"devwiki-code-to-doc",
		"devwiki-query",
	})
}

func (s *Service) installBuiltinDevwikiSkillsForAgents(projectRoot string, agentKeys []string, lang string, names []string) error {
	skillsRoot, cleanup, err := devwiki.ExtractBuiltinSkills(lang)
	if err != nil {
		return err
	}
	defer cleanup()

	found, err := skills.Discover(skillsRoot)
	if err != nil {
		return err
	}
	selected := found
	if len(names) > 0 {
		selected = selectDevwikiSkillsByName(found, names...)
		if len(selected) != len(names) {
			return fmt.Errorf("missing built-in DevWiki skills: %s", strings.Join(names, ", "))
		}
	}
	if err := s.installSelectedSkillsForAgents(projectRoot, agentKeys, false, lang, selected); err != nil {
		return err
	}
	return ensureProjectInstallGitignoreForAgents(projectRoot, agentKeys)
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

func normalizeRepoAgents(input []string) ([]string, error) {
	if len(input) == 0 {
		return []string{"codex"}, nil
	}
	return common.NormalizeAgentKeys(input, skills.SkillAsset, ui.Messages().UnsupportedAgentFmt)
}

func selectRepoInitAgents() ([]string, error) {
	copy := ui.Messages()
	items := make([]ui.Option, 0, len(agents.Supported()))
	initial := make([]string, 0, len(agents.Supported()))
	for _, agent := range agents.Supported() {
		if _, ok := agent.ProjectDirs[skills.SkillAsset]; !ok {
			continue
		}
		items = append(items, ui.Option{
			Value: agent.Key,
			Label: agent.DisplayName,
		})
		initial = append(initial, agent.Key)
	}
	selected, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:         ui.Bold + copy.PromptSelectAgents + " " + ui.Dim + "(" + copy.MultiSelectHelp + ")" + ui.Reset,
		Items:           items,
		Required:        true,
		MaxVisible:      8,
		InitialSelected: initial,
	})
	if err != nil || cancelled {
		return nil, err
	}
	return common.NormalizeAgentKeys(selected, skills.SkillAsset, copy.UnsupportedAgentFmt)
}

func ensureProjectInstallGitignoreForAgents(projectRoot string, agentKeys []string) error {
	var gitignorePaths []string
	for _, agentKey := range agentKeys {
		installDir, err := agents.ResolveSkillsDir(agentKey, false, projectRoot)
		if err != nil {
			return err
		}
		gitignorePaths = append(gitignorePaths, installDir)
	}
	lockPath, err := devwikiLockPath(projectRoot, false)
	if err != nil {
		return err
	}
	gitignorePaths = append(gitignorePaths, filepath.Join(projectRoot, ".cache"), lockPath)
	sort.Strings(gitignorePaths)
	return common.EnsureProjectGitignore(projectRoot, gitignorePaths...)
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
