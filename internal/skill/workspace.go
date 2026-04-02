// Package skill 提供技能来源解析、项目定位、锁文件读写和安装同步等核心能力。
package skill

import (
	"fmt"
	"os"
	"path/filepath"
)

// Workspace 描述一次技能操作所依赖的项目工作区信息。
type Workspace struct {
	// CWD 是命令执行时的当前工作目录。
	CWD string
	// ProjectRoot 是根据项目标记向上推导出的项目根目录。
	ProjectRoot string
}

// NewWorkspace 根据给定目录构建工作区对象，并自动解析项目根目录。
func NewWorkspace(cwd string) *Workspace {
	if cwd == "" {
		resolved, err := os.Getwd()
		if err != nil {
			cwd = "."
		} else {
			cwd = resolved
		}
	}

	return &Workspace{
		CWD:         cwd,
		ProjectRoot: resolveProjectRoot(cwd),
	}
}

// ProjectDir 返回当前命令应当使用的项目目录。
func (w *Workspace) ProjectDir() string {
	if w.ProjectRoot != "" {
		return w.ProjectRoot
	}
	return w.CWD
}

// LockFilePath 根据作用域返回对应的锁文件路径。
func (w *Workspace) LockFilePath(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, ".agents", LockFileName), nil
	}
	return filepath.Join(w.ProjectDir(), LockFileName), nil
}

// resolveProjectRoot 从当前目录开始向上查找项目标记，直到找到项目根目录。
func resolveProjectRoot(cwd string) string {
	current := cwd
	for {
		if hasProjectMarker(current) {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			return cwd
		}
		current = parent
	}
}

// hasProjectMarker 判断目录是否包含常见的项目级标记文件或目录。
func hasProjectMarker(dir string) bool {
	for _, marker := range []string{
		".git",
		".jj",
		".hg",
		"go.mod",
		"package.json",
		"pyproject.toml",
		"Cargo.toml",
		"Gemfile",
		".agents",
		".cursor",
		".claude",
	} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}
