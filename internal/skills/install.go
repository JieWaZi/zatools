package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// AssetKind 表示可安装资产的类型。
type AssetKind string

const (
	// SkillAsset 表示技能目录资产。
	SkillAsset AssetKind = "skill"
	// RuleAsset 表示规则文件资产。
	RuleAsset AssetKind = "rule"
)

// InstalledAsset 记录已经安装到本地的资产元数据。
// 这份结构会被写入锁文件，作为 list/check/update/remove 的事实来源。
type InstalledAsset struct {
	// DetectedAgent 表示基于资产内容或文件后缀推断出的默认 agent 标签。
	DetectedAgent string `json:"detected_agent,omitempty"`
	// Name 是资产展示名，同时也是锁文件 bucket 中的逻辑主键。
	Name string `json:"name"`
	// Description 是安装时记录下来的描述，便于 list 直接展示。
	Description string `json:"description"`
	// Path 是主安装副本所在的绝对路径。
	// 当只安装到一个 agent 时，其它地方通常也直接复用这个字段。
	Path string `json:"path"`
	// Source 是最初的来源输入，用于后续重新解析来源检查更新。
	Source string `json:"source"`
	// SourceSubdir 是资产相对仓库根目录的目录位置。
	// 目录型资产会优先使用该字段直接定位目录。
	SourceSubdir string `json:"source_subdir,omitempty"`
	// SourceRelpath 是文件型资产相对仓库根目录的路径。
	SourceRelpath string `json:"source_relpath,omitempty"`
	// SourceFiles 记录需要从 SourceRelpath 指向目录中选取的相对文件集合。
	// 当资产只对应目录中的部分文件时，check/update 会基于这份清单做稳定重建。
	SourceFiles []string `json:"source_files,omitempty"`
	// Hash 是安装目标内容哈希，用来判断来源是否有变更。
	Hash string `json:"hash"`
	// Agents 记录当前资产同步到了哪些 agent。
	Agents []string `json:"agents,omitempty"`
	// AgentPaths 记录每个 agent 实际落盘的路径。
	AgentPaths map[string]string `json:"agent_paths,omitempty"`
}

// LockFile 表示锁文件的序列化结构。
// 当前仍以“同一 bucket 内名称唯一”作为逻辑主键。
type LockFile struct {
	// Skills 保存 skill 资产的已安装记录。
	Skills map[string]InstalledAsset `json:"skills,omitempty"`
	// Rules 保存 rule 资产的已安装记录。
	Rules map[string]InstalledAsset `json:"rules,omitempty"`
}

// CheckResult 表示单个资产的更新检查结果。
type CheckResult struct {
	// Asset 是锁文件中的已安装记录。
	Asset InstalledAsset
	// Status 是检查状态，例如 current、outdated、source-error。
	Status string
	// LatestHash 是来源当前内容的哈希；仅检查成功时有意义。
	LatestHash string
	// Message 保存错误说明或补充信息。
	Message string
}

// InstallSpec 描述一次通用资产安装所需的源元数据和目标布局。
type InstallSpec struct {
	// Name 是资产展示名。
	Name string
	// Description 是资产描述。
	Description string
	// SourcePath 是待安装源路径，可以是目录或文件。
	SourcePath string
	// TargetRelativePath 是相对目标根目录的落盘路径。
	TargetRelativePath string
	// SourceRelativeDir 是资产相对仓库根目录的目录位置。
	SourceRelativeDir string
	// SourceRelativePath 是资产相对仓库根目录的文件位置。
	SourceRelativePath string
	// SourceFiles 记录需要从 SourcePath 中安装的相对文件集合。
	// 为空时表示复制整个 SourcePath；非空时只复制列出的文件。
	SourceFiles []string
}

// LockFileName 是项目级和全局级共用的锁文件名。
const LockFileName = ".zatools-lock.json"

// LoadLock 从指定路径加载锁文件；不存在时返回空锁结构。
func LoadLock(path string) (LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewLockFile(), nil
		}
		return LockFile{}, err
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return LockFile{}, err
	}
	lock.Ensure()
	return lock, nil
}

