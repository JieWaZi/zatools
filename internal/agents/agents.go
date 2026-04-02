package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

// Agent 描述一个支持技能安装的代理及其目录约定。
type Agent struct {
	// Key 是命令行和锁文件中使用的稳定标识。
	Key string
	// DisplayName 是面向用户展示的名称。
	DisplayName string
	// ProjectSkills 是项目级技能目录，相对项目根目录解析。
	ProjectSkills string
	// GlobalSkills 是全局级技能目录，相对用户主目录解析。
	GlobalSkills string
}

var supported = []Agent{
	{
		Key:           "codex",
		DisplayName:   "Codex",
		ProjectSkills: ".agents/skills",
		GlobalSkills:  ".codex/skills",
	},
	{
		Key:           "cursor",
		DisplayName:   "Cursor",
		ProjectSkills: ".cursor/skills",
		GlobalSkills:  ".cursor/skills",
	},
	{
		Key:           "claude",
		DisplayName:   "Claude Code",
		ProjectSkills: ".claude/skills",
		GlobalSkills:  ".claude/skills",
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

// ResolveSkillsDir 解析代理在给定作用域下的技能目录。
func ResolveSkillsDir(agentKey string, global bool, cwd string) (string, error) {
	agent, ok := Lookup(agentKey)
	if !ok {
		return "", fmt.Errorf("unsupported agent %q", agentKey)
	}

	if !global {
		return filepath.Join(cwd, agent.ProjectSkills), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, agent.GlobalSkills), nil
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
