package devwikiapp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/skills"
)

func TestGraphCheckDoesNotWriteOutputs(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "config", "project.yaml"), "project_name: Sample\nproject_slug: sample\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "features", "vip.md"), "---\ntitle: VIP\nslug: vip\nsummary: VIP\n---\n")
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Graph(context.Background(), GraphOptions{Root: root, Check: true, Stdout: &out})
	if err != nil {
		t.Fatalf("Graph(check) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".devwiki", "graph")); !os.IsNotExist(err) {
		t.Fatalf("check mode wrote output dir: %v", err)
	}
	if !strings.Contains(out.String(), "DevWiki graph check passed") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestGraphBuildWritesOutputsWithoutServingWhenNoServe(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "config", "project.yaml"), "project_name: Sample\nproject_slug: sample\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "features", "vip.md"), "---\ntitle: VIP\nslug: vip\nsummary: VIP\n---\n")
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Graph(context.Background(), GraphOptions{Root: root, NoServe: true, NoOpen: true, Stdout: &out})
	if err != nil {
		t.Fatalf("Graph(build) error = %v", err)
	}
	for _, rel := range []string{".devwiki/graph/graph.json", ".devwiki/graph/manifest.json", ".devwiki/graph/index.html"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
}
