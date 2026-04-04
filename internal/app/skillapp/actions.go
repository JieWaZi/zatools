package skillapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"zatools/internal/platform/agents"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

// Add 完成技能来源解析、技能选择、目标确认和最终安装。
func (s *Service) Add(ctx context.Context, sourceArg string, opts AddOptions) error {
	copy := ui.Messages()
	spinner := ui.NewStepPrinter()
	spinner.Start(copy.StepParsingSource)
	source, err := skills.ParseSource(sourceArg)
	if err != nil {
		return err
	}
	spinner.Stop(fmt.Sprintf("%s: %s", copy.SourceLabel, formatSourceSummary(source)))

	if source.Type == "local" {
		spinner.Start(copy.StepValidateLocalPath)
	} else {
		spinner.Start(copy.StepCloneRepository)
	}
	resolved, err := skills.ResolveSource(ctx, source)
	if err != nil {
		return err
	}
	defer resolved.Cleanup()
	searchRoot, err := resolved.SearchRoot()
	if err != nil {
		return err
	}
	if source.Type == "local" {
		spinner.Stop(copy.StepLocalPathValidated)
	} else {
		spinner.Stop(copy.StepRepositoryCloned)
	}

	spinner.Start(copy.StepDiscoveringSkills)
	found, err := skills.Discover(searchRoot)
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return fmt.Errorf(copy.NoSkillsFoundInFmt, searchRoot)
	}
	spinner.Stop(ui.FoundSkillsText(len(found)))

	if opts.ListOnly {
		return printAvailableSkills(found)
	}

	// 先确定要安装哪些技能，再确定安装目标，避免用户选择了 scope/agent 后又因为技能为空退出。
	selected, err := s.selectSkills(found, opts)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	selectedAgents, globalScope, proceed, err := s.resolveInstallTargets(opts)
	if err != nil {
		return err
	}
	if !proceed || len(selectedAgents) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.InstallationCancelled, ui.Reset)
		return nil
	}
	opts.Global = globalScope

	ui.Step(copy.StepLoadingAgents)
	fmt.Printf(copy.AgentsCountFmt, ui.Green+"◇"+ui.Reset, len(selectedAgents))

	fmt.Println()
	ui.Note(copy.TitleInstallSummary, s.buildInstallSummary(source, selected, selectedAgents, opts.Global))

	confirmed, err := confirmInstall(opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	lockPath, err := s.runtime.Workspace.LockFilePath(opts.Global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	ui.Step(copy.StepInstallingSkills)
	for _, selectedSkill := range selected {
		// 每个技能都会重新生成安装记录，并覆盖锁文件中同名条目，保证路径和哈希是最新状态。
		entry, err := s.installForAgents(source, selectedSkill, selectedAgents, opts.Global)
		if err != nil {
			return err
		}
		lock.Skills[entry.Name] = entry
	}

	if err := skills.SaveLock(lockPath, lock); err != nil {
		return err
	}

	s.printInstallResults(lock, selected)
	fmt.Println()
	fmt.Printf("%s%s%s\n", ui.Green, copy.DoneReviewPermissions, ui.Reset)
	return nil
}

// List 列出当前作用域下已经安装的技能。
func (s *Service) List(_ context.Context, global bool) error {
	copy := ui.Messages()
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	installed := sortedInstalledSkills(lock)
	if len(installed) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, fmt.Sprintf(copy.NoScopeSkillsFmt, ui.ScopeText(global)), ui.Reset)
		return nil
	}
	title := copy.ProjectSkillsTitle
	if global {
		title = copy.GlobalSkillsTitle
	}
	displayBase := s.runtime.Workspace.ProjectDir()
	if global {
		displayBase = s.runtime.Workspace.CWD
	}
	fmt.Printf("%s%s%s\n\n", ui.Bold, title, ui.Reset)
	for _, entry := range installed {
		fmt.Printf("%s%s%s %s%s%s\n", ui.Cyan, entry.Name, ui.Reset, ui.Dim, shortenPath(entry.Path, displayBase), ui.Reset)
		if entry.Description != "" {
			fmt.Printf("  %s\n", entry.Description)
		}
		if entry.Source != "" {
			fmt.Printf("  %s%s:%s %s\n", ui.Dim, copy.SourceLabel, ui.Reset, entry.Source)
		}
		if len(entry.Agents) > 0 {
			fmt.Printf("  %s%s:%s %s\n", ui.Dim, copy.AgentsLabel, ui.Reset, strings.Join(agents.DisplayNames(entry.Agents), ", "))
		}
		if len(entry.AgentPaths) > 0 {
			agentKeys := append([]string(nil), entry.Agents...)
			sort.Strings(agentKeys)
			for _, agentKey := range agentKeys {
				path := entry.AgentPaths[agentKey]
				if path == "" {
					continue
				}
				if agent, ok := agents.Lookup(agentKey); ok {
					fmt.Printf("  %s%s:%s %s\n", ui.Dim, agent.DisplayName, ui.Reset, shortenPath(path, displayBase))
				}
			}
		}
	}
	fmt.Println()
	return nil
}

