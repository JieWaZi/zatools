package skillcmd

import (
	"github.com/spf13/cobra"

	"zatools/internal/app/skillapp"
	"zatools/internal/ui"
)

// NewCommand 构建 `skill` 子命令及其所有管理动作。
func NewCommand() *cobra.Command {
	copy := ui.Messages()
	service := skillapp.NewService()
	skillCmd := &cobra.Command{
		Use:           "skill",
		Short:         copy.SkillShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	skillCmd.AddCommand(newSkillAddCmd(service))
	skillCmd.AddCommand(newSkillListCmd(service))
	skillCmd.AddCommand(newSkillInitCmd(service))
	skillCmd.AddCommand(newSkillRemoveCmd(service))
	skillCmd.AddCommand(newSkillCheckCmd(service))
	skillCmd.AddCommand(newSkillUpdateCmd(service))
	return skillCmd
}

// newSkillAddCmd 构建安装技能包的命令。
func newSkillAddCmd(service *skillapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts skillapp.AddOptions

	cmd := &cobra.Command{
		Use:   "add <source>",
		Short: copy.AddShort,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ScopeProvided = cmd.Flags().Changed("global")
			return service.Add(cmd.Context(), args[0], opts)
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
func newSkillListCmd(service *skillapp.Service) *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   copy.ListShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.List(cmd.Context(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagListGlobal)
	return cmd
}

// newSkillInitCmd 构建技能模板初始化命令。
func newSkillInitCmd(service *skillapp.Service) *cobra.Command {
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
			return service.Init(cmd.Context(), name)
		},
	}
	return cmd
}

// newSkillRemoveCmd 构建移除技能命令。
func newSkillRemoveCmd(service *skillapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts skillapp.RemoveOptions

	cmd := &cobra.Command{
		Use:     "remove [skills...]",
		Aliases: []string{"rm"},
		Short:   copy.RemoveShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SkillNames = append(opts.SkillNames[:0], args...)
			return service.Remove(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, copy.FlagRemoveGlobal)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipConfirmation)
	cmd.Flags().BoolVar(&opts.All, "all", false, copy.FlagRemoveAll)
	cmd.Flags().StringSliceVarP(&opts.SkillNames, "skill", "s", nil, copy.FlagSpecificSkills)
	return cmd
}

// newSkillCheckCmd 构建检查更新命令。
func newSkillCheckCmd(service *skillapp.Service) *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: copy.CheckShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Check(cmd.Context(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagCheckGlobal)
	return cmd
}

// newSkillUpdateCmd 构建批量更新技能命令。
func newSkillUpdateCmd(service *skillapp.Service) *cobra.Command {
	copy := ui.Messages()
	var global bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: copy.UpdateShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Update(cmd.Context(), global)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, copy.FlagUpdateGlobal)
	return cmd
}
