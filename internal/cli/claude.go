package cli

import (
	"github.com/hellolib/agent-notify/internal/common"
	"github.com/spf13/cobra"
)

func newClaudeCmd(streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude Code hook integration",
	}
	cmd.AddCommand(newClaudePrintHooksCmd(streams), newClaudeInstallHooksCmd())
	return cmd
}

func newClaudePrintHooksCmd(streams Streams) *cobra.Command {
	var binaryPath string

	cmd := &cobra.Command{
		Use:   "print-hooks",
		Short: "Print Claude Code hook settings JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrintClaudeHooks(streams, firstNonEmpty(binaryPath))
		},
	}
	cmd.Flags().StringVar(&binaryPath, "binary", common.ResolveBinaryPath(""), "agent-notify binary path")
	return cmd
}

func newClaudeInstallHooksCmd() *cobra.Command {
	var binaryPath string
	var scope string

	cmd := &cobra.Command{
		Use:   "install-hooks",
		Short: "Install Claude Code hook settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallClaudeHooks(scope, firstNonEmpty(binaryPath))
		},
	}
	cmd.Flags().StringVar(&binaryPath, "binary", common.ResolveBinaryPath(""), "agent-notify binary path")
	cmd.Flags().StringVar(&scope, "scope", "user", "install scope")
	return cmd
}
