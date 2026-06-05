package devwiki

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"zatools/internal/qmd"
)

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

const (
	runtimeBridgeStartMarker = "<!-- zatools:devwiki-runtime:start -->"
	runtimeBridgeEndMarker   = "<!-- zatools:devwiki-runtime:end -->"
	codeLinkStartMarker      = "<!-- zatools:devwiki-link:start -->"
	codeLinkEndMarker        = "<!-- zatools:devwiki-link:end -->"
)

// CodeRepo 描述一个注册到 DevWiki 配置中的代码仓条目。
type CodeRepo struct {
	Name    string `yaml:"name"`
	Slug    string `yaml:"slug"`
	Path    string `yaml:"path"`
	Default bool   `yaml:"default"`
}

// ProjectSpec 描述生成 DevWiki 工程所需的配置。
type ProjectSpec struct {
	ProjectName string
	ProjectSlug string
	Agent       string
	Lang        string
	CodeRepos   []CodeRepo
}

// Slugify 生成稳定、可落盘的 slug。
func Slugify(text string) string {
	normalized := slugPattern.ReplaceAllString(strings.ToLower(text), "-")
	normalized = strings.Trim(normalized, "-")
	if normalized != "" {
		return normalized
	}
	sum := sha1.Sum([]byte(text))
	return "item-" + hex.EncodeToString(sum[:])[:8]
}

// NormalizeCodeRepos 将代码目录归一化为 DevWiki 的 repo 配置。
func NormalizeCodeRepos(baseDir string, codeDirs []string) ([]CodeRepo, error) {
	repos := make([]CodeRepo, 0, len(codeDirs))
	seen := map[string]int{}
	for index, raw := range codeDirs {
		candidate := raw
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(baseDir, candidate)
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(abs)
		if name == "." || name == string(filepath.Separator) || name == "" {
			name = fmt.Sprintf("repo-%d", index+1)
		}
		baseSlug := Slugify(name)
		seen[baseSlug]++
		slug := baseSlug
		if seen[baseSlug] > 1 {
			slug = fmt.Sprintf("%s-%d", baseSlug, seen[baseSlug])
		}
		repos = append(repos, CodeRepo{
			Name:    name,
			Slug:    slug,
			Path:    abs,
			Default: index == 0,
		})
	}
	return repos, nil
}

