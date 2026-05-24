package devwiki

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateProjectCreatesExpectedFiles(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "devwiki-sample")
	err := GenerateProject(target, ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []CodeRepo{
			{Name: "api", Slug: "api", Path: "/tmp/api", Default: true},
			{Name: "web", Slug: "web", Path: "/tmp/web", Default: false},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	for _, rel := range []string{
		"README.md",
		"AGENTS.md",
		"config/project.yaml",
		"config/search.yaml",
		"raw/requirements/.gitkeep",
		"raw/features/.gitkeep",
		"wiki/topics/.gitkeep",
		"wiki/workflows/.gitkeep",
		"wiki/troubleshooting/.gitkeep",
		"wiki/outputs/.gitkeep",
		"wiki/index.md",
		"wiki/glossary.md",
		"wiki/log.md",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}

	for _, rel := range []string{
		"raw/api",
		"raw/code-summaries",
		"raw/postmortems",
		"wiki/documents",
		"wiki/changes",
		"wiki/graph",
		"wiki/sources",
		"wiki/modules",
		"wiki/capabilities",
		"wiki/features",
		"wiki/relations.yml",
		"wiki/open_questions.md",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err == nil {
			t.Fatalf("%s should not be generated in the simplified DevWiki model", rel)
		}
	}
}

func TestGenerateProjectKeepsOnlyUserFacingArtifacts(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "devwiki-sample")
	err := GenerateProject(target, ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "en",
		CodeRepos: []CodeRepo{
			{Name: "api", Slug: "api", Path: "/tmp/api", Default: true},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	for _, rel := range []string{
		"i18n",
		"tools",
		"setup.sh",
		"setup.ps1",
		"requirements.txt",
		"config/project.yaml.example",
		"config/search.yaml.example",
		"config/claude-settings.local.json.example",
		"config/codex-config.example.yaml",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err == nil {
			t.Fatalf("%s should not be generated", rel)
		}
	}

	configEntries, err := os.ReadDir(filepath.Join(target, "config"))
	if err != nil {
		t.Fatalf("ReadDir(config) error = %v", err)
	}
	got := make([]string, 0, len(configEntries))
	for _, entry := range configEntries {
		got = append(got, entry.Name())
	}
	if len(got) != 2 || got[0] != "project.yaml" || got[1] != "search.yaml" {
		t.Fatalf("config entries = %#v, want [project.yaml search.yaml]", got)
	}
}

func TestGenerateProjectCreatesOnlySelectedRuntimeFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		agent       string
		presentFile string
		absentFile  string
	}{
		{name: "codex", agent: "codex", presentFile: "AGENTS.md", absentFile: "CLAUDE.md"},
		{name: "cursor", agent: "cursor", presentFile: "AGENTS.md", absentFile: "CLAUDE.md"},
		{name: "claude", agent: "claude", presentFile: "CLAUDE.md", absentFile: "AGENTS.md"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := filepath.Join(t.TempDir(), "devwiki-sample")
			err := GenerateProject(target, ProjectSpec{
				ProjectName: "Sample",
				ProjectSlug: "sample",
				Agent:       tc.agent,
				Lang:        "zh",
				CodeRepos: []CodeRepo{
					{Name: "api", Slug: "api", Path: "/tmp/api", Default: true},
				},
			})
			if err != nil {
				t.Fatalf("GenerateProject error = %v", err)
			}

			if _, err := os.Stat(filepath.Join(target, tc.presentFile)); err != nil {
				t.Fatalf("missing runtime file %s: %v", tc.presentFile, err)
			}
			if _, err := os.Stat(filepath.Join(target, tc.absentFile)); err == nil {
				t.Fatalf("%s should not be generated for agent %s", tc.absentFile, tc.agent)
			}
		})
	}
}

