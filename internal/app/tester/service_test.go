package tester

import (
	"context"
	"testing"

	"github.com/hellolib/agent-notify/internal/config"
	"github.com/hellolib/agent-notify/internal/notify"
)

// mockFeishuPreparer implements FeishuPreparer for testing
type mockFeishuPreparer struct {
	called bool
	err    error
}

func (m *mockFeishuPreparer) EnsureReady(ctx context.Context) error {
	m.called = true
	return m.err
}

type mockConfigLoader struct {
	defaultPath string
	loadPath    string
	cfg         config.Config
}

func (m *mockConfigLoader) Load(path string) (config.Config, error) {
	m.loadPath = path
	return m.cfg, nil
}

func (m *mockConfigLoader) DefaultPath() (string, error) {
	return m.defaultPath, nil
}

type fakeSender struct {
	called bool
	err    error
}

func (f *fakeSender) Name() string { return "fake" }

func (f *fakeSender) Send(ctx context.Context, msg notify.Message) error {
	f.called = true
	return f.err
}

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestNewServiceWithOptions(t *testing.T) {
	preparer := &mockFeishuPreparer{}
	svc := NewService(WithFeishuPreparer(preparer))
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.feishuPreparer != preparer {
		t.Error("feishuPreparer not set correctly")
	}
}

func TestTestFeishu_UsesInjectedConfigLoader(t *testing.T) {
	loader := &mockConfigLoader{
		defaultPath: "/tmp/injected-config.yaml",
		cfg: config.Config{
			Notify: config.NotifyConfig{
				ClaudeCode: config.AgentNotifyConfig{
					Channels: config.ChannelsConfig{
						Feishu: config.ChannelConfig{Enabled: true},
					},
				},
			},
		},
	}
	preparer := &mockFeishuPreparer{}
	sender := &fakeSender{}
	svc := NewService(
		WithConfigLoader(loader),
		WithFeishuPreparer(preparer),
		WithFeishuSender(sender),
	)

	result, err := svc.TestFeishu(context.Background())
	if err != nil {
		t.Fatalf("TestFeishu() error = %v", err)
	}
	if result == nil || result.Message == "" {
		t.Fatal("expected non-empty result")
	}
	if loader.loadPath != "/tmp/injected-config.yaml" {
		t.Fatalf("loadPath = %q, want %q", loader.loadPath, "/tmp/injected-config.yaml")
	}
	if !preparer.called {
		t.Fatal("expected preparer to be called")
	}
	if !sender.called {
		t.Fatal("expected injected sender to be called")
	}
}

func TestTestSystem_UsesInjectedSender(t *testing.T) {
	sender := &fakeSender{}
	svc := NewService(WithSystemSender(sender))

	result, err := svc.TestSystem(context.Background())
	if err != nil {
		t.Fatalf("TestSystem() error = %v", err)
	}
	if result == nil || result.Message == "" {
		t.Fatal("expected non-empty result")
	}
	if !sender.called {
		t.Fatal("expected injected sender to be called")
	}
}

func TestTestFeishu_Disabled(t *testing.T) {
	// Use a temp directory to avoid loading real config
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	svc := NewService(
		WithFeishuPreparer(&mockFeishuPreparer{}),
	)

	// Without config, feishu is disabled by default
	result, err := svc.TestFeishu(context.Background())
	if err == nil {
		t.Fatal("expected error when feishu is disabled")
	}
	if result != nil {
		t.Error("expected nil result when feishu is disabled")
	}
}

func TestTestSystem(t *testing.T) {
	svc := NewService(WithSystemSender(&fakeSender{}))

	result, err := svc.TestSystem(context.Background())
	if err != nil {
		t.Fatalf("TestSystem() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Message == "" {
		t.Error("expected non-empty message")
	}
}