// Remove 删除当前作用域下指定的已安装技能。
func (s *Service) Remove(_ context.Context, opts RemoveOptions) error {
	copy := ui.Messages()
	lockPath, err := s.runtime.Workspace.LockFilePath(opts.Global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}
	if len(lock.Skills) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.NoSkillsToRemove, ui.Reset)
		return nil
	}

	installed := sortedInstalledSkills(lock)

	names, err := s.resolveRemoveNames(installed, opts)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.RemovalCancelled, ui.Reset)
		return nil
	}

	for _, name := range names {
		entry, ok := lock.Skills[name]
		if !ok {
			return fmt.Errorf(copy.SkillNotInstalledFmt, name)
		}
		allowedRoots, err := s.resolveInstalledPathRoots(entry, opts.Global)
		if err != nil {
			return err
		}
		// 先清理各 agent 的同步副本，再清理主安装目录，兼容新旧锁文件字段差异。
		for _, path := range entry.AgentPaths {
			if err := removeInstalledPath(path, allowedRoots); err != nil {
				return err
			}
		}
		if err := removeInstalledPath(entry.Path, allowedRoots); err != nil {
			return err
		}
		delete(lock.Skills, name)
		fmt.Printf(copy.RemovedFmt, ui.Green, ui.Reset, name)
	}

	if err := skills.SaveLock(lockPath, lock); err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("%s%s%s\n", ui.Green, copy.Done, ui.Reset)
	return nil
}

// Check 检查当前作用域下的技能是否有更新。
func (s *Service) Check(ctx context.Context, global bool) error {
	copy := ui.Messages()
	results, err := s.checkInstalledSkills(ctx, global)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.NoSkillsTracked, ui.Reset)
		return nil
	}
	for _, result := range results {
		color := ui.Green
		if result.Status != "current" {
			color = ui.Yellow
		}
		fmt.Printf("%s%s%s %s\n", color, result.Skill.Name, ui.Reset, ui.StatusText(result.Status))
		if result.Message != "" {
			fmt.Printf("  %s%s%s\n", ui.Dim, result.Message, ui.Reset)
		}
	}
	return nil
}

