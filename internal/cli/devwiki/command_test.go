package devwikicmd

import "testing"

func TestNewCommandIncludesInitAndToolSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	if cmd.Use != "devwiki" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "devwiki")
	}

	if sub, _, err := cmd.Find([]string{"init"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "init", err)
	}
	if sub, _, err := cmd.Find([]string{"tool"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "tool", err)
	}
	if sub, _, err := cmd.Find([]string{"update"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "update", err)
	}
}

func TestDevwikiInitFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	initCmd, _, err := cmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("Find(init) error = %v", err)
	}

	for _, flag := range []string{"agent", "lang", "code-dir", "global", "yes"} {
		if initCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("init command missing flag %q", flag)
		}
	}
}

func TestDevwikiToolCommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	toolCmd, _, err := cmd.Find([]string{"tool"})
	if err != nil {
		t.Fatalf("Find(tool) error = %v", err)
	}
	if toolCmd == nil {
		t.Fatal("tool command is nil")
	}

	resetCmd, _, err := cmd.Find([]string{"tool", "reset"})
	if err != nil {
		t.Fatalf("Find(tool reset) error = %v", err)
	}
	for _, flag := range []string{"scope", "project-root", "yes"} {
		if resetCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("reset command missing flag %q", flag)
		}
	}

	logCmd, _, err := cmd.Find([]string{"tool", "log"})
	if err != nil {
		t.Fatalf("Find(tool log) error = %v", err)
	}
	for _, flag := range []string{"wiki-root", "message"} {
		if logCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("log command missing flag %q", flag)
		}
	}

	for _, sub := range toolCmd.Commands() {
		if sub.Name() == "qmd" {
			t.Fatalf("devwiki tool should not expose qmd subcommand anymore")
		}
	}
}
