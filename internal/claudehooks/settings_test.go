package claudehooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildHookSettings(t *testing.T) {
	got := BuildHookSettings("/tmp/agent-notify")

	hooks, ok := got["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks type = %T, want map[string]any", got["hooks"])
	}

	events := []string{"PermissionRequest", "Notification", "Stop", "PostToolUseFailure"}
	for _, event := range events {
		items, ok := hooks[event].([]map[string]any)
		if !ok || len(items) != 1 {
			t.Fatalf("%s hooks missing or invalid", event)
		}
		entryHooks, ok := items[0]["hooks"].([]map[string]any)
		if !ok || len(entryHooks) != 1 {
			t.Fatalf("%s command hooks missing or invalid", event)
		}
		if entryHooks[0]["command"] != "/tmp/agent-notify handle-claude-hook" {
			t.Fatalf("%s command = %v, want /tmp/agent-notify handle-claude-hook", event, entryHooks[0]["command"])
		}
	}
}

func TestInstallMergesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"dark"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Install(path, "/tmp/agent-notify"); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["theme"] != "dark" {
		t.Fatalf("theme = %v, want dark", got["theme"])
	}
	if _, ok := got["hooks"]; !ok {
		t.Fatal("hooks key missing")
	}
}