// Update 更新当前作用域下所有已经过期的技能。
func (s *Service) Update(ctx context.Context, global bool) error {
	copy := ui.Messages()
	results, err := s.checkInstalledSkills(ctx, global)
	if err != nil {
		return err
	}
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	updated := 0
	for _, result := range results {
		if err := ctx.Err(); err != nil {
			return err
		}
		if result.Status != "outdated" {
			continue
		}
		// 更新时基于锁文件里记录的来源重新解析并重新安装，
		// 这样可以复用 add 的安装路径、哈希计算和多 agent 同步逻辑。
		source, err := skills.ParseSource(result.Skill.Source)
		if err != nil {
			return err
		}
		if result.Skill.SourceSubdir != "" {
			source.Subpath = result.Skill.SourceSubdir
		}
		resolved, err := skills.ResolveSource(ctx, source)
		if err != nil {
			return err
		}
		searchRoot, err := resolved.SearchRoot()
		if err != nil {
			_ = resolved.Cleanup()
			return err
		}
		discovered, err := skills.Discover(searchRoot)
		if err != nil {
			_ = resolved.Cleanup()
			return err
		}
		selectedSkill, ok := findDiscoveredSkill(discovered, result.Skill.Name)
		if !ok {
			_ = resolved.Cleanup()
			return fmt.Errorf(copy.SourceNoLongerContainsFmt, result.Skill.Source, result.Skill.Name)
		}
		agentKeys := result.Skill.Agents
		if len(agentKeys) == 0 {
			// 兼容旧锁文件：历史记录里没有 Agents 字段时，默认只回填到 codex。
			agentKeys = []string{"codex"}
		}
		entry, err := s.installForAgents(source, selectedSkill, agentKeys, global)
		cleanupErr := resolved.Cleanup()
		if err == nil && cleanupErr != nil {
			err = cleanupErr
		}
		if err != nil {
			return err
		}
		lock.Skills[entry.Name] = entry
		updated++
		fmt.Printf(copy.UpdatedFmt, ui.Green, ui.Reset, entry.Name)
	}
	if err := skills.SaveLock(lockPath, lock); err != nil {
		return err
	}
	if updated == 0 {
		fmt.Printf("%s%s%s\n", ui.Text, copy.AllUpToDate, ui.Reset)
	}
	return nil
}

// initSkill 在目标目录中创建一份新的 SKILL.md 模板。
func initSkill(name string) error {
	copy := ui.Messages()
	skillName := "my-skill"
	dir := "."
	if name == "" {
		skillName = "my-skill"
	} else {
		dir = filepath.Clean(name)
		skillName = skills.SanitizeName(filepath.Base(dir))
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf(copy.AlreadyExistsFmt, path)
	}

	content := fmt.Sprintf(skillTemplate, skillName, skillName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf(copy.CreatedFmt, path)
	return nil
}

const skillTemplate = `---
name: %s
description: Describe when this skill should be used
---

# %s

Describe the workflow, constraints, and expected output here.

## When to Use

- Explain the trigger conditions.

## Steps

1. Inspect the current context.
2. Apply the skill-specific workflow.
3. Report the result clearly.
`

// selectSkills 根据命令参数、发现结果和当前终端能力，决定最终要安装哪些技能。
func (s *Service) selectSkills(found []skills.Skill, opts AddOptions) ([]skills.Skill, error) {
	copy := ui.Messages()
	if len(opts.SkillNames) > 0 {
		if len(opts.SkillNames) == 1 && opts.SkillNames[0] == "*" {
			return found, nil
		}
		var selected []skills.Skill
		for _, foundSkill := range found {
			if slices.Contains(opts.SkillNames, foundSkill.Name) {
				selected = append(selected, foundSkill)
			}
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf(copy.NoneRequestedSkillsFmt, opts.SkillNames)
		}
		return selected, nil
	}

	if len(found) == 1 || opts.Yes || !s.runtime.IsTTY {
		// 非交互场景下直接选中全部技能，保证脚本模式不会卡在 TUI。
		return found, nil
	}

	items := make([]ui.Option, 0, len(found))
	for _, foundSkill := range found {
		items = append(items, ui.Option{
			Value: foundSkill.Name,
			Label: foundSkill.Name,
			Hint:  foundSkill.Description,
		})
	}

	selectedNames, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:    ui.Bold + copy.PromptSelectSkills + " " + ui.Dim + "(" + copy.MultiSelectHelp + ")" + ui.Reset,
		Items:      items,
		Required:   true,
		MaxVisible: 8,
	})
	if err != nil || cancelled {
		return nil, err
	}

	var selected []skills.Skill
	for _, name := range selectedNames {
		for _, foundSkill := range found {
			if foundSkill.Name == name {
				selected = append(selected, foundSkill)
				break
			}
		}
	}
	return selected, nil
}