func TestGenerateProjectRendersReadmeAndRuntimeTemplates(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "devwiki-sample")
	err := GenerateProject(target, ProjectSpec{
		ProjectName: "Sample Project",
		ProjectSlug: "sample-project",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []CodeRepo{
			{Name: "go-skills", Slug: "go-skills", Path: "/tmp/go-skills", Default: true},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	readmeData, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	readme := string(readmeData)
	if !containsAll(readme,
		"# DevWiki",
		"结构化 Wiki 与代码检索工作流",
		"topics` 合并能力边界与功能规则",
		"zatools devwiki init",
		"zatools devwiki graph",
		"zatools devwiki check document",
		"zatools devwiki check graph",
		"zatools devwiki read topic <slug> --view card",
		"zatools qmd sync --root . --apply",
		"zatools qmd update",
		"zatools qmd status",
		"zatools devwiki tool reset --scope wiki --project-root .",
		"直接把当前工作目录作为 DevWiki 文档库根目录",
		"zatools devwiki link --root . --code-dir",
		"devwiki-qmd-sync",
		"devwiki-project-router",
		"devwiki-maintain",
		"devwiki-query",
		"devwiki-code-to-doc",
		"/tmp/go-skills",
		"./",
		"├── AGENTS.md",
		"├── config/",
		"├── raw/",
		"└── wiki/",
		"raw/requirements/",
		"raw/designs/",
		"raw/features/",
		"raw/tests/",
		"wiki/topics/",
		"wiki/workflows/",
		"wiki/troubleshooting/",
		"wiki/glossary.md",
	) {
		t.Fatalf("README.md content missing expected latest Go guidance:\n%s", readme)
	}
	if containsAny(readme, "{{", "}}", "Python 3.11+", ".venv", "setup.sh", "setup.ps1", "├── i18n/", "├── search/", "├── tools/", "wiki/documents/", "wiki/changes/", "wiki/graph/", "wiki/sources/", "wiki/modules/", "wiki/open_questions.md", "raw/api/", "raw/code-summaries/", "raw/postmortems/", "devwiki-<project>") {
		t.Fatalf("README.md still contains unresolved placeholders or outdated paths:\n%s", readme)
	}

}

func TestGenerateProjectRendersLatestRuntimeTemplates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		agent       string
		runtimeFile string
	}{
		{name: "codex", agent: "codex", runtimeFile: "AGENTS.md"},
		{name: "cursor", agent: "cursor", runtimeFile: "AGENTS.md"},
		{name: "claude", agent: "claude", runtimeFile: "CLAUDE.md"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			target := filepath.Join(t.TempDir(), "devwiki-sample")
			err := GenerateProject(target, ProjectSpec{
				ProjectName: "Sample Project",
				ProjectSlug: "sample-project",
				Agent:       tc.agent,
				Lang:        "zh",
				CodeRepos: []CodeRepo{
					{Name: "go-skills", Slug: "go-skills", Path: "/tmp/go-skills", Default: true},
				},
			})
			if err != nil {
				t.Fatalf("GenerateProject error = %v", err)
			}

			data, err := os.ReadFile(filepath.Join(target, tc.runtimeFile))
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", tc.runtimeFile, err)
			}
			content := string(data)
			if !containsAll(content,
				"# DevWiki — 运行时规则",
				"由 Claude Code 与 Codex 共同驱动",
				"只定义项目级目录、链接规范、检索顺序与 workflow 约束",
				"./",
				"本目录就是 DevWiki 文档库根目录",
				"代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录",
				"查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源",
				"使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库",
				"使用 `zatools devwiki link`",
				"devwiki-project-router",
				"devwiki-maintain",
				"devwiki-query",
				"devwiki-code-to-doc",
				"wiki/topics/{slug}.md",
				"wiki/workflows/{slug}.md",
				"wiki/troubleshooting/{slug}.md",
				"wiki/log.md",
				"raw/requirements/",
				"raw/designs/",
				"raw/features/",
				"raw/tests/",
				"├── config/",
				"└── wiki/",
				"具体页面模板、字段结构、页面边界和证据写入规则由对应 DevWiki skill 和 `references/` 维护",
			) {
				t.Fatalf("%s content missing expected latest runtime guidance:\n%s", tc.runtimeFile, content)
			}
			if containsAny(content, "{{", "}}", "setup.sh", "setup.ps1", "i18n/", "project.yaml.example", "claude-settings.local.json.example", "codex-config.example.yaml", "search/", "tools/", "wiki/documents/{doc-type}/{slug}.md", "wiki/changes/{slug}.md", "wiki/graph/", "wiki/sources/", "wiki/modules/", "wiki/capabilities/{slug}.md", "wiki/features/{slug}.md", "wiki/relations.yml", "wiki/open_questions.md", "raw/api/", "raw/code-summaries/", "raw/postmortems/", "## 页面模板", "## Source 结构", "## Code Ref 结构") {
				t.Fatalf("%s still contains unresolved placeholders or outdated paths:\n%s", tc.runtimeFile, content)
			}
		})
	}
}

