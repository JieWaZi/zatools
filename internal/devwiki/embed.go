package devwiki

import "embed"

// templateFS 保存 DevWiki 工程模板。
//
//go:embed template/**
var templateFS embed.FS

// TemplateFS 返回 DevWiki 的内置模板文件系统。
func TemplateFS() embed.FS {
	return templateFS
}
