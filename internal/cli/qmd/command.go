package qmdcmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"zatools/internal/qmd"
	"zatools/internal/ui"
)

// NewCommand constructs the top-level `qmd` helper command.
func NewCommand() *cobra.Command {
	copy := ui.Messages()
	defaultModels := qmd.DefaultModels()

	cmd := &cobra.Command{
		Use:                "qmd",
		Short:              copy.QMDShort,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			models, passthrough, err := parseQMDArgs(args, defaultModels)
			if err != nil {
				return err
			}
			if len(passthrough) > 0 && passthrough[0] == "sync" {
				return runQMDSync(cmd, passthrough[1:], models)
			}
			if len(passthrough) > 0 && passthrough[0] == "download" {
				return runQMDDownload(cmd, passthrough[1:], models)
			}
			return qmd.RunCommand(cmd.Context(), passthrough, models, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	return cmd
}

func parseQMDArgs(args []string, defaults qmd.Models) (qmd.Models, []string, error) {
	models := defaults
	passthrough := make([]string, 0, len(args))

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--":
			passthrough = append(passthrough, args[index+1:]...)
			return models, passthrough, nil
		case strings.HasPrefix(arg, "--embed-model="):
			models.EmbedModel = strings.TrimPrefix(arg, "--embed-model=")
		case arg == "--embed-model":
			if index+1 >= len(args) {
				return qmd.Models{}, nil, fmt.Errorf("--embed-model requires a value")
			}
			index++
			models.EmbedModel = args[index]
		case strings.HasPrefix(arg, "--rerank-model="):
			models.RerankModel = strings.TrimPrefix(arg, "--rerank-model=")
		case arg == "--rerank-model":
			if index+1 >= len(args) {
				return qmd.Models{}, nil, fmt.Errorf("--rerank-model requires a value")
			}
			index++
			models.RerankModel = args[index]
		case strings.HasPrefix(arg, "--generate-model="):
			models.GenerateModel = strings.TrimPrefix(arg, "--generate-model=")
		case arg == "--generate-model":
			if index+1 >= len(args) {
				return qmd.Models{}, nil, fmt.Errorf("--generate-model requires a value")
			}
			index++
			models.GenerateModel = args[index]
		default:
			passthrough = append(passthrough, arg)
		}
	}

	return models, passthrough, nil
}

func runQMDSync(cmd *cobra.Command, args []string, models qmd.Models) error {
	root, apply, err := parseQMDSyncArgs(args)
	if err != nil {
		return err
	}

	collections, err := qmd.LoadCollections(root)
	if err != nil {
		return err
	}
	commands, err := qmd.BuildCollectionCommands(root, collections)
	if err != nil {
		return err
	}
	if !apply {
		for _, command := range commands {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), strings.Join(command, " ")); err != nil {
				return err
			}
		}
		return nil
	}
	return qmd.RunCollectionCommands(cmd.Context(), commands, models, cmd.OutOrStdout(), cmd.ErrOrStderr())
}

func runQMDDownload(cmd *cobra.Command, args []string, models qmd.Models) error {
	root, err := parseQMDDownloadArgs(args)
	if err != nil {
		return err
	}
	return qmd.RunDownload(cmd.Context(), root, models, cmd.OutOrStdout(), cmd.ErrOrStderr())
}

func parseQMDSyncArgs(args []string) (string, bool, error) {
	root := "."
	apply := false

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--apply":
			apply = true
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		case arg == "--root":
			if index+1 >= len(args) {
				return "", false, fmt.Errorf("--root requires a value")
			}
			index++
			root = args[index]
		default:
			return "", false, fmt.Errorf("unknown sync argument %q", arg)
		}
	}

	return root, apply, nil
}

func parseQMDDownloadArgs(args []string) (string, error) {
	root := "."

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		case arg == "--root":
			if index+1 >= len(args) {
				return "", fmt.Errorf("--root requires a value")
			}
			index++
			root = args[index]
		default:
			return "", fmt.Errorf("unknown download argument %q", arg)
		}
	}

	return root, nil
}
