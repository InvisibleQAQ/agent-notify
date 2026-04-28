// Package doctor provides the diagnostics service for agent-notify.
// It checks the current notification setup and reports status.
package doctor

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/hellolib/agent-notify/internal/agentintegrations"
	"github.com/hellolib/agent-notify/internal/config"
	"github.com/hellolib/agent-notify/internal/feishucli"
)

// OutputWriter handles output messages.
type OutputWriter interface {
	Writef(format string, args ...any)
}

// Service handles diagnostics for agent-notify.
type Service struct {
	claudeIntegration agentintegrations.Integration
	codexIntegration  agentintegrations.Integration
}

// NewService creates a new doctor service.
func NewService(opts ...Option) *Service {
	s := &Service{
		claudeIntegration: agentintegrations.NewClaudeIntegration(),
		codexIntegration:  agentintegrations.NewCodexIntegration(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Option configures the service.
type Option func(*Service)

// WithClaudeIntegration sets the Claude integration.
func WithClaudeIntegration(i agentintegrations.Integration) Option {
	return func(s *Service) { s.claudeIntegration = i }
}

// WithCodexIntegration sets the Codex integration.
func WithCodexIntegration(i agentintegrations.Integration) Option {
	return func(s *Service) { s.codexIntegration = i }
}

type DiagnosticStatus string

const (
	StatusInstalled          DiagnosticStatus = "installed"
	StatusAgentMissing       DiagnosticStatus = "agent_missing"
	StatusConfigMissing      DiagnosticStatus = "config_missing"
	StatusIntegrationMissing DiagnosticStatus = "integration_missing"
)

// DiagnosticsResult contains diagnostic results.
type DiagnosticsResult struct {
	ConfigPath              string
	ConfigExists            bool
	ClaudeInstalled         bool
	ClaudeHookInstalled     bool
	CodexInstalled          bool
	CodexNotifyInstalled    bool
	SystemNotifyAvailable   bool
	SystemNotifyName        string
	FeishuCLIReady          bool
	ClaudeFeishuEnabled     bool
	ClaudeSystemEnabled     bool
	CodexFeishuEnabled      bool
	CodexSystemEnabled      bool
	ClaudeIntegrationStatus DiagnosticStatus
	CodexIntegrationStatus  DiagnosticStatus
}

// Run executes diagnostics and returns results.
func (s *Service) Run() (*DiagnosticsResult, error) {
	result := &DiagnosticsResult{}

	// Detect agents
	result.ClaudeInstalled = s.claudeIntegration.DetectInstalled()
	result.CodexInstalled = s.codexIntegration.DetectInstalled()

	// System notification detection
	result.SystemNotifyAvailable, result.SystemNotifyName = detectSystemNotification()

	// Config
	cfgPath, _ := config.DefaultPath()
	result.ConfigPath = cfgPath
	cfg, cfgLoadErr := config.Load(cfgPath)
	_, cfgErr := os.Stat(cfgPath)
	result.ConfigExists = cfgErr == nil

	// Claude hooks settings
	claudeSettingsPath, _ := s.claudeIntegration.SettingsPath("user")
	if claudeSettingsPath != "" {
		installed, err := s.claudeIntegration.IsHookInstalled(claudeSettingsPath)
		result.ClaudeHookInstalled = err == nil && installed
	}

	// Codex notify settings
	codexSettingsPath, _ := s.codexIntegration.SettingsPath("user")
	if codexSettingsPath != "" {
		installed, err := s.codexIntegration.IsHookInstalled(codexSettingsPath)
		result.CodexNotifyInstalled = err == nil && installed
	}

	// Config values
	result.ClaudeFeishuEnabled = cfgLoadErr == nil && cfg.Notify.ClaudeCode.Channels.Feishu.Enabled
	result.ClaudeSystemEnabled = cfgLoadErr == nil && cfg.Notify.ClaudeCode.Channels.System.Enabled
	result.CodexFeishuEnabled = cfgLoadErr == nil && cfg.Notify.Codex.Channels.Feishu.Enabled
	result.CodexSystemEnabled = cfgLoadErr == nil && cfg.Notify.Codex.Channels.System.Enabled

	result.ClaudeIntegrationStatus = integrationStatus(result.ConfigExists, result.ClaudeInstalled, result.ClaudeHookInstalled)
	result.CodexIntegrationStatus = integrationStatus(result.ConfigExists, result.CodexInstalled, result.CodexNotifyInstalled)

	// Feishu CLI
	_, feishuCLIConfigErr := feishucli.ParseConfig()
	result.FeishuCLIReady = feishuCLIConfigErr == nil

	return result, nil
}

func integrationStatus(configExists, agentInstalled, integrationInstalled bool) DiagnosticStatus {
	if !agentInstalled {
		return StatusAgentMissing
	}
	if !configExists {
		return StatusConfigMissing
	}
	if !integrationInstalled {
		return StatusIntegrationMissing
	}
	return StatusInstalled
}

// Print outputs the diagnostics result.
func (s *Service) Print(output OutputWriter, result *DiagnosticsResult) {
	// Config path header
	output.Writef("配置文件: %s\n\n", result.ConfigPath)

	// Agent installation status table
	output.Writef("【Agent 安装状态】\n")
	output.Writef("┌─────────────┬──────────┬────────────────┐\n")
	output.Writef("│ Agent       │ 安装状态 │ 集成配置       │\n")
	output.Writef("├─────────────┼──────────┼────────────────┤\n")

	claudeInstallStatus := "❌ 未安装"
	if result.ClaudeInstalled {
		claudeInstallStatus = "✅ 已安装"
	}
	claudeHookStatus := padRight(diagnosticStatusLabel(result.ClaudeIntegrationStatus), 14)
	output.Writef("│ Claude Code │ %s │ %s │\n", claudeInstallStatus, claudeHookStatus)

	codexInstallStatus := "❌ 未安装"
	if result.CodexInstalled {
		codexInstallStatus = "✅ 已安装"
	}
	codexNotifyStatus := padRight(diagnosticStatusLabel(result.CodexIntegrationStatus), 14)
	output.Writef("│ Codex       │ %s │ %s │\n", codexInstallStatus, codexNotifyStatus)

	output.Writef("└─────────────┴──────────┴────────────────┘\n")
	output.Writef("\n")

	// Notification channels table
	output.Writef("【通知渠道状态】\n")
	output.Writef("┌─────────────┬────────┬────────┐\n")
	output.Writef("│ Agent       │ 飞书   │ 系统   │\n")
	output.Writef("├─────────────┼────────┼────────┤\n")

	claudeFeishu := "❌"
	if result.ClaudeFeishuEnabled {
		claudeFeishu = "✅"
	}
	claudeSystem := "❌"
	if result.ClaudeSystemEnabled {
		claudeSystem = "✅"
	}
	output.Writef("│ Claude Code │   %s    │   %s    │\n", claudeFeishu, claudeSystem)

	codexFeishu := "❌"
	if result.CodexFeishuEnabled {
		codexFeishu = "✅"
	}
	codexSystem := "❌"
	if result.CodexSystemEnabled {
		codexSystem = "✅"
	}
	output.Writef("│ Codex       │   %s    │   %s    │\n", codexFeishu, codexSystem)

	output.Writef("└─────────────┴────────┴────────┘\n")
	output.Writef("\n")

	// System environment table
	output.Writef("【系统环境】\n")
	output.Writef("┌─────────────────────┬──────────┐\n")
	output.Writef("│ 检查项              │ 状态     │\n")
	output.Writef("├─────────────────────┼──────────┤\n")

	configStatus := "❌ 不存在"
	if result.ConfigExists {
		configStatus = "✅ 已存在"
	}
	output.Writef("│ 配置文件            │ %s │\n", configStatus)

	systemNotifyStatus := "❌ 不可用"
	if result.SystemNotifyAvailable {
		systemNotifyStatus = "✅ 可用"
	}
	output.Writef("│ %s │ %s │\n", padRight(result.SystemNotifyName, 20), systemNotifyStatus)

	feishuCLIStatus := "❌ 未配置"
	if result.FeishuCLIReady {
		feishuCLIStatus = "✅ 已就绪"
	}
	output.Writef("│ 飞书 CLI            │ %s │\n", feishuCLIStatus)

	output.Writef("└─────────────────────┴──────────┘\n")
}

// detectSystemNotification checks if system notifications are available.
// Returns (available, displayName) where displayName is platform-specific.
func detectSystemNotification() (bool, string) {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("osascript")
		return err == nil, "系统通知"
	case "linux":
		_, err := exec.LookPath("notify-send")
		return err == nil, "系统通知"
	case "windows":
		// PowerShell is always available on Windows
		return true, "系统通知"
	default:
		return false, "系统通知"
	}
}

// visualWidth calculates the visual width of a string, treating Chinese characters as 2 columns.
func visualWidth(s string) int {
	width := 0
	for _, r := range s {
		if utf8.RuneLen(r) > 1 {
			// Chinese and other wide characters
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

// padRight pads a string to the target visual width.
func padRight(s string, targetWidth int) string {
	currentWidth := visualWidth(s)
	if currentWidth >= targetWidth {
		return s
	}
	padding := targetWidth - currentWidth
	return s + strings.Repeat(" ", padding)
}

func diagnosticStatusLabel(status DiagnosticStatus) string {
	switch status {
	case StatusInstalled:
		return "✅ 已安装"
	case StatusAgentMissing:
		return "❌ 未安装 Agent"
	case StatusConfigMissing:
		return "❌ 缺少配置"
	case StatusIntegrationMissing:
		return "❌ 未集成"
	default:
		return "❌ 未知"
	}
}
