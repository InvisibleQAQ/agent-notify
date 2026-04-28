package agentintegrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeIntegration_Name(t *testing.T) {
	c := NewClaudeIntegration()
	if got := c.Name(); got != "Claude Code" {
		t.Errorf("ClaudeIntegration.Name() = %q, want %q", got, "Claude Code")
	}
}

func TestCodexIntegration_Name(t *testing.T) {
	c := NewCodexIntegration()
	if got := c.Name(); got != "Codex" {
		t.Errorf("CodexIntegration.Name() = %q, want %q", got, "Codex")
	}
}

func TestClaudeIntegration_SettingsPath(t *testing.T) {
	c := NewClaudeIntegration()

	t.Run("user scope", func(t *testing.T) {
		path, err := c.SettingsPath("user")
		if err != nil {
			t.Fatalf("SettingsPath(user) error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json")
		if path != expected {
			t.Errorf("SettingsPath(user) = %q, want %q", path, expected)
		}
	})

	t.Run("project scope", func(t *testing.T) {
		path, err := c.SettingsPath("project")
		if err != nil {
			t.Fatalf("SettingsPath(project) error: %v", err)
		}
		expected := filepath.Join(".claude", "settings.json")
		if path != expected {
			t.Errorf("SettingsPath(project) = %q, want %q", path, expected)
		}
	})

	t.Run("invalid scope", func(t *testing.T) {
		_, err := c.SettingsPath("invalid")
		if err == nil {
			t.Error("SettingsPath(invalid) expected error, got nil")
		}
	})
}

func TestCodexIntegration_SettingsPath(t *testing.T) {
	c := NewCodexIntegration()

	t.Run("user scope", func(t *testing.T) {
		path, err := c.SettingsPath("user")
		if err != nil {
			t.Fatalf("SettingsPath(user) error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".codex", "config.toml")
		if path != expected {
			t.Errorf("SettingsPath(user) = %q, want %q", path, expected)
		}
	})

	t.Run("project scope", func(t *testing.T) {
		path, err := c.SettingsPath("project")
		if err != nil {
			t.Fatalf("SettingsPath(project) error: %v", err)
		}
		expected := filepath.Join(".codex", "config.toml")
		if path != expected {
			t.Errorf("SettingsPath(project) = %q, want %q", path, expected)
		}
	})

	t.Run("invalid scope", func(t *testing.T) {
		_, err := c.SettingsPath("invalid")
		if err == nil {
			t.Error("SettingsPath(invalid) expected error, got nil")
		}
	})
}

func TestClaudeIntegration_Install(t *testing.T) {
	c := NewClaudeIntegration()

	t.Run("creates settings file with hooks", func(t *testing.T) {
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")

		err := c.Install(settingsPath, "/usr/local/bin/agent-notify")
		if err != nil {
			t.Fatalf("Install() error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
			t.Fatalf("settings.json not created at %q", settingsPath)
		}

		// Verify hooks are installed
		installed, err := c.IsHookInstalled(settingsPath)
		if err != nil {
			t.Fatalf("IsHookInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsHookInstalled() = false, want true")
		}
	})

	t.Run("preserves existing settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "settings.json")

		// Create existing settings
		existingSettings := `{"apiKey": "test-key", "theme": "dark"}`
		if err := os.WriteFile(settingsPath, []byte(existingSettings), 0o644); err != nil {
			t.Fatalf("failed to write existing settings: %v", err)
		}

		err := c.Install(settingsPath, "/usr/local/bin/agent-notify")
		if err != nil {
			t.Fatalf("Install() error: %v", err)
		}

		// Read the file and verify both old and new keys exist
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings: %v", err)
		}

		content := string(data)
		if !containsAll(content, `"apiKey"`, `"theme"`, `"hooks"`) {
			t.Errorf("settings.json should contain apiKey, theme, and hooks, got:\n%s", content)
		}
	})
}

func TestCodexIntegration_Install(t *testing.T) {
	c := NewCodexIntegration()

	t.Run("creates config file with notify", func(t *testing.T) {
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, ".codex", "config.toml")

		err := c.Install(settingsPath, "/usr/local/bin/agent-notify")
		if err != nil {
			t.Fatalf("Install() error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
			t.Fatalf("config.toml not created at %q", settingsPath)
		}

		// Verify notify is installed
		installed, err := c.IsHookInstalled(settingsPath)
		if err != nil {
			t.Fatalf("IsHookInstalled() error: %v", err)
		}
		if !installed {
			t.Error("IsHookInstalled() = false, want true")
		}
	})

	t.Run("preserves existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "config.toml")

		// Create existing config
		existingConfig := `model = "gpt-4"
api_key = "test-key"
`
		if err := os.WriteFile(settingsPath, []byte(existingConfig), 0o644); err != nil {
			t.Fatalf("failed to write existing config: %v", err)
		}

		err := c.Install(settingsPath, "/usr/local/bin/agent-notify")
		if err != nil {
			t.Fatalf("Install() error: %v", err)
		}

		// Read the file and verify both old and new keys exist
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		content := string(data)
		if !containsAll(content, `model =`, `api_key =`, `notify =`) {
			t.Errorf("config.toml should contain model, api_key, and notify, got:\n%s", content)
		}
	})

	t.Run("updates existing notify", func(t *testing.T) {
		tmpDir := t.TempDir()
		settingsPath := filepath.Join(tmpDir, "config.toml")

		// Create config with existing notify
		existingConfig := `model = "gpt-4"
notify = ["old-binary", "old-command"]
`
		if err := os.WriteFile(settingsPath, []byte(existingConfig), 0o644); err != nil {
			t.Fatalf("failed to write existing config: %v", err)
		}

		err := c.Install(settingsPath, "/usr/local/bin/agent-notify")
		if err != nil {
			t.Fatalf("Install() error: %v", err)
		}

		// Read and verify notify was updated
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		content := string(data)
		if containsAll(content, `"old-binary"`) {
			t.Errorf("config.toml should not contain old-binary, got:\n%s", content)
		}
		if !containsAll(content, `handle-codex-notify`) {
			t.Errorf("config.toml should contain handle-codex-notify, got:\n%s", content)
		}
	})
}

func TestClaudeIntegration_DetectInstalled(t *testing.T) {
	c := NewClaudeIntegration()
	// This test just verifies the method doesn't panic
	// The actual result depends on whether claude is installed
	_ = c.DetectInstalled()
}

func TestCodexIntegration_DetectInstalled(t *testing.T) {
	c := NewCodexIntegration()
	// This test just verifies the method doesn't panic
	// The actual result depends on whether codex is installed
	_ = c.DetectInstalled()
}

func containsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
