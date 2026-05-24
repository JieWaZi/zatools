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
		"marked",
		"DOMPurify",
		"fallbackRenderMarkdown",
		"function renderMarkdown(",
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
	if !strings.Contains(indexHTML, `/assets/vendor/vditor/dist/index.css`) {
		t.Fatal("index.html missing local Vditor stylesheet")
	}
	if !strings.Contains(indexHTML, `/assets/vendor/vditor/dist/index.min.js`) {
		t.Fatal("index.html missing local Vditor script")
	}
	if !strings.Contains(appJS, "openMarkdownPreview") {
		t.Fatal("app.js missing openMarkdownPreview")
	}
	if !strings.Contains(appJS, "Vditor.preview") {
		t.Fatal("app.js should render Markdown through Vditor.preview")
	}
	if !strings.Contains(appJS, "cdn: '/assets/vendor/vditor'") {
		t.Fatal("app.js should point Vditor at local vendor assets")
	}
	if strings.Contains(appJS, "const graphData") {
		t.Fatal("app.js should load generated graph.json instead of embedded graph data")
	}
	if !strings.Contains(appJS, "collectSearchVisibleNodeIDs") {
		t.Fatal("app.js should expand search results to directly related nodes")
	}
	if !strings.Contains(appJS, "search-match") {
		t.Fatal("app.js should mark nodes that directly match the search keyword")
	}
	if !strings.Contains(appJS, "search-related") {
		t.Fatal("app.js should mark nodes related to search matches")
	}
	if strings.Contains(indexHTML, `id="depthSelect"`) {
		t.Fatal("index.html should not render relation depth selector")
	}
	if strings.Contains(indexHTML, `id="layoutSelect"`) {
		t.Fatal("index.html should not render layout selector")
	}
	if !strings.Contains(indexHTML, `<option value="topic" selected>Topic</option>`) {
		t.Fatal("index.html should default the dimension selector to Topic")
	}
	if strings.Contains(indexHTML, `<option value="all"`) {
		t.Fatal("index.html should not render an all-dimensions option")
	}
	if strings.Contains(indexHTML, `<option value="workflow"`) {
		t.Fatal("index.html should not render Workflow as a selectable dimension")
	}
	if !strings.Contains(appJS, "let currentFilter = 'topic'") {
		t.Fatal("app.js should default dimension filtering to Topic")
	}
	if strings.Contains(appJS, "currentFilter === 'all'") {
		t.Fatal("app.js should not keep all-dimensions filtering logic")
	}
	if strings.Contains(appJS, "currentDepth") {
		t.Fatal("app.js should not keep relation depth switching state")
	}
	if strings.Contains(appJS, "runPresetLayout") {
		t.Fatal("app.js should not keep preset layout switching")
	}
	if !strings.Contains(appJS, "layout: forceLayoutOptions(false)") {
		t.Fatal("app.js should initialize with force-directed layout")
	}
	if !strings.Contains(styleCSS, "width: min(90vw, calc(100vw - 48px))") {
		t.Fatal("style.css should size Markdown preview close to full screen width with margins")
	}
	if !strings.Contains(styleCSS, "height: min(90vh, calc(100vh - 48px))") {
		t.Fatal("style.css should size Markdown preview close to full screen height with margins")
	}
	if !strings.Contains(styleCSS, "justify-content: center") {
		t.Fatal("style.css should center Markdown preview horizontally")
	}
	if !strings.Contains(appJS, "markdown-preview-inner") {
		t.Fatal("app.js should render Markdown into an inner preview container")
	}
	if !strings.Contains(styleCSS, ".markdown-preview-inner") {
		t.Fatal("style.css should style the Markdown preview inner container")
	}
	if !strings.Contains(styleCSS, "max-width: 1080px") {
		t.Fatal("style.css should constrain Markdown preview content to a readable width")
	}
	if !strings.Contains(styleCSS, "padding: clamp(24px, 4vw, 48px)") {
		t.Fatal("style.css should keep Markdown content away from the dialog edge")
	}
	forbiddenVisibleSlug := map[string]string{
		"index.html":    `>Slug<`,
		"app.js label":  "+ '\\n' + node.slug",
		"app.js detail": "detailSlug",
		"app.js list":   "related-slug",
		"style.css":     ".related-slug",
	}
	for name, token := range forbiddenVisibleSlug {
		content := appJS
		if name == "index.html" {
			content = indexHTML
		}
		if name == "style.css" {
			content = styleCSS
		}
		if strings.Contains(content, token) {
			t.Fatalf("%s should not visibly render slug token %q", name, token)
		}
	}
}
