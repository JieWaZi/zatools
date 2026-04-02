package ui

import (
	"fmt"
	"os"
	"strings"
)

const (
	// DefaultLang 是默认展示语言。
	DefaultLang = "zh"
	langEN      = "en"
	langZH      = "zh"
)

// Catalog 收口 CLI 与交互层使用的文案。
type Catalog struct {
	RootShort                    string
	SkillShort                   string
	AddShort                     string
	ListShort                    string
	InitShort                    string
	RemoveShort                  string
	CheckShort                   string
	UpdateShort                  string
	CompletionShort              string
	CompletionLong               string
	CompletionBashShort          string
	CompletionBashLong           string
	CompletionZshShort           string
	CompletionZshLong            string
	CompletionFishShort          string
	CompletionFishLong           string
	CompletionPowerShellShort    string
	CompletionPowerShellLong     string
	CompletionNoDescFlag         string
	FlagInstallGlobally          string
	FlagListOnly                 string
	FlagSkipPrompts              string
	FlagSpecificSkills           string
	FlagTargetAgents             string
	FlagListGlobal               string
	FlagRemoveGlobal             string
	FlagSkipConfirmation         string
	FlagRemoveAll                string
	FlagCheckGlobal              string
	FlagUpdateGlobal             string
	StepParsingSource            string
	StepValidateLocalPath        string
	StepCloneRepository          string
	StepLocalPathValidated       string
	StepRepositoryCloned         string
	StepDiscoveringSkills        string
	StepLoadingAgents            string
	StepInstallingSkills         string
	TitleInstallSummary          string
	TitleAvailableSkills         string
	PromptSelectSkills           string
	PromptSelectAgents           string
	PromptSelectScope            string
	PromptRemoveSkills           string
	PromptInstallNow             string
	ProjectLabel                 string
	GlobalLabel                  string
	InstallLabel                 string
	CancelLabel                  string
	SourceLabel                  string
	ScopeLabel                   string
	ProjectDirLabel              string
	AgentsLabel                  string
	SkillsLabel                  string
	Cancelled                    string
	InstallationCancelled        string
	RemovalCancelled             string
	NoMatchesFound               string
	SearchLabel                  string
	SelectedLabel                string
	SelectedNone                 string
	MoreSelectedFmt              string
	AlwaysIncludedSuffix         string
	MultiSelectHelp              string
	SingleSelectHelp             string
	Done                         string
	DoneReviewPermissions        string
	NoSkillsTracked              string
	AllUpToDate                  string
	StatusCurrent                string
	StatusOutdated               string
	StatusInvalidSource          string
	StatusSourceError            string
	StatusHashError              string
	CreatedFmt                   string
	NoSkillsFoundInFmt           string
	FoundSkillsFmt               string
	AgentsCountFmt               string
	NoScopeSkillsFmt             string
	ProjectSkillsTitle           string
	GlobalSkillsTitle            string
	NoSkillsToRemove             string
	RemovedFmt                   string
	UpdatedFmt                   string
	InstalledCountFmt            string
	SkillNotInstalledFmt         string
	NoneRequestedSkillsFmt       string
	SourceNoLongerContainsFmt    string
	AlreadyExistsFmt             string
	UnsupportedAgentFmt          string
	InstallInHome                string
	InstallInProject             string
	InstallInHomeWithPathsFmt    string
	InstallInProjectWithPathsFmt string
	HelpUsage                    string
	HelpAliases                  string
	HelpExamples                 string
	HelpAvailableCommands        string
	HelpAdditionalTopics         string
	HelpFlags                    string
	HelpGlobalFlags              string
	HelpMoreInfoFmt              string
	HelpFlag                     string
}

