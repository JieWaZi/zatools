package common

import (
	"os"

	"zatools/internal/skills"
)

// Runtime 保存一次 CLI 执行中可复用的环境信息。
type Runtime struct {
	// Workspace 描述命令执行时的项目工作区。
	Workspace *skills.Workspace
	// IsTTY 标记当前 stdout 是否连接到交互终端。
	IsTTY bool
}

// DetectRuntime 根据当前进程环境构建运行时信息。
func DetectRuntime() Runtime {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	info, err := os.Stdout.Stat()
	isTTY := err == nil && (info.Mode()&os.ModeCharDevice) != 0

	return Runtime{
		Workspace: skills.NewWorkspace(cwd),
		IsTTY:     isTTY,
	}
}
