package agentintegrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hellolib/agent-notify/internal/common"
)

// ClaudeIntegration implements Integration for Claude Code.
type ClaudeIntegration struct{}

// NewClaudeIntegration creates a new Claude Code integration.
func NewClaudeIntegration() *ClaudeIntegration {
	return &ClaudeIntegration{}
}

// Name returns the display name for Claude Code.
func (c *ClaudeIntegration) Name() string {
	return "Claude Code"
}

// DetectInstalled checks if the claude CLI is installed.
func (c *ClaudeIntegration) DetectInstalled() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// SettingsPath returns the path to Claude Code's settings.json file.
func (c *ClaudeIntegration) SettingsPath(scope string) (string, error) {
	switch scope {
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	case "project":
		return filepath.Join(".claude", "settings.json"), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

// Install configures Claude Code to use agent-notify by setting up hooks.
func (c *ClaudeIntegration) Install(settingsPath, binaryPath string) error {
	binaryPath = common.ResolveBinaryPath(binaryPath)
	command := binaryPath + " handle-claude-hook"

	buildEntry := func() []map[string]any {
		return []map[string]any{
			{
				"hooks": []map[string]any{
					{
						"type":    "command",
						"command": command,
					},
				},
			},
		}
	}

	hooks := map[string]any{
		"PermissionRequest":  buildEntry(),
		"Notification":       buildEntry(),
		"Stop":               buildEntry(),
		"PostToolUseFailure": buildEntry(),
	}

	settings := map[string]any{}

	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse settings.json: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read settings.json: %w", err)
	}

	existingHooks, _ := settings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingHooks = map[string]any{}
	}
	for key, value := range hooks {
		existingHooks[key] = value
	}
	settings["hooks"] = existingHooks

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}

// IsHookInstalled checks if agent-notify hooks are installed in the settings file.
func (c *ClaudeIntegration) IsHookInstalled(settingsPath string) (bool, error) {
	data, err := os.ReadFile(settingsPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	settings := map[string]any{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, fmt.Errorf("failed to parse settings.json: %w", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		return false, nil
	}

	// Check if PermissionRequest hook exists and contains handle-claude-hook
	pr, ok := hooks["PermissionRequest"].([]any)
	if !ok || len(pr) == 0 {
		return false, nil
	}

	entry, ok := pr[0].(map[string]any)
	if !ok {
		return false, nil
	}

	hookList, ok := entry["hooks"].([]any)
	if !ok || len(hookList) == 0 {
		return false, nil
	}

	hook, ok := hookList[0].(map[string]any)
	if !ok {
		return false, nil
	}

	cmd, ok := hook["command"].(string)
	if !ok {
		return false, nil
	}

	return strings.Contains(cmd, "handle-claude-hook"), nil
}
