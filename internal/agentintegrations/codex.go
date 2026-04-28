package agentintegrations

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hellolib/agent-notify/internal/common"
	toml "github.com/pelletier/go-toml/v2"
)

// CodexIntegration implements Integration for Codex.
type CodexIntegration struct{}

// NewCodexIntegration creates a new Codex integration.
func NewCodexIntegration() *CodexIntegration {
	return &CodexIntegration{}
}

// Name returns the display name for Codex.
func (c *CodexIntegration) Name() string {
	return "Codex"
}

// DetectInstalled checks if the codex CLI is installed.
func (c *CodexIntegration) DetectInstalled() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// SettingsPath returns the path to Codex's config.toml file.
func (c *CodexIntegration) SettingsPath(scope string) (string, error) {
	switch scope {
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".codex", "config.toml"), nil
	case "project":
		return filepath.Join(".codex", "config.toml"), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

// Install configures Codex to use agent-notify by setting up the notify command.
func (c *CodexIntegration) Install(settingsPath, binaryPath string) error {
	command := c.notifyCommand(binaryPath)

	// Read existing file content
	data, err := os.ReadFile(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read config.toml: %w", err)
	}

	// Parse existing TOML or start with empty map
	config := make(map[string]any)
	if len(data) > 0 {
		if err := toml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse TOML: %w", err)
		}
	}

	// Set notify key at top level
	config["notify"] = command

	// Marshal back to TOML
	updated, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML: %w", err)
	}

	// Ensure newline at end
	result := string(updated)
	if len(result) > 0 && result[len(result)-1] != '\n' {
		result += "\n"
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(settingsPath, []byte(result), 0o644); err != nil {
		return fmt.Errorf("failed to write config.toml: %w", err)
	}

	return nil
}

// notifyCommand returns the notify command for Codex configuration.
func (c *CodexIntegration) notifyCommand(binaryPath string) []string {
	return []string{common.ResolveBinaryPath(binaryPath), "handle-codex-notify"}
}

// IsHookInstalled checks if agent-notify is configured in the Codex config file.
func (c *CodexIntegration) IsHookInstalled(settingsPath string) (bool, error) {
	data, err := os.ReadFile(settingsPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Parse TOML
	config := make(map[string]any)
	if err := toml.Unmarshal(data, &config); err != nil {
		return false, fmt.Errorf("failed to parse TOML: %w", err)
	}

	// Check for notify array
	notify, ok := config["notify"]
	if !ok {
		return false, nil
	}

	// Check if it's an array containing "handle-codex-notify"
	notifySlice, ok := notify.([]any)
	if !ok {
		return false, nil
	}

	for _, item := range notifySlice {
		if str, ok := item.(string); ok && str == "handle-codex-notify" {
			return true, nil
		}
	}

	return false, nil
}
