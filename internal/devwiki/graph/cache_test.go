package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestBuildManifestIncludesStaticAssetHash(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/a.md", "---\ntitle: A\nslug: a\nsummary: A\n---\n")
	manifest, err := BuildManifest(root)
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Marshal(manifest) error = %v", err)
	}
	if !strings.Contains(string(data), `"asset_hash"`) {
		t.Fatalf("manifest JSON should include asset_hash: %s", string(data))
	}

	current := manifest
	encoded, err := json.Marshal(current)
	if err != nil {
		t.Fatalf("Marshal(current) error = %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal(current) error = %v", err)
	}
	raw["asset_hash"] = "changed"
	mutated, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("Marshal(mutated) error = %v", err)
	}
	var changed Manifest
	if err := json.Unmarshal(mutated, &changed); err != nil {
		t.Fatalf("Unmarshal(mutated) error = %v", err)
	}
	if current.IsFresh(changed) {
		t.Fatal("manifest should not be fresh when embedded static assets change")
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
	for _, rel := range []string{
		"graph.json",
		"manifest.json",
		"index.html",
		"stats.html",
		"assets/app.js",
		"assets/stats.js",
		"assets/style.css",
		"assets/stats.css",
		"assets/wordcloud2.js",
		"assets/wordcloud2-LICENSE.txt",
		"assets/cytoscape.min.js",
		"assets/cytoscape-LICENSE.txt",
		"assets/vendor/vditor/dist/index.css",
		"assets/vendor/vditor/dist/index.min.js",
		"assets/vendor/vditor/dist/js/lute/lute.min.js",
		"assets/vendor/vditor/dist/js/mermaid/mermaid.min.js",
		"assets/vendor/vditor/dist/js/highlight.js/highlight.min.js",
		"assets/vendor/vditor/LICENSE",
	} {
		if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing output %s: %v", rel, err)
		}
	}
}
