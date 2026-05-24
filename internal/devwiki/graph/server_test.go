package graph

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphHandlerServesGeneratedGraphAndRootMarkdown(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, ".devwiki", "graph")
	writeGraphFile(t, root, "wiki/topics/vip.md", "# VIP\n\n真实 Markdown 内容\n")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "graph.json"), []byte(`{"nodes":[]}`), 0o644); err != nil {
		t.Fatalf("WriteFile(graph.json) error = %v", err)
	}

	handler := graphHandler(ServerOptions{Dir: outDir, Root: root})
	assertGraphHandlerBody(t, handler, "/graph.json", `{"nodes":[]}`)
	assertGraphHandlerBody(t, handler, "/wiki/topics/vip.md", "真实 Markdown 内容")
}

func assertGraphHandlerBody(t *testing.T, handler http.Handler, path string, want string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want 200; body = %s", path, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("%s body = %q, want substring %q", path, rec.Body.String(), want)
	}
}
