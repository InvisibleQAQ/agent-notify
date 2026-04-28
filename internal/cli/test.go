package cli

import (
	"context"
	"fmt"

	"github.com/hellolib/agent-notify/internal/app/tester"
	"github.com/spf13/cobra"
)

func newTestCmd(ctx context.Context, streams Streams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Send test notifications",
	}
	cmd.AddCommand(
		newTestFeishuCmd(ctx, streams),
		newTestSystemCmd(ctx, streams),
	)
	return cmd
}

func newTestFeishuCmd(ctx context.Context, streams Streams) *cobra.Command {
	return &cobra.Command{
		Use:   "feishu",
		Short: "Send a Feishu test notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use the tester service
			svc := tester.NewService(
				tester.WithFeishuPreparer(&feishuPreparer{}),
			)
			result, err := svc.TestFeishu(ctx)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Stdout, result.Message)
			return err
		},
	}
}

func newTestSystemCmd(ctx context.Context, streams Streams) *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Send a system test notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := tester.NewService()
			result, err := svc.TestSystem(ctx)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(streams.Stdout, result.Message)
			return err
		},
	}
}

// feishuPreparer implements tester.FeishuPreparer
type feishuPreparer struct{}

func (p *feishuPreparer) EnsureReady(ctx context.Context) error {
	return prepareFeishuCLI(ctx)
}
