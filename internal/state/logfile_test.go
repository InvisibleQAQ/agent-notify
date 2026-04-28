package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent-notify.log")

	if err := AppendLog(path, "first line"); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	if err := AppendLog(path, "second line"); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	if !strings.Contains(got, "first line") || !strings.Contains(got, "second line") {
		t.Fatalf("log content = %q, want both lines", got)
	}
}
