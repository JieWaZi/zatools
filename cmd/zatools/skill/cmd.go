package skill

import (
	"github.com/ryan/go-skills/internal/ui"
	"github.com/spf13/cobra"
)

// NewSkillCmd 构建 `skill` 子命令及其所有管理动作。
func NewSkillCmd() *cobra.Command {
	copy := ui.Messages()
	skillCmd := &cobra.Command{
		Use:           "skill",
		Short:         copy.SkillShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	skillCmd.AddCommand(newSkillAddCmd())
	skillCmd.AddCommand(newSkillListCmd())
	skillCmd.AddCommand(newSkillInitCmd())
	skillCmd.AddCommand(newSkillRemoveCmd())
	skillCmd.AddCommand(newSkillCheckCmd())
	skillCmd.AddCommand(newSkillUpdateCmd())
	return skillCmd
}

// newSkillAddCmd 构建安装技能包的命令。
func newSkillAddCmd() *cobra.Command {
	copy := ui.Messages()
	var opts AddOptions

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: copy.AddShort,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ScopeProvided = cmd.Flags().Changed("global")
			return Add(newCommandContext(), args[0], opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, copy.FlagInstallGlobally)
	cmd.Flags().BoolVarP(&opts.ListOnly, "list", "l", false, copy.FlagListOnly)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipPrompts)
	cmd.Flags().StringSliceVarP(&opts.SkillNames, "skill", "s", nil, copy.FlagSpecificSkills)
	cmd.Flags().StringSliceVarP(&opts.Agents, "agent", "a", nil, copy.FlagTargetAgents)
	return cmd
}

// newSkillListCmd 构建已安装技能列表命令。
func newSkillListCmd() *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   copy.ListShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return List(newCommandContext(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagListGlobal)
	return cmd
}

// newSkillInitCmd 构建技能模板初始化命令。
func newSkillInitCmd() *cobra.Command {
	copy := ui.Messages()
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: copy.InitShort,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return Init(name)
		},
	}
	return cmd
}

// newSkillRemoveCmd 构建移除技能命令。
func newSkillRemoveCmd() *cobra.Command {
	copy := ui.Messages()
	var opts RemoveOptions

	cmd := &cobra.Command{
		Use:     "remove [skills...]",
		Aliases: []string{"rm"},
		Short:   copy.RemoveShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SkillNames = append(opts.SkillNames[:0], args...)
			return Remove(newCommandContext(), opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, copy.FlagRemoveGlobal)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipConfirmation)
	cmd.Flags().BoolVar(&opts.All, "all", false, copy.FlagRemoveAll)
	cmd.Flags().StringSliceVarP(&opts.SkillNames, "skill", "s", nil, copy.FlagSpecificSkills)
	return cmd
}

// newSkillCheckCmd 构建检查更新命令。
func newSkillCheckCmd() *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: copy.CheckShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Check(newCommandContext(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagCheckGlobal)
	return cmd
}

// newSkillUpdateCmd 构建批量更新技能命令。
func newSkillUpdateCmd() *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: copy.UpdateShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Update(newCommandContext(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagUpdateGlobal)
	return cmd
}