// SaveLock 将锁文件持久化到目标路径，并确保父目录存在。
func SaveLock(path string, lock LockFile) error {
	lock.Ensure()
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// NewLockFile 返回一个已初始化 bucket 的锁文件结构。
func NewLockFile() LockFile {
	return LockFile{
		Skills: map[string]InstalledAsset{},
		Rules:  map[string]InstalledAsset{},
	}
}

// Ensure 确保锁文件中的 bucket 都已初始化。
func (l *LockFile) Ensure() {
	if l.Skills == nil {
		l.Skills = map[string]InstalledAsset{}
	}
	if l.Rules == nil {
		l.Rules = map[string]InstalledAsset{}
	}
}

// Entries 返回指定资产类型对应的 bucket。
func (l *LockFile) Entries(kind AssetKind) map[string]InstalledAsset {
	l.Ensure()
	switch kind {
	case RuleAsset:
		return l.Rules
	case SkillAsset:
		fallthrough
	default:
		return l.Skills
	}
}

// EnsureDir 确保目标目录存在。
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// CopyDir 递归复制整个目录；如果目标已存在则先清空再写入。
// 这里保留符号链接本身，而不是把链接目标内容拍平复制。
func CopyDir(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("clear destination %s: %w", dst, err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create destination %s: %w", dst, err)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		return copyFile(path, target, info.Mode())
	})
}

// CopyPath 根据源路径类型复制目录或文件。
func CopyPath(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return CopyDir(src, dst)
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("clear destination %s: %w", dst, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent for %s: %w", dst, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(link, dst)
	}
	return copyFile(src, dst, info.Mode())
}

