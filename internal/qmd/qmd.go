package qmd

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultEmbedModel    = "hf:Qwen/Qwen3-Embedding-0.6B-GGUF/Qwen3-Embedding-0.6B-Q8_0.gguf"
	DefaultRerankModel   = "hf:ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF/qwen3-reranker-0.6b-q8_0.gguf"
	DefaultGenerateModel = "hf:tobil/qmd-query-expansion-1.7B-gguf/qmd-query-expansion-1.7B-q4_k_m.gguf"
	WarmupQuery          = "zatools qmd model warmup"
)

var projectRootMarkers = []string{
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
}

type commandContextKey struct{}

// Collection is one `config/search.yaml` collection entry.
type Collection struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Config is the parsed qmd section from `config/search.yaml`.
type Config struct {
	Collections   []Collection `yaml:"collections"`
	EmbedModel    string       `yaml:"embed_model"`
	RerankModel   string       `yaml:"rerank_model"`
	GenerateModel string       `yaml:"generate_model"`
}

// Models describes the model arguments passed to qmd commands.
type Models struct {
	EmbedModel    string
	RerankModel   string
	GenerateModel string
}

// DefaultModels returns the built-in default qmd model configuration.
func DefaultModels() Models {
	return Models{
		EmbedModel:    DefaultEmbedModel,
		RerankModel:   DefaultRerankModel,
		GenerateModel: DefaultGenerateModel,
	}
}

// ResolveModels loads `config/search.yaml` models when available and applies explicit overrides last.
func ResolveModels(root string, overrides Models) (Models, error) {
	models := DefaultModels()
	config, err := LoadConfig(root)
	switch {
	case err == nil:
		if value := strings.TrimSpace(config.EmbedModel); value != "" {
			models.EmbedModel = value
		}
		if value := strings.TrimSpace(config.RerankModel); value != "" {
			models.RerankModel = value
		}
		if value := strings.TrimSpace(config.GenerateModel); value != "" {
			models.GenerateModel = value
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return Models{}, err
	}

	if value := strings.TrimSpace(overrides.EmbedModel); value != "" {
		models.EmbedModel = value
	}
	if value := strings.TrimSpace(overrides.RerankModel); value != "" {
		models.RerankModel = value
	}
	if value := strings.TrimSpace(overrides.GenerateModel); value != "" {
		models.GenerateModel = value
	}
	return normalizeModels(models), nil
}

// LoadConfig reads the qmd section from `config/search.yaml`.
func LoadConfig(root string) (Config, error) {
	type config struct {
		QMD Config `yaml:"qmd"`
	}

	data, err := os.ReadFile(filepath.Join(root, "config", "search.yaml"))
	if err != nil {
		return Config{}, err
	}

	var parsed config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return Config{}, err
	}
	return parsed.QMD, nil
}

// LoadCollections reads qmd collections from `config/search.yaml`.
func LoadCollections(root string) ([]Collection, error) {
	config, err := LoadConfig(root)
	if err != nil {
		return nil, err
	}
	return config.Collections, nil
}

// BuildCollectionCommands returns the qmd registration commands for the collections.
func BuildCollectionCommands(root string, collections []Collection) ([][]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	commands := make([][]string, 0, len(collections))
	for _, collection := range collections {
		path := collection.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(absRoot, path)
		}
		path = filepath.Clean(path)
		commands = append(commands, []string{"qmd", "collection", "add", path, "--name", collection.Name})
	}
	return commands, nil
}

// BuildDownloadCommands returns the warmup command sequence used to pre-download qmd models.
func BuildDownloadCommands(root string, collections []Collection) ([][]string, error) {
	commands, err := BuildCollectionCommands(root, collections)
	if err != nil {
		return nil, err
	}
	commands = append(commands,
		[]string{"qmd", "update"},
		[]string{"qmd", "embed", "-f"},
		[]string{"qmd", "query", WarmupQuery},
	)
	return commands, nil
}

// BuildCommandEnv returns qmd runtime env vars for command execution.
func BuildCommandEnv(models Models, cwd string) []string {
	models = normalizeModels(models)
	env := make([]string, 0, 4)
	if value := strings.TrimSpace(models.EmbedModel); value != "" {
		env = append(env, "QMD_EMBED_MODEL="+value)
	}
	if value := strings.TrimSpace(models.RerankModel); value != "" {
		env = append(env, "QMD_RERANK_MODEL="+value)
	}
	if value := strings.TrimSpace(models.GenerateModel); value != "" {
		env = append(env, "QMD_GENERATE_MODEL="+value)
	}
	env = append(env, "XDG_CACHE_HOME="+resolveCacheHome(cwd))
	return env
}

