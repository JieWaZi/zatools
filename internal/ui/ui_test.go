package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestI18nHelpers(t *testing.T) {
	t.Setenv("ZATOOLS_LANG", "en")
	if got := CurrentLang(); got != "en" {
		t.Fatalf("CurrentLang = %q, want %q", got, "en")
	}
	if got := Messages().RootShort; got == "" {
		t.Fatal("Messages().RootShort returned empty string")
	}
	if got := StatusText("outdated"); got != "outdated" {
		t.Fatalf("StatusText(outdated) = %q", got)
	}
	if got := ScopeText(true); got != "Global" {
		t.Fatalf("ScopeText(true) = %q", got)
	}
	if got := FoundSkillsText(2); !strings.Contains(got, "2") {
		t.Fatalf("FoundSkillsText(2) = %q", got)
	}
	if got := InstalledSkillsText(2); !strings.Contains(got, "2") {
		t.Fatalf("InstalledSkillsText(2) = %q", got)
	}
	if got := ScopeTargetsText(false, "a, b"); !strings.Contains(got, "a, b") {
		t.Fatalf("ScopeTargetsText = %q", got)
	}
	if got := pluralSuffix(2); got != "s" {
		t.Fatalf("pluralSuffix(2) = %q", got)
	}
}

func TestTemplatesAndHelpLocalization(t *testing.T) {
	t.Setenv("ZATOOLS_LANG", "zh")
	usage := UsageTemplate()
	help := HelpTemplate()
	if !strings.Contains(usage, "用法") || !strings.Contains(help, "可用命令") {
		t.Fatalf("templates not localized: usage=%q help=%q", usage, help)
	}

	cmd := &cobra.Command{Use: "demo"}
	child := &cobra.Command{Use: "child"}
	cmd.AddCommand(child)
	ApplyHelpLocalization(cmd)
	if flag := cmd.Flags().Lookup("help"); flag == nil || flag.Usage != "显示帮助" {
		t.Fatalf("help flag usage = %#v", flag)
	}
	if cmd.HelpTemplate() == "" || cmd.UsageTemplate() == "" || child.HelpTemplate() == "" {
		t.Fatal("localized templates were not applied recursively")
	}
}

func TestLogoAndCommandName(t *testing.T) {
	logo := Logo()
	if !strings.Contains(logo, "███████") {
		t.Fatalf("Logo() = %q", logo)
	}

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"/usr/local/bin/zatools"}
	if got := CommandName(); got != "zatools" {
		t.Fatalf("CommandName = %q, want %q", got, "zatools")
	}

	os.Args = nil
	if got := CommandName(); got != "skill" {
		t.Fatalf("CommandName(nil args) = %q, want %q", got, "skill")
	}
}
