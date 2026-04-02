package skill

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// InstalledSkill 记录已经安装到本地或全局的技能元数据。
type InstalledSkill struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Path         string            `json:"path"`
	Source       string            `json:"source"`
	SourceType   string            `json:"source_type"`
	SourceSubdir string            `json:"source_subdir,omitempty"`
	Hash         string            `json:"hash"`
	Agents       []string          `json:"agents,omitempty"`
	AgentPaths   map[string]string `json:"agent_paths,omitempty"`
	InstalledAt  time.Time         `json:"installed_at"`
}

// LockFile 表示锁文件的序列化结构。
type LockFile struct {
	Skills map[string]InstalledSkill `json:"skills"`
}

// CheckResult 表示单个技能的更新检查结果。
type CheckResult struct {
	Skill         InstalledSkill
	Status        string
	LatestHash    string
	Message       string
	MatchedSource string
}

// LockFileName 是项目级和全局级共用的锁文件名。
const LockFileName = ".zskill-lock.json"

// LoadLock 从指定路径加载锁文件；不存在时返回空锁结构。
func LoadLock(path string) (LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LockFile{Skills: map[string]InstalledSkill{}}, nil
		}
		return LockFile{}, err
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return LockFile{}, err
	}
	if lock.Skills == nil {
		lock.Skills = map[string]InstalledSkill{}
	}
	return lock, nil
}

// SaveLock 将锁文件持久化到目标路径，并确保父目录存在。
func SaveLock(path string, lock LockFile) error {
	if lock.Skills == nil {
		lock.Skills = map[string]InstalledSkill{}
	}
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// EnsureDir 确保目标目录存在。
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// CopyDir 递归复制整个目录；如果目标已存在则先清空再写入。
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

// HashDir 计算目录内所有文件内容的稳定哈希，用于检测更新。
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
func InstallSkill(skillsDir string, source Source, skill Skill) (InstalledSkill, error) {
	targetDir := filepath.Join(skillsDir, SanitizeName(skill.Name))
	if err := CopyDir(skill.Dir, targetDir); err != nil {
		return InstalledSkill{}, fmt.Errorf("install %s: %w", skill.Name, err)
	}

	hash, err := HashDir(targetDir)
	if err != nil {
		return InstalledSkill{}, fmt.Errorf("hash %s: %w", skill.Name, err)
	}

	relative := skill.RelativeDir
	if relative == "." {
		relative = source.Subpath
	} else if source.Subpath != "" {
		relative = filepath.ToSlash(filepath.Join(source.Subpath, relative))
	}

	return InstalledSkill{
		Name:         skill.Name,
		Description:  skill.Description,
		Path:         targetDir,
		Source:       source.Original,
		SourceType:   source.Type,
		SourceSubdir: relative,
		Hash:         hash,
		InstalledAt:  time.Now().UTC(),
	}, nil
}

// SyncInstalledSkill 把主安装目录同步到其它代理对应的技能目录。
func SyncInstalledSkill(canonicalPath string, skillName string, targets map[string]string) (map[string]string, error) {
	paths := make(map[string]string, len(targets))
	for agent, dir := range targets {
		targetPath := filepath.Join(dir, SanitizeName(skillName))
		if err := EnsureDir(dir); err != nil {
			return nil, fmt.Errorf("prepare %s skills dir: %w", agent, err)
		}
		if err := CopyDir(canonicalPath, targetPath); err != nil {
			return nil, fmt.Errorf("sync %s to %s: %w", skillName, agent, err)
		}
		paths[agent] = targetPath
	}
	return paths, nil
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
