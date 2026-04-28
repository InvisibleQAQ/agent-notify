package codex

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/hellolib/agent-notify/internal/agenthooks"
	"github.com/hellolib/agent-notify/internal/config"
	"github.com/hellolib/agent-notify/internal/notify"
	"github.com/hellolib/agent-notify/internal/state"
)

func Handle(ctx context.Context, cfg config.Config, statePath, logPath string, stdin io.Reader) error {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return state.AppendLog(logPath, fmt.Sprintf("read stdin error: %v", err))
	}

	body := strings.TrimSpace(string(data))
	if body == "" {
		body = notify.DefaultBody("run_completed")
	}

	msg := notify.Message{
		Agent:     "codex",
		Event:     "run_completed",
		SessionID: uuid.NewString(),
		Title:     notify.FormatTitle("codex", "run_completed"),
		Body:      body,
	}
	return agenthooks.Dispatch(ctx, cfg, statePath, logPath, msg)
}
