package graph

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed web/index.html
var indexHTML string

//go:embed web/app.js
var appJS string

//go:embed web/style.css
var styleCSS string

//go:embed web/stats.html
var statsHTML string

//go:embed web/stats.js
var statsJS string

//go:embed web/stats.css
var statsCSS string

//go:embed web/wordcloud2.js
var wordcloud2JS string

//go:embed web/wordcloud2-LICENSE.txt
var wordcloud2License string

//go:embed web/cytoscape.min.js
var cytoscapeJS string

//go:embed web/cytoscape-LICENSE.txt
var cytoscapeLicense string

//go:embed web/vendor/vditor
var vendorFS embed.FS

// WriteAssets writes the static graph browser into an output directory.
func WriteAssets(outDir string) error {
	for _, asset := range webAssets() {
		rel := asset.outputPath
		path := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(asset.content), 0o644); err != nil {
			return err
		}
	}
	if err := writeEmbeddedDir(outDir, vendorFS, "web/vendor"); err != nil {
		return err
	}
	return nil
}

// AssetHash returns a content hash for every embedded web asset.
func AssetHash() (string, error) {
	hash := sha256.New()
	for _, asset := range webAssets() {
		hash.Write([]byte(asset.outputPath))
		hash.Write([]byte{0})
		hash.Write([]byte(asset.content))
		hash.Write([]byte{0})
	}
	var vendorFiles []string
	if err := fs.WalkDir(vendorFS, "web/vendor", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			vendorFiles = append(vendorFiles, path)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(vendorFiles)
	for _, path := range vendorFiles {
		data, err := vendorFS.ReadFile(path)
		if err != nil {
			return "", err
		}
		hash.Write([]byte(path))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type webAsset struct {
	outputPath string
	content    string
}

func webAssets() []webAsset {
	return []webAsset{
		{outputPath: "index.html", content: indexHTML},
		{outputPath: "stats.html", content: statsHTML},
		{outputPath: "assets/app.js", content: appJS},
		{outputPath: "assets/stats.js", content: statsJS},
		{outputPath: "assets/style.css", content: styleCSS},
		{outputPath: "assets/stats.css", content: statsCSS},
		{outputPath: "assets/wordcloud2.js", content: wordcloud2JS},
		{outputPath: "assets/wordcloud2-LICENSE.txt", content: wordcloud2License},
		{outputPath: "assets/cytoscape.min.js", content: cytoscapeJS},
		{outputPath: "assets/cytoscape-LICENSE.txt", content: cytoscapeLicense},
	}
}

func writeEmbeddedDir(outDir string, fsys embed.FS, root string) error {
	return fs.WalkDir(fsys, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, root+"/")
		target := filepath.Join(outDir, "assets", "vendor", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
