package rulecmd

import (
	"github.com/spf13/cobra"

	"zatools/internal/app/ruleapp"
	"zatools/internal/ui"
)

// NewCommand 构建 `rule` 子命令及其所有管理动作。
func NewCommand() *cobra.Command {
	copy := ui.Messages()
	service := ruleapp.NewService()
	ruleCmd := &cobra.Command{
		Use:           "rule",
		Short:         copy.RuleShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	ruleCmd.AddCommand(newRuleAddCmd(service))
	ruleCmd.AddCommand(newRuleListCmd(service))
	ruleCmd.AddCommand(newRuleRemoveCmd(service))
	ruleCmd.AddCommand(newRuleCheckCmd(service))
	ruleCmd.AddCommand(newRuleUpdateCmd(service))
	return ruleCmd
}

func newRuleAddCmd(service *ruleapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts ruleapp.AddOptions

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: copy.RuleAddShort,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Add(cmd.Context(), args[0], opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.ListOnly, "list", "l", false, copy.FlagRuleListOnly)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipPrompts)
	cmd.Flags().StringSliceVarP(&opts.RuleNames, "rule", "r", nil, copy.FlagSpecificRules)
	cmd.Flags().StringSliceVarP(&opts.Agents, "agent", "a", nil, copy.FlagTargetRuleAgents)
	return cmd
}

func newRuleListCmd(service *ruleapp.Service) *cobra.Command {
	copy := ui.Messages()
	return &cobra.Command{
		Use:   "list",
		Short: copy.RuleListShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.List(cmd.Context())
		},
	}
}

func newRuleRemoveCmd(service *ruleapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts ruleapp.RemoveOptions

	cmd := &cobra.Command{
		Use:   "remove [rules...]",
		Short: copy.RuleRemoveShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RuleNames = append(opts.RuleNames[:0], args...)
			return service.Remove(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipConfirmation)
	cmd.Flags().BoolVar(&opts.All, "all", false, copy.FlagRemoveAllRules)
	cmd.Flags().StringSliceVarP(&opts.RuleNames, "rule", "r", nil, copy.FlagSpecificRules)
	return cmd
}

func newRuleCheckCmd(service *ruleapp.Service) *cobra.Command {
	copy := ui.Messages()
	return &cobra.Command{
		Use:   "check",
		Short: copy.RuleCheckShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Check(cmd.Context())
		},
	}
}

func newRuleUpdateCmd(service *ruleapp.Service) *cobra.Command {
	copy := ui.Messages()
	return &cobra.Command{
		Use:   "update",
		Short: copy.RuleUpdateShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Update(cmd.Context())
		},
	}
}
