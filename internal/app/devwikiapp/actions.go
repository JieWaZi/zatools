package devwikiapp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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

	targetDir, err := s.resolveTargetDir(resolved.ProjectName)
	if err != nil {
		return err
	}
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf(copy.AlreadyExistsFmt, targetDir)
	} else if !os.IsNotExist(err) {
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
	if err := devwiki.EnsureProjectRuntimeBridge(s.runtime.Workspace.ProjectDir(), targetDir, resolved.Agent, resolved.Lang); err != nil {
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
			if err := s.installSelectedSkills(s.runtime.Workspace.ProjectDir(), resolved.Agent, resolved.Global, resolved.Lang, selected); err != nil {
				return err
			}
			spinner.Stop(fmt.Sprintf(copy.DevwikiInstalledSkillsFmt, len(selected)))
		}
	}

	gitignorePaths := []string{filepath.Join(s.runtime.Workspace.ProjectDir(), ".cache")}
	if !resolved.Global {
		installDir, err := agents.ResolveSkillsDir(resolved.Agent, false, s.runtime.Workspace.ProjectDir())
		if err != nil {
			return err
		}
		gitignorePaths = append(gitignorePaths, installDir)
		lockPath, err := devwikiLockPath(s.runtime.Workspace.ProjectDir(), false)
		if err != nil {
			return err
		}
		gitignorePaths = append(gitignorePaths, lockPath)
	}
	if err := common.EnsureProjectGitignore(s.runtime.Workspace.ProjectDir(), gitignorePaths...); err != nil {
		return err
	}

	fmt.Printf("%s%s%s\n", ui.Green, copy.Done, ui.Reset)
	ui.Note(copy.TitleQMDManualDownload, []string{
		copy.QMDManualDownloadHint,
		copy.QMDManualDownloadCommand,
	})
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

func (s *Service) resolveTargetDir(projectName string) (string, error) {
	if strings.TrimSpace(projectName) == "" {
		return "", errors.New(ui.Messages().DevwikiProjectNameRequired)
	}
	return filepath.Join(s.runtime.Workspace.ProjectDir(), "devwiki-"+devwiki.Slugify(projectName)), nil
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

	var outdated []skills.CheckResult
	for _, result := range devwikiResults {
		if result.Status == "outdated" {
			outdated = append(outdated, result)
		}
	}
	if len(outdated) == 0 {
		fmt.Printf("%s%s%s\n", ui.Text, copy.AllUpToDate, ui.Reset)
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