// ProjectSlugFromRoot reads config/project.yaml and returns the DevWiki project slug.
func ProjectSlugFromRoot(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "config", "project.yaml"))
	if err != nil {
		return "", err
	}
	var parsed struct {
		ProjectSlug string `yaml:"project_slug"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	slug := strings.TrimSpace(parsed.ProjectSlug)
	if slug == "" {
		return "", fmt.Errorf("devwiki project_slug is required in %s", filepath.Join(root, "config", "project.yaml"))
	}
	return slug, nil
}

// GenerateProject 把内置模板渲染到指定 DevWiki 工程目录。
func GenerateProject(target string, spec ProjectSpec) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("target directory is empty")
	}
	if info, err := os.Stat(target); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s already exists and is not a directory", target)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return err
		}
	} else {
		return err
	}
	if err := ensureRepoLayout(target); err != nil {
		return err
	}
	if err := writeProjectDocs(target, spec); err != nil {
		return err
	}
	if err := writeProjectConfig(target, spec); err != nil {
		return err
	}
	if err := writeSearchConfig(target, spec); err != nil {
		return err
	}
	return nil
}

func ensureRepoLayout(root string) error {
	for _, name := range rawDirNames {
		dir := filepath.Join(root, "raw", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := ensureGitkeep(dir); err != nil {
			return err
		}
	}

	for _, name := range []string{"topics", "workflows", "troubleshooting", "outputs"} {
		dir := filepath.Join(root, "wiki", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := ensureGitkeep(dir); err != nil {
			return err
		}
	}

	index := "# Wiki Index\n\n| type | description | slug |\n|---|---|---|\n"
	if err := os.WriteFile(filepath.Join(root, "wiki", "index.md"), []byte(index), 0o644); err != nil {
		return err
	}
	glossary := "# Glossary\n\n| glossary | type | description | slug |\n|---|---|---|---|\n"
	if err := os.WriteFile(filepath.Join(root, "wiki", "glossary.md"), []byte(glossary), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "wiki", "log.md"), []byte("# Wiki Log\n\n> Append-only chronological log.\n"), 0o644)
}

func ensureGitkeep(dir string) error {
	return os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte(""), 0o644)
}

func writeProjectDocs(root string, spec ProjectSpec) error {
	docNames := []string{"README.md"}
	switch spec.Agent {
	case "claude":
		docNames = append(docNames, "CLAUDE.md")
	default:
		docNames = append(docNames, "AGENTS.md")
	}

	for _, name := range docNames {
		content, err := renderProjectDoc(spec, name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeProjectConfig(root string, spec ProjectSpec) error {
	configPath := filepath.Join(root, "config", "project.yaml")
	content := struct {
		ProjectName string     `yaml:"project_name"`
		ProjectSlug string     `yaml:"project_slug"`
		Agent       string     `yaml:"agent"`
		Language    string     `yaml:"language"`
		CodeRepos   []CodeRepo `yaml:"code_repos"`
	}{
		ProjectName: spec.ProjectName,
		ProjectSlug: spec.ProjectSlug,
		Agent:       spec.Agent,
		Language:    spec.Lang,
		CodeRepos:   spec.CodeRepos,
	}
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o644)
}

func writeSearchConfig(root string, spec ProjectSpec) error {
	type collection struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	}
	collections := []collection{
		{Name: fmt.Sprintf("devwiki-%s-wiki", spec.ProjectSlug), Path: "wiki"},
	}
	content := map[string]any{
		"qmd": map[string]any{
			"collections":    collections,
			"embed_model":    qmd.DefaultEmbedModel,
			"rerank_model":   qmd.DefaultRerankModel,
			"generate_model": qmd.DefaultGenerateModel,
		},
	}
	data, err := yaml.Marshal(content)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "config", "search.yaml"), data, 0o644)
}

func renderProjectDoc(spec ProjectSpec, name string) (string, error) {
	data, err := fs.ReadFile(templateFS, path.Join("template", "docs", name))
	if err != nil {
		return "", err
	}

	runtimeFile := "AGENTS.md"
	runtimeLabel := "Codex"
	runtimeEntry := "Codex runtime entry"
	if spec.Agent == "cursor" {
		runtimeLabel = "Cursor"
		runtimeEntry = "Cursor runtime entry"
	}
	if spec.Agent == "claude" {
		runtimeFile = "CLAUDE.md"
		runtimeLabel = "Claude Code"
		runtimeEntry = "Claude runtime entry"
	}

	primaryRepo := spec.CodeRepos[0]
	replacer := strings.NewReplacer(
		"{{PROJECT_NAME}}", spec.ProjectName,
		"{{PROJECT_SLUG}}", spec.ProjectSlug,
		"{{WORKSPACE_DIR}}", ".",
		"{{AGENT}}", spec.Agent,
		"{{LANG}}", spec.Lang,
		"{{PRIMARY_CODE_DIR}}", primaryRepo.Path,
		"{{PRIMARY_CODE_SLUG}}", primaryRepo.Slug,
		"{{RUNTIME_FILE}}", runtimeFile,
		"{{RUNTIME_LABEL}}", runtimeLabel,
		"{{RUNTIME_ENTRY}}", runtimeEntry,
	)
	return replacer.Replace(string(data)), nil
}

// EnsureCodeRepoDevwikiLink writes managed DevWiki association text into a code repository.
func EnsureCodeRepoDevwikiLink(codeRoot string, devwikiRoot string, projectSlug string, agent string, lang string) error {
	runtimeFile, err := runtimeFilenameForAgent(agent)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(codeRoot, 0o755); err != nil {
		return err
	}

	codeRoot, err = filepath.Abs(codeRoot)
	if err != nil {
		return err
	}
	devwikiRoot = strings.TrimSpace(devwikiRoot)
	if devwikiRoot != "" {
		devwikiRoot, err = filepath.Abs(devwikiRoot)
		if err != nil {
			return err
		}
	}
	runtimePath := ""
	if devwikiRoot != "" {
		runtimePath = filepath.Join(devwikiRoot, runtimeFile)
	}
	block := renderCodeRepoDevwikiLinkBlock(devwikiRoot, runtimePath, projectSlug, lang)
	wrote := false
	for _, filename := range []string{"AGENTS.md", "CLAUDE.md"} {
		targetPath := filepath.Join(codeRoot, filename)
		if _, err := os.Stat(targetPath); err == nil {
			if err := upsertManagedFileBlock(targetPath, block, lang); err != nil {
				return err
			}
			wrote = true
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if wrote {
		return nil
	}
	return upsertManagedFileBlock(filepath.Join(codeRoot, "AGENTS.md"), block, lang)
}

func renderCodeRepoDevwikiLinkBlock(devwikiRoot string, runtimePath string, projectSlug string, lang string) string {
	_ = lang

	lines := []string{
		codeLinkStartMarker,
		"## 关联 DevWiki",
		fmt.Sprintf("DevWiki project：`%s`。", projectSlug),
		fmt.Sprintf("统一查询命令使用：`--project %s`。", projectSlug),
	}
	if strings.TrimSpace(devwikiRoot) != "" {
		lines = append(lines,
			fmt.Sprintf("DevWiki 文档库根目录：`%s`。", devwikiRoot),
			fmt.Sprintf("在本代码库回答项目知识问题、修改代码，或使用 `devwiki-query` / `devwiki-code` / `devwiki-code-to-doc` 前，必须先阅读 `%s`。", runtimePath),
			"查询时以关联 DevWiki 根目录下的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源。",
			"`devwiki-code-to-doc` 生成的 workflow / topic 相关页面必须写入关联 DevWiki 文档库，不要写入本代码库。",
		)
	}
	lines = append(lines,
		"使用 DevWiki skills 前先判定目标产物和定位锚点：领域/功能/特性描述且缺少代码锚点时，可显式使用 `$devwiki-code` 定位代码入口和规则边界。",
		"使用 `devwiki-query` 或 `devwiki-code` 时，必须严格遵循对应 Skill.md 的查询和定位步骤；禁止绕过 skill 流程自行做全仓广泛搜索或自由发挥式检索。",
		"用户已经给出具体文件、函数、代码块、当前 diff、完整 patch 或明确替换方式时，不自动进入 `devwiki-code`，按普通编辑任务处理。",
		codeLinkEndMarker,
		"",
	)
	return strings.Join(lines, "\n")
}

func upsertManagedFileBlock(targetPath string, block string, lang string) error {
	_ = lang
	data, err := os.ReadFile(targetPath)
	if os.IsNotExist(err) {
		return os.WriteFile(targetPath, []byte("# 仓库运行时入口\n\n"+block), 0o644)
	}
	if err != nil {
		return err
	}
	updated := upsertDelimitedBlock(string(data), block, codeLinkStartMarker, codeLinkEndMarker)
	return os.WriteFile(targetPath, []byte(updated), 0o644)
}

// EnsureProjectRuntimeBridge ensures the project root runtime file points agents at the generated DevWiki runtime file.
func EnsureProjectRuntimeBridge(projectRoot string, workspaceDir string, agent string, lang string) error {
	runtimeFile, err := runtimeFilenameForAgent(agent)
	if err != nil {
		return err
	}

	relativePath, err := filepath.Rel(projectRoot, filepath.Join(workspaceDir, runtimeFile))
	if err != nil {
		return err
	}
	relativePath = filepath.ToSlash(relativePath)
	if !strings.HasPrefix(relativePath, ".") {
		relativePath = "./" + relativePath
	}

	targetPath := filepath.Join(projectRoot, runtimeFile)
	block := renderRuntimeBridgeBlock(relativePath, lang)
	data, err := os.ReadFile(targetPath)
	if os.IsNotExist(err) {
		return os.WriteFile(targetPath, []byte(renderNewRuntimeBridgeFile(block, lang)), 0o644)
	}
	if err != nil {
		return err
	}

	updated := upsertRuntimeBridgeBlock(string(data), block)
	return os.WriteFile(targetPath, []byte(updated), 0o644)
}

func runtimeFilenameForAgent(agent string) (string, error) {
	switch agent {
	case "codex", "cursor":
		return "AGENTS.md", nil
	case "claude":
		return "CLAUDE.md", nil
	default:
		return "", fmt.Errorf("unsupported agent %q", agent)
	}
}

func renderRuntimeBridgeBlock(relativePath string, lang string) string {
	_ = lang

	return strings.Join([]string{
		runtimeBridgeStartMarker,
		"## DevWiki Runtime",
		fmt.Sprintf("处理本仓库中的 DevWiki 相关任务前，必须先阅读并遵守 `%s`。", relativePath),
		"如果该 DevWiki 运行时文件与当前仓库中的其他说明冲突，在 DevWiki 工作中以它为准。",
		runtimeBridgeEndMarker,
		"",
	}, "\n")
}

func renderNewRuntimeBridgeFile(block string, lang string) string {
	_ = lang
	return "# 仓库运行时入口\n\n" + block
}

func upsertRuntimeBridgeBlock(content string, block string) string {
	return upsertDelimitedBlock(content, block, runtimeBridgeStartMarker, runtimeBridgeEndMarker)
}

func upsertDelimitedBlock(content string, block string, startMarker string, endMarker string) string {
	start := strings.Index(content, startMarker)
	if start == -1 {
		content = strings.TrimRight(content, "\n")
		if content != "" {
			content += "\n\n"
		}
		return content + block
	}

	end := strings.Index(content[start:], endMarker)
	if end == -1 {
		content = strings.TrimRight(content[:start], "\n")
		if content != "" {
			content += "\n\n"
		}
		return content + block
	}

	end += start + len(endMarker)
	prefix := strings.TrimRight(content[:start], "\n")
	suffix := strings.TrimLeft(content[end:], "\n")

	parts := make([]string, 0, 3)
	if prefix != "" {
		parts = append(parts, prefix)
	}
	parts = append(parts, strings.TrimRight(block, "\n"))
	if suffix != "" {
		parts = append(parts, suffix)
	}
	return strings.Join(parts, "\n\n") + "\n"
}
