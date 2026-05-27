package devwikiapp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	common "zatools/internal/app/common"
	"zatools/internal/qmd"
	"zatools/internal/skills"
)

func TestSearchTopicFiltersQMDOutputAndWritesJSON(t *testing.T) {
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

	var got []SearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal output error = %v, output=%q", err, out.String())
	}
	want := []SearchResult{
		{File: "ha-brain-split-protection.md", Slug: "ha-brain-split-protection", Title: "HA 脑裂监控与防护", Score: "84%"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("results = %#v, want %#v", got, want)
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

	var got []SearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal output error = %v, output=%q", err, out.String())
	}
	want := []SearchResult{
		{File: "workflow-ha-brain-split-protection.md", Slug: "workflow-ha-brain-split-protection", Title: "HA 脑裂监控与防护实现定位", Score: "100%"},
		{File: "workflow-ha-gateway-config.md", Slug: "workflow-ha-gateway-config", Title: "HA 网关配置实现定位", Score: "49%"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("results = %#v, want %#v", got, want)
	}
}

func TestSearchRunsQueryTermsConcurrently(t *testing.T) {
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
		mu       sync.Mutex
		started  int
		release  = make(chan struct{})
		timedOut bool
	)
	runner := devwikiSearchQMDHelperRunner(t, root)
	ctx := qmd.WithCommandRunner(context.Background(), func(ctx context.Context, name string, args ...string) *exec.Cmd {
		mu.Lock()
		started++
		if started == 2 {
			close(release)
		}
		mu.Unlock()

		select {
		case <-release:
		case <-time.After(2 * time.Second):
			mu.Lock()
			timedOut = true
			mu.Unlock()
		case <-ctx.Done():
		}
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
	gotTimedOut := timedOut
	mu.Unlock()
	if gotStarted != 2 {
		t.Fatalf("started qmd searches = %d, want 2", gotStarted)
	}
	if gotTimedOut {
		t.Fatal("qmd searches did not overlap before timeout")
	}
}

func TestSearchIndexParsesTableAndWritesJSON(t *testing.T) {
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

	var got []IndexSearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal output error = %v, output=%q", err, out.String())
	}
	want := []IndexSearchResult{
		{Type: "topic", Description: "HA 脑裂监控与防护的业务规则入口", Slug: "ha-brain-split-protection"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("results = %#v, want %#v", got, want)
	}
	if strings.Contains(out.String(), "score") {
		t.Fatalf("index search output should not include score: %s", out.String())
	}
}

func TestSearchGlossaryParsesTableAndWritesJSON(t *testing.T) {
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

	var got []GlossarySearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal output error = %v, output=%q", err, out.String())
	}
	want := []GlossarySearchResult{
		{Glossary: "网关配置", Type: "workflow", Description: "HA 网关配置下发和持久化实现链路", Slug: "workflow-ha-gateway-config"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("results = %#v, want %#v", got, want)
	}
	if strings.Contains(out.String(), "score") {
		t.Fatalf("glossary search output should not include score: %s", out.String())
	}
}

func TestSearchIndexWritesEmptyJSONArrayWhenNoRowsMatch(t *testing.T) {
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
	if strings.TrimSpace(out.String()) != "[]" {
		t.Fatalf("Search() output = %q, want []", out.String())
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

	got := parseQMDSearchOutput(input, "workflow")
	want := []SearchResult{{File: "ha-brain-split-handler.md", Title: "HA 脑裂处理链路", Score: "79%"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("workflow results = %#v, want %#v", got, want)
	}
}

func TestParseQMDSearchFileLineAcceptsQMDURICollectionPaths(t *testing.T) {
	got, ok := parseQMDSearchFileLine("qmd://devwiki-huawei-zddi-wiki/workflows/workflow-ha-brain-split-protection.md:2 #b46301", "workflows")
	if !ok {
		t.Fatal("parseQMDSearchFileLine did not accept qmd URI workflow path")
	}
	if got != "workflow-ha-brain-split-protection.md" {
		t.Fatalf("file = %q, want workflow-ha-brain-split-protection.md", got)
	}
}

func TestSearchWritesEmptyJSONArrayWhenNoHitsMatchKind(t *testing.T) {
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
	if strings.TrimSpace(out.String()) != "[]" {
		t.Fatalf("Search() output = %q, want []", out.String())
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
