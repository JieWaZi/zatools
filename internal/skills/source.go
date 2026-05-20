package skills

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"zatools/internal/devwiki"
)

// Source 表示一条技能来源配置。
// 它既承载用户输入，也承载 CLI 后续 clone、发现和安装阶段要用到的解析结果。
type Source struct {
	// Original 是用户输入的原始来源字符串。
	// 锁文件会保留这个值，后续 check/update 会基于它重新解析来源。
	Original string
	// Type 表示来源类型，例如 local、github、gitlab、git。
	// 这个字段决定 ResolveSource 走本地目录校验还是远端 git 拉取流程。
	Type string
	// RepoURL 是解析后的远端仓库地址。
	// 仅远端来源会使用该字段，本地来源保持为空。
	RepoURL string
	// LocalDir 是本地目录来源的绝对路径。
	// ParseSource 在解析本地来源时会提前转成绝对路径，便于后续稳定复用。
	LocalDir string
	// Ref 是用户指定的远端引用名，例如分支、标签或提交哈希。
	// ResolveSource 会尽量把仓库定位到这个引用对应的内容。
	Ref string
	// Builtin 是内置库标识，例如 devwiki。
	// 仅 Type=builtin 时使用。
	Builtin string
	// Subpath 是仓库中技能所在的子目录。
	// 它表示“搜索技能时从仓库根往下进入的相对路径”，不是最终安装目录。
	Subpath string
}

// Skill 表示从来源目录中发现的一项技能定义。
type Skill struct {
	// Name 是技能名称，对应 SKILL.md frontmatter 里的 name。
	Name string
	// Description 是技能描述，对应 SKILL.md frontmatter 里的 description。
	Description string
	// Dir 是技能目录的绝对路径。
	// 安装阶段会直接把这个目录复制到目标 skills 目录下。
	Dir string
	// RelativeDir 是技能目录相对“本次搜索根目录”的路径。
	// 当来源本身已经指向某个子目录时，它可能是 "."。
	RelativeDir string
}

// ResolvedSource 表示已经解析并准备好的技能来源。
// 对远端仓库来说，这个结构还负责持有临时 clone 目录的清理函数。
type ResolvedSource struct {
	// Source 是原始来源元数据。
	Source Source
	// RootDir 是本地目录或临时 clone 目录。
	// 后续 SearchRoot 会在这个根目录下进一步定位实际扫描目录。
	RootDir string
	cleanup func() error
}

// skillFrontmatter 只映射 SKILL.md 里当前需要消费的 YAML 字段。
// 其它 frontmatter 字段即使存在，也不会参与发现和安装逻辑。
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

const fallbackSkillDescription = "没有相关信息"

var githubShorthandPattern = regexp.MustCompile(`^([^/\s]+)/([^/\s]+)(?:/(.+))?$`)
var windowsAbsPattern = regexp.MustCompile(`^[a-zA-Z]:[/\\]`)

const builtinSourceNamespace = "zatools"

var sourceAliases = map[string]string{
	"coinbase/agentWallet": "coinbase/agentic-wallet-skills",
}

// cloneTimeout 限制远端 clone 的最长时间，避免命令无限阻塞。
const cloneTimeout = 60 * time.Second
const defaultBuiltinDevwikiLang = "zh"

