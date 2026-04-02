package ui

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyHelpLocalization 为命令树设置本地化帮助模板和帮助参数说明。
func ApplyHelpLocalization(cmd *cobra.Command) {
	copy := Messages()
	cmd.InitDefaultHelpFlag()
	if flag := cmd.Flags().Lookup("help"); flag != nil {
		flag.Usage = copy.HelpFlag
	}
	cmd.SetHelpTemplate(HelpTemplate())
	cmd.SetUsageTemplate(UsageTemplate())
	for _, child := range cmd.Commands() {
		ApplyHelpLocalization(child)
	}
}

// UsageTemplate 返回当前语言下的 Cobra 用法模板。
func UsageTemplate() string {
	copy := Messages()
	return fmt.Sprintf(`%s:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s:{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s{{end}}
`, copy.HelpUsage, copy.HelpAliases, copy.HelpExamples, copy.HelpAvailableCommands, copy.HelpFlags, copy.HelpGlobalFlags, copy.HelpAdditionalTopics, fmt.Sprintf(copy.HelpMoreInfoFmt, "{{.CommandPath}} [command] --help"))
}

// HelpTemplate 返回当前语言下的 Cobra 帮助模板。
func HelpTemplate() string {
	copy := Messages()
	return fmt.Sprintf(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

%s:
  {{.UseLine}}{{if .HasAvailableSubCommands}}

%s:{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if gt (len .Aliases) 0}}

%s:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s:
{{.Example}}{{end}}{{if .HasAvailableLocalFlags}}

%s:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s{{end}}
`, copy.HelpUsage, copy.HelpAvailableCommands, copy.HelpAliases, copy.HelpExamples, copy.HelpFlags, copy.HelpGlobalFlags, copy.HelpAdditionalTopics, fmt.Sprintf(copy.HelpMoreInfoFmt, "{{.CommandPath}} [command] --help"))
}
