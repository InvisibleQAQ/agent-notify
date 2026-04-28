package notify

import (
	"context"
	"fmt"
)

type UnsupportedSender struct {
	platform string
}

func NewUnsupportedSender(platform string) Sender {
	return &UnsupportedSender{platform: platform}
}

func (s *UnsupportedSender) Name() string { return "system" }

func (s *UnsupportedSender) Send(context.Context, Message) error {
	return fmt.Errorf("unsupported platform for system notifications: %s", s.platform)
}