// ParseSource 把用户输入的来源字符串解析为统一的 Source 结构。
func ParseSource(input string) (Source, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Source{}, fmt.Errorf("source is required")
	}

	input, ref := parseFragmentRef(input)

	if alias, ok := sourceAliases[input]; ok {
		input = alias
	}

	if builtin, ok, err := parseBuiltinSource(input, ref); ok || err != nil {
		return builtin, err
	}

	if isLocalPath(input) {
		abs, err := filepath.Abs(input)
		if err != nil {
			return Source{}, fmt.Errorf("resolve local path: %w", err)
		}
		return Source{
			Original: input,
			Type:     "local",
			LocalDir: abs,
			Ref:      ref,
		}, nil
	}

	if match := strings.TrimPrefix(input, "github:"); match != input {
		return ParseSource(appendFragmentRef(match, ref))
	}

	if match := strings.TrimPrefix(input, "gitlab:"); match != input {
		return ParseSource(appendFragmentRef("https://gitlab.com/"+match, ref))
	}

	if strings.HasPrefix(input, "https://github.com/") || strings.HasPrefix(input, "http://github.com/") {
		return parseGitHubURL(input, ref)
	}

	if strings.HasPrefix(input, "https://gitlab.com/") || strings.HasPrefix(input, "http://gitlab.com/") {
		return parseGitLabURL(input, ref)
	}

	if !strings.Contains(input, ":") && !strings.HasPrefix(input, ".") && !strings.HasPrefix(input, "/") {
		if match := githubShorthandPattern.FindStringSubmatch(input); match != nil {
			subpath, err := sanitizeSubpath(match[3])
			if err != nil {
				return Source{}, err
			}
			return Source{
				Original: input,
				Type:     "github",
				RepoURL:  fmt.Sprintf("https://github.com/%s/%s.git", match[1], strings.TrimSuffix(match[2], ".git")),
				Ref:      ref,
				Subpath:  subpath,
			}, nil
		}
	}

	if isDirectGitURL(input) {
		return Source{
			Original: input,
			Type:     "git",
			RepoURL:  input,
			Ref:      ref,
		}, nil
	}

	return Source{}, fmt.Errorf("unsupported source %q", input)
}

// NewBuiltinSource 构造一个稳定的内置库来源描述。
func NewBuiltinSource(library string, variant string) Source {
	library = strings.TrimSpace(library)
	if library == "devwiki" || variant == "" {
		variant = defaultBuiltinDevwikiLang
	}
	base := builtinSourceNamespace + "/" + library
	return Source{
		Original: appendFragmentRef(base, variant),
		Type:     "builtin",
		Ref:      variant,
		Builtin:  library,
	}
}

func stableSourceString(source Source) string {
	if source.Type == "local" && source.LocalDir != "" {
		return source.LocalDir
	}
	return source.Original
}

// ResolveSource 把来源解析成本地可访问目录；远端来源会先 clone 到临时目录。
// 这里会统一关闭 git 的交互式凭证提示，避免 CLI 在无人值守场景下卡住。
func ResolveSource(ctx context.Context, source Source) (ResolvedSource, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if source.Type == "builtin" {
		return resolveBuiltinSource(source)
	}

	if source.Type == "local" {
		info, err := os.Stat(source.LocalDir)
		if err != nil {
			return ResolvedSource{}, fmt.Errorf("stat local source: %w", err)
		}
		if !info.IsDir() {
			return ResolvedSource{}, fmt.Errorf("local source %q is not a directory", source.LocalDir)
		}
		return ResolvedSource{
			Source:  source,
			RootDir: source.LocalDir,
		}, nil
	}

	tempDir, err := os.MkdirTemp("", "go-skills-*")
	if err != nil {
		return ResolvedSource{}, fmt.Errorf("create temp dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cloneTimeout)
	defer cancel()

	if err := cloneIntoTempDir(ctx, source, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return ResolvedSource{}, err
	}

	return ResolvedSource{
		Source:  source,
		RootDir: tempDir,
		cleanup: func() error { return os.RemoveAll(tempDir) },
	}, nil
}

func resolveBuiltinSource(source Source) (ResolvedSource, error) {
	switch source.Builtin {
	case "devwiki":
		variant := source.Ref
		if variant != defaultBuiltinDevwikiLang {
			variant = defaultBuiltinDevwikiLang
		}
		root, cleanup, err := devwiki.ExtractBuiltinSkills(variant)
		if err != nil {
			return ResolvedSource{}, err
		}
		return ResolvedSource{
			Source:  source,
			RootDir: root,
			cleanup: func() error {
				cleanup()
				return nil
			},
		}, nil
	default:
		return ResolvedSource{}, fmt.Errorf("unsupported builtin source %q", source.Builtin)
	}
}

// SearchRoot 返回真正用于技能发现的根目录。
// 当来源带有 Subpath 时，这里会做一次最终的规范化和越界校验，
// 避免旧锁文件或手工构造输入把搜索范围带出 RootDir。
func (r ResolvedSource) SearchRoot() (string, error) {
	if r.Source.Subpath == "" {
		return r.RootDir, nil
	}

	subpath, err := sanitizeSubpath(r.Source.Subpath)
	if err != nil {
		return "", err
	}
	root := filepath.Clean(r.RootDir)
	searchRoot := filepath.Join(root, filepath.FromSlash(subpath))
	rel, err := filepath.Rel(root, searchRoot)
	if err != nil {
		return "", fmt.Errorf("resolve search root: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe subpath %q escapes source root", r.Source.Subpath)
	}
	return searchRoot, nil
}

// Cleanup 释放解析来源时创建的临时资源。
func (r ResolvedSource) Cleanup() error {
	if r.cleanup == nil {
		return nil
	}
	return r.cleanup()
}

// Discover 递归扫描来源目录，找到所有包含 SKILL.md 的技能目录。
func Discover(root string) ([]Skill, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve source root: %w", err)
	}

	var found []Skill
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}

		// 常见依赖目录里通常不应继续递归，否则发现速度和结果噪声都会明显变差。
		if shouldSkipDir(d.Name()) && path != root {
			return filepath.SkipDir
		}

		skillFile := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			return nil
		}

		skill, err := ParseSkillFile(skillFile)
		if err != nil {
			return err
		}
		skill.Dir = path
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		skill.RelativeDir = filepath.ToSlash(rel)
		found = append(found, skill)
		// 一个目录一旦命中 SKILL.md，就视为一个完整技能根。
		// 继续向下扫描只会把嵌套示例目录误判成独立技能，因此直接剪枝。
		if path != root {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].Name < found[j].Name
	})

	return found, nil
}