// RunCollectionCommands executes qmd registration commands sequentially.
func RunCollectionCommands(ctx context.Context, commands [][]string, models Models, stdout io.Writer, stderr io.Writer) error {
	env, err := currentCommandEnv(models)
	if err != nil {
		return err
	}
	return runCommands(ctx, commands, "", stdout, stderr, env)
}

// RunCommand executes one qmd command with the provided arguments.
func RunCommand(ctx context.Context, args []string, models Models, stdout io.Writer, stderr io.Writer) error {
	command := append([]string{"qmd"}, args...)
	env, err := currentCommandEnv(models)
	if err != nil {
		return err
	}
	return runCommands(ctx, [][]string{command}, "", stdout, stderr, env)
}

// RunCollectionCommandsInDir executes qmd registration commands sequentially from the provided directory.
func RunCollectionCommandsInDir(ctx context.Context, dir string, commands [][]string, models Models, stdout io.Writer, stderr io.Writer) error {
	env := BuildCommandEnv(models, dir)
	return runCommands(ctx, commands, dir, stdout, stderr, env)
}

// RunCommandInDir executes one qmd command from the provided directory.
func RunCommandInDir(ctx context.Context, dir string, args []string, models Models, stdout io.Writer, stderr io.Writer) error {
	command := append([]string{"qmd"}, args...)
	env := BuildCommandEnv(models, dir)
	return runCommands(ctx, [][]string{command}, dir, stdout, stderr, env)
}

// RunDownload warms qmd models for a DevWiki workspace and writes them into the rooted cache.
func RunDownload(ctx context.Context, root string, overrides Models, stdout io.Writer, stderr io.Writer) error {
	collections, err := LoadCollections(root)
	if err != nil {
		return err
	}
	models, err := ResolveModels(root, overrides)
	if err != nil {
		return err
	}
	commands, err := BuildDownloadCommands(root, collections)
	if err != nil {
		return err
	}
	return runDownloadCommands(ctx, root, commands, models, stdout, stderr)
}

// WithCommandRunner installs a custom exec runner in context for tests.
func WithCommandRunner(ctx context.Context, runner func(context.Context, string, ...string) *exec.Cmd) context.Context {
	return context.WithValue(ctx, commandContextKey{}, runner)
}

func normalizeModels(models Models) Models {
	defaults := DefaultModels()
	if strings.TrimSpace(models.EmbedModel) == "" {
		models.EmbedModel = defaults.EmbedModel
	}
	if strings.TrimSpace(models.RerankModel) == "" {
		models.RerankModel = defaults.RerankModel
	}
	if strings.TrimSpace(models.GenerateModel) == "" {
		models.GenerateModel = defaults.GenerateModel
	}
	return models
}

func currentCommandEnv(models Models) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return BuildCommandEnv(models, cwd), nil
}

func runDownloadCommands(ctx context.Context, root string, commands [][]string, models Models, stdout io.Writer, stderr io.Writer) error {
	env := BuildCommandEnv(models, root)
	downloadStart := len(commands) - 3
	for index, args := range commands {
		if err := runCommands(ctx, [][]string{args}, root, stdout, stderr, env); err != nil {
			// `download` should be safely rerunnable; duplicate collection registration must not block warmup.
			if index < downloadStart {
				continue
			}
			return err
		}
	}
	return nil
}

func runCommands(ctx context.Context, commands [][]string, dir string, stdout io.Writer, stderr io.Writer, extraEnv []string) error {
	for _, args := range commands {
		if len(args) == 0 {
			continue
		}
		cmd := buildCommandContext(ctx, args[0], args[1:]...)
		if strings.TrimSpace(dir) != "" {
			cmd.Dir = dir
		}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Env = mergeCommandEnv(cmd.Env, extraEnv)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func buildCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if runner, ok := ctx.Value(commandContextKey{}).(func(context.Context, string, ...string) *exec.Cmd); ok && runner != nil {
		return runner(ctx, name, args...)
	}
	return exec.CommandContext(ctx, name, args...)
}

func mergeCommandEnv(base []string, extra []string) []string {
	if len(extra) == 0 {
		if len(base) == 0 {
			return os.Environ()
		}
		return append([]string(nil), base...)
	}

	if len(base) == 0 {
		merged := append([]string{}, os.Environ()...)
		return append(merged, extra...)
	}

	merged := append([]string{}, base...)
	return append(merged, extra...)
}

func resolveCacheHome(cwd string) string {
	root := resolveProjectRoot(cwd)
	return filepath.Join(root, ".cache")
}

func resolveProjectRoot(cwd string) string {
	if strings.TrimSpace(cwd) == "" {
		cwd = "."
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return cwd
	}

	current := abs
	for {
		if hasProjectMarker(current) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return abs
		}
		current = parent
	}
}

func hasProjectMarker(dir string) bool {
	for _, marker := range projectRootMarkers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}
