package devwiki

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestDevwikiReferenceGroupsAreConsistent(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	issues, err := CheckDevwikiReferenceGroups(repoRoot)
	if err != nil {
		t.Fatalf("CheckDevwikiReferenceGroups error = %v", err)
	}
	if len(issues) > 0 {
		t.Fatalf("reference group issues:\n%s", FormatReferenceGroupIssues(issues))
	}
}

func TestDevwikiSkillReferencesAreMinimal(t *testing.T) {
	t.Parallel()

	skillsRoot := filepath.Clean(filepath.Join("..", "..", "skills", "devwiki"))
	expected := map[string][]string{
		"code-to-doc": {
			"code-tracing.md",
			"common-file-format.md",
			"evidence-grounding.md",
			"mutation-safety.md",
			"zatools-devwiki.md",
			"zatools-qmd.md",
		},
		"ingest": {
			"code-tracing.md",
			"common-file-format.md",
			"evidence-grounding.md",
			"knowledge-placement.md",
			"mutation-safety.md",
			"zatools-devwiki.md",
			"zatools-qmd.md",
		},
		"query": {
			"zatools-devwiki.md",
			"zatools-qmd.md",
		},
		"topic": {
			"common-file-format.md",
			"evidence-grounding.md",
			"knowledge-placement.md",
			"mutation-safety.md",
			"topic_template.md",
			"zatools-devwiki.md",
		},
		"workflow": {
			"code-tracing.md",
			"common-file-format.md",
			"evidence-grounding.md",
			"knowledge-placement.md",
			"mutation-safety.md",
			"workflow_template.md",
			"zatools-devwiki.md",
		},
	}

	for skill, want := range expected {
		entries, err := os.ReadDir(filepath.Join(skillsRoot, skill, "references"))
		if err != nil {
			t.Fatalf("ReadDir(%s references) error = %v", skill, err)
		}
		var got []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			got = append(got, entry.Name())
		}
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s references = %#v, want %#v", skill, got, want)
		}
	}
}
