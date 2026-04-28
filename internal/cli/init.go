package cli

import (
	"context"
	"strings"

	"github.com/hellolib/agent-notify/internal/common"
	"github.com/hellolib/agent-notify/internal/feishucli"
	"github.com/spf13/cobra"
)

var prepareFeishuCLI = func(ctx context.Context) error {
	_, err := feishucli.EnsureReady(ctx)
	return err
}

func newInitCmd(streams Streams) *cobra.Command {
	var configPath string
	var settingsPath string
	var binaryPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize config and install Claude hooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			prompter, err := newPrompter(streams)
			if err != nil {
				return err
			}
			return runInitFlow(cmd.Context(), streams, prompter, configPath, settingsPath, firstNonEmpty(binaryPath))
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "config path")
	cmd.Flags().StringVar(&settingsPath, "settings", "", "claude settings path")
	cmd.Flags().StringVar(&binaryPath, "binary", common.ResolveBinaryPath(""), "installed agent-notify binary path")
	return cmd
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
