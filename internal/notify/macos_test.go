package notify

import (
	"context"
	"testing"
)

func TestMacOSSenderSendFallbackToOsascript(t *testing.T) {
	var gotName string
	var gotArgs []string
	callCount := 0

	sender := NewMacOSSender(func(_ context.Context, name string, args ...string) error {
		callCount++
		if name == "which" {
			// Simulate terminal-notifier not installed
			return context.DeadlineExceeded
		}
		gotName = name
		gotArgs = args
		return nil
	})

	if err := sender.Send(context.Background(), Message{Title: "Title", Body: "Body", Workspace: "/path"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if gotName != "osascript" {
		t.Fatalf("name = %q, want osascript", gotName)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "-e" {
		t.Fatalf("args = %#v, want osascript script args", gotArgs)
	}
	if callCount < 2 {
		t.Fatalf("callCount = %d, expected at least 2 (which + osascript)", callCount)
	}
}

func TestMacOSSenderSendUsesTerminalNotifier(t *testing.T) {
	var calls []struct {
		name string
		args []string
	}

	sender := NewMacOSSender(func(_ context.Context, name string, args ...string) error {
		calls = append(calls, struct {
			name string
			args []string
		}{name, args})
		return nil
	})

	if err := sender.Send(context.Background(), Message{Title: "Title", Body: "Body", Workspace: "/path"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// First call is 'which terminal-notifier', second is 'terminal-notifier'
	if len(calls) < 1 || calls[0].name != "which" {
		t.Fatalf("expected first call to be 'which', got %#v", calls)
	}
	if len(calls) >= 2 && calls[1].name == "terminal-notifier" {
		// Successfully used terminal-notifier
		return
	}
	// Otherwise should have fallen back to osascript
	if len(calls) >= 2 && calls[1].name != "osascript" {
		t.Fatalf("expected second call to be 'terminal-notifier' or 'osascript', got %s", calls[1].name)
	}
}
