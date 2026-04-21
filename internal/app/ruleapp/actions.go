package ruleapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	common "zatools/internal/app/common"
	"zatools/internal/platform/agents"
	"zatools/internal/rules"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

func (s *Service) Add(ctx context.Context, sourceArg string, opts AddOptions) error {
	copy := ui.Messages()
	spinner := ui.NewStepPrinter()
	spinner.Start(copy.StepParsingSource)
	source, err := skills.ParseSource(sourceArg)
	if err != nil {
		return err
	}
	spinner.Stop(fmt.Sprintf("%s: %s", copy.SourceLabel, common.FormatSourceSummary(source)))

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
	rulesRoot := resolveDiscoverRoot(searchRoot, source.Subpath)
	sourcePrefix, err := sourceRelativePrefix(resolved.RootDir, rulesRoot)
	if err != nil {
		return err
	}

	if source.Type == "local" {
		spinner.Stop(copy.StepLocalPathValidated)
	} else {
		spinner.Stop(copy.StepRepositoryCloned)
	}

	spinner.Start(copy.StepDiscoveringRules)
	found, err := rules.Discover(rulesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(copy.NoRulesFoundInFmt, rulesRoot)
		}
		return err
	}
	if len(found) == 0 {
		return fmt.Errorf(copy.NoRulesFoundInFmt, rulesRoot)
	}
	spinner.Stop(ui.FoundRulesText(len(found)))

	if opts.ListOnly {
		return printAvailableRules(found)
	}

	selected, err := s.selectRules(found, opts)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.Cancelled, ui.Reset)
		return nil
	}

	agentKeys, proceed, err := s.resolveAgents(opts)
	if err != nil {
		return err
	}
	if !proceed || len(agentKeys) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.InstallationCancelled, ui.Reset)
		return nil
	}

	fmt.Println()
	ui.Note(copy.TitleInstallSummary, s.buildInstallSummary(source, selected, agentKeys))

	confirmed, err := common.ConfirmInstall(opts.Yes, copy.PromptInstallRulesNow)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	lockPath, err := s.runtime.Workspace.LockFilePath(false)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	ui.Step(copy.StepInstallingRules)
	entries := lock.Entries(skills.RuleAsset)
	for _, selectedRule := range selected {
		entry, err := s.installForAgents(source, sourcePrefix, selectedRule, agentKeys)
		if err != nil {
			return err
		}
		entries[entry.Name] = entry
	}

	if err := skills.SaveLock(lockPath, lock); err != nil {
		return err
	}
	targetDirs, err := common.ResolveAgentDirectories(skills.RuleAsset, agentKeys, false, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return err
	}
	gitignorePaths := make([]string, 0, len(targetDirs)+1)
	for _, agentKey := range agentKeys {
		gitignorePaths = append(gitignorePaths, targetDirs[agentKey])
	}
	gitignorePaths = append(gitignorePaths, lockPath)
	if err := common.EnsureProjectGitignore(s.runtime.Workspace.ProjectDir(), gitignorePaths...); err != nil {
		return err
	}

	s.printInstallResults(lock, selected)
	fmt.Println()
	fmt.Printf("%s%s%s\n", ui.Green, copy.DoneReviewPermissions, ui.Reset)
	return nil
}

func (s *Service) List(_ context.Context) error {
	copy := ui.Messages()
	lockPath, err := s.runtime.Workspace.LockFilePath(false)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	installed := common.SortedInstalledAssets(lock, skills.RuleAsset)
	if len(installed) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.NoRulesTracked, ui.Reset)
		return nil
	}

	fmt.Printf("%s%s%s\n\n", ui.Bold, copy.ProjectRulesTitle, ui.Reset)
	for _, entry := range installed {
		fmt.Printf("%s%s%s %s[%s]%s %s%s%s\n",
			ui.Cyan, entry.Name, ui.Reset,
			ui.Dim, entry.DetectedAgent, ui.Reset,
			ui.Dim, common.ShortenPath(entry.Path, s.runtime.Workspace.ProjectDir()), ui.Reset,
		)
		if entry.Description != "" {
			fmt.Printf("  %s\n", entry.Description)
		}
		if entry.Source != "" {
			fmt.Printf("  %s%s:%s %s\n", ui.Dim, copy.SourceLabel, ui.Reset, entry.Source)
		}
		if len(entry.Agents) > 0 {
			fmt.Printf("  %s%s:%s %s\n", ui.Dim, copy.AgentsLabel, ui.Reset, strings.Join(agents.DisplayNames(entry.Agents), ", "))
		}
	}
	fmt.Println()
	return nil
}

