package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hellolib/agent-notify/internal/state"
)

type Dispatcher struct {
	store   *state.Store
	window  time.Duration
	senders []Sender
}

func NewDispatcher(store *state.Store, window time.Duration, senders ...Sender) *Dispatcher {
	return &Dispatcher{
		store:   store,
		window:  window,
		senders: senders,
	}
}

func (d *Dispatcher) SendAll(ctx context.Context, msg Message) error {
	var errs []string
	for _, sender := range d.senders {
		now := time.Now()
		key := fmt.Sprintf("%s:%s:%s:%s", msg.Agent, msg.Event, msg.SessionID, sender.Name())
		allow, err := d.store.ReserveSend(key, d.window, now)
		if err != nil {
			return err
		}
		if !allow {
			continue
		}
		if err := sender.Send(ctx, msg); err != nil {
			_ = d.store.ClearReservation(key)
			errs = append(errs, fmt.Sprintf("%s: %v", sender.Name(), err))
			continue
		}
		if err := d.store.MarkSent(key, now); err != nil {
			return err
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "; "))
}