// CopySelectedFiles 从源目录复制指定相对文件集合到目标目录。
func CopySelectedFiles(srcRoot, dstRoot string, files []string) error {
	if err := os.RemoveAll(dstRoot); err != nil {
		return fmt.Errorf("clear destination %s: %w", dstRoot, err)
	}
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return fmt.Errorf("create destination %s: %w", dstRoot, err)
	}

	selected, err := normalizeSelectedFiles(files)
	if err != nil {
		return err
	}
	for _, rel := range selected {
		srcPath := filepath.Join(srcRoot, filepath.FromSlash(rel))
		info, err := os.Lstat(srcPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("selected path %q is a directory", rel)
		}

		dstPath := filepath.Join(dstRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return fmt.Errorf("create parent for %s: %w", dstPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			if err := os.Symlink(link, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

// HashDir 计算目录内所有文件内容的稳定哈希，用于检测更新。
// 哈希同时包含相对路径和文件内容，避免“内容相同但文件结构不同”时被误判为未变化。
func HashDir(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	var files []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	hash := sha256.New()
	for _, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(filepath.ToSlash(rel))); err != nil {
			return "", err
		}

		f, err := os.Open(file)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(hash, f); err != nil {
			_ = f.Close()
			return "", err
		}
		_ = f.Close()
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HashFile 计算单文件内容的稳定哈希。
func HashFile(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := hash.Write([]byte(filepath.Base(path))); err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte(link)); err != nil {
			return "", err
		}
		return hex.EncodeToString(hash.Sum(nil)), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HashPath 根据路径类型对文件或目录计算哈希。
func HashPath(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return HashDir(path)
	}
	return HashFile(path)
}

// HashSelectedFiles 对目录中指定相对文件集合计算稳定哈希。
func HashSelectedFiles(root string, files []string) (string, error) {
	selected, err := normalizeSelectedFiles(files)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	for _, rel := range selected {
		fullPath := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Lstat(fullPath)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return "", fmt.Errorf("selected path %q is a directory", rel)
		}
		if _, err := hash.Write([]byte(rel)); err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(fullPath)
			if err != nil {
				return "", err
			}
			if _, err := hash.Write([]byte(link)); err != nil {
				return "", err
			}
			continue
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(hash, f); err != nil {
			_ = f.Close()
			return "", err
		}
		_ = f.Close()
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SanitizeName 把技能名转换成适合落盘的目录名。
func SanitizeName(name string) string {
	var out []rune
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
			lastDash = false
		case r >= '0' && r <= '9':
			out = append(out, r)
			lastDash = false
		case r == '-' || r == '_' || r == '.':
			out = append(out, r)
			lastDash = false
		default:
			if !lastDash {
				out = append(out, '-')
				lastDash = true
			}
		}
	}
	s := strings.Trim(string(out), "-.")
	if s == "" {
		return "unnamed-skill"
	}
	return s
}

// InstallSkill 把技能复制到目标目录，并返回安装记录。
// SourceSubdir 会尽量落成“相对仓库根目录的技能目录”，这样 update/check 可以直接命中该技能。
func InstallSkill(skillsDir string, source Source, skill Skill) (InstalledAsset, error) {
	return InstallAsset(skillsDir, source, InstallSpec{
		Name:               skill.Name,
		Description:        skill.Description,
		SourcePath:         skill.Dir,
		TargetRelativePath: SanitizeName(skill.Name),
		SourceRelativeDir:  skill.RelativeDir,
	})
}

// InstallAsset 把资产复制到目标目录，并返回安装记录。
func InstallAsset(targetRoot string, source Source, spec InstallSpec) (InstalledAsset, error) {
	targetPath := filepath.Join(targetRoot, spec.TargetRelativePath)
	var err error
	if len(spec.SourceFiles) > 0 {
		err = CopySelectedFiles(spec.SourcePath, targetPath, spec.SourceFiles)
	} else {
		err = CopyPath(spec.SourcePath, targetPath)
	}
	if err != nil {
		return InstalledAsset{}, fmt.Errorf("install %s: %w", spec.Name, err)
	}

	hash, err := HashPath(targetPath)
	if err != nil {
		return InstalledAsset{}, fmt.Errorf("hash %s: %w", spec.Name, err)
	}

	return InstalledAsset{
		Name:          spec.Name,
		Description:   spec.Description,
		Path:          targetPath,
		Source:        stableSourceString(source),
		SourceSubdir:  joinSourceRelative(source.Subpath, spec.SourceRelativeDir),
		SourceRelpath: joinSourceRelative(source.Subpath, spec.SourceRelativePath),
		SourceFiles:   append([]string(nil), spec.SourceFiles...),
		Hash:          hash,
	}, nil
}

// SyncInstalledSkill 把主安装目录同步到其它代理对应的技能目录。
// 这样可以保证多个 agent 安装的是同一份已落盘内容，避免每个 agent 都单独从来源重新复制。
func SyncInstalledSkill(canonicalPath string, skillName string, targets map[string]string) (map[string]string, error) {
	return SyncInstalledPath(canonicalPath, SanitizeName(skillName), targets)
}

// SyncInstalledPath 把主安装路径同步到其它代理对应的目录。
func SyncInstalledPath(canonicalPath string, targetRelativePath string, targets map[string]string) (map[string]string, error) {
	paths := make(map[string]string, len(targets))
	for agent, dir := range targets {
		targetPath := filepath.Join(dir, targetRelativePath)
		if err := EnsureDir(dir); err != nil {
			return nil, fmt.Errorf("prepare %s install dir: %w", agent, err)
		}
		if err := CopyPath(canonicalPath, targetPath); err != nil {
			return nil, fmt.Errorf("sync %s to %s: %w", targetRelativePath, agent, err)
		}
		paths[agent] = targetPath
	}
	return paths, nil
}

func joinSourceRelative(prefix string, rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	switch rel {
	case "", ".":
		return filepath.ToSlash(strings.TrimSpace(prefix))
	}
	if prefix == "" {
		return rel
	}
	return filepath.ToSlash(filepath.Join(prefix, rel))
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func normalizeSelectedFiles(files []string) ([]string, error) {
	selected := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		normalized, err := normalizeRelativeFile(file)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		selected = append(selected, normalized)
	}
	sort.Strings(selected)
	return selected, nil
}

func normalizeRelativeFile(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("selected file path is empty")
	}
	normalized := path.Clean(filepath.ToSlash(rel))
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") || path.IsAbs(normalized) {
		return "", fmt.Errorf("invalid selected file path %q", rel)
	}
	return normalized, nil
}
