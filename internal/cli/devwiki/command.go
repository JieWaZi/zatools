package devwikicmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"zatools/internal/app/devwikiapp"
	"zatools/internal/devwiki"
	"zatools/internal/ui"
)

// SuppressLogoAnnotation marks commands that should not emit the CLI logo before running.
const SuppressLogoAnnotation = "zatools.io/suppress-logo"

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
	devwikiCmd.AddCommand(newUpdateCmd(service))
	devwikiCmd.AddCommand(newReadCmd(service))
	devwikiCmd.AddCommand(newSearchCmd(service))
	devwikiCmd.AddCommand(newRepoCmd(service))
	devwikiCmd.AddCommand(newCheckCmd(service))
	devwikiCmd.AddCommand(newGraphCmd(service))
	devwikiCmd.AddCommand(newServerCmd(service))
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

func newReadCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.ReadOptions

	cmd := &cobra.Command{
		Use:         "read <topic|workflow> <slug>",
		Short:       copy.DevwikiReadShort,
		Args:        cobra.ExactArgs(2),
		Annotations: map[string]string{SuppressLogoAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Kind = args[0]
			opts.Slug = args[1]
			opts.Stdout = cmd.OutOrStdout()
			return service.Read(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Root, "root", ".", copy.FlagDevwikiRoot)
	cmd.Flags().StringVar(&opts.Project, "project", "", copy.FlagDevwikiProject)
	cmd.Flags().StringVar(&opts.View, "view", "card", copy.FlagDevwikiReadView)
	cmd.Flags().StringVar(&opts.Format, "format", "text", copy.FlagDevwikiReadFormat)
	return cmd
}

func newSearchCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.SearchOptions

	cmd := &cobra.Command{
		Use:         "search <index|glossary|topic|workflow> <query...>",
		Short:       copy.DevwikiSearchShort,
		Args:        cobra.MinimumNArgs(2),
		Annotations: map[string]string{SuppressLogoAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Kind = args[0]
			opts.QueryTerms = args[1:]
			opts.Stdout = cmd.OutOrStdout()
			return service.Search(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Root, "root", ".", copy.FlagDevwikiRoot)
	cmd.Flags().StringVar(&opts.Project, "project", "", copy.FlagDevwikiProject)
	return cmd
}

func newRepoCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	cmd := &cobra.Command{
		Use:           "repo",
		Short:         copy.DevwikiRepoShort,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	cmd.AddCommand(newRepoAddCmd(service))
	cmd.AddCommand(newRepoLinkCmd(service))
	cmd.AddCommand(newRepoInfoCmd(service))
	return cmd
}

func newRepoAddCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.RepoAddOptions
	cmd := &cobra.Command{
		Use:   "add <project> [path]",
		Short: copy.DevwikiRepoAddShort,
		Args:  repoArgs("DevWiki repo add", cobra.RangeArgs(1, 2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ProjectSlug = args[0]
			if len(args) > 1 {
				opts.LocalPath = args[1]
			}
			opts.Stdout = cmd.OutOrStdout()
			return runRepoCommand(cmd, "DevWiki repo add", func() error {
				return service.RepoAdd(cmd.Context(), opts)
			})
		},
	}
	cmd.Flags().StringVar(&opts.RemoteURL, "remote", "", copy.FlagDevwikiRemote)
	return cmd
}

func newRepoLinkCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.RepoLinkOptions
	return &cobra.Command{
		Use:   "link <project> <repo-slug> <path>",
		Short: copy.DevwikiRepoLinkShort,
		Args:  repoArgs("DevWiki repo link", cobra.ExactArgs(3)),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ProjectSlug = args[0]
			opts.RepoSlug = args[1]
			opts.Path = args[2]
			opts.Stdout = cmd.OutOrStdout()
			return runRepoCommand(cmd, "DevWiki repo link", func() error {
				return service.RepoLink(cmd.Context(), opts)
			})
		},
	}
}

func runRepoCommand(cmd *cobra.Command, label string, run func() error) error {
	err := run()
	if err == nil {
		return nil
	}
	_, _ = cmd.ErrOrStderr().Write([]byte(formatRepoFailure(label, err)))
	return err
}

func repoArgs(label string, validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := validate(cmd, args)
		if err != nil {
			_, _ = cmd.ErrOrStderr().Write([]byte(formatRepoFailure(label, err)))
		}
		return err
	}
}

func formatRepoFailure(label string, err error) string {
	return fmt.Sprintf(ui.Messages().DevwikiRepoFailureFmt, label, err)
}

func newRepoInfoCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.RepoInfoOptions
	cmd := &cobra.Command{
		Use:         "info [project]",
		Short:       copy.DevwikiRepoInfoShort,
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{SuppressLogoAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ProjectSlug = args[0]
			}
			opts.Stdout = cmd.OutOrStdout()
			return service.RepoInfo(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Format, "format", "json", copy.FlagDevwikiFormat)
	return cmd
}

func newCheckCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.CheckOptions
	cmd := &cobra.Command{
		Use:   "check [document|graph] [path...]",
		Short: copy.DevwikiCheckShort,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && (args[0] == "document" || args[0] == "graph") {
				opts.Types = []string{args[0]}
				opts.Paths = args[1:]
			} else {
				opts.Paths = args
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Stdout = cmd.OutOrStdout()
			return service.Check(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Root, "root", ".", ui.Messages().FlagDevwikiRoot)
	return cmd
}

func newGraphCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.GraphOptions

	cmd := &cobra.Command{
		Use:   "graph",
		Short: copy.DevwikiGraphShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Graph(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Root, "root", ".", copy.FlagDevwikiRoot)
	cmd.Flags().StringVar(&opts.Project, "project", "", copy.FlagDevwikiProject)
	cmd.Flags().StringVar(&opts.Host, "host", "127.0.0.1", copy.FlagDevwikiGraphHost)
	cmd.Flags().IntVar(&opts.Port, "port", 0, copy.FlagDevwikiGraphPort)
	cmd.Flags().BoolVar(&opts.NoOpen, "no-open", false, copy.FlagDevwikiGraphNoOpen)
	cmd.Flags().BoolVar(&opts.Force, "force", false, copy.FlagDevwikiGraphForce)
	cmd.Flags().BoolVar(&opts.Check, "check", false, copy.FlagDevwikiGraphCheck)
	return cmd
}

func newServerCmd(service *devwikiapp.Service) *cobra.Command {
	copy := ui.Messages()
	var opts devwikiapp.ServerOptions

	cmd := &cobra.Command{
		Use:   "server",
		Short: copy.DevwikiServerShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Stdout = cmd.OutOrStdout()
			return service.Server(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Root, "root", ".", copy.FlagDevwikiRoot)
	cmd.Flags().StringVar(&opts.Project, "project", "", copy.FlagDevwikiProject)
	cmd.Flags().StringVar(&opts.Host, "host", "0.0.0.0", copy.FlagDevwikiGraphHost)
	cmd.Flags().IntVar(&opts.Port, "port", 5697, copy.FlagDevwikiGraphPort)
	return cmd
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
