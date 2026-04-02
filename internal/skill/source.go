package skill

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Source 表示技能来源信息。
type Source struct {
	// Original 是用户输入的原始来源字符串。
	Original string
	// Type 表示来源类型，例如 local、github、gitlab、git。
	Type string
	// RepoURL 是解析后的远端仓库地址。
	RepoURL string
	// LocalDir 是本地目录来源的绝对路径。
	LocalDir string
	// Ref 是分支、标签或提交引用。
	Ref string
	// Subpath 是仓库中技能所在的子目录。
	Subpath string
}

// Skill 表示从来源中发现的一项技能定义。
type Skill struct {
	// Name 是技能名称。
	Name string
	// Description 是技能描述。
	Description string
	// Dir 是技能目录的绝对路径。
	Dir string
	// RelativeDir 是技能目录相对来源根目录的路径。
	RelativeDir string
}

// ResolvedSource 表示已经解析并准备好的技能来源。
type ResolvedSource struct {
	// Source 是原始来源元数据。
	Source Source
	// RootDir 是本地目录或临时 clone 目录。
	RootDir string
	cleanup func() error
}

type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var githubShorthandPattern = regexp.MustCompile(`^([^/\s]+)/([^/\s]+)(?:/(.+))?$`)
var windowsAbsPattern = regexp.MustCompile(`^[a-zA-Z]:[/\\]`)

var sourceAliases = map[string]string{
	"coinbase/agentWallet": "coinbase/agentic-wallet-skills",
}

// cloneTimeout 限制远端 clone 的最长时间，避免命令无限阻塞。
const cloneTimeout = 60 * time.Second

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
			return Source{
				Original: input,
				Type:     "github",
				RepoURL:  fmt.Sprintf("https://github.com/%s/%s.git", match[1], strings.TrimSuffix(match[2], ".git")),
				Ref:      ref,
				Subpath:  match[3],
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

// ResolveSource 把来源解析成本地可访问目录；远端来源会先 clone 到临时目录。
func ResolveSource(source Source) (ResolvedSource, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), cloneTimeout)
	defer cancel()

	args := []string{"clone", "--depth", "1"}
	if source.Ref != "" {
		args = append(args, "--branch", source.Ref)
	}
	args = append(args, source.RepoURL, tempDir)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return ResolvedSource{}, fmt.Errorf("git clone failed: %v\n%s", err, string(output))
	}

	return ResolvedSource{
		Source:  source,
		RootDir: tempDir,
		cleanup: func() error { return os.RemoveAll(tempDir) },
	}, nil
}

// SearchRoot 返回真正用于技能发现的根目录。
func (r ResolvedSource) SearchRoot() string {
	if r.Source.Subpath == "" {
		return r.RootDir
	}
	return filepath.Join(r.RootDir, filepath.FromSlash(r.Source.Subpath))
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
	if strings.TrimSpace(meta.Name) == "" || strings.TrimSpace(meta.Description) == "" {
		return Skill{}, fmt.Errorf("%s: frontmatter requires name and description", path)
	}

	return Skill{
		Name:        meta.Name,
		Description: meta.Description,
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
		source.Ref = parts[3]
		source.Subpath = strings.Join(parts[4:], "/")
	}

	return source, nil
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
	normalized := strings.ReplaceAll(subpath, "\\", "/")
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return "", fmt.Errorf("unsafe subpath %q contains path traversal segments", subpath)
		}
	}
	return subpath, nil
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
