package devwikiapp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"zatools/internal/skills"
)

func useLocalDevwikiSkills(t *testing.T, service *Service) skills.Source {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", "..", "..", "skills", "devwiki"))
	if _, err := os.Stat(filepath.Join(root, "query", "SKILL.md")); err != nil {
		t.Fatalf("missing root DevWiki skills: %v", err)
	}
	source := skills.NewDevwikiSkillsSource("")
	service.devwikiSkillsResolver = func(ctx context.Context) (devwikiSkillsBundle, error) {
		_ = ctx
		found, err := skills.Discover(root)
		if err != nil {
			return devwikiSkillsBundle{}, err
		}
		return devwikiSkillsBundle{
			source: source,
			skills: found,
		}, nil
	}
	return source
}

func rootDevwikiSkills(t *testing.T) []skills.Skill {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", "..", "..", "skills", "devwiki"))
	found, err := skills.Discover(root)
	if err != nil {
		t.Fatalf("Discover root DevWiki skills error = %v", err)
	}
	return found
}
