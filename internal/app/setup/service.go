// Package setup provides the setup/init flow service for agent-notify.
// It handles agent configuration and hook installation.
package setup

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hellolib/agent-notify/internal/agentintegrations"
	"github.com/hellolib/agent-notify/internal/common"
	"github.com/hellolib/agent-notify/internal/config"
)

// Prompter interface for user interactions.
type Prompter interface {
	Select(message string, options []PromptOption, defaultValue string) (string, error)
	MultiSelect(message string, options []PromptOption, defaults []string) ([]string, error)
	Confirm(message string, defaultValue bool) (bool, error)
	Input(message, defaultValue string) (string, error)
}

// PromptOption represents a selectable option in prompts.
type PromptOption struct {
	Label string
	Value string
}

// FeishuPreparer prepares the Feishu CLI for use.
type FeishuPreparer interface {
	EnsureReady(ctx context.Context) error
}

// OutputWriter handles output messages.
type OutputWriter interface {
	Writef(format string, args ...any)
}

// Service handles the init/setup flow for agent-notify.
type Service struct {
	claudeIntegration agentintegrations.Integration
	codexIntegration  agentintegrations.Integration
	feishuPreparer    FeishuPreparer
	configLoader      ConfigLoader
}

// ConfigLoader loads and saves configuration.
type ConfigLoader interface {
	Load(path string) (config.Config, error)
	Save(path string, cfg config.Config) error
	DefaultPath() (string, error)
}

// SetupResult contains the result of a setup operation.
type SetupResult struct {
	Agent        string
	ConfigPath   string
	SettingsPath string
}

// NewService creates a new setup service.
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

// WithFeishuPreparer sets the Feishu preparer.
func WithFeishuPreparer(p FeishuPreparer) Option {
	return func(s *Service) { s.feishuPreparer = p }
}

// WithConfigLoader sets the config loader.
func WithConfigLoader(l ConfigLoader) Option {
	return func(s *Service) { s.configLoader = l }
}

var notificationEventOptions = []PromptOption{
	{Label: "需要授权 (permission_required)", Value: "permission_required"},
	{Label: "等待输入 (input_required)", Value: "input_required"},
	{Label: "任务完成 (run_completed)", Value: "run_completed"},
	{Label: "任务失败 (run_failed)", Value: "run_failed"},
}

