package notify

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type Runner func(ctx context.Context, name string, args ...string) error

type MacOSSender struct {
	run Runner
}

func NewMacOSSender(run Runner) *MacOSSender {
	return &MacOSSender{run: run}
}

func DefaultRunner(ctx context.Context, name string, args ...string) error {
	return exec.CommandContext(ctx, name, args...).Run()
}

func (s *MacOSSender) Name() string { return "system" }

func (s *MacOSSender) Send(ctx context.Context, msg Message) error {
	// Use terminal-notifier if available for better notifications with icon support
	if s.tryTerminalNotifier(ctx, msg) {
		return nil
	}

	// Fallback to osascript with improved content
	formattedBody := s.formatBody(msg)
	script := fmt.Sprintf(`display notification %q with title %q sound name "Submarine"`, formattedBody, msg.Title)
	return s.run(ctx, "osascript", "-e", script)
}

// tryTerminalNotifier attempts to use terminal-notifier for richer notifications
func (s *MacOSSender) tryTerminalNotifier(ctx context.Context, msg Message) bool {
	// Check if terminal-notifier is installed
	if err := s.run(ctx, "which", "terminal-notifier"); err != nil {
		return false
	}

	args := []string{
		"-title", msg.Title,
		"-message", s.formatBody(msg),
		"-sound", "Submarine",
		"-group", "com.claude-code.notify",
	}

	// Use Claude app icon if available
	args = append(args, "-appIcon", "/Applications/Claude.app/Contents/Resources/AppIcon.icns")

	// terminal-notifier returns 0 on success, non-zero on failure
	return s.run(ctx, "terminal-notifier", args...) == nil
}

// formatBody formats the notification body with timestamp
func (s *MacOSSender) formatBody(msg Message) string {
	timestamp := time.Now().Format("15:04:05")
	if msg.Workspace != "" {
		return fmt.Sprintf("📁 %s\n%s\n⏰ %s", msg.Workspace, msg.Body, timestamp)
	}
	return fmt.Sprintf("%s\n⏰ %s", msg.Body, timestamp)
}
