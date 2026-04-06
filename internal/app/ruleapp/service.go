package ruleapp

import (
	common "zatools/internal/app/common"
	"zatools/internal/skills"
	"zatools/internal/ui"
)

// Service 编排规则文件的安装、查询和移除流程。
type Service struct {
	runtime common.Runtime
}

// AddOptions 描述 `rule add` 的命令参数。
type AddOptions struct {
	// ListOnly 只列出发现到的 rules，不执行安装。
	ListOnly bool
	// Yes 表示跳过交互确认，按默认行为继续执行。
	Yes bool
	// RuleNames 用于按名称筛选要安装的 rules。
	RuleNames []string
	// Agents 指定目标 agent；当前支持 cursor 和 claude。
	Agents []string
}

// RemoveOptions 描述 `rule remove` 的命令参数。
type RemoveOptions struct {
	// Yes 表示跳过确认；如果又没有明确给出规则名，则不会执行删除。
	Yes bool
	// All 表示删除当前项目下的全部已安装 rules。
	All bool
	// RuleNames 指定要删除的规则名称列表。
	RuleNames []string
}

// NewService 构建使用当前终端环境的规则服务。
func NewService() *Service {
	return NewServiceWithRuntime(common.DetectRuntime())
}

// NewServiceWithRuntime 允许测试或上层装配自定义运行环境。
func NewServiceWithRuntime(runtime common.Runtime) *Service {
	return &Service{runtime: runtime}
}

func defaultAgents() []string {
	return []string{"claude", "cursor"}
}

func normalizeAgents(input []string) ([]string, error) {
	return common.NormalizeAgentKeys(input, skills.RuleAsset, ui.Messages().UnsupportedRuleAgentFmt)
}

// Runtime 返回服务正在使用的运行环境副本。
func (s *Service) Runtime() common.Runtime {
	return s.runtime
}
