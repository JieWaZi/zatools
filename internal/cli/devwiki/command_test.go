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
	if sub, _, err := cmd.Find([]string{"read"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "read", err)
	}
	if sub, _, err := cmd.Find([]string{"graph"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "graph", err)
	}
	if sub, _, err := cmd.Find([]string{"check"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "check", err)
	}
	if sub, _, err := cmd.Find([]string{"search"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "search", err)
	}
}

func TestDevwikiInitFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	initCmd, _, err := cmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("Find(init) error = %v", err)
	}

	for _, flag := range []string{"agent", "code-dir", "global", "yes"} {
		if initCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("init command missing flag %q", flag)
		}
	}
	if initCmd.Flags().Lookup("lang") != nil {
		t.Fatal("init command should not expose lang flag")
	}
}

func TestDevwikiLinkDoesNotExposeLangFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	linkCmd, _, err := cmd.Find([]string{"link"})
	if err != nil {
		t.Fatalf("Find(link) error = %v", err)
	}
	if linkCmd.Flags().Lookup("lang") != nil {
		t.Fatal("link command should not expose lang flag")
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

func TestDevwikiGraphCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	graphCmd, _, err := cmd.Find([]string{"graph"})
	if err != nil {
		t.Fatalf("Find(graph) error = %v", err)
	}
	if graphCmd == nil {
		t.Fatal("graph command is nil")
	}
	for _, flag := range []string{"root", "host", "port", "no-open", "force", "check"} {
		if graphCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("graph command missing flag %q", flag)
		}
	}
}

func TestDevwikiCheckCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	checkCmd, _, err := cmd.Find([]string{"check"})
	if err != nil {
		t.Fatalf("Find(check) error = %v", err)
	}
	if checkCmd == nil {
		t.Fatal("check command is nil")
	}
	if checkCmd.Flags().Lookup("root") == nil {
		t.Fatal("check command missing flag root")
	}
}

func TestDevwikiReadCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	readCmd, _, err := cmd.Find([]string{"read"})
	if err != nil {
		t.Fatalf("Find(read) error = %v", err)
	}
	if readCmd == nil {
		t.Fatal("read command is nil")
	}
	for _, flag := range []string{"root", "view", "format"} {
		if readCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("read command missing flag %q", flag)
		}
	}
	if readCmd.Annotations[SuppressLogoAnnotation] != "true" {
		t.Fatalf("read command should suppress logo annotation")
	}
	if initCmd, _, err := cmd.Find([]string{"init"}); err != nil || initCmd.Annotations[SuppressLogoAnnotation] == "true" {
		t.Fatalf("init command should not suppress logo: cmd=%v err=%v", initCmd, err)
	}
}

func TestDevwikiSearchCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	searchCmd, _, err := cmd.Find([]string{"search"})
	if err != nil {
		t.Fatalf("Find(search) error = %v", err)
	}
	if searchCmd == nil {
		t.Fatal("search command is nil")
	}
	if searchCmd.Flags().Lookup("root") == nil {
		t.Fatal("search command missing flag root")
	}
	if searchCmd.Annotations[SuppressLogoAnnotation] != "true" {
		t.Fatalf("search command should suppress logo annotation")
	}
}

func TestDevwikiSearchAcceptsMultipleQueryTerms(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	searchCmd, _, err := cmd.Find([]string{"search", "workflow", "防脑裂", "网关"})
	if err != nil {
		t.Fatalf("Find(search with multiple terms) error = %v", err)
	}
	if err := searchCmd.Args(searchCmd, []string{"workflow", "防脑裂", "网关"}); err != nil {
		t.Fatalf("search Args() error = %v", err)
	}
}

func TestDevwikiSearchAcceptsIndexAndGlossaryKinds(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	for _, kind := range []string{"index", "glossary"} {
		searchCmd, _, err := cmd.Find([]string{"search", kind, "脑裂"})
		if err != nil {
			t.Fatalf("Find(search %s) error = %v", kind, err)
		}
		if err := searchCmd.Args(searchCmd, []string{kind, "脑裂"}); err != nil {
			t.Fatalf("search Args(%s) error = %v", kind, err)
		}
	}
}
