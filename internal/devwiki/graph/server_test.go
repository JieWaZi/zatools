package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"zatools/internal/qmd"
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

func TestGraphHandlerDoesNotServeDevwikiAPI(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, ".devwiki", "graph")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	handler := graphHandler(ServerOptions{Dir: outDir, Root: root})
	req := httptest.NewRequest(http.MethodGet, "/api/devwiki/project", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("graph handler should not serve DevWiki API; body = %s", rec.Body.String())
	}
}

func TestAPIHandlerRequiresBasicAuth(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Huawei ZDDI\nproject_slug: huawei-zddi\nlanguage: zh\n")

	handler := APIHandler(root)
	req := httptest.NewRequest(http.MethodGet, "/api/devwiki/project", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("missing WWW-Authenticate header")
	}
}

func TestAPIHandlerServesDevwikiReadEndpointWithBasicAuth(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/topics/vip.md", `---
title: VIP
slug: vip
kind: topic
status: active
summary: VIP topic
formatter: markdown
confidence: high
---
# VIP

<!-- devwiki:section id=card -->
card body
<!-- /devwiki:section -->
`)

	handler := APIHandler(root)
	body := bytes.NewBufferString(`{"kind":"topic","slug":"vip","view":"card"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/read", body)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal response error = %v, body=%q", err, rec.Body.String())
	}
	for _, want := range []string{"title: VIP", "status: active", "summary: VIP topic", "confidence: high", "card body"} {
		if !strings.Contains(got.Text, want) {
			t.Fatalf("read response missing %q: %q", want, got.Text)
		}
	}
	if strings.Contains(got.Text, "formatter") || strings.Contains(got.Text, "slug:") {
		t.Fatalf("read response leaked non-card metadata: %q", got.Text)
	}
}

func TestAPIHandlerServesDevwikiProjectEndpointWithBasicAuth(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "config/project.yaml", "project_name: Huawei ZDDI\nproject_slug: huawei-zddi\nlanguage: zh\n")

	handler := APIHandler(root)
	req := httptest.NewRequest(http.MethodGet, "/api/devwiki/project", nil)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal response error = %v, body=%q", err, rec.Body.String())
	}
	if got["project_slug"] != "huawei-zddi" || got["project_name"] != "Huawei ZDDI" || got["language"] != "zh" {
		t.Fatalf("project response = %#v", got)
	}
}

func TestAPIHandlerServesDevwikiSearchIndexEndpointWithBasicAuth(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/index.md", `# Wiki Index

| type | description | slug |
|---|---|---|
| topic | VIP 业务规则入口 | vip |
| workflow | HA 网关配置实现入口 | workflow-ha-gateway |
`)

	handler := APIHandler(root)
	body := bytes.NewBufferString(`{"kind":"index","query":["VIP"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/search", body)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var got []map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal response error = %v, body=%q", err, rec.Body.String())
	}
	if len(got) != 1 || got[0]["slug"] != "vip" || got[0]["type"] != "topic" {
		t.Fatalf("search response = %#v", got)
	}
}

func TestAPIHandlerServesDevwikiSearchWorkflowWithSharedQMDSearch(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/workflows/workflow-ha-brain-split-protection.md", `---
title: "HA 脑裂监控与防护实现定位"
slug: workflow-ha-brain-split-protection
kind: workflow
summary: "HA workflow"
---
# HA 脑裂监控与防护实现定位
`)
	writeGraphFile(t, root, "wiki/workflows/workflow-ha-gateway-config.md", `---
title: "HA 网关配置实现定位"
slug: workflow-ha-gateway-config
kind: workflow
summary: "Gateway workflow"
---
# HA 网关配置实现定位
`)

	ctx := qmd.WithCommandRunner(context.Background(), devwikiGraphSearchQMDHelperRunner(t, root))
	handler := APIHandlerWithContext(ctx, root)
	body := bytes.NewBufferString(`{"kind":"workflow","query":["防脑裂","网关"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/search", body)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	var got []map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal response error = %v, body=%q", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("search response = %#v, want 2 results", got)
	}
	if got[0]["slug"] != "workflow-ha-brain-split-protection" || got[0]["score"] != "100%" {
		t.Fatalf("first result = %#v", got[0])
	}
	if got[1]["slug"] != "workflow-ha-gateway-config" || got[1]["score"] != "49%" {
		t.Fatalf("second result = %#v", got[1])
	}
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

func devwikiGraphSearchQMDHelperRunner(t *testing.T, root string) func(context.Context, string, ...string) *exec.Cmd {
	t.Helper()

	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		commandArgs := []string{"-test.run=TestDevwikiGraphSearchQMDHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_DEVWIKI_GRAPH_SEARCH_QMD_HELPER=1")
		cmd.Dir = root
		return cmd
	}
}

func TestDevwikiGraphSearchQMDHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_DEVWIKI_GRAPH_SEARCH_QMD_HELPER") != "1" {
		return
	}

	args := os.Args
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	if len(args) < 3 || args[0] != "qmd" || args[1] != "search" {
		_, _ = os.Stderr.WriteString("unexpected argv=" + strings.Join(args, " ") + "\n")
		os.Exit(2)
	}
	switch args[2] {
	case "防脑裂":
		_, _ = os.Stdout.WriteString(`workflows/workflow-ha-brain-split-protection.md:2 #b46301
Title: HA 脑裂监控与防护实现定位
Score:  83%
`)
	case "网关":
		_, _ = os.Stdout.WriteString(`workflows/workflow-ha-brain-split-protection.md:2 #b46301
Title: HA 脑裂监控与防护实现定位
Score:  83%

workflows/workflow-ha-gateway-config.md:8 #ff0000
Title: HA 网关配置实现定位
Score:  81%
`)
	default:
		_, _ = os.Stderr.WriteString("unexpected query=" + args[2] + "\n")
		os.Exit(2)
	}
	os.Exit(0)
}