var catalogs = map[string]Catalog{
	langZH: {
		RootShort:                    "AI 工具命令行",
		SkillShort:                   "管理 Agent 技能",
		AddShort:                     "安装技能包",
		ListShort:                    "列出已安装技能",
		InitShort:                    "创建新的 SKILL.md 模板",
		RemoveShort:                  "删除已安装技能",
		CheckShort:                   "检查技能是否有可用更新",
		UpdateShort:                  "更新已安装技能",
		CompletionShort:              "生成指定 Shell 的自动补全脚本",
		CompletionLong:               "为指定的 Shell 生成自动补全脚本。请查看各子命令帮助以了解具体用法。",
		CompletionBashShort:          "生成 bash 自动补全脚本",
		CompletionBashLong:           "为 bash 生成自动补全脚本。\n\n当前会话加载：\n\n  source <(%[1]s completion bash)\n\n持久生效：\n\n  %[1]s completion bash > /etc/bash_completion.d/%[1]s\n",
		CompletionZshShort:           "生成 zsh 自动补全脚本",
		CompletionZshLong:            "为 zsh 生成自动补全脚本。\n\n当前会话加载：\n\n  source <(%[1]s completion zsh)\n\n持久生效：\n\n  %[1]s completion zsh > \"${fpath[1]}/_%[1]s\"\n",
		CompletionFishShort:          "生成 fish 自动补全脚本",
		CompletionFishLong:           "为 fish 生成自动补全脚本。\n\n当前会话加载：\n\n  %[1]s completion fish | source\n\n持久生效：\n\n  %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish\n",
		CompletionPowerShellShort:    "生成 powershell 自动补全脚本",
		CompletionPowerShellLong:     "为 powershell 生成自动补全脚本。\n\n当前会话加载：\n\n  %[1]s completion powershell | Out-String | Invoke-Expression\n",
		CompletionNoDescFlag:         "禁用补全描述信息",
		FlagInstallGlobally:          "安装到全局作用域",
		FlagListOnly:                 "仅列出技能，不执行安装",
		FlagSkipPrompts:              "跳过交互提示",
		FlagSpecificSkills:           "指定技能名称",
		FlagTargetAgents:             "目标 Agent（codex、cursor、claude）",
		FlagListGlobal:               "列出全局技能",
		FlagRemoveGlobal:             "从全局作用域删除",
		FlagSkipConfirmation:         "跳过确认",
		FlagRemoveAll:                "删除全部已安装技能",
		FlagCheckGlobal:              "检查全局技能",
		FlagUpdateGlobal:             "更新全局技能",
		StepParsingSource:            "解析来源...",
		StepValidateLocalPath:        "校验本地路径...",
		StepCloneRepository:          "克隆仓库...",
		StepLocalPathValidated:       "本地路径已校验",
		StepRepositoryCloned:         "仓库已克隆",
		StepDiscoveringSkills:        "发现技能...",
		StepLoadingAgents:            "加载 Agent...",
		StepInstallingSkills:         "安装技能...",
		TitleInstallSummary:          "安装摘要",
		TitleAvailableSkills:         "可用技能",
		PromptSelectSkills:           "选择要安装的技能",
		PromptSelectAgents:           "选择要安装到哪些 Agent",
		PromptSelectScope:            "选择安装范围",
		PromptRemoveSkills:           "选择要删除的技能",
		PromptInstallNow:             "现在安装已选技能？",
		ProjectLabel:                 "项目级",
		GlobalLabel:                  "全局",
		InstallLabel:                 "安装",
		CancelLabel:                  "取消",
		SourceLabel:                  "来源",
		ScopeLabel:                   "范围",
		ProjectDirLabel:              "项目",
		AgentsLabel:                  "Agent",
		SkillsLabel:                  "技能",
		Cancelled:                    "已取消",
		InstallationCancelled:        "安装已取消",
		RemovalCancelled:             "删除已取消",
		NoMatchesFound:               "没有匹配项",
		SearchLabel:                  "搜索",
		SelectedLabel:                "已选择",
		SelectedNone:                 "无",
		MoreSelectedFmt:              "%s 等另外 %d 项",
		AlwaysIncludedSuffix:         "始终包含",
		MultiSelectHelp:              "↑↓ 移动，空格选择，回车确认",
		SingleSelectHelp:             "↑↓ 移动，回车确认",
		Done:                         "完成！",
		DoneReviewPermissions:        "完成！  请在使用前检查技能内容；它们会以 Agent 的完整权限运行。",
		NoSkillsTracked:              "锁文件中没有记录任何技能。",
		AllUpToDate:                  "所有技能都已是最新。",
		StatusCurrent:                "最新",
		StatusOutdated:               "可更新",
		StatusInvalidSource:          "来源无效",
		StatusSourceError:            "来源错误",
		StatusHashError:              "哈希错误",
		CreatedFmt:                   "已创建 %s\n",
		NoSkillsFoundInFmt:           "%s 中未发现任何技能",
		FoundSkillsFmt:               "发现 %s%d%s 个技能",
		AgentsCountFmt:               "%s  %d 个 Agent\n",
		NoScopeSkillsFmt:             "没有找到%s技能。",
		ProjectSkillsTitle:           "项目技能",
		GlobalSkillsTitle:            "全局技能",
		NoSkillsToRemove:             "没有可删除的技能。",
		RemovedFmt:                   "%s✓%s 已删除 %s\n",
		UpdatedFmt:                   "%s✓%s 已更新 %s\n",
		InstalledCountFmt:            "%s已安装 %d 个技能%s\n",
		SkillNotInstalledFmt:         "技能 %q 未安装",
		NoneRequestedSkillsFmt:       "未找到请求的技能：%v",
		SourceNoLongerContainsFmt:    "来源 %s 中已不存在 %s",
		AlreadyExistsFmt:             "%s 已存在",
		UnsupportedAgentFmt:          "不支持的 Agent %q",
		InstallInHome:                "安装到主目录",
		InstallInProject:             "安装到当前项目",
		InstallInHomeWithPathsFmt:    "安装到主目录（%s）",
		InstallInProjectWithPathsFmt: "安装到当前项目（%s）",
		HelpUsage:                    "用法",
		HelpAliases:                  "别名",
		HelpExamples:                 "示例",
		HelpAvailableCommands:        "可用命令",
		HelpAdditionalTopics:         "更多帮助主题",
		HelpFlags:                    "参数",
		HelpGlobalFlags:              "全局参数",
		HelpMoreInfoFmt:              "使用 %q 查看更多信息。",
		HelpFlag:                     "显示帮助",
	},
	langEN: {
		RootShort:                    "AI tools CLI",
		SkillShort:                   "Manage agent skills",
		AddShort:                     "Add a skill package",
		ListShort:                    "List installed skills",
		InitShort:                    "Create a new SKILL.md template",
		RemoveShort:                  "Remove installed skills",
		CheckShort:                   "Check for available skill updates",
		UpdateShort:                  "Update installed skills",
		CompletionShort:              "Generate the autocompletion script for the specified shell",
		CompletionLong:               "Generate the autocompletion script for the specified shell. See each sub-command's help for details.",
		CompletionBashShort:          "Generate the autocompletion script for bash",
		CompletionBashLong:           "Generate the autocompletion script for the bash shell.\n\nLoad for the current session:\n\n  source <(%[1]s completion bash)\n",
		CompletionZshShort:           "Generate the autocompletion script for zsh",
		CompletionZshLong:            "Generate the autocompletion script for the zsh shell.\n\nLoad for the current session:\n\n  source <(%[1]s completion zsh)\n",
		CompletionFishShort:          "Generate the autocompletion script for fish",
		CompletionFishLong:           "Generate the autocompletion script for the fish shell.\n\nLoad for the current session:\n\n  %[1]s completion fish | source\n",
		CompletionPowerShellShort:    "Generate the autocompletion script for powershell",
		CompletionPowerShellLong:     "Generate the autocompletion script for powershell.\n\nLoad for the current session:\n\n  %[1]s completion powershell | Out-String | Invoke-Expression\n",
		CompletionNoDescFlag:         "disable completion descriptions",
		FlagInstallGlobally:          "Install globally",
		FlagListOnly:                 "List skills without installing",
		FlagSkipPrompts:              "Skip prompts",
		FlagSpecificSkills:           "Specific skill names",
		FlagTargetAgents:             "Target agents (codex, cursor, claude)",
		FlagListGlobal:               "List global skills",
		FlagRemoveGlobal:             "Remove from global scope",
		FlagSkipConfirmation:         "Skip confirmation",
		FlagRemoveAll:                "Remove all installed skills",
		FlagCheckGlobal:              "Check global skills",
		FlagUpdateGlobal:             "Update global skills",
		StepParsingSource:            "Parsing source...",
		StepValidateLocalPath:        "Validating local path...",
		StepCloneRepository:          "Cloning repository...",
		StepLocalPathValidated:       "Local path validated",
		StepRepositoryCloned:         "Repository cloned",
		StepDiscoveringSkills:        "Discovering skills...",
		StepLoadingAgents:            "Loading agents...",
		StepInstallingSkills:         "Installing skills...",
		TitleInstallSummary:          "Installation Summary",
		TitleAvailableSkills:         "Available Skills",
		PromptSelectSkills:           "Select skills to install",
		PromptSelectAgents:           "Which agents do you want to install to?",
		PromptSelectScope:            "Select installation scope",
		PromptRemoveSkills:           "Select skills to remove",
		PromptInstallNow:             "Install selected skills now?",
		ProjectLabel:                 "Project",
		GlobalLabel:                  "Global",
		InstallLabel:                 "Install",
		CancelLabel:                  "Cancel",
		SourceLabel:                  "Source",
		ScopeLabel:                   "Scope",
		ProjectDirLabel:              "Project",
		AgentsLabel:                  "Agents",
		SkillsLabel:                  "Skills",
		Cancelled:                    "Cancelled",
		InstallationCancelled:        "Installation cancelled",
		RemovalCancelled:             "Removal cancelled",
		NoMatchesFound:               "No matches found",
		SearchLabel:                  "Search",
		SelectedLabel:                "Selected",
		SelectedNone:                 "(none)",
		MoreSelectedFmt:              "%s +%d more",
		AlwaysIncludedSuffix:         "always included",
		MultiSelectHelp:              "↑↓ move, space select, enter confirm",
		SingleSelectHelp:             "↑↓ move, enter confirm",
		Done:                         "Done!",
		DoneReviewPermissions:        "Done!  Review skills before use; they run with full agent permissions.",
		NoSkillsTracked:              "No skills tracked in lock file.",
		AllUpToDate:                  "All skills are up to date",
		StatusCurrent:                "current",
		StatusOutdated:               "outdated",
		StatusInvalidSource:          "invalid-source",
		StatusSourceError:            "source-error",
		StatusHashError:              "hash-error",
		CreatedFmt:                   "created %s\n",
		NoSkillsFoundInFmt:           "no skills found in %s",
		FoundSkillsFmt:               "Found %s%d%s skill%s",
		AgentsCountFmt:               "%s  %d agents\n",
		NoScopeSkillsFmt:             "No %s skills found.",
		ProjectSkillsTitle:           "Project Skills",
		GlobalSkillsTitle:            "Global Skills",
		NoSkillsToRemove:             "No skills found to remove.",
		RemovedFmt:                   "%s✓%s removed %s\n",
		UpdatedFmt:                   "%s✓%s updated %s\n",
		InstalledCountFmt:            "%sInstalled %d skill%s%s\n",
		SkillNotInstalledFmt:         "skill %q is not installed",
		NoneRequestedSkillsFmt:       "none of the requested skills were found: %v",
		SourceNoLongerContainsFmt:    "source %s no longer contains %s",
		AlreadyExistsFmt:             "%s already exists",
		UnsupportedAgentFmt:          "unsupported agent %q",
		InstallInHome:                "Install in home directory",
		InstallInProject:             "Install in current project",
		InstallInHomeWithPathsFmt:    "Install in home directory (%s)",
		InstallInProjectWithPathsFmt: "Install in current project (%s)",
		HelpUsage:                    "Usage",
		HelpAliases:                  "Aliases",
		HelpExamples:                 "Examples",
		HelpAvailableCommands:        "Available Commands",
		HelpAdditionalTopics:         "Additional Help Topics",
		HelpFlags:                    "Flags",
		HelpGlobalFlags:              "Global Flags",
		HelpMoreInfoFmt:              "Use %q for more information.",
		HelpFlag:                     "show help",
	},
}

