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
	RootShort                     string
	SkillShort                    string
	RuleShort                     string
	DevwikiShort                  string
	QMDShort                      string
	AddShort                      string
	ListShort                     string
	InitShort                     string
	DevwikiInitShort              string
	RemoveShort                   string
	CheckShort                    string
	UpdateShort                   string
	RuleAddShort                  string
	RuleListShort                 string
	RuleRemoveShort               string
	RuleCheckShort                string
	RuleUpdateShort               string
	CompletionShort               string
	CompletionLong                string
	CompletionBashShort           string
	CompletionBashLong            string
	CompletionZshShort            string
	CompletionZshLong             string
	CompletionFishShort           string
	CompletionFishLong            string
	CompletionPowerShellShort     string
	CompletionPowerShellLong      string
	CompletionNoDescFlag          string
	FlagInstallGlobally           string
	FlagListOnly                  string
	FlagSkipPrompts               string
	FlagSpecificSkills            string
	FlagTargetAgents              string
	FlagListGlobal                string
	FlagRemoveGlobal              string
	FlagSkipConfirmation          string
	FlagRemoveAll                 string
	FlagCheckGlobal               string
	FlagUpdateGlobal              string
	FlagSpecificRules             string
	FlagTargetRuleAgents          string
	FlagRuleListOnly              string
	FlagRemoveAllRules            string
	FlagDevwikiAgent              string
	FlagDevwikiLang               string
	FlagDevwikiCodeDir            string
	StepParsingSource             string
	StepValidateLocalPath         string
	StepCloneRepository           string
	StepLocalPathValidated        string
	StepRepositoryCloned          string
	StepDiscoveringSkills         string
	StepLoadingAgents             string
	StepInstallingSkills          string
	StepDiscoveringRules          string
	StepInstallingRules           string
	StepCreatingDevwikiProject    string
	StepInstallingDevwikiSkills   string
	StepDownloadingQMDModels      string
	TitleInstallSummary           string
	TitleAvailableSkills          string
	TitleAvailableRules           string
	TitleDevwikiSummary           string
	PromptSelectSkills            string
	PromptSelectAgents            string
	PromptSelectScope             string
	PromptRemoveSkills            string
	PromptInstallNow              string
	PromptSelectRules             string
	PromptRemoveRules             string
	PromptInstallRulesNow         string
	PromptDevwikiProjectName      string
	PromptDevwikiAgent            string
	PromptDevwikiLang             string
	PromptDevwikiCodeDirs         string
	PromptDevwikiScope            string
	PromptSelectDevwikiSkills     string
	PromptSelectDevwikiUpdates    string
	PromptCreateDevwikiNow        string
	ProjectLabel                  string
	GlobalLabel                   string
	InstallLabel                  string
	CancelLabel                   string
	SourceLabel                   string
	ScopeLabel                    string
	ProjectDirLabel               string
	AgentsLabel                   string
	SkillsLabel                   string
	RulesLabel                    string
	DevwikiCodeDirsLabel          string
	DevwikiInstalledSkillsFmt     string
	DevwikiNoSkillsTracked        string
	Cancelled                     string
	InstallationCancelled         string
	RemovalCancelled              string
	NoMatchesFound                string
	SearchLabel                   string
	SelectedLabel                 string
	SelectedNone                  string
	MoreSelectedFmt               string
	AlwaysIncludedSuffix          string
	MultiSelectHelp               string
	SingleSelectHelp              string
	Done                          string
	DoneReviewPermissions         string
	QMDModelsReady                string
	NoSkillsTracked               string
	NoRulesTracked                string
	AllUpToDate                   string
	StatusCurrent                 string
	StatusOutdated                string
	StatusInvalidSource           string
	StatusSourceError             string
	StatusHashError               string
	CreatedFmt                    string
	NoSkillsFoundInFmt            string
	FoundSkillsFmt                string
	NoRulesFoundInFmt             string
	FoundRulesFmt                 string
	AgentsCountFmt                string
	NoScopeSkillsFmt              string
	ProjectSkillsTitle            string
	GlobalSkillsTitle             string
	NoSkillsToRemove              string
	ProjectRulesTitle             string
	NoRulesToRemove               string
	RemovedFmt                    string
	UpdatedFmt                    string
	InstalledCountFmt             string
	InstalledRulesCountFmt        string
	SkillNotInstalledFmt          string
	RuleNotInstalledFmt           string
	NoneRequestedSkillsFmt        string
	NoneRequestedRulesFmt         string
	SourceNoLongerContainsFmt     string
	RuleSourceNoLongerContainsFmt string
	TrackedAssetMissingAgentsFmt  string
	AlreadyExistsFmt              string
	UnsupportedAgentFmt           string
	UnsupportedRuleAgentFmt       string
	DevwikiUnsupportedLangFmt     string
	DevwikiProjectNameRequired    string
	DevwikiCodeDirRequired        string
	DevwikiCodeDirNotDirectoryFmt string
	InstallInHome                 string
	InstallInProject              string
	InstallInHomeWithPathsFmt     string
	InstallInProjectWithPathsFmt  string
	HelpUsage                     string
	HelpAliases                   string
	HelpExamples                  string
	HelpAvailableCommands         string
	HelpAdditionalTopics          string
	HelpFlags                     string
	HelpGlobalFlags               string
	HelpMoreInfoFmt               string
	HelpFlag                      string
}