func (s *Service) Remove(_ context.Context, opts RemoveOptions) error {
	copy := ui.Messages()
	lockPath, err := s.runtime.Workspace.LockFilePath(false)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}
	entries := lock.Entries(skills.RuleAsset)
	if len(entries) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.NoRulesToRemove, ui.Reset)
		return nil
	}

	names, err := s.resolveRemoveNames(common.SortedInstalledAssets(lock, skills.RuleAsset), opts)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.RemovalCancelled, ui.Reset)
		return nil
	}

	for _, name := range names {
		entry, ok := entries[name]
		if !ok {
			return fmt.Errorf(copy.RuleNotInstalledFmt, name)
		}
		allowedRoots, err := s.resolveInstalledPathRoots(entry)
		if err != nil {
			return err
		}
		for _, path := range entry.AgentPaths {
			if err := common.RemoveInstalledPath(path, allowedRoots); err != nil {
				return err
			}
		}
		if err := common.RemoveInstalledPath(entry.Path, allowedRoots); err != nil {
			return err
		}
		delete(entries, name)
		fmt.Printf(copy.RemovedFmt, ui.Green, ui.Reset, name)
	}

	if err := skills.SaveLock(lockPath, lock); err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("%s%s%s\n", ui.Green, copy.Done, ui.Reset)
	return nil
}

func (s *Service) Check(ctx context.Context) error {
	copy := ui.Messages()
	results, err := s.checkInstalledRules(ctx)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Printf("%s%s%s\n", ui.Dim, copy.NoRulesTracked, ui.Reset)
		return nil
	}
	for _, result := range results {
		color := ui.Green
		if result.Status != "current" {
			color = ui.Yellow
		}
		fmt.Printf("%s%s%s %s %s[%s]%s\n",
			color, result.Asset.Name, ui.Reset,
			ui.StatusText(result.Status),
			ui.Dim, result.Asset.DetectedAgent, ui.Reset,
		)
		if result.Message != "" {
			fmt.Printf("  %s%s%s\n", ui.Dim, result.Message, ui.Reset)
		}
	}
	return nil
}

