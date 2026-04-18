package qmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadSearchConfigAndBuildQMDCommands(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteQMDFile(t, filepath.Join(root, "config", "search.yaml"), `qmd:
  embed_model: hf:test/embed
  rerank_model: hf:test/rerank
  generate_model: hf:test/generate
  collections:
    - name: devwiki-demo-raw
      path: raw
    - name: devwiki-demo-code-app
      path: /tmp/app
`)

	collections, err := LoadCollections(root)
	if err != nil {
		t.Fatalf("LoadCollections error = %v", err)
	}
	if len(collections) != 2 {
		t.Fatalf("collections len = %d, want 2", len(collections))
	}

	commands, err := BuildCollectionCommands(root, collections)
	if err != nil {
		t.Fatalf("BuildCollectionCommands error = %v", err)
	}
	if len(commands) != 2 {
		t.Fatalf("commands len = %d, want 2", len(commands))
	}

	wantFirst := []string{"qmd", "collection", "add", filepath.Join(root, "raw"), "--name", "devwiki-demo-raw"}
	if !reflect.DeepEqual(commands[0], wantFirst) {
		t.Fatalf("first command = %#v, want %#v", commands[0], wantFirst)
	}
	wantSecond := []string{"qmd", "collection", "add", "/tmp/app", "--name", "devwiki-demo-code-app"}
	if !reflect.DeepEqual(commands[1], wantSecond) {
		t.Fatalf("second command = %#v, want %#v", commands[1], wantSecond)
	}
}

func TestLoadConfigIncludesModels(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteQMDFile(t, filepath.Join(root, "config", "search.yaml"), `qmd:
  embed_model: hf:test/embed
  rerank_model: hf:test/rerank
  generate_model: hf:test/generate
  collections:
    - name: devwiki-demo-raw
      path: raw
`)

	config, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig error = %v", err)
	}

	if config.EmbedModel != "hf:test/embed" {
		t.Fatalf("EmbedModel = %q", config.EmbedModel)
	}
	if config.RerankModel != "hf:test/rerank" {
		t.Fatalf("RerankModel = %q", config.RerankModel)
	}
	if config.GenerateModel != "hf:test/generate" {
		t.Fatalf("GenerateModel = %q", config.GenerateModel)
	}
	if len(config.Collections) != 1 {
		t.Fatalf("Collections len = %d, want 1", len(config.Collections))
	}
}

func TestResolveModelsUsesConfigAndOverrides(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteQMDFile(t, filepath.Join(root, "config", "search.yaml"), `qmd:
  embed_model: hf:config/embed
  rerank_model: hf:config/rerank
  generate_model: hf:config/generate
`)

	models, err := ResolveModels(root, Models{RerankModel: "hf:override/rerank"})
	if err != nil {
		t.Fatalf("ResolveModels error = %v", err)
	}

	if models.EmbedModel != "hf:config/embed" {
		t.Fatalf("EmbedModel = %q", models.EmbedModel)
	}
	if models.RerankModel != "hf:override/rerank" {
		t.Fatalf("RerankModel = %q", models.RerankModel)
	}
	if models.GenerateModel != "hf:config/generate" {
		t.Fatalf("GenerateModel = %q", models.GenerateModel)
	}
}

func TestResolveModelsFallsBackToDefaultsWhenConfigMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	models, err := ResolveModels(root, Models{})
	if err != nil {
		t.Fatalf("ResolveModels error = %v", err)
	}

	if models != DefaultModels() {
		t.Fatalf("models = %#v, want %#v", models, DefaultModels())
	}
}

func TestBuildDownloadCommands(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	commands, err := BuildDownloadCommands(root, []Collection{
		{Name: "devwiki-demo-raw", Path: "raw"},
	})
	if err != nil {
		t.Fatalf("BuildDownloadCommands error = %v", err)
	}

	want := [][]string{
		{"qmd", "collection", "add", filepath.Join(root, "raw"), "--name", "devwiki-demo-raw"},
		{"qmd", "update"},
		{"qmd", "embed", "-f"},
		{"qmd", "query", "zatools qmd model warmup"},
	}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands = %#v, want %#v", commands, want)
	}
}

func TestBuildCommandEnvUsesProjectRootCacheAndDefaultModels(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	mustWriteQMDFile(t, filepath.Join(root, ".git"), "")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll(nested) error = %v", err)
	}

	env := BuildCommandEnv(Models{}, nested)
	text := strings.Join(env, "\n")
	for _, want := range []string{
		"QMD_EMBED_MODEL=" + DefaultEmbedModel,
		"QMD_RERANK_MODEL=" + DefaultRerankModel,
		"QMD_GENERATE_MODEL=" + DefaultGenerateModel,
		"XDG_CACHE_HOME=" + filepath.Join(root, ".cache"),
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("env missing %q, got %q", want, text)
		}
	}
}

func TestRunCommandInjectsModelEnv(t *testing.T) {
	t.Parallel()

	restore := stubExecCommandContext(t)
	var stdout bytes.Buffer
	err := RunCommand(restore, []string{"status"}, Models{
		EmbedModel:    "hf:test/embed",
		RerankModel:   "hf:test/rerank",
		GenerateModel: "hf:test/generate",
	}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunCommand error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "argv=qmd status") {
		t.Fatalf("helper output missing argv, got %q", output)
	}
	for _, want := range []string{
		"QMD_EMBED_MODEL=hf:test/embed",
		"QMD_RERANK_MODEL=hf:test/rerank",
		"QMD_GENERATE_MODEL=hf:test/generate",
		"XDG_CACHE_HOME=",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("helper output missing %q, got %q", want, output)
		}
	}
}

