package agents

import (
	"fmt"
	"os"
	"path/filepath"

	"zatools/internal/skills"
)

// Agent 描述一个支持技能安装的代理及其目录约定。
type Agent struct {
	// Key 是命令行和锁文件中使用的稳定标识。
	Key string
	// DisplayName 是面向用户展示的名称。
	DisplayName string
	// ProjectDirs 按资产类型声明项目级安装目录，相对项目根目录解析。
	ProjectDirs map[skills.AssetKind]string
	// GlobalDirs 按资产类型声明全局级安装目录，相对用户主目录解析。
	GlobalDirs map[skills.AssetKind]string
}

var supported = []Agent{
	{
		Key:         "codex",
		DisplayName: "Codex",
		ProjectDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".agents/skills",
		},
		GlobalDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".codex/skills",
		},
	},
	{
		Key:         "cursor",
		DisplayName: "Cursor",
		ProjectDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".cursor/skills",
			skills.RuleAsset:  ".cursor/rules",
		},
		GlobalDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".cursor/skills",
		},
	},
	{
		Key:         "claude",
		DisplayName: "Claude Code",
		ProjectDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".claude/skills",
			skills.RuleAsset:  ".claude/rules",
		},
		GlobalDirs: map[skills.AssetKind]string{
			skills.SkillAsset: ".claude/skills",
		},
	},
}

// Supported 返回当前内置支持的代理列表副本。
func Supported() []Agent {
	out := make([]Agent, len(supported))
	copy(out, supported)
	return out
}

// Lookup 根据代理标识查找代理定义。
func Lookup(key string) (Agent, bool) {
	for _, agent := range supported {
		if agent.Key == key {
			return agent, true
		}
	}
	return Agent{}, false
}

// ResolveInstallDir 解析代理在给定作用域下的目标资产目录。
func ResolveInstallDir(agentKey string, kind skills.AssetKind, global bool, cwd string) (string, error) {
	agent, ok := Lookup(agentKey)
	if !ok {
		return "", fmt.Errorf("unsupported agent %q", agentKey)
	}

	dir, ok := agent.ProjectDirs[kind]
	if !global {
		if !ok {
			return "", fmt.Errorf("agent %q does not support %s assets in project scope", agentKey, kind)
		}
		return filepath.Join(cwd, dir), nil
	}

	dir, ok = agent.GlobalDirs[kind]
	if !ok {
		return "", fmt.Errorf("agent %q does not support %s assets in global scope", agentKey, kind)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, dir), nil
}

// ResolveSkillsDir 解析代理在给定作用域下的技能目录。
func ResolveSkillsDir(agentKey string, global bool, cwd string) (string, error) {
	return ResolveInstallDir(agentKey, skills.SkillAsset, global, cwd)
}

// DisplayNames 把代理标识列表转换成展示名称列表。
func DisplayNames(keys []string) []string {
	names := make([]string, 0, len(keys))
	for _, key := range keys {
		if agent, ok := Lookup(key); ok {
			names = append(names, agent.DisplayName)
		}
	}
	return names
}
