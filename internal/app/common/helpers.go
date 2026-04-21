package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zatools/internal/platform/agents"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

// ConfirmInstall 在交互模式下给用户最后一次确认机会。
func ConfirmInstall(skip bool, prompt string) (bool, error) {
	copy := ui.Messages()
	if skip {
		return true, nil
	}

	confirm, cancelled, err := ui.SelectOne(ui.SelectOneOptions{
		Message: prompt,
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

// FormatSourceSummary 把来源信息格式化为人类可读摘要。
func FormatSourceSummary(source skills.Source) string {
	location := source.RepoURL
	if source.Type == "local" {
		location = source.LocalDir
	}
	if source.Type == "builtin" {
		location = source.Original
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

// ShortenPath 优先把绝对路径缩写成相对项目目录或 home 目录的可读形式。
func ShortenPath(fullPath, cwd string) string {
	if cwd != "" {
		if rel, ok := RelativeToRoot(fullPath, cwd, "."); ok {
			return rel
		}
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		if rel, ok := RelativeToRoot(fullPath, home, "~"); ok {
			return rel
		}
	}
	return fullPath
}

// RelativeToRoot 把绝对路径缩写为某个根目录下的相对显示形式。
func RelativeToRoot(path, root, prefix string) (string, bool) {
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

// SupportedAgentKeys 返回支持指定资产类型的 agent 列表。
func SupportedAgentKeys(kind skills.AssetKind) []string {
	var out []string
	for _, agent := range agents.Supported() {
		if _, ok := agent.ProjectDirs[kind]; ok {
			out = append(out, agent.Key)
		}
	}
	sort.Strings(out)
	return out
}

// NormalizeAgentKeys 校验并标准化 agent 标识，同时去重后排序。
func NormalizeAgentKeys(input []string, kind skills.AssetKind, unsupportedFmt string) ([]string, error) {
	if len(input) == 0 {
		return SupportedAgentKeys(kind), nil
	}

	allowed := make(map[string]struct{})
	for _, key := range SupportedAgentKeys(kind) {
		allowed[key] = struct{}{}
	}

	seen := map[string]bool{}
	var out []string
	for _, item := range input {
		if item == "claude-code" {
			item = "claude"
		}
		if _, ok := allowed[item]; !ok {
			return nil, fmt.Errorf(unsupportedFmt, item)
		}
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out, nil
}

// ResolveAgentDirectories 把 agent 列表映射到实际落盘目录。
func ResolveAgentDirectories(kind skills.AssetKind, agentKeys []string, global bool, cwd string) (map[string]string, error) {
	dirs := make(map[string]string, len(agentKeys))
	for _, key := range agentKeys {
		dir, err := agents.ResolveInstallDir(key, kind, global, cwd)
		if err != nil {
			return nil, err
		}
		dirs[key] = dir
	}
	return dirs, nil
}

// ResolveInstalledPathRoots 计算某个已安装资产允许删除的根目录集合。
func ResolveInstalledPathRoots(entry skills.InstalledAsset, kind skills.AssetKind, global bool, cwd string) ([]string, error) {
	agentKeys, err := RequiredAgentKeys(entry)
	if err != nil {
		return nil, err
	}

	targetDirs, err := ResolveAgentDirectories(kind, agentKeys, global, cwd)
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

// RequiredAgentKeys 返回锁文件中记录的 agent 列表；缺失时直接报错。
func RequiredAgentKeys(entry skills.InstalledAsset) ([]string, error) {
	if len(entry.Agents) == 0 {
		return nil, fmt.Errorf(ui.Messages().TrackedAssetMissingAgentsFmt, entry.Name)
	}
	return append([]string(nil), entry.Agents...), nil
}

// RemoveInstalledPath 删除已经安装的路径，并校验其位于允许的根目录下。
func RemoveInstalledPath(path string, allowedRoots []string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := ValidateInstalledPath(path, allowedRoots); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

// ValidateInstalledPath 校验待删除路径位于允许的安装根目录内。
func ValidateInstalledPath(path string, allowedRoots []string) error {
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

// SortedInstalledAssets 返回指定资产 bucket 中按名称排序后的条目。
func SortedInstalledAssets(lock skills.LockFile, kind skills.AssetKind) []skills.InstalledAsset {
	entries := lock.Entries(kind)
	installed := make([]skills.InstalledAsset, 0, len(entries))
	for _, entry := range entries {
		installed = append(installed, entry)
	}
	sort.Slice(installed, func(i, j int) bool { return installed[i].Name < installed[j].Name })
	return installed
}

// GitignoreEntryForProjectPath converts an absolute or project-relative path into a
// stable .gitignore entry. Top-level hidden directories are collapsed to the
// directory root, so ".agents/skills" becomes ".agents".
func GitignoreEntryForProjectPath(projectRoot, path string) (string, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return "", fmt.Errorf("project root is required")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is required")
	}

	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(projectRoot, candidate)
	}

	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root %q: %w", projectRoot, err)
	}
	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("relativize %q to %q: %w", absPath, absRoot, err)
	}
	if rel == "." {
		return "", fmt.Errorf("refuse to ignore the project root directly")
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside project root %q", path, projectRoot)
	}

	entry := filepath.ToSlash(filepath.Clean(rel))
	if strings.HasPrefix(entry, "./") {
		entry = strings.TrimPrefix(entry, "./")
	}
	parts := strings.Split(entry, "/")
	if len(parts) > 1 && strings.HasPrefix(parts[0], ".") {
		return parts[0], nil
	}
	return entry, nil
}

// EnsureProjectGitignore makes sure the provided project paths are ignored in the
// project-root .gitignore. Missing files are created automatically, and existing
// entries are preserved without duplication.
func EnsureProjectGitignore(projectRoot string, paths ...string) error {
	entries := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		entry, err := GitignoreEntryForProjectPath(projectRoot, path)
		if err != nil {
			return err
		}
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil
	}

	gitignorePath := filepath.Join(projectRoot, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	existing := map[string]struct{}{}
	if len(data) > 0 {
		for _, line := range strings.Split(string(data), "\n") {
			normalized := normalizeGitignoreEntry(line)
			if normalized == "" {
				continue
			}
			existing[normalized] = struct{}{}
		}
	}

	missing := make([]string, 0, len(entries))
	for _, entry := range entries {
		if _, ok := existing[normalizeGitignoreEntry(entry)]; ok {
			continue
		}
		missing = append(missing, entry)
	}
	if len(missing) == 0 {
		return nil
	}

	content := string(data)
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += strings.Join(missing, "\n") + "\n"
	return os.WriteFile(gitignorePath, []byte(content), 0o644)
}

func normalizeGitignoreEntry(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimSuffix(trimmed, "/")
	return trimmed
}
