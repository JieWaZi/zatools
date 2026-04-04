package skillapp

import (
	"context"
	"os"

	"zatools/internal/skills"
)

// Runtime 保存一次 CLI 执行中可复用的环境信息。
type Runtime struct {
	Workspace *skills.Workspace
	IsTTY     bool
}

// Service 编排技能安装、查询和移除等应用层流程。
type Service struct {
	runtime Runtime
}

// AddOptions 描述 `skill add` 的命令参数。
type AddOptions struct {
	// Global 表示安装到全局作用域，而不是当前项目作用域。
	Global bool
	// ScopeProvided 记录用户是否显式传入了作用域参数，用于决定是否还需要交互询问。
	ScopeProvided bool
	// ListOnly 只列出来源中的技能，不执行后续安装。
	ListOnly bool
	// Yes 表示跳过交互确认，使用默认行为继续执行。
	Yes bool
	// SkillNames 用于按名称筛选要安装的技能；为空时进入自动或交互选择。
	SkillNames []string
	// Agents 指定要安装到哪些代理；为空时使用默认或交互选择。
	Agents []string
}

// RemoveOptions 描述 `skill remove` 的命令参数。
type RemoveOptions struct {
	// Global 表示从全局作用域删除，而不是当前项目作用域删除。
	Global bool
	// Yes 表示跳过交互确认；如果又没有明确给出技能名，则不会执行删除。
	Yes bool
	// All 表示删除当前作用域内的全部已安装技能。
	All bool
	// SkillNames 指定要删除的技能名称列表。
	SkillNames []string
}

// NewService 构建使用当前终端环境的应用服务。
func NewService() *Service {
	return NewServiceWithRuntime(detectRuntime())
}

// NewServiceWithRuntime 允许测试或上层装配自定义运行环境。
func NewServiceWithRuntime(runtime Runtime) *Service {
	return &Service{runtime: runtime}
}

// Runtime 返回服务正在使用的运行环境副本。
func (s *Service) Runtime() Runtime {
	return s.runtime
}

func detectRuntime() Runtime {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	info, err := os.Stdout.Stat()
	isTTY := err == nil && (info.Mode()&os.ModeCharDevice) != 0

	return Runtime{
		Workspace: skills.NewWorkspace(cwd),
		IsTTY:     isTTY,
	}
}

// Init 在目标目录中创建一份新的 SKILL.md 模板。
func (s *Service) Init(_ context.Context, name string) error {
	return initSkill(name)
}
