package qmdcmd

import (
	"reflect"
	"testing"

	"zatools/internal/qmd"
	"zatools/internal/ui"
)

func TestTopLevelQMDCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	if cmd.Use != "qmd" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "qmd")
	}
	if cmd.Short != ui.Messages().QMDShort {
		t.Fatalf("Short = %q, want %q", cmd.Short, ui.Messages().QMDShort)
	}
	if !cmd.DisableFlagParsing {
		t.Fatal("qmd command should disable cobra flag parsing for passthrough mode")
	}
	if got := len(cmd.Commands()); got != 0 {
		t.Fatalf("qmd command should not define cobra subcommands, got %d", got)
	}
}

func TestParseQMDArgsExtractsModelFlagsAndPreservesPassthrough(t *testing.T) {
	t.Parallel()

	defaults := qmd.DefaultModels()
	models, passthrough, err := parseQMDArgs([]string{
		"--embed-model", "hf:custom/embed",
		"--rerank-model=hf:custom/rerank",
		"query",
		"--limit", "5",
		"--",
		"foo bar",
	}, defaults)
	if err != nil {
		t.Fatalf("parseQMDArgs error = %v", err)
	}

	if models.EmbedModel != "hf:custom/embed" {
		t.Fatalf("EmbedModel = %q", models.EmbedModel)
	}
	if models.RerankModel != "hf:custom/rerank" {
		t.Fatalf("RerankModel = %q", models.RerankModel)
	}
	if models.GenerateModel != defaults.GenerateModel {
		t.Fatalf("GenerateModel = %q, want default %q", models.GenerateModel, defaults.GenerateModel)
	}

	wantArgs := []string{"query", "--limit", "5", "foo bar"}
	if !reflect.DeepEqual(passthrough, wantArgs) {
		t.Fatalf("passthrough = %#v, want %#v", passthrough, wantArgs)
	}
}

func TestParseQMDArgsErrorsWithoutFlagValue(t *testing.T) {
	t.Parallel()

	_, _, err := parseQMDArgs([]string{"--embed-model"}, qmd.DefaultModels())
	if err == nil {
		t.Fatal("parseQMDArgs should fail when embed-model value is missing")
	}
}

func TestParseQMDSyncArgs(t *testing.T) {
	t.Parallel()

	root, apply, err := parseQMDSyncArgs([]string{"--root", "devwiki-test", "--apply"})
	if err != nil {
		t.Fatalf("parseQMDSyncArgs error = %v", err)
	}
	if root != "devwiki-test" {
		t.Fatalf("root = %q, want %q", root, "devwiki-test")
	}
	if !apply {
		t.Fatal("apply = false, want true")
	}
}

func TestParseQMDSyncArgsRejectsUnknownArgs(t *testing.T) {
	t.Parallel()

	_, _, err := parseQMDSyncArgs([]string{"status"})
	if err == nil {
		t.Fatal("parseQMDSyncArgs should reject unknown args")
	}
}

func TestParseQMDDownloadArgs(t *testing.T) {
	t.Parallel()

	root, err := parseQMDDownloadArgs([]string{"--root", "devwiki-test"})
	if err != nil {
		t.Fatalf("parseQMDDownloadArgs error = %v", err)
	}
	if root != "devwiki-test" {
		t.Fatalf("root = %q, want %q", root, "devwiki-test")
	}
}

func TestParseQMDDownloadArgsRejectsUnknownArgs(t *testing.T) {
	t.Parallel()

	_, err := parseQMDDownloadArgs([]string{"query"})
	if err == nil {
		t.Fatal("parseQMDDownloadArgs should reject unknown args")
	}
}