// resolveInstallTargets 把安装目标拆成两个维度处理：
// 1. 安装到哪些 agent。
// 2. 安装到 project 还是 global 作用域。
func (s *Service) resolveInstallTargets(opts AddOptions) ([]string, bool, bool, error) {
	selectedAgents, proceed, err := s.resolveAgents(opts)
	if err != nil {
		return nil, false, false, err
	}
	globalScope, proceedScope, err := s.resolveScope(opts, selectedAgents)
	if err != nil {
		return nil, false, false, err
	}
	return selectedAgents, globalScope, proceed && proceedScope, nil
}

// resolveAgents 决定要写入哪些代理的技能目录。
func (s *Service) resolveAgents(opts AddOptions) ([]string, bool, error) {
	copy := ui.Messages()
	if len(opts.Agents) > 0 {
		agents, err := normalizeAgents(opts.Agents)
		return agents, true, err
	}
	if opts.Yes || !s.runtime.IsTTY {
		// 无交互时选择全部内置代理，确保一次安装即可覆盖常见使用环境。
		return []string{"codex", "cursor", "claude"}, true, nil
	}

	items := make([]ui.Option, 0, len(agents.Supported()))
	for _, agent := range agents.Supported() {
		items = append(items, ui.Option{
			Value: agent.Key,
			Label: agent.DisplayName,
		})
	}

	selected, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:    ui.Bold + copy.PromptSelectAgents + " " + ui.Dim + "(" + copy.MultiSelectHelp + ")" + ui.Reset,
		Items:      items,
		Required:   true,
		MaxVisible: 8,
	})
	if err != nil {
		return nil, false, err
	}
	if cancelled {
		return nil, false, nil
	}
	agents, err := normalizeAgents(selected)
	return agents, true, err
}

// resolveScope 决定技能写入项目目录还是用户主目录下的全局目录。
func (s *Service) resolveScope(opts AddOptions, agentKeys []string) (bool, bool, error) {
	copy := ui.Messages()
	if opts.ScopeProvided || opts.Yes || !s.runtime.IsTTY {
		return opts.Global, true, nil
	}

	// 先把两个 scope 对应的真实目标目录展示出来，帮助用户理解写入位置。
	projectHint, err := s.describeScopeTargets(agentKeys, false)
	if err != nil {
		return false, false, err
	}
	globalHint, err := s.describeScopeTargets(agentKeys, true)
	if err != nil {
		return false, false, err
	}

	selected, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
		Message: copy.PromptSelectScope,
		Items: []ui.Option{
			{Value: "project", Label: copy.ProjectLabel, Hint: projectHint},
			{Value: "global", Label: copy.GlobalLabel, Hint: globalHint},
		},
	})
	if err != nil {
		return false, false, err
	}
	if cancelled {
		return false, false, nil
	}
	return selected == "global", true, nil
}

// resolveRemoveNames 决定 remove 命令最终删除哪些技能。
func (s *Service) resolveRemoveNames(installed []skills.InstalledSkill, opts RemoveOptions) ([]string, error) {
	copy := ui.Messages()
	if opts.All {
		names := make([]string, 0, len(installed))
		for _, entry := range installed {
			names = append(names, entry.Name)
		}
		return names, nil
	}

	if len(opts.SkillNames) > 0 {
		return opts.SkillNames, nil
	}

	if !s.runtime.IsTTY || opts.Yes {
		// remove 不像 add 那样有“全部默认值”可安全采用；
		// 非交互且未显式给出删除目标时，返回空结果让上层按取消处理。
		return nil, nil
	}

	items := make([]ui.Option, 0, len(installed))
	for _, entry := range installed {
		items = append(items, ui.Option{
			Value: entry.Name,
			Label: entry.Name,
			Hint:  entry.Description,
		})
	}
	selected, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:    copy.PromptRemoveSkills,
		Items:      items,
		Required:   true,
		MaxVisible: 8,
	})
	if err != nil || cancelled {
		return nil, err
	}
	return selected, nil
}

