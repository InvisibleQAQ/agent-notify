package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigPathUser(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	got, err := ConfigPath("user")
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}

	want := filepath.Join(dir, ".codex", "config.toml")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestInstallCreatesNotifyCommand(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	// go-toml uses single quotes in output
	if !strings.Contains(got, `notify = ['/tmp/agent-notify', 'handle-codex-notify']`) {
		t.Fatalf("config = %q, want codex notify array command", got)
	}
}

func TestInstallPreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := "model = \"gpt-5.4\"\n\n[features]\nmulti_agent = true\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	// go-toml uses single quotes in output
	if !strings.Contains(got, "model = 'gpt-5.4'") {
		t.Fatalf("config = %q, want existing model preserved", got)
	}
	if !strings.Contains(got, "[features]") {
		t.Fatalf("config = %q, want existing features section preserved", got)
	}
	if !strings.Contains(got, `notify = ['/tmp/agent-notify', 'handle-codex-notify']`) {
		t.Fatalf("config = %q, want codex notify array command", got)
	}
}

func TestInstallUpdatesExistingNotifyCommand(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := "notify = [\"old-command\"]\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	// go-toml uses single quotes in output
	if strings.Contains(got, `notify = ['old-command']`) {
		t.Fatalf("config = %q, want old command replaced", got)
	}
	if !strings.Contains(got, `notify = ['/tmp/agent-notify', 'handle-codex-notify']`) {
		t.Fatalf("config = %q, want codex notify array command", got)
	}
}

func TestInstallPlacesNotifyAtTopLevelBeforeTables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := "personality = \"pragmatic\"\n\n[features]\nmulti_agent = true\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	// go-toml uses single quotes in output
	wantPrefix := "notify = ['/tmp/agent-notify', 'handle-codex-notify']\npersonality = 'pragmatic'\n"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("config = %q, want notify command at top before tables", got)
	}
}

func TestInstallPreservesNestedTables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	// Complex TOML with nested tables and arrays
	content := `model = "gpt-5.4"

[features]
multi_agent = true

[features.advanced]
reasoning = "chain-of-thought"
tools = ["search", "code"]

[mcp.servers]
server1 = "http://localhost:8080"

[[mcp.tools]]
name = "weather"
command = "/usr/bin/weather-cli"

[[mcp.tools]]
name = "calendar"
command = "/usr/bin/calendar-cli"
args = ["--verbose"]

[ui]
theme = "dark"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)

	// Verify notify is added (go-toml uses single quotes)
	if !strings.Contains(got, `notify = ['/tmp/agent-notify', 'handle-codex-notify']`) {
		t.Fatalf("config = %q, want codex notify array command", got)
	}

	// Verify nested tables are preserved
	if !strings.Contains(got, `[features]`) {
		t.Fatalf("config = %q, want [features] section preserved", got)
	}
	if !strings.Contains(got, `[features.advanced]`) {
		t.Fatalf("config = %q, want [features.advanced] nested section preserved", got)
	}
	if !strings.Contains(got, `reasoning = 'chain-of-thought'`) {
		t.Fatalf("config = %q, want reasoning value preserved", got)
	}
	if !strings.Contains(got, `tools = ['search', 'code']`) {
		t.Fatalf("config = %q, want tools array preserved", got)
	}

	// Verify MCP servers section
	if !strings.Contains(got, `[mcp.servers]`) {
		t.Fatalf("config = %q, want [mcp.servers] section preserved", got)
	}

	// Verify array of tables ([[mcp.tools]])
	if !strings.Contains(got, `[[mcp.tools]]`) {
		t.Fatalf("config = %q, want [[mcp.tools]] array of tables preserved", got)
	}
	if !strings.Contains(got, `name = 'weather'`) {
		t.Fatalf("config = %q, want weather tool preserved", got)
	}
	if !strings.Contains(got, `name = 'calendar'`) {
		t.Fatalf("config = %q, want calendar tool preserved", got)
	}
	if !strings.Contains(got, `args = ['--verbose']`) {
		t.Fatalf("config = %q, want calendar args preserved", got)
	}

	// Verify ui section
	if !strings.Contains(got, `[ui]`) {
		t.Fatalf("config = %q, want [ui] section preserved", got)
	}
	if !strings.Contains(got, `theme = 'dark'`) {
		t.Fatalf("config = %q, want theme value preserved", got)
	}
}

func TestIsNotifyInstalled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Test with non-existent file
	installed, err := IsNotifyInstalled(path)
	if err != nil {
		t.Fatalf("IsNotifyInstalled() error = %v", err)
	}
	if installed {
		t.Fatal("IsNotifyInstalled() = true for non-existent file, want false")
	}

	// Test with file but no notify
	content := "model = \"gpt-5.4\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	installed, err = IsNotifyInstalled(path)
	if err != nil {
		t.Fatalf("IsNotifyInstalled() error = %v", err)
	}
	if installed {
		t.Fatal("IsNotifyInstalled() = true for file without notify, want false")
	}

	// Test with notify installed
	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	installed, err = IsNotifyInstalled(path)
	if err != nil {
		t.Fatalf("IsNotifyInstalled() error = %v", err)
	}
	if !installed {
		t.Fatal("IsNotifyInstalled() = false after Install(), want true")
	}
}
