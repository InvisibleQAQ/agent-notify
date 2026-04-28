package codex

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hellolib/agent-notify/internal/common"
	toml "github.com/pelletier/go-toml/v2"
)

func ConfigPath(scope string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch scope {
	case "user":
		return filepath.Join(home, ".codex", "config.toml"), nil
	case "project":
		return filepath.Join(".codex", "config.toml"), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

func NotifyCommand(binaryPath string) []string {
	return []string{common.ResolveBinaryPath(binaryPath), "handle-codex-notify"}
}

func Install(path string, binaryPath string) error {
	command := NotifyCommand(binaryPath)

	// Read existing file content
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(result), 0o644)
}

func IsNotifyInstalled(path string) (bool, error) {
	data, err := os.ReadFile(path)
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
