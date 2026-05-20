package devwikicmd

import (
	"encoding/json"
	"io"
	"time"

	"github.com/spf13/cobra"

	"zatools/internal/app/devwikiapp"
	"zatools/internal/devwiki"
	"zatools/internal/ui"
)

// NewCommand 构建 `devwiki` 子命令及其初始化入口。
func NewCommand() *cobra.Command {
	copy := ui.Messages()
	service := devwikiapp.NewService()

	devwikiCmd := &cobra.Command{
		Use:           "devwiki",
		Short:         copy.DevwikiShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	devwikiCmd.AddCommand(newInitCmd(service))
	devwikiCmd.AddCommand(newLinkCmd(service))
	devwikiCmd.AddCommand(newUpdateCmd(service))
	devwikiCmd.AddCommand(newToolCmd())
	return devwikiCmd
}

func newInitCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.InitOptions

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: copy.DevwikiInitShort,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ProjectName = args[0]
			}
			opts.ScopeProvided = cmd.Flags().Changed("global")
			return service.Init(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Agent, "agent", "", copy.FlagDevwikiAgent)
	cmd.Flags().StringSliceVar(&opts.CodeDirs, "code-dir", nil, copy.FlagDevwikiCodeDir)
	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, copy.FlagInstallGlobally)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipPrompts)
	return cmd
}

func newLinkCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.LinkOptions

	cmd := &cobra.Command{
		Use:   "link",
		Short: copy.DevwikiLinkShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Link(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.DevwikiRoot, "root", "", copy.FlagDevwikiRoot)
	cmd.Flags().StringVar(&opts.Agent, "agent", "", copy.FlagDevwikiAgent)
	cmd.Flags().StringSliceVar(&opts.CodeDirs, "code-dir", nil, copy.FlagDevwikiCodeDir)
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, copy.FlagSkipPrompts)
	return cmd
}

func newUpdateCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	return &cobra.Command{
		Use:   "update",
		Short: copy.UpdateShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Update(cmd.Context())
		},
	}
}

func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "tool",
		Short:         "Run built-in DevWiki maintenance tools",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(newToolResetCmd())
	cmd.AddCommand(newToolLogCmd())
	return cmd
}

func newToolResetCmd() *cobra.Command {
	var scope string
	var projectRoot string
	var yes bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset generated wiki content for selected scopes",
		RunE: func(cmd *cobra.Command, args []string) error {
			scopes, err := devwiki.ParseResetScopes(scope)
			if err != nil {
				return err
			}

			plan, err := devwiki.BuildResetPlan(projectRoot, scopes)
			if err != nil {
				return err
			}

			if !yes {
				return writeJSON(cmd.OutOrStdout(), plan)
			}

			result, err := devwiki.ApplyResetPlan(plan)
			if err != nil {
				return err
			}
			return writeJSON(cmd.OutOrStdout(), result)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "", "Reset scopes: wiki,raw,log,checkpoints,all")
	cmd.Flags().StringVar(&projectRoot, "project-root", ".", "Project root")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Apply the reset plan")
	_ = cmd.MarkFlagRequired("scope")
	return cmd
}

func newToolLogCmd() *cobra.Command {
	var wikiRoot string
	var message string

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Append an entry to wiki/log.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwiki.AppendLog(wikiRoot, message, time.Now())
		},
	}
	cmd.Flags().StringVar(&wikiRoot, "wiki-root", "wiki", "Wiki root")
	cmd.Flags().StringVar(&message, "message", "", "Log message")
	_ = cmd.MarkFlagRequired("message")
	return cmd
}

func writeJSON(w io.Writer, payload any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
