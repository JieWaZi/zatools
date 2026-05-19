package devwikiapp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"

	common "zatools/internal/app/common"
	"zatools/internal/app/skillapp"
	"zatools/internal/devwiki"
	"zatools/internal/platform/agents"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

func (s *Service) runProject(ctx context.Context, opts InitOptions, installSkills bool) error {
	_ = ctx

	copy := ui.Messages()

	resolved, err := s.collectInitOptions(opts, installSkills)
	if err != nil {
		return err
	}
	if strings.TrimSpace(resolved.ProjectName) == "" {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	targetDir, err := s.resolveTargetDir()
	if err != nil {
		return err
	}

	codeRepos, err := devwiki.NormalizeCodeRepos(s.runtime.Workspace.CWD, resolved.CodeDirs)
	if err != nil {
		return err
	}

	slug := devwiki.Slugify(resolved.ProjectName)
	spec := devwiki.ProjectSpec{
		ProjectName: resolved.ProjectName,
		ProjectSlug: slug,
		Agent:       resolved.Agent,
		Lang:        resolved.Lang,
		CodeRepos:   codeRepos,
	}

	lines := []string{
		fmt.Sprintf("%s: %s", copy.ProjectLabel, resolved.ProjectName),
		fmt.Sprintf("%s: %s", copy.SourceLabel, targetDir),
		fmt.Sprintf("%s: %s", copy.AgentsLabel, resolved.Agent),
		fmt.Sprintf("%s: %s", copy.DevwikiCodeDirsLabel, strings.Join(resolved.CodeDirs, ", ")),
	}
	if installSkills {
		lines = append(lines, fmt.Sprintf("%s: %s", copy.ScopeLabel, ui.ScopeText(resolved.Global)))
	}
	ui.Note(copy.TitleDevwikiSummary, lines)

	spinner := ui.NewStepPrinter()
	spinner.Start(copy.StepCreatingDevwikiProject)
	if err := devwiki.GenerateProject(targetDir, spec); err != nil {
		return err
	}
	spinner.Stop(fmt.Sprintf(copy.CreatedFmt, targetDir))

	if installSkills {
		skillsRoot, cleanup, err := devwiki.ExtractBuiltinSkills(resolved.Lang)
		if err != nil {
			return err
		}
		defer cleanup()

		found, err := skills.Discover(skillsRoot)
		if err != nil {
			return err
		}

		selected, err := s.resolveSelectedSkills(found, resolved)
		if err != nil {
			return err
		}
		if selected != nil && len(selected) > 0 {
			spinner.Start(copy.StepInstallingDevwikiSkills)
			if err := s.installSelectedSkills(targetDir, resolved.Agent, resolved.Global, resolved.Lang, selected); err != nil {
				return err
			}
			spinner.Stop(fmt.Sprintf(copy.DevwikiInstalledSkillsFmt, len(selected)))
		}
	}

	gitignorePaths := []string{filepath.Join(targetDir, ".cache")}
	if !resolved.Global {
		installDir, err := agents.ResolveSkillsDir(resolved.Agent, false, targetDir)
		if err != nil {
			return err
		}
		gitignorePaths = append(gitignorePaths, installDir)
		lockPath, err := devwikiLockPath(targetDir, false)
		if err != nil {
			return err
		}
		gitignorePaths = append(gitignorePaths, lockPath)
	}
	if err := common.EnsureProjectGitignore(targetDir, gitignorePaths...); err != nil {
		return err
	}

	fmt.Printf("%s%s%s\n", ui.Green, copy.Done, ui.Reset)
	ui.Note(copy.TitleQMDManualDownload, []string{
		copy.QMDManualDownloadHint,
		copy.QMDManualDownloadCommand,
	})
	return nil
}

func (s *Service) linkCodeRepositories(ctx context.Context, opts LinkOptions) error {
	_ = ctx

	resolved, err := s.normalizeLinkOptions(opts)
	if err != nil {
		return err
	}

	devwikiRoot := resolved.DevwikiRoot
	codeRepos, err := devwiki.NormalizeCodeRepos(devwikiRoot, resolved.CodeDirs)
	if err != nil {
		return err
	}

	skillsRoot, cleanup, err := devwiki.ExtractBuiltinSkills(resolved.Lang)
	if err != nil {
		return err
	}
	defer cleanup()

	found, err := skills.Discover(skillsRoot)
	if err != nil {
		return err
	}
	codeSkills, err := s.resolveCodeRepoSkills(found, resolved.Yes)
	if err != nil {
		return err
	}

	spinner := ui.NewStepPrinter()
	for _, repo := range codeRepos {
		spinner.Start(ui.Messages().StepLinkingDevwikiCodeRepo)
		if err := devwiki.EnsureCodeRepoDevwikiLink(repo.Path, devwikiRoot, resolved.Agent, resolved.Lang); err != nil {
			return err
		}
		spinner.Stop(fmt.Sprintf(ui.Messages().DevwikiLinkedCodeRepoFmt, repo.Path))
		if len(codeSkills) == 0 {
			continue
		}
		spinner.Start(ui.Messages().StepInstallingDevwikiSkills)
		if err := s.installSelectedSkills(repo.Path, resolved.Agent, false, resolved.Lang, codeSkills); err != nil {
			return err
		}
		spinner.Stop(fmt.Sprintf(ui.Messages().DevwikiInstalledSkillsFmt, len(codeSkills)))
		if err := ensureProjectInstallGitignore(repo.Path, resolved.Agent); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) collectInitOptions(opts InitOptions, includeInstall bool) (InitOptions, error) {
	copy := ui.Messages()

	if strings.TrimSpace(opts.ProjectName) == "" {
		if !s.runtime.IsTTY {
			return opts, errors.New(copy.DevwikiProjectNameRequired)
		}
		value, err := promptLine(copy.PromptDevwikiProjectName, "")
		if err != nil {
			return opts, err
		}
		opts.ProjectName = value
	}

	if strings.TrimSpace(opts.Agent) == "" {
		if s.runtime.IsTTY {
			value, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
				Message: copy.PromptDevwikiAgent,
				Items: []ui.Option{
					{Value: "codex", Label: "Codex"},
					{Value: "cursor", Label: "Cursor"},
					{Value: "claude", Label: "Claude Code"},
				},
			})
			if err != nil {
				return opts, err
			}
			if cancelled {
				return InitOptions{}, nil
			}
			opts.Agent = value
		} else {
			opts.Agent = "codex"
		}
	}

	if strings.TrimSpace(opts.Lang) == "" {
		if s.runtime.IsTTY {
			value, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
				Message: copy.PromptDevwikiLang,
				Items: []ui.Option{
					{Value: "zh", Label: "zh"},
					{Value: "en", Label: "en"},
				},
			})
			if err != nil {
				return opts, err
			}
			if cancelled {
				return InitOptions{}, nil
			}
			opts.Lang = value
		} else {
			opts.Lang = ui.DefaultLang
		}
	}

	if len(opts.CodeDirs) == 0 {
		if !s.runtime.IsTTY {
			opts.CodeDirs = []string{"."}
		} else {
			value, err := promptLine(copy.PromptDevwikiCodeDirs, s.runtime.Workspace.CWD)
			if err != nil {
				return opts, err
			}
			opts.CodeDirs = splitCommaValues(value)
		}
	}

	if includeInstall && !opts.ScopeProvided && s.runtime.IsTTY {
		value, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
			Message: copy.PromptDevwikiScope,
			Items: []ui.Option{
				{Value: "project", Label: copy.InstallInProject},
				{Value: "global", Label: copy.InstallInHome},
			},
		})
		if err != nil {
			return opts, err
		}
		if cancelled {
			return InitOptions{}, nil
		}
		opts.Global = value == "global"
	}

	return s.normalizeInitOptions(opts)
}