func (s *Service) Update(ctx context.Context) error {
	copy := ui.Messages()
	results, err := s.checkInstalledRules(ctx)
	if err != nil {
		return err
	}
	lockPath, err := s.runtime.Workspace.LockFilePath(false)
	if err != nil {
		return err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return err
	}

	entries := lock.Entries(skills.RuleAsset)
	updated := 0
	for _, result := range results {
		if err := ctx.Err(); err != nil {
			return err
		}
		if result.Status != "outdated" {
			continue
		}

		source, err := skills.ParseSource(result.Asset.Source)
		if err != nil {
			return err
		}
		resolved, err := skills.ResolveSource(ctx, source)
		if err != nil {
			return err
		}
		sourcePath := filepath.Join(resolved.RootDir, filepath.FromSlash(result.Asset.SourceRelpath))
		if _, err := os.Stat(sourcePath); err != nil {
			_ = resolved.Cleanup()
			return fmt.Errorf(copy.RuleSourceNoLongerContainsFmt, result.Asset.Source, result.Asset.Name)
		}

		entry, err := s.installRuleAtPath(source, result.Asset, sourcePath)
		cleanupErr := resolved.Cleanup()
		if err == nil && cleanupErr != nil {
			err = cleanupErr
		}
		if err != nil {
			return err
		}
		entries[entry.Name] = entry
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

func (s *Service) checkInstalledRules(ctx context.Context) ([]skills.CheckResult, error) {
	lockPath, err := s.runtime.Workspace.LockFilePath(false)
	if err != nil {
		return nil, err
	}
	lock, err := skills.LoadLock(lockPath)
	if err != nil {
		return nil, err
	}

	var results []skills.CheckResult
	for _, entry := range common.SortedInstalledAssets(lock, skills.RuleAsset) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source, err := skills.ParseSource(entry.Source)
		if err != nil {
			results = append(results, skills.CheckResult{Asset: entry, Status: "invalid-source", Message: err.Error()})
			continue
		}
		resolved, err := skills.ResolveSource(ctx, source)
		if err != nil {
			results = append(results, skills.CheckResult{Asset: entry, Status: "source-error", Message: err.Error()})
			continue
		}
		sourcePath := filepath.Join(resolved.RootDir, filepath.FromSlash(entry.SourceRelpath))
		hash, hashErr := hashRuleSource(sourcePath, entry)
		_ = resolved.Cleanup()
		if hashErr != nil {
			results = append(results, skills.CheckResult{Asset: entry, Status: "hash-error", Message: hashErr.Error()})
			continue
		}

		status := "current"
		if hash != entry.Hash {
			status = "outdated"
		}
		results = append(results, skills.CheckResult{
			Asset:      entry,
			Status:     status,
			LatestHash: hash,
		})
	}
	return results, nil
}

func (s *Service) installForAgents(source skills.Source, sourcePrefix string, rule rules.Rule, agentKeys []string) (skills.InstalledAsset, error) {
	targetDirs, err := common.ResolveAgentDirectories(skills.RuleAsset, agentKeys, false, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return skills.InstalledAsset{}, err
	}
	primaryAgent := agentKeys[0]
	entry, err := installRule(targetDirs[primaryAgent], source, sourcePrefix, rule)
	if err != nil {
		return skills.InstalledAsset{}, err
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
		synced, err := skills.SyncInstalledPath(entry.Path, filepath.Base(entry.Path), remaining)
		if err != nil {
			return skills.InstalledAsset{}, err
		}
		for key, path := range synced {
			agentPaths[key] = path
		}
	}

	entry.Agents = append([]string(nil), agentKeys...)
	entry.AgentPaths = agentPaths
	return entry, nil
}

func installRule(targetRoot string, source skills.Source, sourcePrefix string, rule rules.Rule) (skills.InstalledAsset, error) {
	entry, err := skills.InstallAsset(targetRoot, source, skills.InstallSpec{
		Name:               rule.Name,
		Description:        rule.Description,
		SourcePath:         rule.Dir,
		TargetRelativePath: rule.InstallName,
		SourceRelativePath: joinRelative(sourcePrefix, rule.RelativeDir),
		SourceFiles:        ruleSourceFiles(rule),
	})
	if err != nil {
		return skills.InstalledAsset{}, err
	}
	entry.DetectedAgent = strings.Join(rule.DetectedAgents, ",")
	return entry, nil
}

func (s *Service) installRuleAtPath(source skills.Source, existing skills.InstalledAsset, sourcePath string) (skills.InstalledAsset, error) {
	agentKeys, err := common.RequiredAgentKeys(existing)
	if err != nil {
		return skills.InstalledAsset{}, err
	}

	targetDirs, err := common.ResolveAgentDirectories(skills.RuleAsset, agentKeys, false, s.runtime.Workspace.ProjectDir())
	if err != nil {
		return skills.InstalledAsset{}, err
	}

	primaryAgent := agentKeys[0]
	entry, err := skills.InstallAsset(targetDirs[primaryAgent], source, skills.InstallSpec{
		Name:               existing.Name,
		Description:        existing.Description,
		SourcePath:         sourcePath,
		TargetRelativePath: filepath.Base(existing.Path),
		SourceFiles:        existing.SourceFiles,
	})
	if err != nil {
		return skills.InstalledAsset{}, err
	}
	entry.DetectedAgent = existing.DetectedAgent
	entry.SourceSubdir = existing.SourceSubdir
	entry.SourceRelpath = existing.SourceRelpath
	entry.Agents = append([]string(nil), agentKeys...)

	agentPaths := map[string]string{primaryAgent: entry.Path}
	if len(targetDirs) > 1 {
		remaining := make(map[string]string, len(targetDirs)-1)
		for agentKey, dir := range targetDirs {
			if agentKey == primaryAgent {
				continue
			}
			remaining[agentKey] = dir
		}
		synced, err := skills.SyncInstalledPath(entry.Path, filepath.Base(entry.Path), remaining)
		if err != nil {
			return skills.InstalledAsset{}, err
		}
		for key, path := range synced {
			agentPaths[key] = path
		}
	}
	entry.AgentPaths = agentPaths
	return entry, nil
}

func (s *Service) selectRules(found []rules.Rule, opts AddOptions) ([]rules.Rule, error) {
	copy := ui.Messages()
	if len(opts.RuleNames) > 0 {
		if len(opts.RuleNames) == 1 && opts.RuleNames[0] == "*" {
			return found, nil
		}
		var selected []rules.Rule
		for _, rule := range found {
			if slices.Contains(opts.RuleNames, rule.Name) {
				selected = append(selected, rule)
			}
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf(copy.NoneRequestedRulesFmt, opts.RuleNames)
		}
		return selected, nil
	}

	if len(found) == 1 || opts.Yes || !s.runtime.IsTTY {
		return found, nil
	}

	items := make([]ui.Option, 0, len(found))
	for _, rule := range found {
		items = append(items, ui.Option{
			Value: rule.Name,
			Label: rule.Name + " [" + strings.Join(rule.DetectedAgents, ",") + "]",
			Hint:  rule.Description,
		})
	}

	selected, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:    ui.Bold + copy.PromptSelectRules + " " + ui.Dim + "(" + copy.MultiSelectHelp + ")" + ui.Reset,
		Items:      items,
		Required:   true,
		MaxVisible: 8,
	})
	if err != nil || cancelled {
		return nil, err
	}

	out := make([]rules.Rule, 0, len(selected))
	for _, name := range selected {
		for _, rule := range found {
			if rule.Name == name {
				out = append(out, rule)
				break
			}
		}
	}
	return out, nil
}

