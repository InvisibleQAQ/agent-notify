package notify

import (
	"context"
	"strings"
	"testing"
)

func TestWindowsSenderSendCallsPowerShell(t *testing.T) {
	var gotName string
	var gotArgs []string

	sender := NewWindowsSender(func(_ context.Context, name string, args ...string) error {
		gotName = name
		gotArgs = args
		return nil
	})

	msg := Message{Title: "Test Title", Body: "Test Body", Workspace: "/path/to/project"}
	if err := sender.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if gotName != "powershell" {
		t.Fatalf("name = %q, want powershell", gotName)
	}

	// Verify expected arguments structure
	// args: -Command <script>
	if len(gotArgs) < 2 {
		t.Fatalf("args = %#v, want at least 2 args", gotArgs)
	}
	if gotArgs[0] != "-Command" {
		t.Fatalf("args[0] = %q, want -Command", gotArgs[0])
	}

	// Script should contain the title and body
	script := gotArgs[1]
	if !strings.Contains(script, "Test Title") {
		t.Errorf("script = %q, want to contain title %q", script, "Test Title")
	}
	if !strings.Contains(script, "Test Body") {
		t.Errorf("script = %q, want to contain body %q", script, "Test Body")
	}
	if !strings.Contains(script, "/path/to/project") {
		t.Errorf("script = %q, want to contain workspace path", script)
	}
}

func TestWindowsSenderSendWithoutWorkspace(t *testing.T) {
	var gotArgs []string

	sender := NewWindowsSender(func(_ context.Context, name string, args ...string) error {
		gotArgs = args
		return nil
	})

	msg := Message{Title: "Title", Body: "Body", Workspace: ""}
	if err := sender.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	script := gotArgs[1]
	if !strings.Contains(script, "Body") {
		t.Errorf("script = %q, want to contain %q", script, "Body")
	}
}

func TestWindowsSenderFormatBody(t *testing.T) {
	sender := &WindowsSender{}

	tests := []struct {
		name      string
		msg       Message
		wantParts []string
		dontWant  []string
	}{
		{
			name:      "with workspace",
			msg:       Message{Body: "Test message", Workspace: "/home/user/project"},
			wantParts: []string{"/home/user/project", "Test message"},
			dontWant:  []string{},
		},
		{
			name:      "without workspace",
			msg:       Message{Body: "Test message", Workspace: ""},
			wantParts: []string{"Test message"},
			dontWant:  []string{"/home"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sender.formatBody(tt.msg)

			for _, want := range tt.wantParts {
				if !strings.Contains(result, want) {
					t.Errorf("formatBody() = %q, want to contain %q", result, want)
				}
			}

			for _, dontWant := range tt.dontWant {
				if strings.Contains(result, dontWant) {
					t.Errorf("formatBody() = %q, should not contain %q", result, dontWant)
				}
			}

			// Should always contain timestamp
			// Timestamp format is "15:04:05"
			if len(result) < 8 { // minimum: "x\nHH:MM:SS"
				t.Errorf("formatBody() = %q, too short to contain timestamp", result)
			}
		})
	}
}

func TestWindowsSenderName(t *testing.T) {
	sender := &WindowsSender{}
	if sender.Name() != "system" {
		t.Fatalf("Name() = %q, want system", sender.Name())
	}
}

func TestWindowsSenderScriptContainsNotifyIcon(t *testing.T) {
	var gotArgs []string

	sender := NewWindowsSender(func(_ context.Context, name string, args ...string) error {
		gotArgs = args
		return nil
	})

	msg := Message{Title: "Title", Body: "Body"}
	if err := sender.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	script := gotArgs[1]

	// Verify the script uses the correct PowerShell components
	expectedParts := []string{
		"System.Windows.Forms",
		"System.Windows.Forms.NotifyIcon",
		"BalloonTipTitle",
		"BalloonTipText",
		"ShowBalloonTip",
	}

	for _, part := range expectedParts {
		if !strings.Contains(script, part) {
			t.Errorf("script should contain %q", part)
		}
	}
}
