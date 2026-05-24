package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildManifestChangesWhenInputChanges(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/a.md", "---\ntitle: A\nslug: a\nsummary: A\n---\n")
	first, err := BuildManifest(root)
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	writeGraphFile(t, root, "wiki/topics/a.md", "---\ntitle: A\nslug: a\nsummary: changed\n---\n")
	second, err := BuildManifest(root)
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	if first.InputHash == second.InputHash {
		t.Fatalf("InputHash did not change: %s", first.InputHash)
	}
}

func TestManifestFreshness(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/a.md", "---\ntitle: A\nslug: a\nsummary: A\n---\n")
	manifest, err := BuildManifest(root)
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	if !manifest.IsFresh(manifest) {
		t.Fatal("manifest should be fresh against itself")
	}
	changed := manifest
	changed.BuilderVersion++
	if manifest.IsFresh(changed) {
		t.Fatal("manifest should not be fresh when builder version changes")
	}
}

func TestWriteOutputsCreatesGraphFiles(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, ".devwiki", "graph")
	graph := Graph{SchemaVersion: SchemaVersion, Project: Project{Name: "Sample", Slug: "sample", Root: root}, Documents: map[string]Document{}}
	manifest := Manifest{SchemaVersion: SchemaVersion, BuilderVersion: BuilderVersion, InputHash: "abc"}
	if err := WriteOutputs(outDir, graph, manifest); err != nil {
		t.Fatalf("WriteOutputs() error = %v", err)
	}
	for _, rel := range []string{"graph.json", "manifest.json", "index.html", "assets/app.js", "assets/style.css", "assets/cytoscape.min.js", "assets/cytoscape-LICENSE.txt"} {
		if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing output %s: %v", rel, err)
		}
	}
}
