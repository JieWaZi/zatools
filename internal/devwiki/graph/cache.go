package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Manifest records the graph input hash and output locations.
type Manifest struct {
	SchemaVersion  int            `json:"schema_version"`
	BuilderVersion int            `json:"builder_version"`
	InputHash      string         `json:"input_hash"`
	AssetHash      string         `json:"asset_hash"`
	BuiltAt        time.Time      `json:"built_at"`
	Files          []ManifestFile `json:"files"`
	Outputs        ManifestOutput `json:"outputs"`
}

// ManifestFile records one graph input file's content identity.
type ManifestFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// ManifestOutput records generated graph output paths.
type ManifestOutput struct {
	Graph string `json:"graph"`
	Index string `json:"index"`
	Stats string `json:"stats"`
}

// IsFresh reports whether the old manifest is fresh for the current inputs.
func (m Manifest) IsFresh(current Manifest) bool {
	return m.SchemaVersion == current.SchemaVersion &&
		m.BuilderVersion == current.BuilderVersion &&
		m.InputHash == current.InputHash &&
		m.AssetHash == current.AssetHash
}

// BuildManifest computes the current graph input manifest for a DevWiki root.
func BuildManifest(root string) (Manifest, error) {
	files, err := inputFiles(root)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		SchemaVersion:  SchemaVersion,
		BuilderVersion: BuilderVersion,
		BuiltAt:        time.Now(),
		Outputs: ManifestOutput{
			Graph: ".devwiki/graph/graph.json",
			Index: ".devwiki/graph/index.html",
			Stats: ".devwiki/graph/stats.html",
		},
	}
	assetHash, err := AssetHash()
	if err != nil {
		return Manifest{}, err
	}
	manifest.AssetHash = assetHash
	hash := sha256.New()
	for _, rel := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		data, err := os.ReadFile(path)
		if err != nil {
			return Manifest{}, err
		}
		sum := sha256.Sum256(data)
		info, err := os.Stat(path)
		if err != nil {
			return Manifest{}, err
		}
		file := ManifestFile{Path: rel, Size: info.Size(), SHA256: hex.EncodeToString(sum[:])}
		manifest.Files = append(manifest.Files, file)
		hash.Write([]byte(file.Path))
		hash.Write([]byte(file.SHA256))
	}
	manifest.InputHash = hex.EncodeToString(hash.Sum(nil))
	return manifest, nil
}

// ReadManifest reads a JSON manifest from disk.
func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// WriteOutputs writes graph JSON, manifest JSON, and static frontend assets.
func WriteOutputs(outDir string, graph Graph, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		return err
	}
	graphData, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "graph.json"), graphData, 0o644); err != nil {
		return err
	}
	manifest.BuiltAt = time.Now()
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "manifest.json"), manifestData, 0o644); err != nil {
		return err
	}
	return WriteAssets(outDir)
}

func inputFiles(root string) ([]string, error) {
	var files []string
	for _, dir := range []string{"wiki/topics", "wiki/workflows"} {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				files = append(files, filepath.ToSlash(filepath.Join(dir, entry.Name())))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}