func (s *Service) resolveAgents(opts AddOptions) ([]string, bool, error) {
	copy := ui.Messages()
	if len(opts.Agents) > 0 {
		agentKeys, err := normalizeAgents(opts.Agents)
		return agentKeys, true, err
	}
	if opts.Yes || !s.runtime.IsTTY {
		return defaultAgents(), true, nil
	}

	items := make([]ui.Option, 0, len(common.SupportedAgentKeys(skills.RuleAsset)))
	for _, agentKey := range common.SupportedAgentKeys(skills.RuleAsset) {
		agent, _ := agents.Lookup(agentKey)
		items = append(items, ui.Option{
			Value: agentKey,
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
	agentKeys, err := normalizeAgents(selected)
	return agentKeys, true, err
}

func (s *Service) resolveRemoveNames(installed []skills.InstalledAsset, opts RemoveOptions) ([]string, error) {
	if opts.All {
		names := make([]string, 0, len(installed))
		for _, entry := range installed {
			names = append(names, entry.Name)
		}
		return names, nil
	}
	if len(opts.RuleNames) > 0 {
		return opts.RuleNames, nil
	}
	if !s.runtime.IsTTY || opts.Yes {
		return nil, nil
	}

	copy := ui.Messages()
	items := make([]ui.Option, 0, len(installed))
	for _, entry := range installed {
		items = append(items, ui.Option{
			Value: entry.Name,
			Label: entry.Name + " [" + entry.DetectedAgent + "]",
			Hint:  entry.Description,
		})
	}
	selected, cancelled, err := ui.SearchMultiselect(ui.SearchMultiselectOptions{
		Message:    copy.PromptRemoveRules,
		Items:      items,
		Required:   true,
		MaxVisible: 8,
	})
	if err != nil || cancelled {
		return nil, err
	}
	return selected, nil
}

func (s *Service) resolveInstalledPathRoots(entry skills.InstalledAsset) ([]string, error) {
	return common.ResolveInstalledPathRoots(entry, skills.RuleAsset, false, s.runtime.Workspace.ProjectDir())
}

func (s *Service) printInstallResults(lock skills.LockFile, selected []rules.Rule) {
	entries := lock.Entries(skills.RuleAsset)
	fmt.Println()
	fmt.Print(ui.InstalledRulesText(len(selected)))
	for _, rule := range selected {
		entry := entries[rule.Name]
		fmt.Printf("%s✓%s %s %s[%s]%s\n", ui.Green, ui.Reset, entry.Name, ui.Dim, entry.DetectedAgent, ui.Reset)
		agentKeys := append([]string(nil), entry.Agents...)
		sort.Strings(agentKeys)
		for _, agentKey := range agentKeys {
			path := entry.AgentPaths[agentKey]
			if path == "" {
				path = entry.Path
			}
			agent, ok := agents.Lookup(agentKey)
			if !ok {
				continue
			}
			fmt.Printf("  %s→%s %s: %s\n", ui.Dim, ui.Reset, agent.DisplayName, common.ShortenPath(path, s.runtime.Workspace.ProjectDir()))
		}
	}
}

func (s *Service) buildInstallSummary(source skills.Source, selected []rules.Rule, agentKeys []string) []string {
	copy := ui.Messages()
	names := make([]string, 0, len(selected))
	for _, rule := range selected {
		names = append(names, rule.Name+" ["+strings.Join(rule.DetectedAgents, ",")+"]")
	}
	return []string{
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.SourceLabel, ui.Reset, common.FormatSourceSummary(source)),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.ProjectDirLabel, ui.Reset, s.runtime.Workspace.ProjectDir()),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.AgentsLabel, ui.Reset, strings.Join(agents.DisplayNames(agentKeys), ", ")),
		fmt.Sprintf("  %s%s:%s %s", ui.Dim, copy.RulesLabel, ui.Reset, strings.Join(names, ", ")),
	}
}