// checkInstalledSkills 通过重新解析来源并比较目录哈希，判断每个已安装技能是否过期。
func (s *Service) checkInstalledSkills(ctx context.Context, global bool) ([]skills.CheckResult, error) {
	lockPath, err := s.runtime.Workspace.LockFilePath(global)
	if err != nil {
		return nil, err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return nil, err
	}

	var results []skills.CheckResult
	for _, entry := range sortedInstalledSkills(lock) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source, err := skills.ParseSource(entry.Source)
		if err != nil {
			results = append(results, skills.CheckResult{Skill: entry, Status: "invalid-source", Message: err.Error()})
			continue
		}
		if entry.SourceSubdir != "" {
			source.Subpath = entry.SourceSubdir
		}
		resolved, err := skills.ResolveSource(ctx, source)
		if err != nil {
			results = append(results, skills.CheckResult{Skill: entry, Status: "source-error", Message: err.Error()})
			continue
		}
		searchRoot, err := resolved.SearchRoot()
		if err != nil {
			_ = resolved.Cleanup()
			results = append(results, skills.CheckResult{Skill: entry, Status: "source-error", Message: err.Error()})
			continue
		}
		// 哈希基于来源目录当前内容计算；与锁文件记录的安装哈希不同则判定为 outdated。
		hash, hashErr := skills.HashDir(searchRoot)
		_ = resolved.Cleanup()
		if hashErr != nil {
			results = append(results, skills.CheckResult{Skill: entry, Status: "hash-error", Message: hashErr.Error()})
			continue
		}
		status := "current"
		if hash != entry.Hash {
			status = "outdated"
		}
		results = append(results, skills.CheckResult{
			Skill:      entry,
			Status:     status,
			LatestHash: hash,
		})
	}
	return results, nil
}

// printAvailableSkills 以统一格式展示来源中发现的技能列表。
func printAvailableSkills(found []skills.Skill) error {
	copy := ui.Messages()
	fmt.Println()
	ui.Step(ui.Bold + copy.TitleAvailableSkills + ui.Reset)
	for _, foundSkill := range found {
		location := foundSkill.RelativeDir
		if location == "." {
			location = filepath.Base(foundSkill.Dir)
		}
		fmt.Printf("%s%s%s %s(%s)%s\n", ui.Cyan, foundSkill.Name, ui.Reset, ui.Dim, location, ui.Reset)
		fmt.Printf("  %s%s%s\n", ui.Dim, foundSkill.Description, ui.Reset)
	}
	return nil
}

// confirmInstall 在交互模式下给用户最后一次确认机会。
func confirmInstall(skip bool) (bool, error) {
	copy := ui.Messages()
	if skip {
		return true, nil
	}

	confirm, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
		Message: copy.PromptInstallNow,
		Items: []ui.Option{
			{Value: "install", Label: copy.InstallLabel},
			{Value: "cancel", Label: copy.CancelLabel},
		},
	})
	if err != nil {
		return false, err
	}
	if cancelled || confirm != "install" {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.InstallationCancelled, ui.Reset)
		return false, nil
	}
	return true, nil
}