// ParseSkillFile 解析单个 SKILL.md 的 YAML 前置信息。
func ParseSkillFile(path string) (Skill, error) {
	file, err := os.Open(path)
	if err != nil {
		return Skill{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return Skill{}, fmt.Errorf("%s: missing YAML frontmatter", path)
	}

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return Skill{}, fmt.Errorf("read %s: %w", path, err)
	}

	var meta skillFrontmatter
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &meta); err != nil {
		return Skill{}, fmt.Errorf("%s: parse frontmatter: %w", path, err)
	}

	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = filepath.Base(filepath.Dir(path))
	}

	description := strings.TrimSpace(meta.Description)
	if description == "" {
		description = fallbackSkillDescription
	}

	return Skill{
		Name:        name,
		Description: description,
	}, nil
}

func parseGitHubURL(input, ref string) (Source, error) {
	u, err := url.Parse(input)
	if err != nil {
		return Source{}, fmt.Errorf("parse github url: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return Source{}, fmt.Errorf("invalid github url %q", input)
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	source := Source{
		Original: input,
		Type:     "github",
		RepoURL:  fmt.Sprintf("%s://%s/%s/%s.git", u.Scheme, u.Host, owner, repo),
		Ref:      ref,
	}

	if len(parts) >= 5 && parts[2] == "tree" {
		subpath, err := sanitizeSubpath(strings.Join(parts[4:], "/"))
		if err != nil {
			return Source{}, err
		}
		source.Ref = parts[3]
		source.Subpath = subpath
	}

	return source, nil
}

func parseBuiltinSource(input, ref string) (Source, bool, error) {
	if !strings.HasPrefix(input, builtinSourceNamespace+"/") {
		return Source{}, false, nil
	}

	parts := strings.Split(strings.Trim(input, "/"), "/")
	if len(parts) != 2 {
		return Source{}, true, fmt.Errorf("builtin source %q does not support nested paths", input)
	}
	if ref == "" {
		ref = defaultBuiltinDevwikiLang
	}
	return NewBuiltinSource(parts[1], ref), true, nil
}

func parseGitLabURL(input, ref string) (Source, error) {
	u, err := url.Parse(input)
	if err != nil {
		return Source{}, fmt.Errorf("parse gitlab url: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return Source{}, fmt.Errorf("invalid gitlab url %q", input)
	}

	repoPath := parts
	source := Source{
		Original: input,
		Type:     "gitlab",
		Ref:      ref,
	}

	for i := 0; i < len(parts); i++ {
		if parts[i] == "-" && i+2 < len(parts) && parts[i+1] == "tree" {
			source.Ref = parts[i+2]
			repoPath = parts[:i]
			if i+3 < len(parts) {
				subpath, err := sanitizeSubpath(strings.Join(parts[i+3:], "/"))
				if err != nil {
					return Source{}, err
				}
				source.Subpath = subpath
			}
			break
		}
	}

	if len(repoPath) < 2 {
		return Source{}, fmt.Errorf("invalid gitlab url %q", input)
	}

	source.RepoURL = fmt.Sprintf("%s://%s/%s.git", u.Scheme, u.Host, strings.TrimSuffix(strings.Join(repoPath, "/"), ".git"))
	return source, nil
}

func isLocalPath(input string) bool {
	if filepath.IsAbs(input) {
		return true
	}
	if strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") || input == "." || input == ".." {
		return true
	}
	if windowsAbsPattern.MatchString(input) {
		return true
	}
	if _, err := os.Stat(input); err == nil {
		return true
	}
	return false
}

// parseFragmentRef 解析来源末尾的 #ref 语法。
// 这里只消费 fragment 中的引用名；如果 fragment 里还带 @subpath，当前实现会忽略 @ 之后的内容。
func parseFragmentRef(input string) (string, string) {
	hashIndex := strings.Index(input, "#")
	if hashIndex < 0 {
		return input, ""
	}

	base := input[:hashIndex]
	fragment := input[hashIndex+1:]
	if fragment == "" || !looksLikeGitSource(base) {
		return input, ""
	}

	if atIndex := strings.Index(fragment, "@"); atIndex >= 0 {
		fragment = fragment[:atIndex]
	}
	return base, decodeFragmentValue(fragment)
}

func appendFragmentRef(input, ref string) string {
	if ref == "" {
		return input
	}
	return input + "#" + ref
}

func looksLikeGitSource(input string) bool {
	if strings.HasPrefix(input, "github:") || strings.HasPrefix(input, "gitlab:") || strings.HasPrefix(input, "git@") {
		return true
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err == nil {
			path := u.Path
			if u.Host == "github.com" {
				matched, _ := regexp.MatchString(`^/[^/]+/[^/]+(?:\.git)?(?:/tree/[^/]+(?:/.*)?)?/?$`, path)
				return matched
			}
			if u.Host == "gitlab.com" {
				matched, _ := regexp.MatchString(`^/.+?/[^/]+(?:\.git)?(?:/-/tree/[^/]+(?:/.*)?)?/?$`, path)
				return matched
			}
		}
	}

	if matched, _ := regexp.MatchString(`^https?://.+\.git(?:$|[/?])`, input); matched {
		return true
	}

	return !strings.Contains(input, ":") &&
		!strings.HasPrefix(input, ".") &&
		!strings.HasPrefix(input, "/") &&
		githubShorthandPattern.MatchString(input)
}

func isDirectGitURL(input string) bool {
	if strings.HasPrefix(input, "git@") || strings.HasPrefix(input, "ssh://") {
		return true
	}
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		if strings.HasSuffix(input, ".git") || strings.Contains(input, ".git?") {
			return true
		}
	}
	return false
}

func sanitizeSubpath(subpath string) (string, error) {
	if strings.TrimSpace(subpath) == "" {
		return "", nil
	}

	normalized := strings.ReplaceAll(strings.TrimSpace(subpath), "\\", "/")
	if strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("unsafe subpath %q is absolute", subpath)
	}

	cleaned := path.Clean(normalized)
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("unsafe subpath %q contains path traversal segments", subpath)
	}
	return cleaned, nil
}

func decodeFragmentValue(value string) string {
	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".idea", ".vscode":
		return true
	default:
		return false
	}
}

func cloneIntoTempDir(ctx context.Context, source Source, tempDir string) error {
	if source.Ref == "" {
		return runGit(ctx, "", "clone", "--depth", "1", source.RepoURL, tempDir)
	}

	if !looksLikeCommitHash(source.Ref) {
		return runGit(ctx, "", "clone", "--depth", "1", "--branch", source.Ref, source.RepoURL, tempDir)
	}

	// commit 哈希不能可靠地通过 `git clone --branch` 解析，因此先 clone 默认分支，
	// 再显式 fetch 目标提交并切到 FETCH_HEAD。
	if err := runGit(ctx, "", "clone", "--depth", "1", source.RepoURL, tempDir); err != nil {
		return err
	}
	if err := runGit(ctx, tempDir, "fetch", "--depth", "1", "origin", source.Ref); err != nil {
		return err
	}
	return runGit(ctx, tempDir, "checkout", "--detach", "FETCH_HEAD")
}

func runGit(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return nil
}

func looksLikeCommitHash(ref string) bool {
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, r := range ref {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