func printAvailableRules(found []rules.Rule) error {
	copy := ui.Messages()
	fmt.Println()
	ui.Step(ui.Bold + copy.TitleAvailableRules + ui.Reset)
	for _, rule := range found {
		fmt.Printf("%s%s%s %s(%s)%s %s[%s]%s\n",
			ui.Cyan, rule.Name, ui.Reset,
			ui.Dim, rule.RelativeDir, ui.Reset,
			ui.Dim, strings.Join(rule.DetectedAgents, ","), ui.Reset,
		)
		if rule.Description != "" {
			fmt.Printf("  %s%s%s\n", ui.Dim, rule.Description, ui.Reset)
		}
	}
	return nil
}

func resolveDiscoverRoot(searchRoot string, subpath string) string {
	if strings.TrimSpace(subpath) != "" {
		return searchRoot
	}
	return rules.DefaultRoot(searchRoot)
}

func sourceRelativePrefix(rootDir string, discoverRoot string) (string, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve source root: %w", err)
	}
	absDiscover, err := filepath.Abs(discoverRoot)
	if err != nil {
		return "", fmt.Errorf("resolve discover root: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absDiscover)
	if err != nil {
		return "", fmt.Errorf("resolve source prefix: %w", err)
	}
	if rel == "." {
		return "", nil
	}
	return filepath.ToSlash(rel), nil
}

func joinRelative(prefix string, rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	switch rel {
	case "", ".":
		return filepath.ToSlash(strings.TrimSpace(prefix))
	}
	if prefix == "" {
		return rel
	}
	return filepath.ToSlash(filepath.Join(prefix, rel))
}

func hashRuleSource(sourcePath string, entry skills.InstalledAsset) (string, error) {
	if len(entry.SourceFiles) > 0 {
		return skills.HashSelectedFiles(sourcePath, entry.SourceFiles)
	}
	return skills.HashPath(sourcePath)
}

func ruleSourceFiles(rule rules.Rule) []string {
	if rule.RelativeDir != "." {
		return nil
	}
	return append([]string(nil), rule.Files...)
}
