package devwikiapp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	common "zatools/internal/app/common"
	"zatools/internal/devwiki/retrieval"
	"zatools/internal/qmd"
	"zatools/internal/skills"
)

func TestSearchTopicFiltersQMDOutputAndWritesTable(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/topics/ha-brain-split-protection.md", `---
title: "HA 脑裂监控与防护"
slug: ha-brain-split-protection
kind: topic
summary: "HA brain split"
---
# HA 脑裂监控与防护
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	ctx := qmd.WithCommandRunner(context.Background(), devwikiSearchQMDHelperRunner(t, root))
	var out bytes.Buffer

	err := service.Search(ctx, SearchOptions{
		Root:   root,
		Kind:   "topic",
		Query:  "脑裂",
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	want := "|file|slug|title|score|\n|ha-brain-split-protection.md|ha-brain-split-protection|HA 脑裂监控与防护|84%|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
}

func TestSearchRunsEachQueryTermAndFusesDuplicateResults(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/workflows/workflow-ha-brain-split-protection.md", `---
title: "HA 脑裂监控与防护实现定位"
slug: workflow-ha-brain-split-protection
kind: workflow
summary: "HA workflow"
---
# HA 脑裂监控与防护实现定位
`)
	writeDevwikiReadFixture(t, root, "wiki/workflows/workflow-ha-gateway-config.md", `---
title: "HA 网关配置实现定位"
slug: workflow-ha-gateway-config
kind: workflow
summary: "Gateway workflow"
---
# HA 网关配置实现定位
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	ctx := qmd.WithCommandRunner(context.Background(), devwikiSearchQMDHelperRunner(t, root))
	var out bytes.Buffer

	err := service.Search(ctx, SearchOptions{
		Root:       root,
		Kind:       "workflow",
		QueryTerms: []string{"防脑裂", "网关"},
		Stdout:     &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	want := "|file|slug|title|score|\n" +
		"|workflow-ha-brain-split-protection.md|workflow-ha-brain-split-protection|HA 脑裂监控与防护实现定位|100%|\n" +
		"|workflow-ha-gateway-config.md|workflow-ha-gateway-config|HA 网关配置实现定位|49%|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
}

func TestSearchRunsQueryTermsSerially(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/workflows/workflow-ha-brain-split-protection.md", `---
title: "HA 脑裂监控与防护实现定位"
slug: workflow-ha-brain-split-protection
kind: workflow
summary: "HA workflow"
---
# HA 脑裂监控与防护实现定位
`)
	writeDevwikiReadFixture(t, root, "wiki/workflows/workflow-ha-gateway-config.md", `---
title: "HA 网关配置实现定位"
slug: workflow-ha-gateway-config
kind: workflow
summary: "Gateway workflow"
---
# HA 网关配置实现定位
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var (
		mu        sync.Mutex
		started   int
		active    int
		maxActive int
		queries   []string
	)
	runner := devwikiSearchQMDHelperRunner(t, root)
	ctx := qmd.WithCommandRunner(context.Background(), func(ctx context.Context, name string, args ...string) *exec.Cmd {
		mu.Lock()
		started++
		active++
		if active > maxActive {
			maxActive = active
		}
		if len(args) >= 2 && args[0] == "search" {
			queries = append(queries, args[1])
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		active--
		mu.Unlock()
		return runner(ctx, name, args...)
	})
	var out bytes.Buffer

	err := service.Search(ctx, SearchOptions{
		Root:       root,
		Kind:       "workflow",
		QueryTerms: []string{"防脑裂", "网关"},
		Stdout:     &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	mu.Lock()
	gotStarted := started
	gotMaxActive := maxActive
	gotQueries := append([]string(nil), queries...)
	mu.Unlock()
	if gotStarted != 2 {
		t.Fatalf("started qmd searches = %d, want 2", gotStarted)
	}
	if gotMaxActive != 1 {
		t.Fatalf("qmd searches ran concurrently, max active = %d, want 1", gotMaxActive)
	}
	wantQueries := []string{"防脑裂", "网关"}
	if !reflect.DeepEqual(gotQueries, wantQueries) {
		t.Fatalf("qmd search queries = %#v, want %#v", gotQueries, wantQueries)
	}

	if gotRows := strings.Count(strings.TrimSpace(out.String()), "\n"); gotRows != 2 {
		t.Fatalf("fused output row count = %d, want 2 data rows plus header; output=%q", gotRows, out.String())
	}
}

func TestSearchIndexParsesTableAndWritesPipeTable(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/index.md", `# Wiki Index

| type | description | slug |
|---|---|---|
| topic | HA 脑裂监控与防护的业务规则入口 | ha-brain-split-protection |
| workflow | HA 网关配置的实现定位入口 | workflow-ha-gateway-config |
| workflow | missing slug row | |
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer

	err := service.Search(context.Background(), SearchOptions{
		Root:       root,
		Kind:       "index",
		QueryTerms: []string{"脑裂", "missing"},
		Stdout:     &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	want := "|type|description|slug|\n|topic|HA 脑裂监控与防护的业务规则入口|ha-brain-split-protection|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
	if strings.Contains(out.String(), "score") {
		t.Fatalf("index search output should not include score: %s", out.String())
	}
}

func TestSearchGlossaryParsesTableAndWritesPipeTable(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/glossary.md", `# Glossary

| glossary | type | description | slug |
|---|---|---|---|
| 脑裂 | topic | HA 集群节点互相误判时的隔离与恢复规则 | ha-brain-split-protection |
| 网关配置 | workflow | HA 网关配置下发和持久化实现链路 | workflow-ha-gateway-config |
| 无效术语 | topic | missing slug | |
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer

	err := service.Search(context.Background(), SearchOptions{
		Root:       root,
		Kind:       "glossary",
		QueryTerms: []string{"持久化"},
		Stdout:     &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	want := "|glossary|type|description|slug|\n|网关配置|workflow|HA 网关配置下发和持久化实现链路|workflow-ha-gateway-config|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
	if strings.Contains(out.String(), "score") {
		t.Fatalf("glossary search output should not include score: %s", out.String())
	}
}

func TestSearchIndexWritesHeaderOnlyTableWhenNoRowsMatch(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/index.md", `# Wiki Index

| type | description | slug |
|---|---|---|
| topic | HA 脑裂监控与防护的业务规则入口 | ha-brain-split-protection |
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer

	err := service.Search(context.Background(), SearchOptions{
		Root:   root,
		Kind:   "index",
		Query:  "不存在",
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	want := "|type|description|slug|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
}

func TestGlossaryKeywordsWritesUniqueTermsAsLines(t *testing.T) {
	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/glossary.md", `# Glossary

| glossary | type | description | slug |
|---|---|---|---|
| 脑裂 | topic | HA 集群节点互相误判时的隔离与恢复规则 | ha-brain-split-protection |
| 网关配置 | workflow | HA 网关配置下发和持久化实现链路 | workflow-ha-gateway-config |
| 脑裂 | workflow | 重复别名应去重 | workflow-ha-brain-split-protection |
| 无效术语 | topic | missing slug | |
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer

	err := service.GlossaryKeywords(context.Background(), GlossaryKeywordsOptions{
		Root:   root,
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("GlossaryKeywords() error = %v", err)
	}
	want := "脑裂\n网关配置\n"
	if out.String() != want {
		t.Fatalf("GlossaryKeywords() output = %q, want %q", out.String(), want)
	}
}

func TestParseQMDSearchOutputKeepsOnlyRequestedKind(t *testing.T) {
	input := `qmd://devwiki-huawei-zddi-wiki/topics/ha-brain-split-protection.md:2 #be1507
Title: HA 脑裂监控与防护
Score:  84%

qmd://devwiki-huawei-zddi-wiki/workflows/ha-brain-split-handler.md:9 #abc123
Title: HA 脑裂处理链路
Score:  79%

wiki/topics/cluster-quorum.md:1
Score: 90%
`

	got := retrieval.ParseQMDSearchOutput(input, "workflow")
	want := []SearchResult{{File: "ha-brain-split-handler.md", Title: "HA 脑裂处理链路", Score: "79%"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("workflow results = %#v, want %#v", got, want)
	}
}

func TestParseQMDSearchFileLineAcceptsQMDURICollectionPaths(t *testing.T) {
	got, ok := retrieval.ParseQMDSearchFileLine("qmd://devwiki-huawei-zddi-wiki/workflows/workflow-ha-brain-split-protection.md:2 #b46301", "workflows")
	if !ok {
		t.Fatal("parseQMDSearchFileLine did not accept qmd URI workflow path")
	}
	if got != "workflow-ha-brain-split-protection.md" {
		t.Fatalf("file = %q, want workflow-ha-brain-split-protection.md", got)
	}
}

func TestSearchWritesHeaderOnlyTableWhenNoHitsMatchKind(t *testing.T) {
	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	ctx := qmd.WithCommandRunner(context.Background(), devwikiSearchQMDHelperRunner(t, root))
	var out bytes.Buffer

	err := service.Search(ctx, SearchOptions{
		Root:   root,
		Kind:   "topic",
		Query:  "workflow-only",
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	want := "|file|slug|title|score|\n"
	if out.String() != want {
		t.Fatalf("Search() output = %q, want %q", out.String(), want)
	}
}

func TestSearchRejectsUnsupportedKind(t *testing.T) {
	root := t.TempDir()
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Search(context.Background(), SearchOptions{
		Root:   root,
		Kind:   "module",
		Query:  "脑裂",
		Stdout: &out,
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported devwiki search kind") {
		t.Fatalf("Search() error = %v, want unsupported kind", err)
	}
}

func TestSearchUsesLocalProjectConfigWhenProjectProvided(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	root := t.TempDir()
	writeDevwikiReadFixture(t, root, "wiki/index.md", `# Wiki Index

| type | description | slug |
|---|---|---|
| topic | VIP 业务规则入口 | vip |
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(t.TempDir())})
	var out bytes.Buffer

	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		LocalPath:   root,
	}); err != nil {
		t.Fatalf("RepoAdd() error = %v", err)
	}
	if err := service.Search(context.Background(), SearchOptions{
		Project: "sample",
		Kind:    "index",
		Query:   "VIP",
		Stdout:  &out,
	}); err != nil {
		t.Fatalf("Search(project) error = %v", err)
	}
	want := "|type|description|slug|\n|topic|VIP 业务规则入口|vip|\n"
	if out.String() != want {
		t.Fatalf("Search(project) output = %q, want %q", out.String(), want)
	}
}

func TestSearchUsesRemoteProjectConfigWhenProjectProvided(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/devwiki/search" || r.Method != http.MethodPost {
			t.Fatalf("unexpected remote request %s %s", r.Method, r.URL.Path)
		}
		user, password, ok := r.BasicAuth()
		if !ok || user != "devwiki" || password != "T19xwxc3n2I38F1A" {
			t.Fatalf("unexpected basic auth user=%q password=%q ok=%v", user, password, ok)
		}
		data, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(data), `"kind":"workflow"`) || !strings.Contains(string(data), "脑裂") {
			t.Fatalf("unexpected remote request body %s", string(data))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"file":"workflow-ha.md","slug":"workflow-ha","title":"HA","score":"100%"}]`))
	}))
	defer server.Close()

	service := NewService()
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		RemoteURL:   server.URL,
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}
	var out bytes.Buffer
	if err := service.Search(context.Background(), SearchOptions{
		Project:    "sample",
		Kind:       "workflow",
		QueryTerms: []string{"脑裂"},
		Stdout:     &out,
	}); err != nil {
		t.Fatalf("Search(remote project) error = %v", err)
	}
	want := "|file|slug|title|score|\n|workflow-ha.md|workflow-ha|HA|100%|\n"
	if out.String() != want {
		t.Fatalf("Search(remote project) output = %q, want %q", out.String(), want)
	}
}

func TestSearchUsesRemoteProjectConfigForIndexAndGlossary(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	requests := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/devwiki/search" || r.Method != http.MethodPost {
			t.Fatalf("unexpected remote request %s %s", r.Method, r.URL.Path)
		}
		data, _ := io.ReadAll(r.Body)
		body := string(data)
		requests = append(requests, body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(body, `"kind":"index"`):
			_, _ = w.Write([]byte(`[{"type":"topic","description":"VIP 业务规则入口","slug":"vip"}]`))
		case strings.Contains(body, `"kind":"glossary"`):
			_, _ = w.Write([]byte(`[{"glossary":"VIP","type":"topic","description":"VIP 业务规则入口","slug":"vip"}]`))
		default:
			t.Fatalf("unexpected remote request body %s", body)
		}
	}))
	defer server.Close()

	service := NewService()
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		RemoteURL:   server.URL,
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}

	var indexOut bytes.Buffer
	if err := service.Search(context.Background(), SearchOptions{
		Project: "sample",
		Kind:    "index",
		Query:   "VIP",
		Stdout:  &indexOut,
	}); err != nil {
		t.Fatalf("Search(remote index) error = %v", err)
	}
	indexWant := "|type|description|slug|\n|topic|VIP 业务规则入口|vip|\n"
	if indexOut.String() != indexWant {
		t.Fatalf("Search(remote index) output = %q, want %q", indexOut.String(), indexWant)
	}

	var glossaryOut bytes.Buffer
	if err := service.Search(context.Background(), SearchOptions{
		Project: "sample",
		Kind:    "glossary",
		Query:   "VIP",
		Stdout:  &glossaryOut,
	}); err != nil {
		t.Fatalf("Search(remote glossary) error = %v", err)
	}
	glossaryWant := "|glossary|type|description|slug|\n|VIP|topic|VIP 业务规则入口|vip|\n"
	if glossaryOut.String() != glossaryWant {
		t.Fatalf("Search(remote glossary) output = %q, want %q", glossaryOut.String(), glossaryWant)
	}
	if len(requests) != 2 {
		t.Fatalf("remote request count = %d, want 2", len(requests))
	}
}

func TestGlossaryKeywordsUsesRemoteProjectConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/devwiki/glossary/keywords" || r.Method != http.MethodPost {
			t.Fatalf("unexpected remote request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"权威记录管理\n授权区管理\n"}`))
	}))
	defer server.Close()

	service := NewService()
	if err := service.RepoAdd(context.Background(), RepoAddOptions{
		ProjectSlug: "sample",
		RemoteURL:   server.URL,
	}); err != nil {
		t.Fatalf("RepoAdd(remote) error = %v", err)
	}
	var out bytes.Buffer
	if err := service.GlossaryKeywords(context.Background(), GlossaryKeywordsOptions{
		Project: "sample",
		Stdout:  &out,
	}); err != nil {
		t.Fatalf("GlossaryKeywords(remote project) error = %v", err)
	}
	want := "权威记录管理\n授权区管理\n"
	if out.String() != want {
		t.Fatalf("GlossaryKeywords(remote project) output = %q, want %q", out.String(), want)
	}
}

func devwikiSearchQMDHelperRunner(t *testing.T, root string) func(context.Context, string, ...string) *exec.Cmd {
	t.Helper()

	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		commandArgs := []string{"-test.run=TestDevwikiSearchQMDHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_DEVWIKI_SEARCH_QMD_HELPER=1")
		cmd.Dir = root
		return cmd
	}
}

func TestDevwikiSearchQMDHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_DEVWIKI_SEARCH_QMD_HELPER") != "1" {
		return
	}

	args := os.Args
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	cwd, _ := os.Getwd()
	if len(args) < 3 || args[0] != "qmd" || args[1] != "search" {
		_, _ = os.Stderr.WriteString("unexpected argv=" + strings.Join(args, " ") + "\n")
		os.Exit(2)
	}
	if filepath.Clean(cwd) == "" {
		_, _ = os.Stderr.WriteString("missing cwd\n")
		os.Exit(2)
	}

	if args[2] == "workflow-only" {
		_, _ = os.Stdout.WriteString(`workflows/ha-brain-split-handler.md:9 #abc123
Title: HA 脑裂处理链路
Score:  79%
`)
		os.Exit(0)
	}
	if args[2] == "防脑裂" {
		_, _ = os.Stdout.WriteString(`workflows/workflow-ha-brain-split-protection.md:2 #b46301
Title: HA 脑裂监控与防护实现定位
Score:  83%
`)
		os.Exit(0)
	}
	if args[2] == "网关" {
		_, _ = os.Stdout.WriteString(`workflows/workflow-ha-brain-split-protection.md:2 #b46301
Title: HA 脑裂监控与防护实现定位
Score:  83%

workflows/workflow-ha-gateway-config.md:8 #ff0000
Title: HA 网关配置实现定位
Score:  81%
`)
		os.Exit(0)
	}
	if args[2] != "脑裂" {
		_, _ = os.Stderr.WriteString("unexpected query=" + args[2] + "\n")
		os.Exit(2)
	}

	_, _ = os.Stdout.WriteString(`topics/ha-brain-split-protection.md:2 #be1507
Title: HA 脑裂监控与防护
Score:  84%

workflows/ha-brain-split-handler.md:9 #abc123
Title: HA 脑裂处理链路
Score:  79%
`)
	os.Exit(0)
}
