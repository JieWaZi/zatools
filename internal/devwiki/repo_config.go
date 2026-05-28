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
	// Source describes where project knowledge is read from.
	Source RepoSource `yaml:"source"`
	// CodeRepos contains local-only code repository bindings.
	CodeRepos []CodeRepo `yaml:"code_repos,omitempty"`
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
	if err := ValidateRepoConfig(cfg); err != nil {
		return RepoConfig{}, err
	}
	return cfg, nil
}

// SaveRepoConfig validates and writes a per-user DevWiki project config.
func SaveRepoConfig(cfg RepoConfig) error {
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
	switch cfg.Source.Type {
	case RepoSourceLocal:
		if strings.TrimSpace(cfg.Source.Path) == "" {
			return fmt.Errorf("local devwiki source requires path")
		}
		if strings.TrimSpace(cfg.Source.URL) != "" {
			return fmt.Errorf("local devwiki source must not set url")
		}
	case RepoSourceRemote:
		if strings.TrimSpace(cfg.Source.URL) == "" {
			return fmt.Errorf("remote devwiki source requires url")
		}
		if strings.TrimSpace(cfg.Source.Path) != "" {
			return fmt.Errorf("remote devwiki source must not set path")
		}
	default:
		return fmt.Errorf("unsupported devwiki source type %q", cfg.Source.Type)
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
