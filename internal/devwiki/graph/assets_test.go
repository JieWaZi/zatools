package graph

import (
	"strings"
	"testing"
)

func TestStaticAssetsRespectMarkdownPreviewContract(t *testing.T) {
	forbidden := []string{
		"eye_icon",
		"icon-eye",
		"👁",
		"previewCurrentIcon",
		"previewPrimaryBtn",
		"detailPath",
		"copyPath",
		"copyInline",
		"copy-path",
		"copy-small",
		"文件路径",
		"source: 'mock'",
	}
	assets := map[string]string{
		"index.html":       indexHTML,
		"assets/app.js":    appJS,
		"assets/style.css": styleCSS,
	}
	for name, content := range assets {
		for _, token := range forbidden {
			if strings.Contains(content, token) {
				t.Fatalf("%s contains forbidden token %q", name, token)
			}
		}
	}
	if !strings.Contains(indexHTML, `id="previewCurrentBtn"`) {
		t.Fatal("index.html missing previewCurrentBtn")
	}
	if !strings.Contains(indexHTML, `id="markdownModal"`) {
		t.Fatal("index.html missing markdownModal")
	}
	if !strings.Contains(appJS, "openMarkdownPreview") {
		t.Fatal("app.js missing openMarkdownPreview")
	}
	if strings.Contains(appJS, "const graphData") {
		t.Fatal("app.js should load generated graph.json instead of embedded graph data")
	}
}
