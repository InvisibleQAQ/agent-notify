package agenthooks

import (
	"testing"

	"github.com/hellolib/agent-notify/internal/config"
	"github.com/hellolib/agent-notify/internal/notify"
)

func TestBuildSendersUsesClaudeCodeConfigByDefault(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.ClaudeCode.Channels.System.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_completed"}
	cfg.Notify.ClaudeCode.Channels.Feishu.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_completed"}
	cfg.Notify.Codex.Channels.System.Enabled = false
	cfg.Notify.Codex.Channels.Feishu.Enabled = false

	senders := buildSenders(cfg, notify.Message{Event: "run_completed"})

	if len(senders) != 2 {
		t.Fatalf("len(senders) = %d, want 2", len(senders))
	}
	if senders[0].Name() != "system" {
		t.Fatalf("senders[0] = %q, want system", senders[0].Name())
	}
	if senders[1].Name() != "feishu" {
		t.Fatalf("senders[1] = %q, want feishu", senders[1].Name())
	}
}

func TestBuildSendersUsesCodexConfigForCodexMessages(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.ClaudeCode.Channels.System.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_completed"}
	cfg.Notify.ClaudeCode.Channels.Feishu.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_completed"}
	cfg.Notify.Codex.Channels.System.Enabled = true
	cfg.Notify.Codex.Channels.Feishu.Enabled = false

	senders := buildSenders(cfg, notify.Message{Agent: "codex", Event: "run_completed"})

	if len(senders) != 1 {
		t.Fatalf("len(senders) = %d, want 1", len(senders))
	}
	if senders[0].Name() != "system" {
		t.Fatalf("senders[0] = %q, want system", senders[0].Name())
	}
}

func TestBuildSendersIgnoresClaudeCodeEventListsForCodexMessages(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.ClaudeCode.Channels.System.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_failed"}
	cfg.Notify.ClaudeCode.Channels.Feishu.Enabled = true
	cfg.Notify.ClaudeCode.Events = []string{"run_failed"}
	cfg.Notify.Codex.Channels.System.Enabled = true
	cfg.Notify.Codex.Channels.Feishu.Enabled = true

	senders := buildSenders(cfg, notify.Message{Agent: "codex", Event: "run_completed"})

	if len(senders) != 2 {
		t.Fatalf("len(senders) = %d, want 2", len(senders))
	}
	if senders[0].Name() != "system" {
		t.Fatalf("senders[0] = %q, want system", senders[0].Name())
	}
	if senders[1].Name() != "feishu" {
		t.Fatalf("senders[1] = %q, want feishu", senders[1].Name())
	}
}
