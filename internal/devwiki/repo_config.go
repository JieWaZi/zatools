package devwiki

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// RepoSourceLocal identifies a local DevWiki document source.
	RepoSourceLocal = "local"
	// RepoSourceRemote identifies a remote DevWiki HTTP API source.
	RepoSourceRemote = "remote"
)

// RepoConfig is the per-user DevWiki project configuration.
type RepoConfig struct {
	// ProjectName is the display name for the DevWiki project.
	ProjectName string `yaml:"project_name"`
	// ProjectSlug is the stable project identifier used by CLI commands.
	ProjectSlug string `yaml:"project_slug"`
	// Language is the default project language.
	Language string `yaml:"language"`
	// ActiveSource is the source type selected for read/search operations.
	ActiveSource string `yaml:"active_source,omitempty"`
	// Sources stores every configured project knowledge source.
	Sources RepoSources `yaml:"sources,omitempty"`
	// CodeRepos contains local-only code repository bindings.
	CodeRepos []CodeRepo `yaml:"code_repos,omitempty"`
}

// RepoSources stores the switchable local and remote DevWiki sources.
type RepoSources struct {
	// Local is the local DevWiki document root source.
	Local *RepoSource `yaml:"local,omitempty" json:"local,omitempty"`
	// Remote is the remote DevWiki HTTP API source.
	Remote *RepoSource `yaml:"remote,omitempty" json:"remote,omitempty"`
}

// RepoSource describes a local or remote DevWiki knowledge source.
type RepoSource struct {
	// Type is "local" or "remote".
	Type string `yaml:"type" json:"type"`
	// URL is used for remote HTTP API sources.
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
	// Path is used for local document roots.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// RepoConfigPath returns the user config path for a project slug.
func RepoConfigPath(projectSlug string) (string, error) {
	slug := strings.TrimSpace(projectSlug)
	if slug == "" {
		return "", fmt.Errorf("devwiki project slug is required")
	}
	configRoot, err := repoConfigRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(configRoot, "devwiki", slug, "config.yaml"), nil
}

func repoConfigRoot() (string, error) {
	configRoot := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if configRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configRoot = filepath.Join(home, ".config")
	}
	return configRoot, nil
}

// ListRepoProjectSlugs returns configured DevWiki project slugs.
func ListRepoProjectSlugs() ([]string, error) {
	configRoot, err := repoConfigRoot()
	if err != nil {
		return nil, err
	}
	devwikiRoot := filepath.Join(configRoot, "devwiki")
	entries, err := os.ReadDir(devwikiRoot)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	slugs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := strings.TrimSpace(entry.Name())
		if slug == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(devwikiRoot, slug, "config.yaml")); err == nil {
			slugs = append(slugs, slug)
		} else if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Strings(slugs)
	return slugs, nil
}

// LoadRepoConfig reads a per-user DevWiki project config.
func LoadRepoConfig(projectSlug string) (RepoConfig, error) {
	path, err := RepoConfigPath(projectSlug)
	if err != nil {
		return RepoConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return RepoConfig{}, err
	}
	var cfg RepoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return RepoConfig{}, err
	}
	normalizeRepoConfig(&cfg)
	if err := ValidateRepoConfig(cfg); err != nil {
		return RepoConfig{}, err
	}
	return cfg, nil
}

// SaveRepoConfig validates and writes a per-user DevWiki project config.
func SaveRepoConfig(cfg RepoConfig) error {
	normalizeRepoConfig(&cfg)
	if err := ValidateRepoConfig(cfg); err != nil {
		return err
	}
	path, err := RepoConfigPath(cfg.ProjectSlug)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ActiveRepoSource returns the configured source selected by ActiveSource.
func ActiveRepoSource(cfg RepoConfig) (RepoSource, error) {
	switch strings.TrimSpace(cfg.ActiveSource) {
	case RepoSourceLocal:
		if cfg.Sources.Local == nil {
			return RepoSource{}, fmt.Errorf("active local devwiki source is not configured")
		}
		return *cfg.Sources.Local, nil
	case RepoSourceRemote:
		if cfg.Sources.Remote == nil {
			return RepoSource{}, fmt.Errorf("active remote devwiki source is not configured")
		}
		return *cfg.Sources.Remote, nil
	default:
		return RepoSource{}, fmt.Errorf("unsupported active devwiki source %q", cfg.ActiveSource)
	}
}

// ValidateRepoConfig validates the stable v1 repo config contract.
func ValidateRepoConfig(cfg RepoConfig) error {
	if strings.TrimSpace(cfg.ProjectSlug) == "" {
		return fmt.Errorf("devwiki project slug is required")
	}
	if strings.TrimSpace(cfg.ProjectName) == "" {
		return fmt.Errorf("devwiki project name is required")
	}
	if strings.TrimSpace(cfg.Language) == "" {
		return fmt.Errorf("devwiki project language is required")
	}
	active := strings.TrimSpace(cfg.ActiveSource)
	if active == "" {
		return fmt.Errorf("active devwiki source is required")
	}
	switch active {
	case RepoSourceLocal:
		if cfg.Sources.Local == nil {
			return fmt.Errorf("active local devwiki source is not configured")
		}
	case RepoSourceRemote:
		if cfg.Sources.Remote == nil {
			return fmt.Errorf("active remote devwiki source is not configured")
		}
	default:
		return fmt.Errorf("unsupported active devwiki source %q", active)
	}
	if cfg.Sources.Local != nil {
		if cfg.Sources.Local.Type != RepoSourceLocal {
			return fmt.Errorf("local devwiki source must have type %q", RepoSourceLocal)
		}
		if strings.TrimSpace(cfg.Sources.Local.Path) == "" {
			return fmt.Errorf("local devwiki source requires path")
		}
		if strings.TrimSpace(cfg.Sources.Local.URL) != "" {
			return fmt.Errorf("local devwiki source must not set url")
		}
	}
	if cfg.Sources.Remote != nil {
		if cfg.Sources.Remote.Type != RepoSourceRemote {
			return fmt.Errorf("remote devwiki source must have type %q", RepoSourceRemote)
		}
		if strings.TrimSpace(cfg.Sources.Remote.URL) == "" {
			return fmt.Errorf("remote devwiki source requires url")
		}
		if strings.TrimSpace(cfg.Sources.Remote.Path) != "" {
			return fmt.Errorf("remote devwiki source must not set path")
		}
	}
	for _, repo := range cfg.CodeRepos {
		if strings.TrimSpace(repo.Slug) == "" {
			return fmt.Errorf("code repo slug is required")
		}
		if strings.TrimSpace(repo.Path) == "" {
			return fmt.Errorf("code repo path is required")
		}
	}
	return nil
}

func normalizeRepoConfig(cfg *RepoConfig) {
	cfg.ActiveSource = strings.TrimSpace(cfg.ActiveSource)
}