func (s *Service) normalizeInitOptions(opts InitOptions) (InitOptions, error) {
	copy := ui.Messages()

	opts.ProjectName = strings.TrimSpace(opts.ProjectName)
	if opts.ProjectName == "" {
		return opts, errors.New(copy.DevwikiProjectNameRequired)
	}

	switch opts.Agent {
	case "codex", "cursor", "claude":
	case "":
		opts.Agent = "codex"
	default:
		return opts, fmt.Errorf(copy.UnsupportedAgentFmt, opts.Agent)
	}

	switch opts.Lang {
	case "", "zh", "en":
		if opts.Lang == "" {
			opts.Lang = ui.DefaultLang
		}
	default:
		return opts, fmt.Errorf(copy.DevwikiUnsupportedLangFmt, opts.Lang)
	}

	codeDirs := make([]string, 0, len(opts.CodeDirs))
	for _, raw := range opts.CodeDirs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		candidate := raw
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(s.runtime.Workspace.CWD, candidate)
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return opts, err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return opts, err
		}
		if !info.IsDir() {
			return opts, fmt.Errorf(copy.DevwikiCodeDirNotDirectoryFmt, abs)
		}
		codeDirs = append(codeDirs, abs)
	}
	if len(codeDirs) == 0 {
		return opts, errors.New(copy.DevwikiCodeDirRequired)
	}
	opts.CodeDirs = codeDirs
	return opts, nil
}

