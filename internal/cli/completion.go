package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"zatools/internal/ui"
)

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	copy := ui.Messages()
	noDesc := root.CompletionOptions.DisableDescriptions

	cmd := &cobra.Command{
		Use:   "completion",
		Short: copy.CompletionShort,
		Long:  copy.CompletionLong,
		Args:  cobra.NoArgs,
	}

	addNoDesc := func(shellCmd *cobra.Command) {
		if root.CompletionOptions.DisableNoDescFlag || root.CompletionOptions.DisableDescriptions {
			return
		}
		shellCmd.Flags().BoolVar(&noDesc, "no-descriptions", false, copy.CompletionNoDescFlag)
	}

	bashCmd := &cobra.Command{
		Use:                   "bash",
		Short:                 copy.CompletionBashShort,
		Long:                  fmt.Sprintf(copy.CompletionBashLong, root.Name()),
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenBashCompletionV2(root.OutOrStdout(), !noDesc)
		},
	}
	addNoDesc(bashCmd)

	zshCmd := &cobra.Command{
		Use:   "zsh",
		Short: copy.CompletionZshShort,
		Long:  fmt.Sprintf(copy.CompletionZshLong, root.Name()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDesc {
				return root.GenZshCompletionNoDesc(root.OutOrStdout())
			}
			return root.GenZshCompletion(root.OutOrStdout())
		},
	}
	addNoDesc(zshCmd)

	fishCmd := &cobra.Command{
		Use:   "fish",
		Short: copy.CompletionFishShort,
		Long:  fmt.Sprintf(copy.CompletionFishLong, root.Name()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return root.GenFishCompletion(root.OutOrStdout(), !noDesc)
		},
	}
	addNoDesc(fishCmd)

	powershellCmd := &cobra.Command{
		Use:   "powershell",
		Short: copy.CompletionPowerShellShort,
		Long:  fmt.Sprintf(copy.CompletionPowerShellLong, root.Name()),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if noDesc {
				return root.GenPowerShellCompletion(root.OutOrStdout())
			}
			return root.GenPowerShellCompletionWithDesc(root.OutOrStdout())
		},
	}
	addNoDesc(powershellCmd)

	cmd.AddCommand(bashCmd, zshCmd, fishCmd, powershellCmd)
	return cmd
}
