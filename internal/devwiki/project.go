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

// BuiltinSkillsPath 返回内置技能模板在嵌入文件系统中的路径。
func BuiltinSkillsPath(lang string) string {
	return path.Join("template", "i18n", lang, "skills")
}

// GenerateProject 把内置模板渲染成一个独立的 DevWiki 工程目录。
func GenerateProject(target string, spec ProjectSpec) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("target directory is empty")
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s already exists", target)
	} else if !os.IsNotExist(err) {
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

// ExtractBuiltinSkills 将内置技能模板解压到临时目录，供现有发现/安装流程复用。
func ExtractBuiltinSkills(lang string) (string, func(), error) {
	root := BuiltinSkillsPath(lang)
	tempRoot, err := os.MkdirTemp("", "zatools-devwiki-skills-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tempRoot) }

	err = fs.WalkDir(templateFS, root, func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == root {
			return nil
		}
		rel := strings.TrimPrefix(name, root+"/")
		dest := filepath.Join(tempRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := fs.ReadFile(templateFS, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if err := materializeSharedReferences(tempRoot, lang); err != nil {
		cleanup()
		return "", nil, err
	}
	return tempRoot, cleanup, nil
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

	for _, name := range []string{"capabilities", "features", "outputs", "graph"} {
		dir := filepath.Join(root, "wiki", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := ensureGitkeep(dir); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(root, "wiki", "index.md"), []byte("# Wiki Index\n"), 0o644); err != nil {
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
		{Name: fmt.Sprintf("devwiki-%s-raw", spec.ProjectSlug), Path: "raw"},
		{Name: fmt.Sprintf("devwiki-%s-wiki", spec.ProjectSlug), Path: "wiki"},
	}
	for _, repo := range spec.CodeRepos {
		collections = append(collections, collection{
			Name: fmt.Sprintf("devwiki-%s-code-%s", spec.ProjectSlug, repo.Slug),
			Path: repo.Path,
		})
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
		"{{WORKSPACE_DIR}}", "devwiki-"+spec.ProjectSlug,
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
	if lang == "en" {
		return strings.Join([]string{
			runtimeBridgeStartMarker,
			"## DevWiki Runtime",
			fmt.Sprintf("Before handling DevWiki tasks in this repository, always read and follow `%s`.", relativePath),
			"If that DevWiki runtime file conflicts with other repository notes here, prefer it for DevWiki work.",
			runtimeBridgeEndMarker,
			"",
		}, "\n")
	}

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
	title := "# Repository Runtime\n\n"
	if lang != "en" {
		title = "# 仓库运行时入口\n\n"
	}
	return title + block
}

func upsertRuntimeBridgeBlock(content string, block string) string {
	start := strings.Index(content, runtimeBridgeStartMarker)
	if start == -1 {
		content = strings.TrimRight(content, "\n")
		if content != "" {
			content += "\n\n"
		}
		return content + block
	}

	end := strings.Index(content[start:], runtimeBridgeEndMarker)
	if end == -1 {
		content = strings.TrimRight(content[:start], "\n")
		if content != "" {
			content += "\n\n"
		}
		return content + block
	}

	end += start + len(runtimeBridgeEndMarker)
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

func materializeSharedReferences(skillsRoot string, lang string) error {
	sharedRoot := path.Join("template", "i18n", lang, "shared-references")
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		targetRoot := filepath.Join(skillsRoot, entry.Name(), "references")
		err := fs.WalkDir(templateFS, sharedRoot, func(name string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if name == sharedRoot {
				return nil
			}
			rel := strings.TrimPrefix(name, sharedRoot+"/")
			dest := filepath.Join(targetRoot, filepath.FromSlash(rel))
			if d.IsDir() {
				return os.MkdirAll(dest, 0o755)
			}
			data, err := fs.ReadFile(templateFS, name)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			return os.WriteFile(dest, data, 0o644)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