var catalogs = map[string]Catalog{
	langZH: {
		RootShort:                     "AI 工具命令行",
		SkillShort:                    "管理 Agent 技能",
		RuleShort:                     "管理 Agent 规则",
		DevwikiShort:                  "创建和初始化 DevWiki 工程",
		QMDShort:                      "使用 zatools 管理的环境执行 qmd",
		AddShort:                      "安装技能包",
		ListShort:                     "列出已安装技能",
		InitShort:                     "创建新的 SKILL.md 模板",
		DevwikiInitShort:              "生成 DevWiki 项目并安装所选 runtime skills",
		RemoveShort:                   "删除已安装技能",
		CheckShort:                    "检查技能是否有可用更新",
		UpdateShort:                   "更新已安装技能",
		RuleAddShort:                  "安装规则文件",
		RuleListShort:                 "列出已安装规则",
		RuleRemoveShort:               "删除已安装规则",
		RuleCheckShort:                "检查规则是否有可用更新",
		RuleUpdateShort:               "更新已安装规则",
		CompletionShort:               "生成指定 Shell 的自动补全脚本",
		CompletionLong:                "为指定的 Shell 生成自动补全脚本。请查看各子命令帮助以了解具体用法。",
		CompletionBashShort:           "生成 bash 自动补全脚本",
		CompletionBashLong:            "为 bash 生成自动补全脚本。\n\n当前会话加载：\n\n  source <(%[1]s completion bash)\n\n持久生效：\n\n  %[1]s completion bash > /etc/bash_completion.d/%[1]s\n",
		CompletionZshShort:            "生成 zsh 自动补全脚本",
		CompletionZshLong:             "为 zsh 生成自动补全脚本。\n\n当前会话加载：\n\n  source <(%[1]s completion zsh)\n\n持久生效：\n\n  %[1]s completion zsh > \"${fpath[1]}/_%[1]s\"\n",
		CompletionFishShort:           "生成 fish 自动补全脚本",
		CompletionFishLong:            "为 fish 生成自动补全脚本。\n\n当前会话加载：\n\n  %[1]s completion fish | source\n\n持久生效：\n\n  %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish\n",
		CompletionPowerShellShort:     "生成 powershell 自动补全脚本",
		CompletionPowerShellLong:      "为 powershell 生成自动补全脚本。\n\n当前会话加载：\n\n  %[1]s completion powershell | Out-String | Invoke-Expression\n",
		CompletionNoDescFlag:          "禁用补全描述信息",
		FlagInstallGlobally:           "安装到全局作用域",
		FlagListOnly:                  "仅列出技能，不执行安装",
		FlagSkipPrompts:               "跳过交互提示",
		FlagSpecificSkills:            "指定技能名称",
		FlagTargetAgents:              "目标 Agent（codex、cursor、claude）",
		FlagListGlobal:                "列出全局技能",
		FlagRemoveGlobal:              "从全局作用域删除",
		FlagSkipConfirmation:          "跳过确认",
		FlagRemoveAll:                 "删除全部已安装技能",
		FlagCheckGlobal:               "检查全局技能",
		FlagUpdateGlobal:              "更新全局技能",
		FlagSpecificRules:             "指定规则名称",
		FlagTargetRuleAgents:          "目标 Agent（cursor、claude）",
		FlagRuleListOnly:              "仅列出规则，不执行安装",
		FlagRemoveAllRules:            "删除全部已安装规则",
		FlagDevwikiAgent:              "目标 runtime（codex、cursor、claude）",
		FlagDevwikiLang:               "运行时语言（zh、en）",
		FlagDevwikiCodeDir:            "代码目录，可重复传入",
		StepParsingSource:             "解析来源...",
		StepValidateLocalPath:         "校验本地路径...",
		StepCloneRepository:           "克隆仓库...",
		StepLocalPathValidated:        "本地路径已校验",
		StepRepositoryCloned:          "仓库已克隆",
		StepDiscoveringSkills:         "发现技能...",
		StepLoadingAgents:             "加载 Agent...",
		StepInstallingSkills:          "安装技能...",
		StepDiscoveringRules:          "发现 rules...",
		StepInstallingRules:           "安装 rules...",
		StepCreatingDevwikiProject:    "创建 DevWiki 工程...",
		StepInstallingDevwikiSkills:   "安装 DevWiki skills...",
		StepDownloadingQMDModels:      "预热 QMD models...",
		TitleInstallSummary:           "安装摘要",
		TitleAvailableSkills:          "可用技能",
		TitleAvailableRules:           "可用规则",
		TitleDevwikiSummary:           "DevWiki 初始化摘要",
		PromptSelectSkills:            "选择要安装的技能",
		PromptSelectAgents:            "选择要安装到哪些 Agent",
		PromptSelectScope:             "选择安装范围",
		PromptRemoveSkills:            "选择要删除的技能",
		PromptInstallNow:              "现在安装已选技能？",
		PromptSelectRules:             "选择要安装的规则",
		PromptRemoveRules:             "选择要删除的规则",
		PromptInstallRulesNow:         "现在安装已选规则？",
		PromptDevwikiProjectName:      "DevWiki 项目名称",
		PromptDevwikiAgent:            "选择 DevWiki runtime",
		PromptDevwikiLang:             "选择语言",
		PromptDevwikiCodeDirs:         "代码目录（逗号分隔）",
		PromptDevwikiScope:            "选择 DevWiki skill 安装范围",
		PromptSelectDevwikiSkills:     "选择要安装的 DevWiki skills",
		PromptSelectDevwikiUpdates:    "选择要更新的 DevWiki skills",
		PromptCreateDevwikiNow:        "现在创建这个 DevWiki 工程？",
		ProjectLabel:                  "项目级",
		GlobalLabel:                   "全局",
		InstallLabel:                  "安装",
		CancelLabel:                   "取消",
		SourceLabel:                   "来源",
		ScopeLabel:                    "范围",
		ProjectDirLabel:               "项目",
		AgentsLabel:                   "Agent",
		SkillsLabel:                   "技能",
		RulesLabel:                    "Rules",
		DevwikiCodeDirsLabel:          "代码目录",
		DevwikiInstalledSkillsFmt:     "已安装 %d 个 DevWiki skills",
		DevwikiNoSkillsTracked:        "锁文件中没有记录任何 DevWiki skills。",
		Cancelled:                     "已取消",
		InstallationCancelled:         "安装已取消",
		RemovalCancelled:              "删除已取消",
		NoMatchesFound:                "没有匹配项",
		SearchLabel:                   "搜索",
		SelectedLabel:                 "已选择",
		SelectedNone:                  "无",
		MoreSelectedFmt:               "%s 等另外 %d 项",
		AlwaysIncludedSuffix:          "始终包含",
		MultiSelectHelp:               "↑↓ 移动，空格选择，回车确认",
		SingleSelectHelp:              "↑↓ 移动，回车确认",
		Done:                          "完成！",
		DoneReviewPermissions:         "完成！  请在使用前检查技能内容；它们会以 Agent 的完整权限运行。",
		QMDModelsReady:                "QMD models 已预热",
		NoSkillsTracked:               "锁文件中没有记录任何技能。",
		NoRulesTracked:                "锁文件中没有记录任何规则。",
		AllUpToDate:                   "所有技能都已是最新。",
		StatusCurrent:                 "最新",
		StatusOutdated:                "可更新",
		StatusInvalidSource:           "来源无效",
		StatusSourceError:             "来源错误",
		StatusHashError:               "哈希错误",
		CreatedFmt:                    "已创建 %s\n",
		NoSkillsFoundInFmt:            "%s 中未发现任何技能",
		FoundSkillsFmt:                "发现 %s%d%s 个技能",
		NoRulesFoundInFmt:             "%s 中未发现任何规则",
		FoundRulesFmt:                 "发现 %s%d%s 个规则",
		AgentsCountFmt:                "%s  %d 个 Agent\n",
		NoScopeSkillsFmt:              "没有找到%s技能。",
		ProjectSkillsTitle:            "项目技能",
		GlobalSkillsTitle:             "全局技能",
		NoSkillsToRemove:              "没有可删除的技能。",
		ProjectRulesTitle:             "项目规则",
		NoRulesToRemove:               "没有可删除的规则。",
		RemovedFmt:                    "%s✓%s 已删除 %s\n",
		UpdatedFmt:                    "%s✓%s 已更新 %s\n",
		InstalledCountFmt:             "%s已安装 %d 个技能%s\n",
		InstalledRulesCountFmt:        "%s已安装 %d 个规则%s\n",
		SkillNotInstalledFmt:          "技能 %q 未安装",
		RuleNotInstalledFmt:           "规则 %q 未安装",
		NoneRequestedSkillsFmt:        "未找到请求的技能：%v",
		NoneRequestedRulesFmt:         "未找到请求的规则：%v",
		SourceNoLongerContainsFmt:     "来源 %s 中已不存在 %s",
		RuleSourceNoLongerContainsFmt: "来源 %s 中已不存在 %s",
		TrackedAssetMissingAgentsFmt:  "已安装条目 %q 缺少 agents 字段，请重新安装",
		AlreadyExistsFmt:              "%s 已存在",
		UnsupportedAgentFmt:           "不支持的 Agent %q",
		UnsupportedRuleAgentFmt:       "规则目前仅支持安装到 cursor 或 claude，收到 agent: %q",
		DevwikiUnsupportedLangFmt:     "不支持的语言 %q",
		DevwikiProjectNameRequired:    "项目名称不能为空",
		DevwikiCodeDirRequired:        "至少需要一个代码目录",
		DevwikiCodeDirNotDirectoryFmt: "代码目录不是有效目录：%s",
		InstallInHome:                 "安装到主目录",
		InstallInProject:              "安装到当前项目",
		InstallInHomeWithPathsFmt:     "安装到主目录（%s）",
		InstallInProjectWithPathsFmt:  "安装到当前项目（%s）",
		HelpUsage:                     "用法",
		HelpAliases:                   "别名",
		HelpExamples:                  "示例",
		HelpAvailableCommands:         "可用命令",
		HelpAdditionalTopics:          "更多帮助主题",
		HelpFlags:                     "参数",
		HelpGlobalFlags:               "全局参数",
		HelpMoreInfoFmt:               "使用 %q 查看更多信息。",
		HelpFlag:                      "显示帮助",
	},
	langEN: {
		RootShort:                     "AI tools CLI",
		SkillShort:                    "Manage agent skills",
		RuleShort:                     "Manage agent rules",
		DevwikiShort:                  "Create and initialize DevWiki projects",
		QMDShort:                      "Run qmd with zatools-managed environment",
		AddShort:                      "Add a skill package",
		ListShort:                     "List installed skills",
		InitShort:                     "Create a new SKILL.md template",
		DevwikiInitShort:              "Generate a DevWiki project and install selected runtime skills",
		RemoveShort:                   "Remove installed skills",
		CheckShort:                    "Check for available skill updates",
		UpdateShort:                   "Update installed skills",
		RuleAddShort:                  "Add rule files",
		RuleListShort:                 "List installed rules",
		RuleRemoveShort:               "Remove installed rules",
		RuleCheckShort:                "Check for available rule updates",
		RuleUpdateShort:               "Update installed rules",
		CompletionShort:               "Generate the autocompletion script for the specified shell",
		CompletionLong:                "Generate the autocompletion script for the specified shell. See each sub-command's help for details.",
		CompletionBashShort:           "Generate the autocompletion script for bash",
		CompletionBashLong:            "Generate the autocompletion script for the bash shell.\n\nLoad for the current session:\n\n  source <(%[1]s completion bash)\n",
		CompletionZshShort:            "Generate the autocompletion script for zsh",
		CompletionZshLong:             "Generate the autocompletion script for the zsh shell.\n\nLoad for the current session:\n\n  source <(%[1]s completion zsh)\n",
		CompletionFishShort:           "Generate the autocompletion script for fish",
		CompletionFishLong:            "Generate the autocompletion script for the fish shell.\n\nLoad for the current session:\n\n  %[1]s completion fish | source\n",
		CompletionPowerShellShort:     "Generate the autocompletion script for powershell",
		CompletionPowerShellLong:      "Generate the autocompletion script for powershell.\n\nLoad for the current session:\n\n  %[1]s completion powershell | Out-String | Invoke-Expression\n",
		CompletionNoDescFlag:          "disable completion descriptions",
		FlagInstallGlobally:           "Install globally",
		FlagListOnly:                  "List skills without installing",
		FlagSkipPrompts:               "Skip prompts",
		FlagSpecificSkills:            "Specific skill names",
		FlagTargetAgents:              "Target agents (codex, cursor, claude)",
		FlagListGlobal:                "List global skills",
		FlagRemoveGlobal:              "Remove from global scope",
		FlagSkipConfirmation:          "Skip confirmation",
		FlagRemoveAll:                 "Remove all installed skills",
		FlagCheckGlobal:               "Check global skills",
		FlagUpdateGlobal:              "Update global skills",
		FlagSpecificRules:             "Specific rule names",
		FlagTargetRuleAgents:          "Target agents (cursor, claude)",
		FlagRuleListOnly:              "List rules without installing",
		FlagRemoveAllRules:            "Remove all installed rules",
		FlagDevwikiAgent:              "Target runtime (codex, cursor, claude)",
		FlagDevwikiLang:               "Runtime language (zh, en)",
		FlagDevwikiCodeDir:            "Code directory, repeatable",
		StepParsingSource:             "Parsing source...",
		StepValidateLocalPath:         "Validating local path...",
		StepCloneRepository:           "Cloning repository...",
		StepLocalPathValidated:        "Local path validated",
		StepRepositoryCloned:          "Repository cloned",
		StepDiscoveringSkills:         "Discovering skills...",
		StepLoadingAgents:             "Loading agents...",
		StepInstallingSkills:          "Installing skills...",
		StepDiscoveringRules:          "Discovering rules...",
		StepInstallingRules:           "Installing rules...",
		StepCreatingDevwikiProject:    "Creating DevWiki project...",
		StepInstallingDevwikiSkills:   "Installing DevWiki skills...",
		StepDownloadingQMDModels:      "Warming qmd models...",
		TitleInstallSummary:           "Installation Summary",
		TitleAvailableSkills:          "Available Skills",
		TitleAvailableRules:           "Available Rules",
		TitleDevwikiSummary:           "DevWiki Init Summary",
		PromptSelectSkills:            "Select skills to install",
		PromptSelectAgents:            "Which agents do you want to install to?",
		PromptSelectScope:             "Select installation scope",
		PromptRemoveSkills:            "Select skills to remove",
		PromptInstallNow:              "Install selected skills now?",
		PromptSelectRules:             "Select rules to install",
		PromptRemoveRules:             "Select rules to remove",
		PromptInstallRulesNow:         "Install selected rules now?",
		PromptDevwikiProjectName:      "DevWiki project name",
		PromptDevwikiAgent:            "Select DevWiki runtime",
		PromptDevwikiLang:             "Select language",
		PromptDevwikiCodeDirs:         "Code directories (comma-separated)",
		PromptDevwikiScope:            "Select DevWiki skill install scope",
		PromptSelectDevwikiSkills:     "Select DevWiki skills to install",
		PromptSelectDevwikiUpdates:    "Select DevWiki skills to update",
		PromptCreateDevwikiNow:        "Create this DevWiki project now?",
		ProjectLabel:                  "Project",
		GlobalLabel:                   "Global",
		InstallLabel:                  "Install",
		CancelLabel:                   "Cancel",
		SourceLabel:                   "Source",
		ScopeLabel:                    "Scope",
		ProjectDirLabel:               "Project",
		AgentsLabel:                   "Agents",
		SkillsLabel:                   "Skills",
		RulesLabel:                    "Rules",
		DevwikiCodeDirsLabel:          "Code Directories",
		DevwikiInstalledSkillsFmt:     "Installed %d DevWiki skills",
		DevwikiNoSkillsTracked:        "No DevWiki skills are tracked in the lock file.",
		Cancelled:                     "Cancelled",
		InstallationCancelled:         "Installation cancelled",
		RemovalCancelled:              "Removal cancelled",
		NoMatchesFound:                "No matches found",
		SearchLabel:                   "Search",
		SelectedLabel:                 "Selected",
		SelectedNone:                  "(none)",
		MoreSelectedFmt:               "%s +%d more",
		AlwaysIncludedSuffix:          "always included",
		MultiSelectHelp:               "↑↓ move, space select, enter confirm",
		SingleSelectHelp:              "↑↓ move, enter confirm",
		Done:                          "Done!",
		DoneReviewPermissions:         "Done!  Review skills before use; they run with full agent permissions.",
		QMDModelsReady:                "qmd models warmed",
		NoSkillsTracked:               "No skills tracked in lock file.",
		NoRulesTracked:                "No rules tracked in lock file.",
		AllUpToDate:                   "All skills are up to date",
		StatusCurrent:                 "current",
		StatusOutdated:                "outdated",
		StatusInvalidSource:           "invalid-source",
		StatusSourceError:             "source-error",
		StatusHashError:               "hash-error",
		CreatedFmt:                    "created %s\n",
		NoSkillsFoundInFmt:            "no skills found in %s",
		FoundSkillsFmt:                "Found %s%d%s skill%s",
		NoRulesFoundInFmt:             "no rules found in %s",
		FoundRulesFmt:                 "Found %s%d%s rule%s",
		AgentsCountFmt:                "%s  %d agents\n",
		NoScopeSkillsFmt:              "No %s skills found.",
		ProjectSkillsTitle:            "Project Skills",
		GlobalSkillsTitle:             "Global Skills",
		NoSkillsToRemove:              "No skills found to remove.",
		ProjectRulesTitle:             "Project Rules",
		NoRulesToRemove:               "No rules found to remove.",
		RemovedFmt:                    "%s✓%s removed %s\n",
		UpdatedFmt:                    "%s✓%s updated %s\n",
		InstalledCountFmt:             "%sInstalled %d skill%s%s\n",
		InstalledRulesCountFmt:        "%sInstalled %d rule%s%s\n",
		SkillNotInstalledFmt:          "skill %q is not installed",
		RuleNotInstalledFmt:           "rule %q is not installed",
		NoneRequestedSkillsFmt:        "none of the requested skills were found: %v",
		NoneRequestedRulesFmt:         "none of the requested rules were found: %v",
		SourceNoLongerContainsFmt:     "source %s no longer contains %s",
		RuleSourceNoLongerContainsFmt: "source %s no longer contains %s",
		TrackedAssetMissingAgentsFmt:  "installed entry %q is missing tracked agents; reinstall it",
		AlreadyExistsFmt:              "%s already exists",
		UnsupportedAgentFmt:           "unsupported agent %q",
		UnsupportedRuleAgentFmt:       "rules currently support only the cursor and claude agents, got: %q",
		DevwikiUnsupportedLangFmt:     "unsupported language %q",
		DevwikiProjectNameRequired:    "project name is required",
		DevwikiCodeDirRequired:        "at least one code directory is required",
		DevwikiCodeDirNotDirectoryFmt: "code directory is not a directory: %s",
		InstallInHome:                 "Install in home directory",
		InstallInProject:              "Install in current project",
		InstallInHomeWithPathsFmt:     "Install in home directory (%s)",
		InstallInProjectWithPathsFmt:  "Install in current project (%s)",
		HelpUsage:                     "Usage",
		HelpAliases:                   "Aliases",
		HelpExamples:                  "Examples",
		HelpAvailableCommands:         "Available Commands",
		HelpAdditionalTopics:          "Additional Help Topics",
		HelpFlags:                     "Flags",
		HelpGlobalFlags:               "Global Flags",
		HelpMoreInfoFmt:               "Use %q for more information.",
		HelpFlag:                      "show help",
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

// FoundRulesText 返回“发现规则数量”的本地化文案。
func FoundRulesText(count int) string {
	m := Messages()
	if CurrentLang() == langZH {
		return fmt.Sprintf(m.FoundRulesFmt, Green, count, Reset)
	}
	return fmt.Sprintf(m.FoundRulesFmt, Green, count, Reset, pluralSuffix(count))
}

// InstalledSkillsText 返回“已安装技能数量”的本地化文案。
func InstalledSkillsText(count int) string {
	m := Messages()
	if CurrentLang() == langZH {
		return fmt.Sprintf(m.InstalledCountFmt, Green, count, Reset)
	}
	return fmt.Sprintf(m.InstalledCountFmt, Green, count, pluralSuffix(count), Reset)
}

// InstalledRulesText 返回“已安装规则数量”的本地化文案。
func InstalledRulesText(count int) string {
	m := Messages()
	if CurrentLang() == langZH {
		return fmt.Sprintf(m.InstalledRulesCountFmt, Green, count, Reset)
	}
	return fmt.Sprintf(m.InstalledRulesCountFmt, Green, count, pluralSuffix(count), Reset)
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
