package devwiki

import (
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
		"wiki/documents/designs/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
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
		"为什么不是临时 RAG",
		"zatools devwiki init",
		"zatools qmd sync --root . --apply",
		"zatools qmd update",
		"zatools qmd status",
		"zatools devwiki tool reset --scope wiki --project-root .",
		"devwiki-qmd-sync",
		"devwiki-sample-project-raw",
		"/tmp/go-skills",
		"devwiki-sample-project/",
		"├── AGENTS.md",
		"├── config/",
		"└── wiki/",
		"项目级 skill 安装状态、桥接用运行时文件和 `.zatools-lock.json`",
	) {
		t.Fatalf("README.md content missing expected latest Go guidance:\n%s", readme)
	}
	if containsAny(readme, "{{", "}}", "Python 3.11+", ".venv", "setup.sh", "setup.ps1", "├── i18n/", "├── search/", "├── tools/") {
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
				"# DevWiki — 运行时 Schema",
				"由 Claude Code 与 Codex 共同驱动",
				"devwiki-sample-project/",
				"当前项目根的桥接运行时文件会要求 agent 在处理 DevWiki 任务前先阅读 `./devwiki-sample-project/"+tc.runtimeFile+"`",
				"使用 `zatools devwiki init` 初始化 DevWiki 工作区",
				"wiki/documents/{doc-type}/{slug}.md",
				"├── config/",
				"└── wiki/",
			) {
				t.Fatalf("%s content missing expected latest runtime guidance:\n%s", tc.runtimeFile, content)
			}
			if containsAny(content, "{{", "}}", "setup.sh", "setup.ps1", "i18n/", "project.yaml.example", "claude-settings.local.json.example", "codex-config.example.yaml", "search/", "tools/") {
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
}

func TestSlugifyProducesStableDirectoryNames(t *testing.T) {
	t.Parallel()

	if got := Slugify("Sample Project"); got != "sample-project" {
		t.Fatalf("Slugify = %q, want %q", got, "sample-project")
	}
}

func TestExtractBuiltinSkillsMaterializesSharedReferencesIntoEachSkill(t *testing.T) {
	t.Parallel()

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

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for _, shared := range []string{"code-tracing.md", "zatools-qmd.md"} {
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