// installForAgents 先安装到主 agent，再把结果同步到其它 agent 目录。
// 这样可以只做一次“从来源复制到本地并生成元数据”的工作，其余目录都从主副本同步。
func (s *Service) installForAgents(source skills.Source, selectedSkill skills.Skill, agentKeys []string, global bool) (skills.InstalledSkill, error) {
	targetDirs, err := resolveAgentDirectories(agentKeys, global, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return skills.InstalledSkill{}, err
	}
	primaryAgent := agentKeys[0]
	entry, err := skills.InstallSkill(targetDirs[primaryAgent], source, selectedSkill)
	if err != nil {
		return skills.InstalledSkill{}, err
	}

	agentPaths := map[string]string{primaryAgent: entry.Path}
	if len(targetDirs) > 1 {
		remaining := make(map[string]string, len(targetDirs)-1)
		for agentKey, dir := range targetDirs {
			if agentKey == primaryAgent {
				continue
			}
			remaining[agentKey] = dir
		}
		synced, err := skills.SyncInstalledSkill(entry.Path, entry.Name, remaining)
		if err != nil {
			return skills.InstalledSkill{}, err
		}
		for key, path := range synced {
			agentPaths[key] = path
		}
	}

	entry.Agents = append([]string(nil), agentKeys...)
	entry.AgentPaths = agentPaths
	return entry, nil
}

// printInstallResults 输出安装完成后的逐技能结果和各 agent 对应路径。
func (s *Service) printInstallResults(lock skills.LockFile, selected []skills.Skill) {
	fmt.Println()
	fmt.Print(ui.InstalledSkillsText(len(selected)))
	for _, selectedSkill := range selected {
		entry := lock.Skills[selectedSkill.Name]
		fmt.Printf("%s✓%s %s\n", ui.Green, ui.Reset, entry.Name)
		for _, agentKey := range entry.Agents {
			path := entry.AgentPaths[agentKey]
			if path == "" {
				path = entry.Path
			}
			if path == "" {
				continue
			}
			if agent, ok := agents.Lookup(agentKey); ok {
				fmt.Printf("  %s→%s %s: %s\n", ui.Dim, ui.Reset, agent.DisplayName, shortenPath(path, s.runtime.Workspace.ProjectDir()))
			}
		}
	}
}

// formatSourceSummary 把来源信息格式化为人类可读摘要，用于步骤提示和安装确认。
func formatSourceSummary(source skills.Source) string {
	location := source.RepoURL
	if source.Type == "local" {
		location = source.LocalDir
	}

	var extra []string
	if source.Ref != "" {
		extra = append(extra, "@ "+ui.Yellow+source.Ref+ui.Reset)
	}
	if source.Subpath != "" {
		extra = append(extra, "("+source.Subpath+")")
	}

	if len(extra) == 0 {
		return location
	}
	return location + " " + strings.Join(extra, " ")
}

// describeScopeTargets 返回给定 scope 下各 agent 的实际安装目录摘要。
func (s *Service) describeScopeTargets(agentKeys []string, global bool) (string, error) {
	targetDirs, err := resolveAgentDirectories(agentKeys, global, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return "", err
	}

	var paths []string
	seen := map[string]bool{}
	for _, agentKey := range agentKeys {
		path := targetDirs[agentKey]
		short := shortenPath(path, s.runtime.Workspace.ProjectDir())
		if short == "" || seen[short] {
			continue
		}
		seen[short] = true
		paths = append(paths, short)
	}
	if len(paths) == 0 {
		return ui.ScopeTargetsText(global, ""), nil
	}

	return ui.ScopeTargetsText(global, strings.Join(paths, ", ")), nil
}

// buildInstallSummary 汇总来源、作用域、项目目录、agent 和技能名，用于最终确认展示。
func (s *Service) buildInstallSummary(source skills.Source, selected []skills.Skill, agentKeys []string, global bool) []string {
	copy := ui.Messages()
	scope := ui.ScopeText(global)

	lines := []string{
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.SourceLabel, ui.Reset, formatSourceSummary(source)),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.ScopeLabel, ui.Reset, scope),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.ProjectDirLabel, ui.Reset, s.runtime.Workspace.ProjectDir()),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.AgentsLabel, ui.Reset, strings.Join(agents.DisplayNames(agentKeys), ", ")),
	}

	names := make([]string, 0, len(selected))
	for _, selectedSkill := range selected {
		names = append(names, selectedSkill.Name)
	}
	lines = append(lines, fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.SkillsLabel, ui.Reset, strings.Join(names, ", ")))
	return lines
}