func (s *Service) normalizeLinkOptions(opts LinkOptions) (LinkOptions, error) {
	if strings.TrimSpace(opts.DevwikiRoot) == "" {
		opts.DevwikiRoot = s.runtime.Workspace.CWD
	}
	devwikiRoot, err := filepath.Abs(opts.DevwikiRoot)
	if err != nil {
		return opts, err
	}
	info, err := os.Stat(devwikiRoot)
	if err != nil {
		return opts, err
	}
	if !info.IsDir() {
		return opts, fmt.Errorf("%s is not a directory", devwikiRoot)
	}
	opts.DevwikiRoot = devwikiRoot

	if strings.TrimSpace(opts.Agent) == "" {
		opts.Agent = "codex"
	}
	switch opts.Agent {
	case "codex", "cursor", "claude":
	default:
		return opts, fmt.Errorf(ui.Messages().UnsupportedAgentFmt, opts.Agent)
	}
	switch opts.Lang {
	case "", "zh", "en":
		if opts.Lang == "" {
			opts.Lang = ui.DefaultLang
		}
	default:
		return opts, fmt.Errorf(ui.Messages().DevwikiUnsupportedLangFmt, opts.Lang)
	}
	if len(opts.CodeDirs) == 0 {
		codeDirs, err := readConfiguredCodeDirs(devwikiRoot)
		if err != nil {
			return opts, err
		}
		opts.CodeDirs = codeDirs
	}
	initOpts, err := s.normalizeInitOptions(InitOptions{
		ProjectName: "link",
		Agent:       opts.Agent,
		Lang:        opts.Lang,
		CodeDirs:    opts.CodeDirs,
	})
	if err != nil {
		return opts, err
	}
	opts.CodeDirs = initOpts.CodeDirs
	return opts, nil
}

