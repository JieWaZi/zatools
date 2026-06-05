package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"zatools/internal/qmd"
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
	if sub, _, err := cmd.Find([]string{"update"}); err != nil || sub == nil {
		t.Fatalf("root command missing update subcommand: %v", err)
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

func TestDevwikiInitDoesNotPrintLogo(t *testing.T) {
	t.Chdir(t.TempDir())

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"devwiki", "init", "sample", "--code-dir", t.TempDir(), "--yes"})
	var commandOut bytes.Buffer
	cmd.SetOut(&commandOut)
	cmd.SetErr(io.Discard)

	processOut := captureStdoutRootTest(t, func() {
		_ = cmd.Execute()
	})
	if strings.Contains(processOut, "███████") || strings.Contains(commandOut.String(), "███████") {
		t.Fatalf("devwiki init output should not include logo: stdout=%q commandOut=%q", processOut, commandOut.String())
	}
}

func TestDevwikiRootDoesNotPrintLogo(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"devwiki"})
	var commandOut bytes.Buffer
	cmd.SetOut(&commandOut)
	cmd.SetErr(io.Discard)

	processOut := captureStdoutRootTest(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
	if strings.Contains(processOut, "███████") || strings.Contains(commandOut.String(), "███████") {
		t.Fatalf("devwiki root output should not include logo: stdout=%q commandOut=%q", processOut, commandOut.String())
	}
}

func TestQMDDoesNotPrintLogo(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"qmd", "status"})
	cmd.SetContext(qmd.WithCommandRunner(context.Background(), rootQMDHelperRunner(t)))
	var commandOut bytes.Buffer
	cmd.SetOut(&commandOut)
	cmd.SetErr(io.Discard)

	processOut := captureStdoutRootTest(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
	if strings.Contains(processOut, "███████") || strings.Contains(commandOut.String(), "███████") {
		t.Fatalf("qmd output should not include logo: stdout=%q commandOut=%q", processOut, commandOut.String())
	}
	if !strings.Contains(commandOut.String(), "argv=qmd status") {
		t.Fatalf("qmd command did not run through helper: %q", commandOut.String())
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

func rootQMDHelperRunner(t *testing.T) func(context.Context, string, ...string) *exec.Cmd {
	t.Helper()

	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		commandArgs := []string{"-test.run=TestRootQMDHelperProcess", "--", name}
		commandArgs = append(commandArgs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		cmd.Env = append(os.Environ(), "GO_WANT_ROOT_QMD_HELPER=1")
		return cmd
	}
}

func TestRootQMDHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_ROOT_QMD_HELPER") != "1" {
		return
	}

	args := os.Args
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]
			break
		}
	}
	_, _ = os.Stdout.WriteString("argv=" + strings.Join(args, " ") + "\n")
	os.Exit(0)
}
