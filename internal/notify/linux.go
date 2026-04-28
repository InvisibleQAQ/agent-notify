package notify

import (
	"context"
	"fmt"
	"time"
)

type LinuxSender struct {
	run Runner
}

func NewLinuxSender(run Runner) *LinuxSender {
	return &LinuxSender{run: run}
}

func (s *LinuxSender) Name() string { return "system" }

func (s *LinuxSender) Send(ctx context.Context, msg Message) error {
	// Use notify-send for Linux notifications
	// Format: notify-send "Title" "Body" [options]

	formattedBody := s.formatBody(msg)

	// notify-send arguments:
	// -a "Claude Code" sets app name
	// -u normal sets urgency
	// -t 5000 sets timeout in milliseconds (5 seconds)
	return s.run(ctx, "notify-send",
		"-a", "Claude Code",
		"-u", "normal",
		"-t", "5000",
		msg.Title,
		formattedBody,
	)
}

func (s *LinuxSender) formatBody(msg Message) string {
	timestamp := time.Now().Format("15:04:05")
	if msg.Workspace != "" {
		return fmt.Sprintf("%s\n%s\n%s", msg.Workspace, msg.Body, timestamp)
	}
	return fmt.Sprintf("%s\n%s", msg.Body, timestamp)
}
