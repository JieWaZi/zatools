package updatecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"zatools/internal/app/updateapp"
	"zatools/internal/ui"
)

type updater interface {
	Update(context.Context) (updateapp.Result, error)
}

// NewCommand constructs the top-level `update` command.
func NewCommand() *cobra.Command {
	return NewCommandWithService(updateapp.NewService(updateapp.ServiceOptions{}))
}

// NewCommandWithService constructs the command with an injected service for tests.
func NewCommandWithService(service updater) *cobra.Command {
	copy := ui.Messages()
	cmd := &cobra.Command{
		Use:   "update",
		Short: copy.SelfUpdateShort,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := service.Update(cmd.Context())
			if err != nil {
				return err
			}
			if result.Deferred {
				fmt.Fprintf(cmd.OutOrStdout(), copy.SelfUpdateDeferredFmt, ui.Green, ui.Reset, result.Asset, result.Path)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), copy.SelfUpdateDoneFmt, ui.Green, ui.Reset, result.Asset, result.Path)
			return nil
		},
	}
	return cmd
}
