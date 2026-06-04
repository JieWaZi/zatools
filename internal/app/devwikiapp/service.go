package devwikiapp

import (
	"context"

	common "zatools/internal/app/common"
)

// InitOptions 描述 `devwiki init` 的命令参数。
type InitOptions struct {
	ProjectName   string
	Agent         string
	Lang          string
	CodeDirs      []string
	Global        bool
	ScopeProvided bool
	Yes           bool
}

// Service 编排 DevWiki 工程初始化与 runtime skill 安装。
type Service struct {
	runtime common.Runtime
}

// NewService 构建使用当前终端环境的 DevWiki 应用服务。
func NewService() *Service {
	return NewServiceWithRuntime(common.DetectRuntime())
}

// NewServiceWithRuntime 允许测试注入自定义运行环境。
func NewServiceWithRuntime(runtime common.Runtime) *Service {
	return &Service{
		runtime: runtime,
	}
}

// Init 创建一个新的 DevWiki 工程并安装选定的 runtime skills。
func (s *Service) Init(ctx context.Context, opts InitOptions) error {
	return s.runProject(ctx, opts, true)
}

// Update 更新当前作用域下已安装的 DevWiki builtin skills。
func (s *Service) Update(ctx context.Context) error {
	return s.updateSkills(ctx)
}

// Graph validates, builds, and serves the DevWiki graph view.
func (s *Service) Graph(ctx context.Context, opts GraphOptions) error {
	return s.runGraph(ctx, opts)
}

// Server serves the read-only DevWiki HTTP API.
func (s *Service) Server(ctx context.Context, opts ServerOptions) error {
	return s.runServer(ctx, opts)
}

// Check validates DevWiki documents, graph relations, or both.
func (s *Service) Check(ctx context.Context, opts CheckOptions) error {
	return s.runCheck(ctx, opts)
}

// Read prints one DevWiki page view.
func (s *Service) Read(ctx context.Context, opts ReadOptions) error {
	return s.runRead(ctx, opts)
}

// Search searches DevWiki entries and prints compact pipe-table hits.
func (s *Service) Search(ctx context.Context, opts SearchOptions) error {
	return s.runSearch(ctx, opts)
}