func TestGenerateProjectWritesSlimProjectConfigSchema(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "devwiki-sample")
	err := GenerateProject(target, ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []CodeRepo{
			{Name: "go-skills", Slug: "go-skills", Path: "/tmp/go-skills", Default: true},
			{Name: "api", Slug: "api", Path: "/tmp/api", Default: false},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(target, "config", "project.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(project.yaml) error = %v", err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("yaml.Unmarshal error = %v", err)
	}

	for _, key := range []string{"project_name", "project_slug", "agent", "language", "code_repos"} {
		if _, ok := config[key]; !ok {
			t.Fatalf("project.yaml missing key %q: %#v", key, config)
		}
	}
	for _, key := range []string{"default_agent", "default_language", "default_code_repo", "default_code_dir", "confirmation_mode"} {
		if _, ok := config[key]; ok {
			t.Fatalf("project.yaml should not contain %q: %#v", key, config)
		}
	}
}

func TestGenerateProjectWritesQmdSearchConfig(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "devwiki-sample")
	err := GenerateProject(target, ProjectSpec{
		ProjectName: "Sample",
		ProjectSlug: "sample",
		Agent:       "codex",
		Lang:        "zh",
		CodeRepos: []CodeRepo{
			{Name: "go-skills", Slug: "go-skills", Path: "/tmp/go-skills", Default: true},
			{Name: "api", Slug: "api", Path: "/tmp/api", Default: false},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(target, "config", "search.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(search.yaml) error = %v", err)
	}

	var config map[string]map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("yaml.Unmarshal error = %v", err)
	}

	qmdConfig, ok := config["qmd"]
	if !ok {
		t.Fatalf("search.yaml missing qmd config: %#v", config)
	}

	const wantEmbedModel = "hf:Qwen/Qwen3-Embedding-0.6B-GGUF/Qwen3-Embedding-0.6B-Q8_0.gguf"
	const wantRerankModel = "hf:ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF/qwen3-reranker-0.6b-q8_0.gguf"
	const wantGenerateModel = "hf:tobil/qmd-query-expansion-1.7B-gguf/qmd-query-expansion-1.7B-q4_k_m.gguf"
	if got := qmdConfig["embed_model"]; got != wantEmbedModel {
		t.Fatalf("embed_model = %#v, want %q", got, wantEmbedModel)
	}
	if got := qmdConfig["rerank_model"]; got != wantRerankModel {
		t.Fatalf("rerank_model = %#v, want %q", got, wantRerankModel)
	}
	if got := qmdConfig["generate_model"]; got != wantGenerateModel {
		t.Fatalf("generate_model = %#v, want %q", got, wantGenerateModel)
	}

	rawCollections, ok := qmdConfig["collections"].([]any)
	if !ok {
		t.Fatalf("collections has unexpected type: %#v", qmdConfig["collections"])
	}
	if len(rawCollections) != 1 {
		t.Fatalf("collections len = %d, want only wiki collection: %#v", len(rawCollections), rawCollections)
	}
	collection, ok := rawCollections[0].(map[string]any)
	if !ok {
		t.Fatalf("collection has unexpected type: %#v", rawCollections[0])
	}
	if collection["name"] != "devwiki-sample-wiki" || collection["path"] != "wiki" {
		t.Fatalf("collection = %#v, want devwiki-sample-wiki -> wiki", collection)
	}
}

func TestSlugifyProducesStableDirectoryNames(t *testing.T) {
	t.Parallel()

	if got := Slugify("Sample Project"); got != "sample-project" {
		t.Fatalf("Slugify = %q, want %q", got, "sample-project")
	}
}

func TestExtractBuiltinSkillsMaterializesSharedReferencesIntoEachSkill(t *testing.T) {
	t.Parallel()

	if got := BuiltinSkillsPath("zh"); got != "template/skills" {
		t.Fatalf("BuiltinSkillsPath = %q, want template/skills", got)
	}
	if _, err := fs.Stat(TemplateFS(), "template/skills/shared-references/code-tracing.md"); err != nil {
		t.Fatalf("missing flat shared reference path: %v", err)
	}

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", root, err)
	}
	if len(entries) == 0 {
		t.Fatal("expected extracted builtin skills")
	}
	gotNames := make([]string, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "shared-references" {
			t.Fatal("extracted builtin skills should not include shared-references as an installable directory")
		}
		gotNames = append(gotNames, entry.Name())
		for _, shared := range []string{"code-tracing.md", "common-file-format.md", "zatools-qmd.md"} {
			target := filepath.Join(root, entry.Name(), "references", shared)
			if _, err := os.Stat(target); err != nil {
				t.Fatalf("missing shared reference %s for %s: %v", shared, entry.Name(), err)
			}
		}
	}

	if _, err := os.Stat(filepath.Join(root, "qmd-sync", "SKILL.md")); err != nil {
		t.Fatalf("missing qmd-sync builtin skill: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "qmd-sync", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(qmd-sync/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "`references/zatools-qmd.md`") {
		t.Fatalf("qmd-sync skill should reference shared zatools-qmd guidance:\n%s", content)
	}
	if strings.Contains(content, "## zatools qmd Environment") || strings.Contains(content, "## zatools qmd 环境准备") {
		t.Fatalf("qmd-sync skill should rely on shared qmd guidance instead of duplicating the section:\n%s", content)
	}

	wantNames := []string{"code-to-doc", "ingest", "maintain", "project-router", "qmd-sync", "query", "topic", "workflow"}
	if strings.Join(gotNames, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("builtin skill dirs = %#v, want %#v", gotNames, wantNames)
	}
}

func TestExtractBuiltinSkillsIncludesProjectRouterNewWorkflow(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "project-router", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(project-router/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		`name: "devwiki-project-router"`,
		"判断：这是 [意图类型]，命中 [目标 Skill]，需要/不需要 qmd，需要/不需要代码搜索。",
		"devwiki-ingest",
		"devwiki-topic",
		"devwiki-workflow",
		"devwiki-maintain",
		"devwiki-query",
		"devwiki-code-to-doc",
		"devwiki-qmd-sync",
		"DevWiki 的总入口",
	) {
		t.Fatalf("project-router/SKILL.md missing new DevWiki workflow guidance:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesMaintainGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "maintain", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(maintain/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		`name: "devwiki-maintain"`,
		"证据一致性",
		"知识健康",
		"Query 污染风险",
		"差异报告误落盘",
		"exclude_from_query: true",
		"# Maintain Proposal",
		"这是维护过程报告，不是功能事实来源",
		"glossary.md",
		"references/common-file-format.md",
		"zatools qmd update",
		"zatools devwiki check document",
		"zatools devwiki check graph",
	) {
		t.Fatalf("maintain/SKILL.md missing maintain guidance:\n%s", content)
	}
	if containsAny(content, "relations.yml", "relations/index/glossary", "index/relations/glossary") {
		t.Fatalf("maintain/SKILL.md still references relations.yml:\n%s", content)
	}
	if containsAny(content,
		"不使用 `devwiki:section`",
		"不要给 `wiki/index.md`",
		"全局术语和稳定入口概念",
		"业务能力、系统能力、跨页面主题",
	) {
		t.Fatalf("maintain/SKILL.md should delegate support file format rules to common-file-format:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesStructuredIngestGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "ingest", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(ingest/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		`name: "devwiki-ingest"`,
		"Topic 是主题页",
		"Workflow 是工程实现页",
		"默认只读本文件，不要在任务开始时一次性读取所有 references",
		"只有触发条件满足时再读取对应 reference",
		"写入、重命名、拆分、合并或改主关系前",
		"本地 Wiki 命中低置信、噪声过大或需要 qmd 时",
		"需要结合当前代码、核对实现或写入代码定位时",
		"判断内容应该放入 card、core、explain 还是保留 raw 时",
		"确认要生成 Topic 页面正文时",
		"确认要生成 Workflow 页面正文时",
		"更新 `wiki/index.md`、`wiki/glossary.md` 或 `wiki/log.md` 时",
		"`references/common-file-format.md`",
		"Ingest 是编排器，只生成 TopicTask / WorkflowTask / TroubleshootingTask，不写完整 Topic 或 Workflow 正文",
		"加载 `devwiki-topic`",
		"加载 `devwiki-workflow`",
		"wiki/topics/<slug>.md",
		"wiki/workflows/<slug>.md",
		"wiki/troubleshooting/<slug>.md",
		"wiki/glossary.md",
		"## 需要你确认的问题",
		"落盘前检查",
		"# Ingest Proposal",
		"discussion_only",
		"confirmed_write",
		"确认落盘",
		"按 proposal 写入",
		"用户要求“生成 Wiki / 导入资料 / 构建知识库”只表示启动 ingest 分析流程，不等于允许落盘",
		"实际写入路径是否完全包含在 Ingest Proposal 的“拟写入文件”表内",
		"zatools devwiki check document",
		"zatools devwiki check graph",
	) {
		t.Fatalf("ingest/SKILL.md missing structured ingest guidance:\n%s", content)
	}
	if containsAny(content,
		"wiki/relations.yml",
		"relations.yml",
		"wiki/capabilities/",
		"wiki/features/",
		"wiki/sources/<source-id>.md",
		"wiki/modules/<slug>.md",
		"wiki/open_questions.md",
		"CREATE / EDIT `wiki/sources",
		"CREATE / EDIT `wiki/modules",
		"EDIT `wiki/open_questions.md`",
		"| `discussion_only` |",
		"| `dry_run` |",
		"先阅读通用约束",
		"Capability → 列出并链接 Feature",
		"| 信息类型 | 写入位置 |",
		"`code_refs` 顶层 `note` 只写文件级职责",
		"命名时先问：",
		"如果原文只给出低层名词，要先归并到更稳定的能力或主题",
		"| 原文低层名词 | glossary 候选入口 |",
		"DNS 整机导入导出",
		"link_owners、NTP 自动模式",
	) {
		t.Fatalf("ingest/SKILL.md still references removed wiki paths:\n%s", content)
	}
	if containsAny(content,
		"不使用 `devwiki:section`",
		"全局术语和稳定入口概念",
		"业务能力、系统能力、跨页面主题",
		"单个进程、文件名、字段名",
		"好例子：`告警采集与外发`",
		"坏例子：`告警节点进程`",
	) {
		t.Fatalf("ingest/SKILL.md should delegate support file format rules to common-file-format:\n%s", content)
	}
	for _, rel := range []string{
		"topic/references/topic_template.md",
		"workflow/references/workflow_template.md",
	} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("missing page writing reference %s: %v", rel, err)
		}
	}
}

func TestExtractBuiltinSkillsIncludesCommonFileFormatGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "ingest", "references", "common-file-format.md"))
	if err != nil {
		t.Fatalf("ReadFile(common-file-format.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		"# DevWiki 通用文件格式",
		"`wiki/index.md`",
		"`wiki/glossary.md`",
		"`wiki/log.md`",
		"不使用 YAML frontmatter",
		"不使用 `devwiki:section`",
		"# Wiki Index",
		"# Glossary",
		"# Wiki Log",
		"Append-only chronological log",
		"业务能力、系统能力、跨页面主题、稳定领域概念和常用别名",
		"不要把单个进程、文件名、字段名、配置键、CSV 列、动作码、函数名、日志关键字或检索词当作 glossary 术语",
		"一个术语说明要回答“这个能力/主题解决什么问题、覆盖哪些关键行为、和哪些场景相关”",
	) {
		t.Fatalf("common-file-format.md missing support file format guidance:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesHardMutationSafetyGate(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "ingest", "references", "mutation-safety.md"))
	if err != nil {
		t.Fatalf("ReadFile(mutation-safety.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		"默认写入模式为 `discussion_only`",
		"任何写入都必须先输出 proposal",
		"所有写入，无论风险高低，都必须等用户在 proposal 之后显式确认",
		"用户要求“生成 / 导入 / 整理 / 更新 / 维护 Wiki”只表示进入分析和 proposal 流程，不等于落盘确认",
		"确认落盘",
		"按 proposal 写入",
		"风险等级只影响 proposal 的详细程度，不影响是否需要确认",
	) {
		t.Fatalf("mutation-safety.md missing hard write gate:\n%s", content)
	}
	if containsAny(content, "低风险且确定性的维护动作，在 workflow 已经隐含授权时可以直接执行", "中风险和高风险变更，写入前必须拿到") {
		t.Fatalf("mutation-safety.md still allows implicit or risk-limited writes:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesPageWritingSkillsAndPlacementRules(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	checks := []struct {
		relative string
		wants    []string
		rejects  []string
	}{
		{
			relative: filepath.Join("topic", "SKILL.md"),
			wants: []string{
				`name: "devwiki-topic"`,
				"`devwiki-topic` 只负责生成和维护 Topic",
				"不写代码路径、函数名、handler、调用链",
				"references/knowledge-placement.md",
				"低频、低价值、高体积内容不进入 Wiki，保留 raw",
				"按 `references/knowledge-placement.md` 分配 card/core/explain/raw",
			},
		},
		{
			relative: filepath.Join("workflow", "SKILL.md"),
			wants: []string{
				`name: "devwiki-workflow"`,
				"`devwiki-workflow` 只负责生成和维护 Workflow",
				"代码定位以文件 `path` 为粒度，一个文件一行",
				"references/knowledge-placement.md",
				"代码定位与关键逻辑统一写入 `core` section",
				"只列关键文件、关键入口和关键逻辑点",
				"按 `references/knowledge-placement.md` 分配 card/core/explain/raw",
			},
		},
		{
			relative: filepath.Join("topic", "references", "topic_template.md"),
			wants: []string{
				"# DevWiki Topic 编写模板",
				"<!-- devwiki:section id=card -->",
				"<!-- devwiki:section id=core -->",
				"<!-- devwiki:section id=explain -->",
				"高频高价值内容进入 card/core",
			},
			rejects: []string{"<!-- engwiki:section"},
		},
		{
			relative: filepath.Join("workflow", "references", "workflow_template.md"),
			wants: []string{
				"# DevWiki Workflow 编写模板",
				"<!-- devwiki:section id=card -->",
				"<!-- devwiki:section id=core -->",
				"<!-- devwiki:section id=explain -->",
				"## 关键代码与逻辑",
				"只列关键文件、关键入口和关键逻辑点",
				"辅助函数、普通校验、局部工具函数和顺手读到的方法不要列入表格",
			},
			rejects: []string{"<!-- engwiki:section", "section id=code", "### 代码引用", "| 逻辑点 | 对应代码 | 说明 |"},
		},
		{
			relative: filepath.Join("query", "references", "knowledge-placement.md"),
			wants: []string{
				"# 知识经济学放置规则",
				"高频、高价值、低体积 -> card",
				"高频、高价值、中体积 -> core",
				"低频、高价值、高体积 -> explain",
				"低频、低价值、高体积 -> 不进入 Wiki，保留 raw",
				"读取频率、决策价值和体积",
			},
			rejects: []string{
				"## Topic 内容放置",
				"## Workflow 内容放置",
				"关键文件、关键入口、关键 symbol、关键逻辑点",
				"| 内容类型 | 位置 |",
			},
		},
	}

	for _, check := range checks {
		data, err := os.ReadFile(filepath.Join(root, check.relative))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", check.relative, err)
		}
		content := string(data)
		if !containsAll(content, check.wants...) {
			t.Fatalf("%s missing page writing or knowledge placement guidance:\n%s", check.relative, content)
		}
		if containsAny(content, check.rejects...) {
			t.Fatalf("%s still contains rejected page writing guidance:\n%s", check.relative, content)
		}
	}
}

func TestExtractBuiltinSkillsIncludesQueryGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "query", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(query/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		`name: "devwiki-query"`,
		"explain_topic",
		"locate_code",
		"public_answer",
		"wiki/glossary.md",
		"代码定位线索",
		"## 目录选择规则",
		"## 去重与权威来源规则",
		"能力/功能问题：`topics`",
		"实现问题：`workflows → topics → rg`",
		"如果文档已经足够回答，就不要为了“更稳”再默认展开代码阅读。",
		"召回分档、低置信升档和 qmd fallback 统一遵守 `references/zatools-qmd.md`",
		"本轮 qmd 不可用，已降级",
		"### Step 5: 按需沉淀答案",
		"当前 Project Brain 没有足够信息支持该结论。",
		"只有用户明确要求保存回答、沉淀结论、写入报告时",
	) {
		t.Fatalf("query/SKILL.md missing query guidance:\n%s", content)
	}
	if containsAny(content,
		"wiki/relations.yml",
		"relations.yml",
		"wiki/sources/",
		"wiki/modules/",
		"wiki/open_questions.md",
		"modules →",
		"modules ->",
		"low：0 命中、短词命中过泛、超过 20 条散点",
		"`ssh`、`vip`、`auth`、`token`、`query`、`sync` 这类短词不是强锚点",
	) {
		t.Fatalf("query/SKILL.md still references removed wiki paths:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesLocalWikiFirstQmdFallbackGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "query", "references", "zatools-qmd.md"))
	if err != nil {
		t.Fatalf("ReadFile(zatools-qmd.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		"本地 Wiki 优先，低置信升档",
		"这里的“本地优先”首先指 `wiki/`，不是代码仓全局搜索",
		"意图识别 → 本地 Wiki 搜索 → 命中质量判断 → qmd search → qmd query → raw/code 核对",
		"`locate_exact`",
		"`explain_topic`",
		"`trace_implementation`",
		"`troubleshoot`",
		"`design_intent`",
		"不要把所有关键词都当成精确锚点",
		"`ssh`、`vip`、`auth`、`token`、`query`、`sync` 这类短词只是中锚点",
		"默认先检索 DevWiki 文档层",
		"low | 0 命中；超过 20 条散点命中",
		"必须升到 `zatools qmd search`",
		"qmd 失败 fallback",
		"本轮 qmd 不可用，已降级为本地 Wiki 搜索",
		"raw/code 仍需本地核对",
	) {
		t.Fatalf("zatools-qmd.md missing local-wiki-first qmd fallback guidance:\n%s", content)
	}
	if containsAny(content,
		"当问题里已经包含具体锚点时，**优先本地搜索**，不走 `zatools qmd ...`",
		"若未检测到 GPU / 加速器，或确认当前环境只能在 CPU 上跑 embed / rerank，直接报告",
	) {
		t.Fatalf("zatools-qmd.md still contains old overly strict local-first or GPU gate guidance:\n%s", content)
	}
}

func TestExtractBuiltinSkillsIncludesCodeRefsFileLevelGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	checks := []struct {
		relative string
		wants    []string
		rejects  []string
	}{
		{
			relative: filepath.Join("query", "references", "evidence-grounding.md"),
			wants: []string{
				"每个重要结论必须能回溯到真实来源",
				"`qmd` 只是召回加速器，不是真相源",
				"事实和推断必须拆开表达",
				"如果 `wiki/` 与 `raw/` 冲突",
				"低置信处理协议",
			},
			rejects: []string{
				"## 页面层级边界",
				"## 如何使用代码定位表",
				"Topic = 系统能力边界",
				"| 层级 | 权威内容 | 不应该写什么 |",
				"| 路径 | 文件职责 | 关键入口 / Symbol |",
				"代码定位表以代码文件 `path` 为唯一粒度",
				"是整个文件相关，还是只有某个 symbol 相关",
				"symbol_notes",
				"code_refs:",
				"UserService#updateProfile#method",
			},
		},
		{
			relative: filepath.Join("workflow", "references", "workflow_template.md"),
			wants: []string{
				"一个文件只写一行",
				"关键入口最多列 8 个",
				"不列全量方法",
				"代码定位与关键逻辑统一写入 `core` section",
				"只列关键文件、关键入口和关键逻辑点",
				"Workflow frontmatter 不再放 `code_refs`、`symbols`、`api_entries`、`test_refs`",
			},
			rejects: []string{
				`symbol: "<类/函数/方法/常量>"`,
				"| 路径 | 符号 | 类型 | 说明 |",
				"| 逻辑点 | 对应代码 | 说明 |",
				"### 代码引用",
				"symbol_notes",
			},
		},
		{
			relative: filepath.Join("code-to-doc", "SKILL.md"),
			wants: []string{
				"默认写入 `wiki/workflows/<slug>.md`",
				"Topic 的 sources 不写代码文件路径或 `kind: code`",
			},
			rejects: []string{
				"写入 `code_refs` 前必须先按 `path` 分组",
				"顶层 `note` 只写文件级职责",
				"`symbols` 使用 `<symbol>#<kind>: \"<短说明>\"` 格式",
				"symbol_notes",
			},
		},
		{
			relative: filepath.Join("query", "references", "code-tracing.md"),
			wants: []string{
				"按代码文件归并",
				"按 Workflow 模板的 `关键代码与逻辑` 规则归并",
				"不要列出文件内所有方法",
			},
			rejects: []string{
				"代码定位表以代码文件 `path` 为粒度",
				"同一文件只能一行",
				"最多 8 个关键入口 symbol",
			},
		},
		{
			relative: filepath.Join("query", "references", "evidence-grounding.md"),
			wants: []string{
				"DevWiki 输出必须落到真实来源、已核对代码证据，或被明确标注为推断",
				"不要虚构来源或代码证据",
			},
			rejects: []string{"symbol_notes", "code_refs:", "`UserService#class`, `updateProfile#method`"},
		},
	}

	for _, check := range checks {
		data, err := os.ReadFile(filepath.Join(root, check.relative))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", check.relative, err)
		}
		content := string(data)
		if !containsAll(content, check.wants...) {
			t.Fatalf("%s missing file-level code location guidance:\n%s", check.relative, content)
		}
		if containsAny(content, check.rejects...) {
			t.Fatalf("%s still contains old symbol-level code location guidance:\n%s", check.relative, content)
		}
	}
}

func TestGenerateProjectRuntimeDocsStayProjectLevel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		agent    string
		filename string
	}{
		{agent: "codex", filename: "AGENTS.md"},
		{agent: "claude", filename: "CLAUDE.md"},
	}

	for _, tt := range cases {
		root := filepath.Join(t.TempDir(), tt.agent)
		spec := ProjectSpec{
			ProjectName: "Demo",
			ProjectSlug: "demo",
			Agent:       tt.agent,
			Lang:        "zh",
			CodeRepos: []CodeRepo{
				{Name: "api", Slug: "api", Path: "/tmp/api", Default: true},
			},
		}
		if err := GenerateProject(root, spec); err != nil {
			t.Fatalf("GenerateProject(%s) error = %v", tt.agent, err)
		}

		data, err := os.ReadFile(filepath.Join(root, tt.filename))
		if err != nil {
			t.Fatalf("ReadFile(%s/%s) error = %v", tt.agent, tt.filename, err)
		}
		content := string(data)
		if !containsAll(content,
			"本目录就是 DevWiki 文档库根目录",
			"## 目录结构",
			"## Wiki 目录",
			"## 链接语法",
			"## 检索顺序",
			"## Workflow 约束",
			"## 操作说明",
			"项目知识任务先由 `devwiki-project-router` 判断意图",
			"具体页面模板、字段结构、页面边界和证据写入规则由对应 DevWiki skill 和 `references/` 维护",
		) {
			t.Fatalf("%s missing project-level runtime guidance:\n%s", tt.filename, content)
		}
		if containsAny(content,
			"## 页面模板",
			"## Source 结构",
			"## Code Ref 结构",
			"## 来源和代码引用",
			"### 角色约束",
			"capability 页是业务能力与系统能力的中心页",
			"feature 页是功能设计与行为说明页",
			"workflow 页是工程定位、代码线索和修改影响页",
			"troubleshooting 页是排障现象、诊断路径、证据和修复建议页",
			"页面职责",
			"某个业务能力或系统能力的中心页",
			"某个具体功能的设计页",
			"面向编程的工程定位页",
			"排障现象、诊断路径、证据、修复建议和适用版本",
			"```yaml",
			"sources:",
			"code_refs:",
			"api_entries:",
			"test_refs:",
			"UserService#updateProfile#method",
			"正文建议分区",
			"正文：`",
			"`code_refs` 以代码文件 `path` 为唯一粒度",
			"顶层 `note` 只写文件级职责",
			"`symbols` 是关键入口索引",
		) {
			t.Fatalf("%s still duplicates page template or source/code_refs schema:\n%s", tt.filename, content)
		}
	}
}

func TestExtractBuiltinSkillsIncludesCodeToDocGuidance(t *testing.T) {
	t.Parallel()

	root, cleanup, err := ExtractBuiltinSkills("zh")
	if err != nil {
		t.Fatalf("ExtractBuiltinSkills error = %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(filepath.Join(root, "code-to-doc", "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(code-to-doc/SKILL.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		`name: "devwiki-code-to-doc"`,
		"直接转交对应页面 Skill",
		"加载 `devwiki-topic`",
		"加载 `devwiki-workflow`",
		"内容放置必须遵守 `references/knowledge-placement.md`",
		"默认写入 `wiki/workflows/<slug>.md`",
		"wiki/topics/",
		"wiki/workflows/",
		"wiki/troubleshooting/",
		"涉及 `wiki/index.md`、`wiki/glossary.md` 或 `wiki/log.md` 更新时，再读 `references/common-file-format.md`",
		"Topic 的 sources 不写代码文件路径或 `kind: code`",
		"代码、wiki、raw 冲突",
		"zatools devwiki check document",
		"zatools devwiki check graph",
		"zatools qmd update",
		"zatools qmd status",
	) {
		t.Fatalf("code-to-doc/SKILL.md missing code-to-doc guidance:\n%s", content)
	}
	if containsAny(content, "wiki/relations.yml", "relations.yml", "wiki/modules/", "Source Card") {
		t.Fatalf("code-to-doc/SKILL.md still references removed module/source behavior:\n%s", content)
	}
	if containsAny(content,
		"不使用 `devwiki:section`",
		"不要给 `wiki/index.md`",
		"全局术语和稳定入口概念",
		"业务能力、系统能力、跨页面主题",
	) {
		t.Fatalf("code-to-doc/SKILL.md should delegate support file format rules to common-file-format:\n%s", content)
	}
}

func TestEnsureProjectRuntimeBridgeCreatesSelectedRootRuntimeFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workspaceDir := filepath.Join(root, "devwiki-sample")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	err := EnsureProjectRuntimeBridge(root, workspaceDir, "codex", "zh")
	if err != nil {
		t.Fatalf("EnsureProjectRuntimeBridge error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	content := string(data)
	if !containsAll(content,
		"./devwiki-sample/AGENTS.md",
		runtimeBridgeStartMarker,
		runtimeBridgeEndMarker,
	) {
		t.Fatalf("AGENTS.md missing runtime bridge block:\n%s", content)
	}
	if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); err == nil {
		t.Fatal("codex bridge should not create CLAUDE.md")
	}
}

func TestEnsureProjectRuntimeBridgePreservesExistingContentAndUpdatesManagedBlock(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workspaceDir := filepath.Join(root, "devwiki-sample")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	original := "# Existing Rules\n\nKeep this content.\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile(existing AGENTS.md) error = %v", err)
	}

	if err := EnsureProjectRuntimeBridge(root, workspaceDir, "codex", "zh"); err != nil {
		t.Fatalf("EnsureProjectRuntimeBridge first call error = %v", err)
	}
	if err := EnsureProjectRuntimeBridge(root, workspaceDir, "codex", "zh"); err != nil {
		t.Fatalf("EnsureProjectRuntimeBridge second call error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Keep this content.") {
		t.Fatalf("existing content should be preserved:\n%s", content)
	}
	if strings.Count(content, runtimeBridgeStartMarker) != 1 {
		t.Fatalf("managed block should appear once:\n%s", content)
	}
	if !strings.Contains(content, "./devwiki-sample/AGENTS.md") {
		t.Fatalf("managed block missing runtime path:\n%s", content)
	}
}

func TestEnsureCodeRepoDevwikiLinkCreatesAgentsWhenMissing(t *testing.T) {
	t.Parallel()

	codeRoot := t.TempDir()
	devwikiRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(devwikiRoot, "AGENTS.md"), []byte("# DevWiki\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(devwiki AGENTS.md) error = %v", err)
	}

	if err := EnsureCodeRepoDevwikiLink(codeRoot, devwikiRoot, "codex", "zh"); err != nil {
		t.Fatalf("EnsureCodeRepoDevwikiLink error = %v", err)
	}

	agentsData, err := os.ReadFile(filepath.Join(codeRoot, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(code AGENTS.md) error = %v", err)
	}
	agents := string(agentsData)
	if !containsAll(agents,
		codeLinkStartMarker,
		codeLinkEndMarker,
		devwikiRoot,
		filepath.Join(devwikiRoot, "AGENTS.md"),
		"devwiki-query",
		"devwiki-code-to-doc",
		"必须写入关联 DevWiki 文档库",
	) {
		t.Fatalf("code AGENTS.md missing DevWiki link block:\n%s", agents)
	}
}

func TestEnsureCodeRepoDevwikiLinkUpdatesAgentsAndClaudeIdempotently(t *testing.T) {
	t.Parallel()

	codeRoot := t.TempDir()
	devwikiRoot := t.TempDir()
	agentsOriginal := "# Existing Agents\n\nKeep AGENTS.\n"
	claudeOriginal := "# Existing Claude\n\nKeep CLAUDE.\n"
	if err := os.WriteFile(filepath.Join(codeRoot, "AGENTS.md"), []byte(agentsOriginal), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(codeRoot, "CLAUDE.md"), []byte(claudeOriginal), 0o644); err != nil {
		t.Fatalf("WriteFile(CLAUDE.md) error = %v", err)
	}

	if err := EnsureCodeRepoDevwikiLink(codeRoot, devwikiRoot, "claude", "zh"); err != nil {
		t.Fatalf("EnsureCodeRepoDevwikiLink first call error = %v", err)
	}
	if err := EnsureCodeRepoDevwikiLink(codeRoot, devwikiRoot, "claude", "zh"); err != nil {
		t.Fatalf("EnsureCodeRepoDevwikiLink second call error = %v", err)
	}

	for _, filename := range []string{"AGENTS.md", "CLAUDE.md"} {
		data, err := os.ReadFile(filepath.Join(codeRoot, filename))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", filename, err)
		}
		content := string(data)
		if strings.Count(content, codeLinkStartMarker) != 1 {
			t.Fatalf("%s managed block should appear once:\n%s", filename, content)
		}
		if !strings.Contains(content, filepath.Join(devwikiRoot, "CLAUDE.md")) {
			t.Fatalf("%s missing Claude runtime path:\n%s", filename, content)
		}
	}

	agentsData, err := os.ReadFile(filepath.Join(codeRoot, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile(AGENTS.md) error = %v", err)
	}
	if !strings.Contains(string(agentsData), "Keep AGENTS.") {
		t.Fatalf("AGENTS.md should preserve existing content:\n%s", string(agentsData))
	}
	claudeData, err := os.ReadFile(filepath.Join(codeRoot, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile(CLAUDE.md) error = %v", err)
	}
	if !strings.Contains(string(claudeData), "Keep CLAUDE.") {
		t.Fatalf("CLAUDE.md should preserve existing content:\n%s", string(claudeData))
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}

func containsAny(text string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(text, part) {
			return true
		}
	}
	return false
}
