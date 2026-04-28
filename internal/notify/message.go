package notify

import "context"

type Message struct {
	Agent     string
	Event     string
	SessionID string
	Workspace string
	Title     string
	Body      string
}

type Sender interface {
	Name() string
	Send(ctx context.Context, msg Message) error
}