// Messages 返回当前语言对应的文案集合。默认使用中文。
func Messages() Catalog {
	lang := strings.ToLower(strings.TrimSpace(os.Getenv("ZATOOLS_LANG")))
	if _, ok := catalogs[lang]; !ok {
		lang = DefaultLang
	}
	return catalogs[lang]
}

// StatusText 将内部状态值转换为当前语言的展示文案。
func StatusText(status string) string {
	m := Messages()
	switch status {
	case "current":
		return m.StatusCurrent
	case "outdated":
		return m.StatusOutdated
	case "invalid-source":
		return m.StatusInvalidSource
	case "source-error":
		return m.StatusSourceError
	case "hash-error":
		return m.StatusHashError
	default:
		return status
	}
}

// ScopeText 返回当前语言下的 scope 展示名称。
func ScopeText(global bool) string {
	if global {
		return Messages().GlobalLabel
	}
	return Messages().ProjectLabel
}

// FoundSkillsText 返回“发现技能数量”的本地化文案。
func FoundSkillsText(count int) string {
	m := Messages()
	if CurrentLang() == langZH {
		return fmt.Sprintf(m.FoundSkillsFmt, Green, count, Reset)
	}
	return fmt.Sprintf(m.FoundSkillsFmt, Green, count, Reset, pluralSuffix(count))
}

// InstalledSkillsText 返回“已安装技能数量”的本地化文案。
func InstalledSkillsText(count int) string {
	m := Messages()
	if CurrentLang() == langZH {
		return fmt.Sprintf(m.InstalledCountFmt, Green, count, Reset)
	}
	return fmt.Sprintf(m.InstalledCountFmt, Green, count, pluralSuffix(count), Reset)
}

// ScopeTargetsText 返回当前语言下的安装路径摘要。
func ScopeTargetsText(global bool, joinedPaths string) string {
	m := Messages()
	if joinedPaths == "" {
		if global {
			return m.InstallInHome
		}
		return m.InstallInProject
	}
	if global {
		return fmt.Sprintf(m.InstallInHomeWithPathsFmt, joinedPaths)
	}
	return fmt.Sprintf(m.InstallInProjectWithPathsFmt, joinedPaths)
}

// CurrentLang 返回当前使用的语言代码。
func CurrentLang() string {
	lang := strings.ToLower(strings.TrimSpace(os.Getenv("ZATOOLS_LANG")))
	if _, ok := catalogs[lang]; ok {
		return lang
	}
	return DefaultLang
}

func pluralSuffix(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