func TestRunCommandUsesDefaultModelEnv(t *testing.T) {
	t.Parallel()

	restore := stubExecCommandContext(t)
	var stdout bytes.Buffer
	err := RunCommand(restore, []string{"status"}, Models{}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunCommand error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"QMD_EMBED_MODEL=" + DefaultEmbedModel,
		"QMD_RERANK_MODEL=" + DefaultRerankModel,
		"QMD_GENERATE_MODEL=" + DefaultGenerateModel,
		"XDG_CACHE_HOME=",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("helper output missing %q, got %q", want, output)
		}
	}
}

func TestRunCollectionCommandsInjectModelEnv(t *testing.T) {
	t.Parallel()

	restore := stubExecCommandContext(t)
	var stdout bytes.Buffer
	err := RunCollectionCommands(restore, [][]string{{"qmd", "update"}}, Models{
		EmbedModel:    "hf:test/embed",
		RerankModel:   "hf:test/rerank",
		GenerateModel: "hf:test/generate",
	}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunCollectionCommands error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "argv=qmd update") {
		t.Fatalf("helper output missing argv, got %q", output)
	}
	for _, want := range []string{
		"QMD_EMBED_MODEL=hf:test/embed",
		"QMD_RERANK_MODEL=hf:test/rerank",
		"QMD_GENERATE_MODEL=hf:test/generate",
		"XDG_CACHE_HOME=",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("helper output missing %q, got %q", want, output)
		}
	}
}

func TestRunCommandInDirUsesProvidedWorkingDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	canonicalRoot := canonicalQMDPath(root)
	restore := stubExecCommandContext(t)
	var stdout bytes.Buffer
	err := RunCommandInDir(restore, root, []string{"status"}, Models{}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunCommandInDir error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "cwd="+canonicalRoot) {
		t.Fatalf("helper output missing cwd, got %q", output)
	}
	if !strings.Contains(output, "XDG_CACHE_HOME="+filepath.Join(root, ".cache")) {
		t.Fatalf("helper output missing rooted cache env, got %q", output)
	}
}

func TestRunDownloadExecutesWarmupSequenceFromRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	canonicalRoot := canonicalQMDPath(root)
	mustWriteQMDFile(t, filepath.Join(root, "config", "search.yaml"), `qmd:
  embed_model: hf:test/embed
  rerank_model: hf:test/rerank
  generate_model: hf:test/generate
  collections:
    - name: devwiki-demo-raw
      path: raw
`)

	restore := stubExecCommandContext(t)
	var stdout bytes.Buffer
	err := RunDownload(restore, root, Models{}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunDownload error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"argv=qmd collection add " + filepath.Join(root, "raw") + " --name devwiki-demo-raw",
		"argv=qmd update",
		"argv=qmd embed -f",
		"argv=qmd query zatools qmd model warmup",
		"cwd=" + canonicalRoot,
		"QMD_EMBED_MODEL=hf:test/embed",
		"QMD_RERANK_MODEL=hf:test/rerank",
		"QMD_GENERATE_MODEL=hf:test/generate",
		"XDG_CACHE_HOME=" + filepath.Join(root, ".cache"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("helper output missing %q, got %q", want, output)
		}
	}
}

func TestRunDownloadIgnoresCollectionRegistrationErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteQMDFile(t, filepath.Join(root, "config", "search.yaml"), `qmd:
  collections:
    - name: devwiki-demo-raw
      path: raw
`)

	restore := WithCommandRunner(context.Background(), func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if len(args) >= 2 && args[0] == "collection" && args[1] == "add" {
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestQMDCollectionAddFailureHelperProcess", "--", name)
			cmd.Env = append(os.Environ(), "GO_WANT_QMD_HELPER_PROCESS=collection-add-fail")
			cmd.Dir = root
			return cmd
		}
		commandArgs := []string{"-test.run=TestQMDCommandHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_QMD_HELPER_PROCESS=1")
		cmd.Dir = root
		return cmd
	})

	var stdout bytes.Buffer
	err := RunDownload(restore, root, Models{}, &stdout, &stdout)
	if err != nil {
		t.Fatalf("RunDownload error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"argv=qmd update",
		"argv=qmd embed -f",
		"argv=qmd query zatools qmd model warmup",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("helper output missing %q, got %q", want, output)
		}
	}
}

func TestRunDownloadFailsWithoutSearchConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := RunDownload(context.Background(), root, Models{}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("RunDownload should fail when config/search.yaml is missing")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("RunDownload error = %v, want os.ErrNotExist", err)
	}
}

func TestQMDCommandHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_QMD_HELPER_PROCESS") != "1" {
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
	_, _ = os.Stdout.WriteString("argv=" + strings.Join(args, " ") + "\n")
	_, _ = os.Stdout.WriteString("cwd=" + cwd + "\n")
	for _, key := range []string{"QMD_EMBED_MODEL", "QMD_RERANK_MODEL", "QMD_GENERATE_MODEL", "XDG_CACHE_HOME"} {
		_, _ = os.Stdout.WriteString(key + "=" + os.Getenv(key) + "\n")
	}
	os.Exit(0)
}

func TestQMDCollectionAddFailureHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_QMD_HELPER_PROCESS") != "collection-add-fail" {
		return
	}
	_, _ = os.Stderr.WriteString("collection already exists\n")
	os.Exit(1)
}

func stubExecCommandContext(t *testing.T) context.Context {
	t.Helper()

	return WithCommandRunner(context.Background(), func(ctx context.Context, name string, args ...string) *exec.Cmd {
		commandArgs := []string{"-test.run=TestQMDCommandHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_QMD_HELPER_PROCESS=1")
		return cmd
	})
}

func mustWriteQMDFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func canonicalQMDPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}
