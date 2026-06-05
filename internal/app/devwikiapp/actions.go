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

	common "zatools/internal/app/common"
	"zatools/internal/devwiki"
	"zatools/internal/platform/agents"
	"zatools/internal/qmd"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

func (s *Service) runProject(ctx context.Context, opts InitOptions, installSkills bool) error {
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
		bundle, err := s.resolveDevwikiSkills(ctx)
		if err != nil {
			return err
		}
		if bundle.cleanup != nil {
			defer bundle.cleanup()
		}

		selected, err := s.resolveSelectedSkills(bundle.skills, resolved)
		if err != nil {
			return err
		}
		if selected != nil && len(selected) > 0 {
			spinner.Start(copy.StepInstallingDevwikiSkills)
			if err := s.installSelectedSkills(targetDir, resolved.Agent, resolved.Global, bundle.source, selected); err != nil {
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

	opts.Lang = ui.DefaultLang

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

	opts.Lang = ui.DefaultLang

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

func (s *Service) installSelectedSkills(projectRoot, agent string, global bool, source skills.Source, selected []skills.Skill) error {
	return s.installSelectedSkillsForAgents(projectRoot, []string{agent}, global, source, selected)
}

func (s *Service) installSelectedSkillsForAgents(projectRoot string, agentKeys []string, global bool, source skills.Source, selected []skills.Skill) error {
	lockPath, err := devwikiLockPath(projectRoot, global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	entries := lock.Entries(skills.SkillAsset)
	for _, skill := range selected {
		var merged skills.InstalledAsset
		for index, agentKey := range agentKeys {
			installDir, err := agents.ResolveSkillsDir(agentKey, global, projectRoot)
			if err != nil {
				return err
			}
			if err := skills.EnsureDir(installDir); err != nil {
				return err
			}
			entry, err := skills.InstallSkill(installDir, source, skill)
			if err != nil {
				return err
			}
			if index == 0 {
				merged = entry
				merged.Agents = []string{}
				merged.AgentPaths = map[string]string{}
			}
			merged.Agents = append(merged.Agents, agentKey)
			merged.AgentPaths[agentKey] = entry.Path
		}
		entries[merged.Name] = merged
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

	devwikiEntries, err := s.installedDevwikiEntries(global)
	if err != nil {
		return err
	}
	if len(devwikiEntries) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.DevwikiNoSkillsTracked, ui.Reset)
		return s.refreshDevwikiQMD(ctx)
	}

	bundle, err := s.resolveDevwikiSkills(ctx)
	if err != nil {
		return err
	}
	if bundle.cleanup != nil {
		defer bundle.cleanup()
	}

	devwikiResults, err := s.checkInstalledDevwikiSkills(ctx, devwikiEntries, bundle)
	if err != nil {
		return err
	}

	installedMissing, err := s.installMissingDevwikiSkills(global, devwikiResults, bundle)
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
		return s.refreshDevwikiQMD(ctx)
	}

	selected, err := s.selectDevwikiUpdates(outdated)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	updated, err := s.updateDevwikiResults(ctx, global, selected, bundle)
	if err != nil {
		return err
	}
	if updated == 0 {
		fmt.Printf("%s%s%s\n", ui.Text, copy.AllUpToDate, ui.Reset)
	}
	return s.refreshDevwikiQMD(ctx)
}

func (s *Service) refreshDevwikiQMD(ctx context.Context) error {
	projectRoot := s.runtime.Workspace.ProjectDir()
	if !isDevwikiDocumentRoot(projectRoot) {
		return nil
	}

	models, err := qmd.ResolveModels(projectRoot, qmd.Models{})
	if err != nil {
		printDevwikiQMDWarning("qmd models", err)
		models = qmd.DefaultModels()
	}

	collections, err := qmd.LoadCollections(projectRoot)
	if err != nil {
		printDevwikiQMDWarning("qmd sync", err)
	} else {
		commands, err := qmd.BuildCollectionCommands(projectRoot, collections)
		if err != nil {
			printDevwikiQMDWarning("qmd sync", err)
		} else if err := qmd.RunCollectionCommandsInDir(ctx, projectRoot, commands, models, os.Stdout, os.Stderr); err != nil {
			printDevwikiQMDWarning("qmd sync", err)
		}
	}

	if err := qmd.RunCommandInDir(ctx, projectRoot, []string{"update"}, models, os.Stdout, os.Stderr); err != nil {
		printDevwikiQMDWarning("qmd update", err)
	}
	if err := qmd.RunCommandInDir(ctx, projectRoot, []string{"embed"}, models, os.Stdout, os.Stderr); err != nil {
		printDevwikiQMDWarning("qmd embed", err)
	}
	return nil
}

func (s *Service) installedDevwikiEntries(global bool) ([]skills.CheckResult, error) {
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return nil, err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return nil, err
	}
	var results []skills.CheckResult
	for _, entry := range common.SortedInstalledAssets(lock, skills.SkillAsset) {
		if isDevwikiInstalledSkill(entry) {
			results = append(results, skills.CheckResult{Asset: entry})
		}
	}
	return results, nil
}

func (s *Service) checkInstalledDevwikiSkills(ctx context.Context, entries []skills.CheckResult, bundle devwikiSkillsBundle) ([]skills.CheckResult, error) {
	byName := make(map[string]skills.Skill, len(bundle.skills))
	for _, skill := range bundle.skills {
		byName[skill.Name] = skill
	}
	results := make([]skills.CheckResult, 0, len(entries))
	for _, result := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		skill, ok := byName[result.Asset.Name]
		if !ok {
			result.Status = "source-error"
			result.Message = fmt.Sprintf(ui.Messages().SourceNoLongerContainsFmt, bundle.source.Original, result.Asset.Name)
			results = append(results, result)
			continue
		}
		hash, err := skills.HashDir(skill.Dir)
		if err != nil {
			result.Status = "hash-error"
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		result.LatestHash = hash
		result.Status = "current"
		if hash != result.Asset.Hash {
			result.Status = "outdated"
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) updateDevwikiResults(ctx context.Context, global bool, results []skills.CheckResult, bundle devwikiSkillsBundle) (int, error) {
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return 0, err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return 0, err
	}
	byName := make(map[string]skills.Skill, len(bundle.skills))
	for _, skill := range bundle.skills {
		byName[skill.Name] = skill
	}
	entries := lock.Entries(skills.SkillAsset)
	updated := 0
	for _, result := range results {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if result.Status != "outdated" {
			continue
		}
		skill, ok := byName[result.Asset.Name]
		if !ok {
			return 0, fmt.Errorf(ui.Messages().SourceNoLongerContainsFmt, bundle.source.Original, result.Asset.Name)
		}
		agentKeys, err := common.RequiredAgentKeys(result.Asset)
		if err != nil {
			return 0, err
		}
		if err := s.installSelectedSkillsForAgents(s.runtime.Workspace.ProjectDir(), agentKeys, global, bundle.source, []skills.Skill{skill}); err != nil {
			return 0, err
		}
		updatedLock, err := skills.LoadLock(lockPath)
		if err != nil {
			return 0, err
		}
		entries = updatedLock.Entries(skills.SkillAsset)
		lock = updatedLock
		if entry, ok := entries[result.Asset.Name]; ok {
			entries[result.Asset.Name] = entry
		}
		updated++
		fmt.Printf(ui.Messages().UpdatedFmt, ui.Green, ui.Reset, result.Asset.Name)
	}
	if err := skills.SaveLock(lockPath, lock); err != nil {
		return 0, err
	}
	return updated, nil
}

func printDevwikiQMDWarning(step string, err error) {
	fmt.Printf("%s!%s %s\n", ui.Yellow, ui.Reset, fmt.Sprintf(ui.Messages().DevwikiQMDRefreshFailedFmt, step, err))
}

func (s *Service) installMissingDevwikiSkills(global bool, devwikiResults []skills.CheckResult, bundle devwikiSkillsBundle) (int, error) {
	projectRoot := s.runtime.Workspace.ProjectDir()
	if !isDevwikiDocumentRoot(projectRoot) {
		return 0, nil
	}

	installedNames := make(map[string]struct{}, len(devwikiResults))
	for _, result := range devwikiResults {
		installedNames[result.Asset.Name] = struct{}{}
	}
	missing := make([]skills.Skill, 0)
	for _, skill := range bundle.skills {
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
			entry, err := skills.InstallSkill(installDir, bundle.source, skill)
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
		if err == nil && skills.IsDevwikiSkillsSource(source) {
			return ui.DefaultLang
		}
	}
	return ui.DefaultLang
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
		if !isDevwikiInstalledSkill(entry) || isCurrentDevwikiSource(entry.Source) {
			continue
		}
		entry.Source = skills.NewDevwikiSkillsSource("").Original
		entries[name] = entry
		changed = true
	}
	if !changed {
		return nil
	}
	return skills.SaveLock(lockPath, lock)
}

func isDevwikiInstalledSkill(entry skills.InstalledAsset) bool {
	return strings.HasPrefix(entry.Name, "devwiki-") || strings.HasPrefix(entry.Source, "zatools/devwiki") || isCurrentDevwikiSource(entry.Source)
}

func isCurrentDevwikiSource(sourceText string) bool {
	source, err := skills.ParseSource(sourceText)
	return err == nil && skills.IsDevwikiSkillsSource(source) && !strings.HasPrefix(sourceText, "zatools/devwiki")
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