// Run executes the init flow.
func (s *Service) Run(ctx context.Context, prompter Prompter, output OutputWriter, configPath, binaryPath string) (*SetupResult, error) {
	cfg, path, err := s.loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Detect installed agents
	var agentOptions []PromptOption
	var defaultAgent string
	if s.claudeIntegration.DetectInstalled() {
		agentOptions = append(agentOptions, PromptOption{Label: "Claude Code", Value: "claude"})
		if cfg.Agent.ClaudeCode.Enabled {
			defaultAgent = "claude"
		}
	}
	if s.codexIntegration.DetectInstalled() {
		agentOptions = append(agentOptions, PromptOption{Label: "Codex", Value: "codex"})
		if cfg.Agent.Codex.Enabled && defaultAgent == "" {
			defaultAgent = "codex"
		}
	}

	if len(agentOptions) == 0 {
		return nil, errors.New("未检测到 Claude Code 或 Codex，请先安装其中一个")
	}

	// If no agent is enabled, default to the first detected agent
	if defaultAgent == "" {
		defaultAgent = agentOptions[0].Value
	}

	// Step 1: Single-select for agent (default to enabled agent or first detected)
	selectedAgent, err := prompter.Select("选择要配置的 Agent", agentOptions, defaultAgent)
	if err != nil {
		return nil, err
	}

	// Step 2: Multi-select channels (default from current config)
	var currentChannels []string
	if selectedAgent == "claude" {
		if cfg.Notify.ClaudeCode.Channels.Feishu.Enabled {
			currentChannels = append(currentChannels, "feishu")
		}
		if cfg.Notify.ClaudeCode.Channels.System.Enabled {
			currentChannels = append(currentChannels, "system")
		}
	} else {
		if cfg.Notify.Codex.Channels.Feishu.Enabled {
			currentChannels = append(currentChannels, "feishu")
		}
		if cfg.Notify.Codex.Channels.System.Enabled {
			currentChannels = append(currentChannels, "system")
		}
	}

	channelChoices, err := prompter.MultiSelect(
		"启用通知渠道",
		[]PromptOption{{Label: "飞书", Value: "feishu"}, {Label: "系统通知", Value: "system"}},
		currentChannels,
	)
	if err != nil {
		return nil, err
	}

	// Step 3: Check if any channel selected
	feishuEnabled := slices.Contains(channelChoices, "feishu")
	systemEnabled := slices.Contains(channelChoices, "system")
	hasChannel := feishuEnabled || systemEnabled

	// If no channel selected, disable the agent's notification and return early
	if !hasChannel {
		return s.disableAgentNotification(cfg, path, selectedAgent, output)
	}

	// Step 4: If Claude Code: select events (default from current config)
	// If Codex: skip event selection entirely
	var events []string
	if selectedAgent == "claude" {
		// Get current events from config
		currentEvents := cfg.Notify.ClaudeCode.Events

		events, err = s.selectNotificationEvents(prompter, "通知事件", currentEvents)
		if err != nil {
			return nil, err
		}
		// If no events selected, disable the agent's notification and return early
		if len(events) == 0 {
			return s.disableAgentNotification(cfg, path, selectedAgent, output)
		}
	}

	// Step 5: Update the selected agent's notify config

	switch selectedAgent {
	case "claude":
		cfg.Notify.ClaudeCode.Channels.Feishu.Enabled = feishuEnabled
		cfg.Notify.ClaudeCode.Channels.System.Enabled = systemEnabled
		cfg.Notify.ClaudeCode.Events = dedupeStrings(events)

		if feishuEnabled {
			if err := s.prepareFeishu(ctx); err != nil {
				return nil, fmt.Errorf("飞书初始化失败: %w", err)
			}
		}

		agentScope := "user"
		if cfg.Agent.ClaudeCode.InstallScope == "project" {
			agentScope = "project"
		}

		agentSettingsPath, err := s.claudeIntegration.SettingsPath(agentScope)
		if err != nil {
			return nil, fmt.Errorf("获取 claude settings 路径失败: %w", err)
		}

		resolvedBinary := common.ResolveBinaryPath(binaryPath)
		if err := s.claudeIntegration.Install(agentSettingsPath, resolvedBinary); err != nil {
			return nil, fmt.Errorf("安装 claude hooks 失败: %w", err)
		}
		output.Writef("claude hooks 安装: %s\n", agentSettingsPath)
		cfg.Agent.ClaudeCode.InstallScope = agentScope
		cfg.Agent.ClaudeCode.Enabled = true

		if err := s.saveConfig(path, cfg); err != nil {
			return nil, err
		}
		output.Writef("配置文件: %s\n", path)

		return &SetupResult{
			Agent:        selectedAgent,
			ConfigPath:   path,
			SettingsPath: agentSettingsPath,
		}, nil

	case "codex":
		cfg.Notify.Codex.Channels.Feishu.Enabled = feishuEnabled
		cfg.Notify.Codex.Channels.System.Enabled = systemEnabled
		cfg.Notify.Codex.Events = nil // Codex doesn't support events

		if feishuEnabled {
			if err := s.prepareFeishu(ctx); err != nil {
				return nil, fmt.Errorf("飞书初始化失败: %w", err)
			}
		}

		agentScope := "user"
		if cfg.Agent.Codex.InstallScope == "project" {
			agentScope = "project"
		}

		agentSettingsPath, err := s.codexIntegration.SettingsPath(agentScope)
		if err != nil {
			return nil, fmt.Errorf("获取 codex settings 路径失败: %w", err)
		}

		resolvedBinary := common.ResolveBinaryPath(binaryPath)
		if err := s.codexIntegration.Install(agentSettingsPath, resolvedBinary); err != nil {
			return nil, fmt.Errorf("安装 codex notify 失败: %w", err)
		}
		output.Writef("codex notify 安装: %s\n", agentSettingsPath)
		cfg.Agent.Codex.InstallScope = agentScope
		cfg.Agent.Codex.Enabled = true

		if err := s.saveConfig(path, cfg); err != nil {
			return nil, err
		}
		output.Writef("配置文件: %s\n", path)

		return &SetupResult{
			Agent:        selectedAgent,
			ConfigPath:   path,
			SettingsPath: agentSettingsPath,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported agent: %s", selectedAgent)
	}
}

func (s *Service) loadConfig(configPath string) (config.Config, string, error) {
	path := configPath
	var err error
	if path == "" {
		path, err = s.defaultConfigPath()
		if err != nil {
			return config.Config{}, "", err
		}
	}
	cfg, err := s.loadConfigFile(path)
	if err != nil {
		return config.Config{}, "", err
	}
	return cfg, path, nil
}

func (s *Service) saveConfig(path string, cfg config.Config) error {
	if s.configLoader != nil {
		return s.configLoader.Save(path, cfg)
	}
	return config.Save(path, cfg)
}

func (s *Service) defaultConfigPath() (string, error) {
	if s.configLoader != nil {
		return s.configLoader.DefaultPath()
	}
	return config.DefaultPath()
}

func (s *Service) loadConfigFile(path string) (config.Config, error) {
	if s.configLoader != nil {
		return s.configLoader.Load(path)
	}
	return config.Load(path)
}

func (s *Service) prepareFeishu(ctx context.Context) error {
	if s.feishuPreparer != nil {
		return s.feishuPreparer.EnsureReady(ctx)
	}
	return nil
}

func (s *Service) selectNotificationEvents(prompter Prompter, message string, defaults []string) ([]string, error) {
	return prompter.MultiSelect(message, notificationEventOptions, defaults)
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// disableAgentNotification disables all notification channels for the given agent
// and saves the configuration. This is called when the user doesn't select any
// channels or events.
func (s *Service) disableAgentNotification(cfg config.Config, path, agent string, output OutputWriter) (*SetupResult, error) {
	switch agent {
	case "claude":
		cfg.Notify.ClaudeCode.Channels.Feishu.Enabled = false
		cfg.Notify.ClaudeCode.Channels.System.Enabled = false
		cfg.Notify.ClaudeCode.Events = nil
		cfg.Agent.ClaudeCode.Enabled = false
	case "codex":
		cfg.Notify.Codex.Channels.Feishu.Enabled = false
		cfg.Notify.Codex.Channels.System.Enabled = false
		cfg.Notify.Codex.Events = nil
		cfg.Agent.Codex.Enabled = false
	}

	if err := s.saveConfig(path, cfg); err != nil {
		return nil, err
	}
	output.Writef("%s 通知已关闭\n", agentName(agent))
	output.Writef("配置文件: %s\n", path)

	return &SetupResult{
		Agent:      agent,
		ConfigPath: path,
	}, nil
}

func agentName(agent string) string {
	switch agent {
	case "claude":
		return "Claude Code"
	case "codex":
		return "Codex"
	default:
		return agent
	}
}