func readConfiguredCodeDirs(devwikiRoot string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(devwikiRoot, "config", "project.yaml"))
	if err != nil {
		return nil, err
	}
	var config struct {
		CodeRepos []devwiki.CodeRepo `yaml:"code_repos"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	codeDirs := make([]string, 0, len(config.CodeRepos))
	for _, repo := range config.CodeRepos {
		if strings.TrimSpace(repo.Path) != "" {
			codeDirs = append(codeDirs, repo.Path)
		}
	}
	if len(codeDirs) == 0 {
		return nil, errors.New(ui.Messages().DevwikiCodeDirRequired)
	}
	return codeDirs, nil
}

func (s *Service) resolveTargetDir() (string, error) {
	if strings.TrimSpace(s.runtime.Workspace.CWD) == "" {
		return "", errors.New("cwd is empty")
	}
	return filepath.Abs(s.runtime.Workspace.CWD)
}

func (s *Service) resolveSelectedSkills(found []skills.Skill, opts InitOptions) ([]skills.Skill, error) {
	if len(found) == 0 {
		return nil, nil
	}
	if !s.runtime.IsTTY || opts.Yes {
		return append([]skills.Skill(nil), found...), nil
	}

	items := make([]ui.Option, 0, len(found))
	initial := make([]string, 0, len(found))
	for _, skill := range found {
		items = append(items, ui.Option{
			Value: skill.Name,
			Label: skill.Name,
			Hint:  skill.Description,
		})
		initial = append(initial, skill.Name)
	}

	selectedNames, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:         ui.Messages().PromptSelectDevwikiSkills,
		Items:           items,
		MaxVisible:      8,
		InitialSelected: initial,
		Required:        false,
	})
	if err != nil {
		return nil, err
	}
	if cancelled {
		return nil, nil
	}

	selectedSet := make(map[string]struct{}, len(selectedNames))
	for _, name := range selectedNames {
		selectedSet[name] = struct{}{}
	}
	selected := make([]skills.Skill, 0, len(selectedNames))
	for _, skill := range found {
		if _, ok := selectedSet[skill.Name]; ok {
			selected = append(selected, skill)
		}
	}
	return selected, nil
}

func (s *Service) resolveCodeRepoSkills(found []skills.Skill, yes bool) ([]skills.Skill, error) {
	candidates := filterDevwikiCodeRepoSkills(found)
	if len(candidates) == 0 {
		return nil, nil
	}
	if !s.runtime.IsTTY {
		if yes {
			return candidates, nil
		}
		return nil, nil
	}

	items := make([]ui.Option, 0, len(candidates))
	initial := make([]string, 0, len(candidates))
	for _, skill := range candidates {
		items = append(items, ui.Option{
			Value: skill.Name,
			Label: skill.Name,
			Hint:  skill.Description,
		})
		initial = append(initial, skill.Name)
	}
	selectedNames, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:         ui.Messages().PromptSelectDevwikiCodeSkills,
		Items:           items,
		MaxVisible:      4,
		InitialSelected: initial,
		Required:        false,
	})
	if err != nil || cancelled {
		return nil, err
	}
	selectedSet := make(map[string]struct{}, len(selectedNames))
	for _, name := range selectedNames {
		selectedSet[name] = struct{}{}
	}
	selected := make([]skills.Skill, 0, len(selectedNames))
	for _, skill := range candidates {
		if _, ok := selectedSet[skill.Name]; ok {
			selected = append(selected, skill)
		}
	}
	return selected, nil
}

func filterDevwikiCodeRepoSkills(found []skills.Skill) []skills.Skill {
	selected := make([]skills.Skill, 0, 2)
	for _, skill := range found {
		switch skill.Name {
		case "devwiki-query", "devwiki-code-to-doc":
			selected = append(selected, skill)
		}
	}
	return selected
}

func ensureProjectInstallGitignore(projectRoot, agent string) error {
	installDir, err := agents.ResolveSkillsDir(agent, false, projectRoot)
	if err != nil {
		return err
	}
	lockPath, err := devwikiLockPath(projectRoot, false)
	if err != nil {
		return err
	}
	return common.EnsureProjectGitignore(projectRoot, filepath.Join(projectRoot, ".cache"), installDir, lockPath)
}

func (s *Service) installSelectedSkills(projectRoot, agent string, global bool, lang string, selected []skills.Skill) error {
	installDir, err := agents.ResolveSkillsDir(agent, global, projectRoot)
	if err != nil {
		return err
	}
	if err := skills.EnsureDir(installDir); err != nil {
		return err
	}

	lockPath, err := devwikiLockPath(projectRoot, global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	source := skills.NewBuiltinSource("devwiki", lang)
	for _, skill := range selected {
		entry, err := skills.InstallSkill(installDir, source, skill)
		if err != nil {
			return err
		}
		entry.Agents = []string{agent}
		entry.AgentPaths = map[string]string{agent: entry.Path}
		lock.Entries(skills.SkillAsset)[entry.Name] = entry
	}
	return skills.SaveLock(lockPath, lock)
}

func (s *Service) updateSkills(ctx context.Context) error {
	copy := ui.Messages()
	global, err := s.resolveUpdateScope()
	if err != nil {
		return err
	}
	if err := s.migrateLegacyDevwikiSources(global); err != nil {
		return err
	}

	skillService := skillapp.NewServiceWithRuntime(s.runtime)
	results, err := skillService.CheckInstalled(ctx, global)
	if err != nil {
		return err
	}

	var devwikiResults []skills.CheckResult
	for _, result := range results {
		if isDevwikiInstalledSkill(result.Asset) {
			devwikiResults = append(devwikiResults, result)
		}
	}
	if len(devwikiResults) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.DevwikiNoSkillsTracked, ui.Reset)
		return nil
	}

	installedMissing, err := s.installMissingDevwikiBuiltinSkills(global, devwikiResults)
	if err != nil {
		return err
	}
	if installedMissing > 0 {
		fmt.Printf("%s%s%s\n", ui.Green, fmt.Sprintf(copy.DevwikiInstalledSkillsFmt, installedMissing), ui.Reset)
	}

	var outdated []skills.CheckResult
	for _, result := range devwikiResults {
		if result.Status == "outdated" {
			outdated = append(outdated, result)
		}
	}
	if len(outdated) == 0 {
		if installedMissing == 0 {
			fmt.Printf("%s%s%s\n", ui.Text, copy.AllUpToDate, ui.Reset)
		}
		return nil
	}

	selected, err := s.selectDevwikiUpdates(outdated)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	updated, err := skillService.UpdateResults(ctx, global, selected)
	if err != nil {
		return err
	}
	if updated == 0 {
		fmt.Printf("%s%s%s\n", ui.Text, copy.AllUpToDate, ui.Reset)
	}
	return nil
}

func (s *Service) installMissingDevwikiBuiltinSkills(global bool, devwikiResults []skills.CheckResult) (int, error) {
	projectRoot := s.runtime.Workspace.ProjectDir()
	if !isDevwikiDocumentRoot(projectRoot) {
		return 0, nil
	}

	lang := inferDevwikiUpdateLang(devwikiResults)
	skillsRoot, cleanup, err := devwiki.ExtractBuiltinSkills(lang)
	if err != nil {
		return 0, err
	}
	defer cleanup()

	found, err := skills.Discover(skillsRoot)
	if err != nil {
		return 0, err
	}

	installedNames := make(map[string]struct{}, len(devwikiResults))
	for _, result := range devwikiResults {
		installedNames[result.Asset.Name] = struct{}{}
	}
	missing := make([]skills.Skill, 0)
	for _, skill := range found {
		if _, ok := installedNames[skill.Name]; ok {
			continue
		}
		missing = append(missing, skill)
	}
	if len(missing) == 0 {
		return 0, nil
	}

	agentKeys := inferDevwikiUpdateAgents(devwikiResults)
	lockPath, err := devwikiLockPath(projectRoot, global)
	if err != nil {
		return 0, err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return 0, err
	}
	entries := lock.Entries(skills.SkillAsset)
	source := skills.NewBuiltinSource("devwiki", lang)
	for _, skill := range missing {
		var merged skills.InstalledAsset
		for i, agentKey := range agentKeys {
			installDir, err := agents.ResolveSkillsDir(agentKey, global, projectRoot)
			if err != nil {
				return 0, err
			}
			if err := skills.EnsureDir(installDir); err != nil {
				return 0, err
			}
			entry, err := skills.InstallSkill(installDir, source, skill)
			if err != nil {
				return 0, err
			}
			if i == 0 {
				merged = entry
				merged.Agents = []string{}
				merged.AgentPaths = map[string]string{}
			}
			merged.Agents = append(merged.Agents, agentKey)
			merged.AgentPaths[agentKey] = entry.Path
		}
		entries[merged.Name] = merged
	}
	if err := skills.SaveLock(lockPath, lock); err != nil {
		return 0, err
	}
	return len(missing), nil
}

func isDevwikiDocumentRoot(root string) bool {
	if strings.TrimSpace(root) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(root, "config", "project.yaml")); err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(root, "wiki")); err != nil || !info.IsDir() {
		return false
	}
	return true
}

func inferDevwikiUpdateLang(results []skills.CheckResult) string {
	for _, result := range results {
		source, err := skills.ParseSource(result.Asset.Source)
		if err == nil && source.Type == "builtin" && source.Builtin == "devwiki" {
			switch source.Ref {
			case "zh", "en":
				return source.Ref
			}
		}
	}
	for _, result := range results {
		switch inferDevwikiSkillLang(result.Asset.Path) {
		case "zh":
			return "zh"
		case "en":
			return "en"
		}
	}
	return ui.CurrentLang()
}

func inferDevwikiUpdateAgents(results []skills.CheckResult) []string {
	seen := map[string]struct{}{}
	for _, result := range results {
		for _, agentKey := range result.Asset.Agents {
			if _, ok := agents.Lookup(agentKey); ok {
				seen[agentKey] = struct{}{}
			}
		}
		for agentKey := range result.Asset.AgentPaths {
			if _, ok := agents.Lookup(agentKey); ok {
				seen[agentKey] = struct{}{}
			}
		}
	}
	if len(seen) == 0 {
		return []string{"codex"}
	}
	out := make([]string, 0, len(seen))
	for agentKey := range seen {
		out = append(out, agentKey)
	}
	sort.Strings(out)
	return out
}

func (s *Service) resolveUpdateScope() (bool, error) {
	projectLock := filepath.Join(s.runtime.Workspace.ProjectDir(), skills.LockFileName)
	if _, err := os.Stat(projectLock); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func (s *Service) selectDevwikiUpdates(results []skills.CheckResult) ([]skills.CheckResult, error) {
	if !s.runtime.IsTTY {
		return append([]skills.CheckResult(nil), results...), nil
	}

	items := make([]ui.Option, 0, len(results))
	initial := make([]string, 0, len(results))
	byName := make(map[string]skills.CheckResult, len(results))
	for _, result := range results {
		items = append(items, ui.Option{
			Value: result.Asset.Name,
			Label: result.Asset.Name,
			Hint:  result.Asset.Description,
		})
		initial = append(initial, result.Asset.Name)
		byName[result.Asset.Name] = result
	}
	selectedNames, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:         ui.Messages().PromptSelectDevwikiUpdates,
		Items:           items,
		MaxVisible:      8,
		InitialSelected: initial,
		Required:        false,
	})
	if err != nil {
		return nil, err
	}
	if cancelled {
		return nil, nil
	}

	selected := make([]skills.CheckResult, 0, len(selectedNames))
	for _, name := range selectedNames {
		if result, ok := byName[name]; ok {
			selected = append(selected, result)
		}
	}
	return selected, nil
}

func (s *Service) migrateLegacyDevwikiSources(global bool) error {
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	changed := false
	entries := lock.Entries(skills.SkillAsset)
	for name, entry := range entries {
		if !isDevwikiInstalledSkill(entry) || strings.HasPrefix(entry.Source, "zatools/devwiki") {
			continue
		}
		lang := inferDevwikiSkillLang(entry.Path)
		entry.Source = skills.NewBuiltinSource("devwiki", lang).Original
		entries[name] = entry
		changed = true
	}
	if !changed {
		return nil
	}
	return skills.SaveLock(lockPath, lock)
}

func isDevwikiInstalledSkill(entry skills.InstalledAsset) bool {
	return strings.HasPrefix(entry.Name, "devwiki-") || strings.HasPrefix(entry.Source, "zatools/devwiki")
}

func inferDevwikiSkillLang(installPath string) string {
	data, err := os.ReadFile(filepath.Join(installPath, "SKILL.md"))
	if err != nil {
		return ui.CurrentLang()
	}
	for _, r := range string(data) {
		if unicode.Is(unicode.Han, r) {
			return "zh"
		}
	}
	return "en"
}

func devwikiLockPath(targetDir string, global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".agents", skills.LockFileName), nil
	}
	return filepath.Join(targetDir, skills.LockFileName), nil
}

func promptLine(prompt string, defaultValue string) (string, error) {
	label := prompt
	if defaultValue != "" {
		label = fmt.Sprintf("%s [%s]", prompt, defaultValue)
	}
	fmt.Printf("%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func splitCommaValues(input string) []string {
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
