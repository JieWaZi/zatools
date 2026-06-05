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
	"time"

	"zatools/internal/devwiki/stats"
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

func TestServeReturnsRequestedHostInURL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url, err := serveHTTP(ctx, "localhost", 0, http.NewServeMux())
	if err != nil {
		t.Fatalf("serveHTTP() error = %v", err)
	}
	if !strings.HasPrefix(url, "http://localhost:") {
		t.Fatalf("url = %q, want requested host", url)
	}
	if strings.Contains(url, "[::]") {
		t.Fatalf("url should not expose wildcard listener address: %q", url)
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

func TestAPIHandlerServesDevwikiGlossaryKeywordsWithBasicAuth(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/glossary.md", `# Glossary

| glossary | type | description | slug |
|---|---|---|---|
| VIP | topic | VIP 业务规则入口 | vip |
| 网关配置 | workflow | HA 网关配置实现入口 | workflow-ha-gateway |
| VIP | workflow | 重复别名 | workflow-vip |
`)

	handler := APIHandler(root)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/glossary/keywords", bytes.NewBufferString(`{}`))
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
	if got.Text != "VIP\n网关配置\n" {
		t.Fatalf("glossary keywords response = %q", got.Text)
	}
}

func TestAPIHandlerRecordsSearchStats(t *testing.T) {
	root := t.TempDir()
	writeGraphFile(t, root, "wiki/index.md", `# Wiki Index

| type | description | slug |
|---|---|---|
| topic | VIP 业务规则入口 | vip |
`)

	ctx := context.Background()
	recorder := stats.NewRecorder(root)
	handler := APIHandlerWithContextAndRecorder(ctx, root, recorder)
	body := bytes.NewBufferString(`{"kind":"index","query":["VIP"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/search", body)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	deadline := time.After(2 * time.Second)
	logPath := filepath.Join(root, stats.DirName)
	for {
		matches, _ := filepath.Glob(filepath.Join(logPath, "queries-*.jsonl"))
		if len(matches) > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for search stats JSONL")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	recorder.Flush()

	summaryData, err := os.ReadFile(filepath.Join(root, stats.DirName, "summary.json"))
	if err != nil {
		t.Fatalf("ReadFile(summary.json) error = %v", err)
	}
	if !strings.Contains(string(summaryData), `"today_search_count": 1`) {
		t.Fatalf("summary = %s", string(summaryData))
	}
}

func TestAPIHandlerRecordsReadStatsAfterFlush(t *testing.T) {
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

	ctx := context.Background()
	recorder := stats.NewRecorder(root)
	handler := APIHandlerWithContextAndRecorder(ctx, root, recorder)
	body := bytes.NewBufferString(`{"kind":"topic","slug":"vip","view":"card"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devwiki/read", body)
	req.SetBasicAuth(DefaultAPIUsername, DefaultAPIPassword)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	recorder.Flush()

	documentsData, err := os.ReadFile(filepath.Join(root, stats.DirName, "documents.json"))
	if err != nil {
		t.Fatalf("ReadFile(documents.json) error = %v", err)
	}
	if !strings.Contains(string(documentsData), `"read_count": 1`) {
		t.Fatalf("documents = %s", string(documentsData))
	}
}

func TestGraphHandlerServesStatsSummary(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, ".devwiki", "graph")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	statsDir := filepath.Join(root, stats.DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	if err := os.WriteFile(filepath.Join(statsDir, "summary.json"), []byte(`{
  "updated_at": "2026-06-05T10:00:00Z",
  "today": "`+today+`",
  "today_search_count": 7,
  "total_search_count": 70,
  "total_read_count": 12
}
`), 0o644); err != nil {
		t.Fatalf("WriteFile(summary.json) error = %v", err)
	}
	logPath := filepath.Join(statsDir, "queries-"+today+".jsonl")
	file, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	encoder := json.NewEncoder(file)
	for i := 0; i < 3; i++ {
		if err := encoder.Encode(stats.Event{
			Timestamp:   now,
			Endpoint:    "search",
			Kind:        "index",
			Queries:     []string{"VIP"},
			ResultCount: 1,
		}); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}
	if err := encoder.Encode(stats.Event{
		Timestamp: now,
		Endpoint:  "read",
		Kind:      "topic",
		Slug:      "vip",
		View:      "card",
	}); err != nil {
		t.Fatalf("Encode(read) error = %v", err)
	}
	file.Close()

	handler := graphHandler(ServerOptions{Dir: outDir, Root: root})
	req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"today_search_count":3`) {
		t.Fatalf("body = %s", body)
	}
	if !strings.Contains(body, `"today_read_count":1`) || !strings.Contains(body, `"today_api_count":4`) {
		t.Fatalf("body = %s", body)
	}
	if !strings.Contains(body, `"total_search_count":70`) || !strings.Contains(body, `"total_api_count":82`) {
		t.Fatalf("body = %s", body)
	}
}

func TestGraphHandlerDisablesBrowserCacheForStaticWebAssets(t *testing.T) {
	root := t.TempDir()
	outDir := filepath.Join(root, ".devwiki", "graph")
	if err := os.MkdirAll(filepath.Join(outDir, "assets"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("WriteFile(index.html) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "assets", "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("WriteFile(app.js) error = %v", err)
	}

	handler := graphHandler(ServerOptions{Dir: outDir, Root: root})
	for _, path := range []string{"/index.html", "/assets/app.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200; body = %s", path, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("%s Cache-Control = %q, want no-store", path, got)
		}
	}
}

func TestServeAPIStartsKeywordUpdates(t *testing.T) {
	root := t.TempDir()
	statsDir := filepath.Join(root, stats.DirName)
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	now := time.Now()
	logPath := filepath.Join(statsDir, "queries-"+now.Format("2006-01-02")+".jsonl")
	file, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := json.NewEncoder(file).Encode(stats.Event{
		Timestamp:   now,
		Endpoint:    "search",
		Kind:        "topic",
		Queries:     []string{"VIP"},
		ResultCount: 1,
	}); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := ServeAPI(ctx, ServerOptions{Root: root, Host: "127.0.0.1", Port: 0}); err != nil {
		t.Fatalf("ServeAPI() error = %v", err)
	}

	keywordsPath := filepath.Join(statsDir, "keywords.json")
	deadline := time.After(2 * time.Second)
	for {
		data, err := os.ReadFile(keywordsPath)
		if err == nil && strings.Contains(string(data), `"VIP"`) {
			return
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for server keyword update")
		default:
			time.Sleep(10 * time.Millisecond)
		}
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
