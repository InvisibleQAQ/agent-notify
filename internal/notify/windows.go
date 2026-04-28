package notify

import (
	"context"
	"fmt"
	"time"
)

type WindowsSender struct {
	run Runner
}

func NewWindowsSender(run Runner) *WindowsSender {
	return &WindowsSender{run: run}
}

func (s *WindowsSender) Name() string { return "system" }

func (s *WindowsSender) Send(ctx context.Context, msg Message) error {
	// Use PowerShell with Windows Toast notification
	// We use a simple approach with Windows.Forms.NotifyIcon

	formattedBody := s.formatBody(msg)

	// PowerShell script to show toast notification
	// Using System.Windows.Forms.NotifyIcon for balloon tip notifications
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$notification = New-Object System.Windows.Forms.NotifyIcon
$notification.Icon = [System.Drawing.SystemIcons]::Information
$notification.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$notification.BalloonTipTitle = %q
$notification.BalloonTipText = %q
$notification.Visible = $true
$notification.ShowBalloonTip(5000)
`, msg.Title, formattedBody)

	return s.run(ctx, "powershell", "-Command", psScript)
}

func (s *WindowsSender) formatBody(msg Message) string {
	timestamp := time.Now().Format("15:04:05")
	if msg.Workspace != "" {
		return fmt.Sprintf("%s\n%s\n%s", msg.Workspace, msg.Body, timestamp)
	}
	return fmt.Sprintf("%s\n%s", msg.Body, timestamp)
}