// shortenPath 优先把绝对路径缩写成相对当前项目目录或 home 目录的可读形式。
func shortenPath(fullPath, cwd string) string {
	if cwd != "" {
		if rel, ok := relativeToRoot(fullPath, cwd, "."); ok {
			return rel
		}
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		if rel, ok := relativeToRoot(fullPath, home, "~"); ok {
			return rel
		}
	}
	return fullPath
}

func sortedInstalledSkills(lock skills.LockFile) []skills.InstalledSkill {
	installed := make([]skills.InstalledSkill, 0, len(lock.Skills))
	for _, entry := range lock.Skills {
		installed = append(installed, entry)
	}
	sort.Slice(installed, func(i, j int) bool { return installed[i].Name < installed[j].Name })
	return installed
}

func findDiscoveredSkill(found []skills.Skill, name string) (skills.Skill, bool) {
	for _, discovered := range found {
		if discovered.Name == name {
			return discovered, true
		}
	}
	return skills.Skill{}, false
}

func (s *Service) resolveInstalledPathRoots(entry skills.InstalledSkill, global bool) ([]string, error) {
	agentKeys := append([]string(nil), entry.Agents...)
	if len(agentKeys) == 0 {
		for _, agent := range agents.Supported() {
			agentKeys = append(agentKeys, agent.Key)
		}
	}

	targetDirs, err := resolveAgentDirectories(agentKeys, global, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return nil, err
	}

	roots := make([]string, 0, len(targetDirs))
	for _, dir := range targetDirs {
		roots = append(roots, dir)
	}
	sort.Strings(roots)
	return roots, nil
}

func removeInstalledPath(path string, allowedRoots []string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := validateInstalledPath(path, allowedRoots); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func validateInstalledPath(path string, allowedRoots []string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve installed path %q: %w", path, err)
	}
	if filepath.Dir(absPath) == absPath {
		return fmt.Errorf("refuse to remove root path %q", path)
	}
	for _, root := range allowedRoots {
		rootPath, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("resolve install root %q: %w", root, err)
		}
		rel, err := filepath.Rel(rootPath, absPath)
		if err != nil {
			return fmt.Errorf("compare installed path %q with root %q: %w", path, root, err)
		}
		if rel == "." {
			return fmt.Errorf("refuse to remove install root %q directly", path)
		}
		if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil
		}
	}
	return fmt.Errorf("refuse to remove unexpected installed path %q", path)
}

func relativeToRoot(path, root, prefix string) (string, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", false
	}
	if rel == "." {
		return prefix, true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(filepath.Join(prefix, rel)), true
}

// normalizeAgents 校验并标准化 agent 标识，同时去重后排序，保证锁文件输出稳定。
func normalizeAgents(input []string) ([]string, error) {
	copy := ui.Messages()
	seen := map[string]bool{}
	var out []string
	for _, item := range input {
		if item == "claude-code" {
			item = "claude"
		}
		if _, ok := agents.Lookup(item); !ok {
			return nil, fmt.Errorf(copy.UnsupportedAgentFmt, item)
		}
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out, nil
}

// resolveAgentDirectories 把 agent 列表映射到实际落盘目录。
func resolveAgentDirectories(agentKeys []string, global bool, cwd string) (map[string]string, error) {
	dirs := make(map[string]string, len(agentKeys))
	for _, key := range agentKeys {
		dir, err := agents.ResolveSkillsDir(key, global, cwd)
		if err != nil {
			return nil, err
		}
		dirs[key] = dir
	}
	return dirs, nil
}
