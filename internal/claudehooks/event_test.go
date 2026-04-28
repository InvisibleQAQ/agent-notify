package claudehooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePermissionRequest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "hooks", "permission_request.json"))
	if err != nil {
		t.Fatal(err)
	}

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	if msg.Event != "permission_required" {
		t.Fatalf("Event = %q, want permission_required", msg.Event)
	}
	if msg.Agent != "claude_code" {
		t.Fatalf("Agent = %q, want claude_code", msg.Agent)
	}
}

func TestParseNotificationWaitingInput(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "hooks", "notification_waiting_input.json"))
	if err != nil {
		t.Fatal(err)
	}

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	if msg.Event != "input_required" {
		t.Fatalf("Event = %q, want input_required", msg.Event)
	}
	if msg.Agent != "claude_code" {
		t.Fatalf("Agent = %q, want claude_code", msg.Agent)
	}
	if msg.Body != "提示: " {
		t.Fatalf("Body = %q, want %q", msg.Body, "提示: ")
	}
}

func TestParseNotificationNeedsInputVariant(t *testing.T) {
	data := []byte(`{"hook_event_name":"Notification","session_id":"s1","cwd":"/tmp/project","message":"needs input: please confirm"}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	if msg.Event != "input_required" {
		t.Fatalf("Event = %q, want input_required", msg.Event)
	}
	if msg.Agent != "claude_code" {
		t.Fatalf("Agent = %q, want claude_code", msg.Agent)
	}
	if msg.Body != "提示: please confirm" {
		t.Fatalf("Body = %q, want %q", msg.Body, "提示: please confirm")
	}
}
