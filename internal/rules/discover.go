package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"zatools/internal/skills"
)

const (
	// CursorTag 表示基于规则文件后缀推断出的 Cursor 标签。
	CursorTag = "cursor"
	// ClaudeTag 表示基于规则文件后缀推断出的 Claude 标签。
	ClaudeTag = "claude"
)

var metadataFilenames = []string{"RULE.yaml", "RULES.yaml"}

// Rule 表示一组可安装的规则目录。
type Rule struct {
	// Name 是规则展示名，优先来自 RULE.yaml/RULES.yaml 的配置。
	Name string
	// Description 是规则描述，优先来自 RULE.yaml/RULES.yaml 的配置。
	Description string
	// Dir 是规则目录的绝对路径。
	Dir string
	// RelativeDir 是规则目录相对发现根目录的路径。
	RelativeDir string
	// InstallName 是安装到目标目录时使用的稳定目录名。
	InstallName string
	// Files 记录规则目录中包含的规则文件相对路径。
	Files []string
	// DetectedAgents 表示根据目录中规则文件推断出的默认 agent 标签集合。
	DetectedAgents []string
}

type ruleMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type metadataSet struct {
	Direct  ruleMeta
	Entries map[string]ruleMeta
}

// Discover 递归扫描根目录，按目录聚合找到所有规则包。
func Discover(root string) ([]Rule, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve rules root: %w", err)
	}

	rootEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	rootMeta, err := loadMetadata(root)
	if err != nil {
		return nil, err
	}

	var found []Rule
	rootFiles := directSupportedFiles(root)
	if len(rootFiles) > 0 {
		rule, err := buildRule(root, root, rootMeta.Direct, rootFiles)
		if err != nil {
			return nil, err
		}
		found = append(found, rule)
	}

	for _, entry := range rootEntries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		files := supportedFiles(dir, root)
		if len(files) == 0 {
			continue
		}

		meta := rootMeta.Entries[entry.Name()]
		dirMeta, err := loadMetadata(dir)
		if err != nil {
			return nil, err
		}
		if dirMeta.Direct.Name != "" || dirMeta.Direct.Description != "" {
			meta = dirMeta.Direct
		}

		rule, err := buildRule(dir, root, meta, files)
		if err != nil {
			return nil, err
		}
		found = append(found, rule)
	}

	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found, nil
}

// DefaultRoot 返回默认规则扫描根目录。
// 如果仓库根目录下存在 rules 子目录，则优先使用它；否则回退到仓库根目录。
func DefaultRoot(root string) string {
	rulesDir := filepath.Join(root, "rules")
	if info, err := os.Stat(rulesDir); err == nil && info.IsDir() {
		return rulesDir
	}
	return root
}

// Find 根据名称查找规则定义。
func Find(found []Rule, name string) (Rule, bool) {
	for _, rule := range found {
		if rule.Name == name {
			return rule, true
		}
	}
	return Rule{}, false
}

func buildRule(dir string, root string, meta ruleMeta, files []string) (Rule, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return Rule{}, err
	}
	rel = filepath.ToSlash(rel)
	defaultName := filepath.Base(dir)
	if rel == "." {
		defaultName = filepath.Base(root)
	}

	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = defaultName
	}
	description := strings.TrimSpace(meta.Description)
	if description == "" {
		description = fallbackDescription(files)
	}

	return Rule{
		Name:           name,
		Description:    description,
		Dir:            dir,
		RelativeDir:    rel,
		InstallName:    targetDirName(rel),
		Files:          files,
		DetectedAgents: detectAgents(files),
	}, nil
}

func supportedFiles(dir string, root string) []string {
	var files []string
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if isMetadataFile(filepath.Base(path)) {
			return nil
		}
		if _, ok := detectAgentByExt(filepath.Ext(path)); !ok {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files
}

func directSupportedFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || isMetadataFile(entry.Name()) {
			continue
		}
		if _, ok := detectAgentByExt(filepath.Ext(entry.Name())); !ok {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files
}

func detectAgents(files []string) []string {
	seen := map[string]bool{}
	for _, file := range files {
		tag, ok := detectAgentByExt(filepath.Ext(file))
		if ok {
			seen[tag] = true
		}
	}

	var tags []string
	for _, item := range []string{ClaudeTag, CursorTag} {
		if seen[item] {
			tags = append(tags, item)
		}
	}
	return tags
}

func fallbackDescription(files []string) string {
	if len(files) == 0 {
		return ""
	}
	limit := len(files)
	if limit > 3 {
		limit = 3
	}
	visible := make([]string, 0, limit)
	for _, file := range files[:limit] {
		visible = append(visible, filepath.Base(file))
	}
	if len(files) > limit {
		return strings.Join(visible, ", ") + fmt.Sprintf(" +%d", len(files)-limit)
	}
	return strings.Join(visible, ", ")
}

func targetDirName(rel string) string {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" || rel == "." {
		return "root-rules"
	}

	parts := strings.Split(rel, "/")
	for i, part := range parts {
		parts[i] = skills.SanitizeName(part)
	}
	return strings.Join(parts, "--")
}

func detectAgentByExt(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".mdc":
		return CursorTag, true
	case ".md":
		return ClaudeTag, true
	default:
		return "", false
	}
}

func loadMetadata(dir string) (metadataSet, error) {
	for _, name := range metadataFilenames {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return metadataSet{}, fmt.Errorf("read rule metadata %s: %w", path, err)
		}
		return parseMetadata(data, path)
	}
	return metadataSet{Entries: map[string]ruleMeta{}}, nil
}

func parseMetadata(data []byte, path string) (metadataSet, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return metadataSet{}, fmt.Errorf("parse rule metadata %s: %w", path, err)
	}
	if len(node.Content) == 0 {
		return metadataSet{Entries: map[string]ruleMeta{}}, nil
	}

	mapping := node.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return metadataSet{}, fmt.Errorf("%s: rule metadata must be a YAML mapping", path)
	}

	if hasDirectMeta(mapping) {
		var meta ruleMeta
		if err := mapping.Decode(&meta); err != nil {
			return metadataSet{}, fmt.Errorf("decode rule metadata %s: %w", path, err)
		}
		return metadataSet{Direct: meta, Entries: map[string]ruleMeta{}}, nil
	}

	entries := make(map[string]ruleMeta, len(mapping.Content)/2)
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := strings.TrimSpace(mapping.Content[i].Value)
		if key == "" {
			continue
		}
		var meta ruleMeta
		if err := mapping.Content[i+1].Decode(&meta); err != nil {
			return metadataSet{}, fmt.Errorf("decode rule metadata entry %s:%s: %w", path, key, err)
		}
		entries[key] = meta
	}
	return metadataSet{Entries: entries}, nil
}

func hasDirectMeta(node *yaml.Node) bool {
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := strings.TrimSpace(node.Content[i].Value)
		if key == "name" || key == "description" {
			return true
		}
	}
	return false
}

func isMetadataFile(name string) bool {
	for _, candidate := range metadataFilenames {
		if name == candidate {
			return true
		}
	}
	return false
}
