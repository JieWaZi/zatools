package skillcmd

import (
	"testing"
)

func TestNewCommandIncludesSkillManagementSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	if cmd.Use != "skill" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "skill")
	}

	for _, name := range []string{"add", "list", "init", "remove", "check", "update"} {
		if sub, _, err := cmd.Find([]string{name}); err != nil || sub == nil {
			t.Fatalf("missing subcommand %q: %v", name, err)
		}
	}
}

func TestSkillSubcommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()

	addCmd, _, err := cmd.Find([]string{"add"})
	if err != nil {
		t.Fatalf("Find(add) error = %v", err)
	}
	for _, flag := range []string{"global", "list", "yes", "skill", "agent"} {
		if addCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("add command missing flag %q", flag)
		}
	}

	removeCmd, _, err := cmd.Find([]string{"remove"})
	if err != nil {
		t.Fatalf("Find(remove) error = %v", err)
	}
	for _, flag := range []string{"global", "yes", "all", "skill"} {
		if removeCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("remove command missing flag %q", flag)
		}
	}
}
