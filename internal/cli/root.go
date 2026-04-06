package cli

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	rulecmd "zatools/internal/cli/rule"
	skillcmd "zatools/internal/cli/skill"
	"zatools/internal/ui"
)

var showLogoOnce sync.Once

// NewRootCmd 构建 CLI 根命令，并把 Logo 合并到 Cobra 的帮助模板中。
func NewRootCmd() *cobra.Command {
	copy := ui.Messages()
	rootCmd := &cobra.Command{
		Use:           ui.CommandName(),
		Short:         copy.RootShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			showLogoOnce.Do(ui.ShowLogo)
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showLogoOnce.Do(ui.ShowLogo)
		writeHelp(cmd.OutOrStdout(), cmd)
	})
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(skillcmd.NewCommand())
	rootCmd.AddCommand(rulecmd.NewCommand())
	rootCmd.AddCommand(newCompletionCmd(rootCmd))
	ui.ApplyHelpLocalization(rootCmd)
	return rootCmd
}

func writeHelp(w io.Writer, cmd *cobra.Command) {
	usage := cmd.Long
	if usage == "" {
		usage = cmd.Short
	}
	usage = strings.TrimRight(usage, "\r\n\t ")
	if usage != "" {
		fmt.Fprintln(w, usage)
		fmt.Fprintln(w)
	}
	if cmd.Runnable() || cmd.HasSubCommands() {
		fmt.Fprint(w, cmd.UsageString())
	}
}
