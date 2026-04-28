package notify

import (
	"runtime"
)

// NewSystemSender returns the appropriate system notification sender for the current platform.
// - darwin: uses macOS notifications (osascript/terminal-notifier)
// - linux: uses notify-send
// - windows: uses PowerShell with Windows Forms
// - other: returns an explicit unsupported sender
func NewSystemSender(run Runner) Sender {
	switch runtime.GOOS {
	case "darwin":
		return NewMacOSSender(run)
	case "linux":
		return NewLinuxSender(run)
	case "windows":
		return NewWindowsSender(run)
	default:
		return NewUnsupportedSender(runtime.GOOS)
	}
}
