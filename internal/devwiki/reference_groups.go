package devwiki

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const devwikiSkillsRelativeRoot = "skills/devwiki"

var skillReferencePattern = regexp.MustCompile(`references/[A-Za-z0-9._/-]+\.md`)

// ReferenceGroupConfig describes duplicated DevWiki skill reference files.
type ReferenceGroupConfig struct {
	// Groups maps a reference group name to the canonical file and its copies.
	Groups map[string]ReferenceGroup `yaml:"groups"`
}

// ReferenceGroup declares one canonical reference file and the copies that must match it.
type ReferenceGroup struct {
	// Canonical is the path, relative to skills/devwiki, used as the source of truth.
	Canonical string `yaml:"canonical"`
	// Files are the paths, relative to skills/devwiki, that must match Canonical.
	Files []string `yaml:"files"`
}

// ReferenceGroupIssue describes one reference consistency problem.
type ReferenceGroupIssue struct {
	// Group is the group key from reference-groups.yaml.
	Group string
	// Canonical is the canonical path configured for Group.
	Canonical string
	// File is the specific file path involved in the issue.
	File string
	// Reason explains the consistency problem.
	Reason string
}

// FindZatoolsRepoRoot walks upward from start and returns the zatools repository root.
func FindZatoolsRepoRoot(start string) (string, error) {
	if strings.TrimSpace(start) == "" {
		start = "."
	}
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(current)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		goMod := filepath.Join(current, "go.mod")
		groups := filepath.Join(current, devwikiSkillsRelativeRoot, "reference-groups.yaml")
		if data, err := os.ReadFile(goMod); err == nil && strings.Contains(string(data), "module zatools") {
			if _, err := os.Stat(groups); err == nil {
				return current, nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("current directory is not inside a zatools repository with %s/reference-groups.yaml", devwikiSkillsRelativeRoot)
}

// LoadReferenceGroupConfig reads skills/devwiki/reference-groups.yaml from repoRoot.
func LoadReferenceGroupConfig(repoRoot string) (ReferenceGroupConfig, error) {
	var cfg ReferenceGroupConfig
	data, err := os.ReadFile(filepath.Join(repoRoot, devwikiSkillsRelativeRoot, "reference-groups.yaml"))
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Groups == nil {
		cfg.Groups = map[string]ReferenceGroup{}
	}
	return cfg, nil
}

// CheckDevwikiReferenceGroups validates duplicated references and SKILL.md reference links.
func CheckDevwikiReferenceGroups(repoRoot string) ([]ReferenceGroupIssue, error) {
	cfg, err := LoadReferenceGroupConfig(repoRoot)
	if err != nil {
		return nil, err
	}
	skillsRoot := filepath.Join(repoRoot, devwikiSkillsRelativeRoot)
	var issues []ReferenceGroupIssue

	for _, groupName := range sortedReferenceGroupNames(cfg.Groups) {
		group := cfg.Groups[groupName]
		canonical := filepath.ToSlash(filepath.Clean(group.Canonical))
		if !referenceFileListContains(group.Files, canonical) {
			issues = append(issues, ReferenceGroupIssue{
				Group:     groupName,
				Canonical: group.Canonical,
				File:      group.Canonical,
				Reason:    "canonical is not listed in files",
			})
		}

		canonicalPath := filepath.Join(skillsRoot, filepath.FromSlash(canonical))
		canonicalData, readCanonicalErr := os.ReadFile(canonicalPath)
		if readCanonicalErr != nil {
			issues = append(issues, ReferenceGroupIssue{
				Group:     groupName,
				Canonical: group.Canonical,
				File:      group.Canonical,
				Reason:    "canonical file is missing or unreadable",
			})
		}

		for _, file := range group.Files {
			normalized := filepath.ToSlash(filepath.Clean(file))
			filePath := filepath.Join(skillsRoot, filepath.FromSlash(normalized))
			data, err := os.ReadFile(filePath)
			if err != nil {
				issues = append(issues, ReferenceGroupIssue{
					Group:     groupName,
					Canonical: group.Canonical,
					File:      file,
					Reason:    "file is missing or unreadable",
				})
				continue
			}
			if readCanonicalErr == nil && !bytes.Equal(data, canonicalData) {
				issues = append(issues, ReferenceGroupIssue{
					Group:     groupName,
					Canonical: group.Canonical,
					File:      file,
					Reason:    "content differs from canonical",
				})
			}
		}
	}

	skillIssues, err := checkSkillReferenceLinks(skillsRoot)
	if err != nil {
		return nil, err
	}
	issues = append(issues, skillIssues...)
	return issues, nil
}

// FixDevwikiReferenceGroups copies canonical reference content to out-of-sync files.
func FixDevwikiReferenceGroups(repoRoot string) ([]string, error) {
	cfg, err := LoadReferenceGroupConfig(repoRoot)
	if err != nil {
		return nil, err
	}
	skillsRoot := filepath.Join(repoRoot, devwikiSkillsRelativeRoot)
	var updated []string
	for _, groupName := range sortedReferenceGroupNames(cfg.Groups) {
		group := cfg.Groups[groupName]
		canonical := filepath.ToSlash(filepath.Clean(group.Canonical))
		if !referenceFileListContains(group.Files, canonical) {
			return nil, fmt.Errorf("reference group %q canonical %q is not listed in files", groupName, group.Canonical)
		}
		canonicalData, err := os.ReadFile(filepath.Join(skillsRoot, filepath.FromSlash(canonical)))
		if err != nil {
			return nil, fmt.Errorf("read canonical %s: %w", group.Canonical, err)
		}
		for _, file := range group.Files {
			normalized := filepath.ToSlash(filepath.Clean(file))
			if normalized == canonical {
				continue
			}
			filePath := filepath.Join(skillsRoot, filepath.FromSlash(normalized))
			current, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("read reference %s: %w", file, err)
			}
			if bytes.Equal(current, canonicalData) {
				continue
			}
			if err := os.WriteFile(filePath, canonicalData, 0o644); err != nil {
				return nil, fmt.Errorf("write reference %s: %w", file, err)
			}
			updated = append(updated, normalized)
		}
	}
	sort.Strings(updated)
	return updated, nil
}

// FormatReferenceGroupIssues renders issues for CLI and test diagnostics.
func FormatReferenceGroupIssues(issues []ReferenceGroupIssue) string {
	if len(issues) == 0 {
		return ""
	}
	lines := make([]string, 0, len(issues))
	for _, issue := range issues {
		lines = append(lines, fmt.Sprintf("%s: %s: %s (canonical: %s)", issue.Group, issue.File, issue.Reason, issue.Canonical))
	}
	return strings.Join(lines, "\n")
}

func checkSkillReferenceLinks(skillsRoot string) ([]ReferenceGroupIssue, error) {
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, err
	}
	var issues []ReferenceGroupIssue
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFile := filepath.Join(skillsRoot, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}
		seen := map[string]struct{}{}
		for _, match := range skillReferencePattern.FindAllString(string(data), -1) {
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			target := filepath.Join(skillsRoot, entry.Name(), filepath.FromSlash(match))
			if _, err := os.Stat(target); err != nil {
				issues = append(issues, ReferenceGroupIssue{
					Group:     "skill-references",
					Canonical: "",
					File:      filepath.ToSlash(filepath.Join(entry.Name(), match)),
					Reason:    "SKILL.md references missing file",
				})
			}
		}
	}
	return issues, nil
}

func sortedReferenceGroupNames(groups map[string]ReferenceGroup) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func referenceFileListContains(files []string, target string) bool {
	for _, file := range files {
		if filepath.ToSlash(filepath.Clean(file)) == target {
			return true
		}
	}
	return false
}
