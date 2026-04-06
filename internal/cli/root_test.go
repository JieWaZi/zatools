package cli

import (
	"bytes"
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
