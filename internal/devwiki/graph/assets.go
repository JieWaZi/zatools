package graph

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed static/index.html
var indexHTML string

//go:embed static/app.js
var appJS string

//go:embed static/style.css
var styleCSS string

//go:embed static/cytoscape.min.js
var cytoscapeJS string

//go:embed static/cytoscape-LICENSE.txt
var cytoscapeLicense string

// WriteAssets writes the static graph browser into an output directory.
func WriteAssets(outDir string) error {
	files := map[string]string{
		"index.html":                   indexHTML,
		"assets/app.js":                appJS,
		"assets/style.css":             styleCSS,
		"assets/cytoscape.min.js":      cytoscapeJS,
		"assets/cytoscape-LICENSE.txt": cytoscapeLicense,
	}
	for rel, content := range files {
		path := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
