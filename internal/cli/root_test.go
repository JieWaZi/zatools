package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmdIncludesExpectedSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewRootCmd()
	if cmd.Use == "" {
		t.Fatal("root command Use is empty")
	}
	if !cmd.CompletionOptions.DisableDefaultCmd {
		t.Fatal("default completion command should be disabled")
	}
	if cmd.CommandPath() == "" {
		t.Fatal("root command path is empty")
	}
	if sub, _, err := cmd.Find([]string{"skill"}); err != nil || sub == nil {
		t.Fatalf("root command missing skill subcommand: %v", err)
	}
	if sub, _, err := cmd.Find([]string{"rule"}); err != nil || sub == nil {
		t.Fatalf("root command missing rule subcommand: %v", err)
	}
	if sub, _, err := cmd.Find([]string{"devwiki"}); err != nil || sub == nil {
		t.Fatalf("root command missing devwiki subcommand: %v", err)
	}
	if sub, _, err := cmd.Find([]string{"qmd"}); err != nil || sub == nil {
		t.Fatalf("root command missing qmd subcommand: %v", err)
	}
	if sub, _, err := cmd.Find([]string{"completion"}); err != nil || sub == nil {
		t.Fatalf("root command missing completion subcommand: %v", err)
	}
}

func TestWriteHelpWritesUsageAndCommandUsageString(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use:   "demo",
		Short: "short help",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	var buf bytes.Buffer
	writeHelp(&buf, cmd)

	out := buf.String()
	if !strings.Contains(out, "short help") || !strings.Contains(out, "demo") {
		t.Fatalf("writeHelp output = %q", out)
	}
}

func TestNewCompletionCmdAddsShellCommands(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "zatools"}
	cmd := newCompletionCmd(root)

	if got := len(cmd.Commands()); got != 4 {
		t.Fatalf("completion subcommands = %d, want 4", got)
	}
	for _, name := range []string{"bash", "zsh", "fish", "powershell"} {
		if sub, _, err := cmd.Find([]string{name}); err != nil || sub == nil {
			t.Fatalf("completion command missing %q: %v", name, err)
		}
	}
}

func TestDevwikiReadDoesNotPrintLogo(t *testing.T) {
	root := t.TempDir()
	mustWriteRootTestFile(t, filepath.Join(root, "wiki/topics/vip.md"), `---
title: "VIP"
slug: "vip"
kind: topic
summary: "VIP topic"
---
# VIP

<!-- devwiki:section id=card -->
## 导航卡
card body
<!-- /devwiki:section -->
`)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"devwiki", "read", "topic", "vip", "--root", root})
	var commandOut bytes.Buffer
	cmd.SetOut(&commandOut)
	cmd.SetErr(io.Discard)

	processOut := captureStdoutRootTest(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
	if strings.Contains(processOut, "███████") || strings.Contains(commandOut.String(), "███████") {
		t.Fatalf("devwiki read output should not include logo: stdout=%q commandOut=%q", processOut, commandOut.String())
	}
	if !strings.Contains(commandOut.String(), "card body") {
		t.Fatalf("devwiki read output missing page content: %q", commandOut.String())
	}
}

func mustWriteRootTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func captureStdoutRootTest(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	os.Stdout = writer
	fn()
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	os.Stdout = original
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("Close reader error = %v", err)
	}
	return string(data)
}
