package devwikicmd

import (
	"bytes"
	"strings"
	"testing"
)

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
	if sub, _, err := cmd.Find([]string{"server"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "server", err)
	}
	if sub, _, err := cmd.Find([]string{"check"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "check", err)
	}
	if sub, _, err := cmd.Find([]string{"search"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "search", err)
	}
	if sub, _, err := cmd.Find([]string{"repo"}); err != nil || sub == nil {
		t.Fatalf("missing subcommand %q: %v", "repo", err)
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

func TestDevwikiDoesNotExposeTopLevelLinkCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "link" {
			t.Fatal("devwiki should not expose top-level link command; use devwiki repo link")
		}
	}
}

func TestDevwikiRepoIncludesUseSubcommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	useCmd, _, err := cmd.Find([]string{"repo", "use"})
	if err != nil {
		t.Fatalf("Find(repo use) error = %v", err)
	}
	if useCmd == nil {
		t.Fatal("repo use command is nil")
	}
	if useCmd.Use != "use <project> <local|remote>" {
		t.Fatalf("repo use Use = %q", useCmd.Use)
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
	for _, flag := range []string{"root", "project", "host", "port", "no-open", "force", "check"} {
		if graphCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("graph command missing flag %q", flag)
		}
	}
	if got := graphCmd.Flags().Lookup("port").DefValue; got != "5696" {
		t.Fatalf("graph port default = %q, want 5696", got)
	}
	if got := graphCmd.Flags().Lookup("host").DefValue; got != "127.0.0.1" {
		t.Fatalf("graph host default = %q, want 127.0.0.1", got)
	}
}

func TestDevwikiServerCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	serverCmd, _, err := cmd.Find([]string{"server"})
	if err != nil {
		t.Fatalf("Find(server) error = %v", err)
	}
	if serverCmd == nil {
		t.Fatal("server command is nil")
	}
	for _, flag := range []string{"root", "project", "host", "port"} {
		if serverCmd.Flags().Lookup(flag) == nil {
			t.Fatalf("server command missing flag %q", flag)
		}
	}
	if got := serverCmd.Flags().Lookup("host").DefValue; got != "0.0.0.0" {
		t.Fatalf("server host default = %q, want 0.0.0.0", got)
	}
	if got := serverCmd.Flags().Lookup("port").DefValue; got != "5697" {
		t.Fatalf("server port default = %q, want 5697", got)
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
	for _, flag := range []string{"root", "project", "view", "format"} {
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
	if searchCmd.Flags().Lookup("project") == nil {
		t.Fatal("search command missing flag project")
	}
	if searchCmd.Annotations[SuppressLogoAnnotation] != "true" {
		t.Fatalf("search command should suppress logo annotation")
	}
}

func TestDevwikiRepoCommandSurface(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	for _, path := range [][]string{
		{"repo", "init"},
		{"repo", "add"},
		{"repo", "link"},
		{"repo", "info"},
	} {
		sub, _, err := cmd.Find(path)
		if err != nil || sub == nil {
			t.Fatalf("missing subcommand %v: cmd=%v err=%v", path, sub, err)
		}
	}

	addCmd, _, err := cmd.Find([]string{"repo", "add"})
	if err != nil {
		t.Fatalf("Find(repo add) error = %v", err)
	}
	if addCmd.Flags().Lookup("remote") == nil {
		t.Fatal("repo add command missing flag remote")
	}

	infoCmd, _, err := cmd.Find([]string{"repo", "info"})
	if err != nil {
		t.Fatalf("Find(repo info) error = %v", err)
	}
	if err := infoCmd.Args(infoCmd, []string{}); err != nil {
		t.Fatalf("repo info should accept zero args: %v", err)
	}
	if err := infoCmd.Args(infoCmd, []string{"huawei-zddi"}); err != nil {
		t.Fatalf("repo info should accept one arg: %v", err)
	}
	format := infoCmd.Flags().Lookup("format")
	if format == nil {
		t.Fatal("repo info command missing flag format")
	}
	if format.DefValue != "json" {
		t.Fatalf("repo info format default = %q, want json", format.DefValue)
	}
	repoCmd, _, err := cmd.Find([]string{"repo"})
	if err != nil {
		t.Fatalf("Find(repo) error = %v", err)
	}
	for _, sub := range repoCmd.Commands() {
		if sub.Name() == "init" && strings.TrimSpace(sub.Short) == "" {
			t.Fatal("repo init command should have short description")
		}
		if sub.Name() == "path" {
			t.Fatal("repo command should not expose path subcommand; use repo info")
		}
		if sub.Name() == "list" {
			t.Fatal("repo command should not expose list subcommand; use repo info")
		}
	}
}

func TestDevwikiGlossaryKeywordsCommandFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand()
	keywordsCmd, _, err := cmd.Find([]string{"glossary", "keywords"})
	if err != nil {
		t.Fatalf("Find(glossary keywords) error = %v", err)
	}
	if keywordsCmd == nil {
		t.Fatal("glossary keywords command is nil")
	}
	if keywordsCmd.Flags().Lookup("root") == nil {
		t.Fatal("glossary keywords command missing flag root")
	}
	if keywordsCmd.Flags().Lookup("project") == nil {
		t.Fatal("glossary keywords command missing flag project")
	}
	if keywordsCmd.Annotations[SuppressLogoAnnotation] != "true" {
		t.Fatalf("glossary keywords command should suppress logo annotation")
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

func TestDevwikiRepoInitPrintsReadableFailure(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"repo", "init"})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want failure")
	}
	if !strings.Contains(errOut.String(), "DevWiki repo init 失败") ||
		!strings.Contains(errOut.String(), "requires an interactive terminal") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestDevwikiRepoAddPrintsReadableFailure(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"repo", "add", "huawei-zddi"})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want failure")
	}
	if !strings.Contains(errOut.String(), "DevWiki repo add 失败") ||
		!strings.Contains(errOut.String(), "requires exactly one local path or --remote url") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestDevwikiRepoAddPrintsReadableArgFailure(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"repo", "add"})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want failure")
	}
	if !strings.Contains(errOut.String(), "DevWiki repo add 失败") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestDevwikiRepoLinkPrintsReadableFailure(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"repo", "link", "huawei-zddi", "zddiv3", "/path/does/not/exist"})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want failure")
	}
	if !strings.Contains(errOut.String(), "DevWiki repo link 失败") ||
		!strings.Contains(errOut.String(), "/path/does/not/exist") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestDevwikiRepoLinkPrintsReadableArgFailure(t *testing.T) {
	cmd := NewCommand()
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"repo", "link", "huawei-zddi"})

	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want failure")
	}
	if !strings.Contains(errOut.String(), "DevWiki repo link 失败") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}
