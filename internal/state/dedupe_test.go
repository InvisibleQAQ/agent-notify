package state

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreShouldSendBlocksRecentlyMarkedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	key := "permission_required:sess-1"

	if err := store.MarkSent(key, now); err != nil {
		t.Fatalf("MarkSent() error = %v, want nil", err)
	}
	if allow, err := store.ShouldSend(key, 60*time.Second, now.Add(30*time.Second)); err != nil || allow {
		t.Fatalf("ShouldSend() within window = (%v, %v), want (false, nil)", allow, err)
	}
	if allow, err := store.ShouldSend(key, 60*time.Second, now.Add(61*time.Second)); err != nil || !allow {
		t.Fatalf("ShouldSend() after window = (%v, %v), want (true, nil)", allow, err)
	}
}

func TestStoreShouldSendDoesNotDedupeUntilMarkedSent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	key := "claude:permission_required:sess-1:system"

	if allow, err := store.ShouldSend(key, 60*time.Second, now); err != nil || !allow {
		t.Fatalf("first ShouldSend() = (%v, %v), want (true, nil)", allow, err)
	}
	if allow, err := store.ShouldSend(key, 60*time.Second, now.Add(30*time.Second)); err != nil || !allow {
		t.Fatalf("second ShouldSend() before mark = (%v, %v), want (true, nil)", allow, err)
	}
	if err := store.MarkSent(key, now.Add(30 * time.Second)); err != nil {
		t.Fatalf("MarkSent() error = %v, want nil", err)
	}
	if allow, err := store.ShouldSend(key, 60*time.Second, now.Add(45*time.Second)); err != nil || allow {
		t.Fatalf("third ShouldSend() after mark = (%v, %v), want (false, nil)", allow, err)
	}
}

func TestStoreReserveSendPreventsDuplicateInFlightSend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewStore(path)
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	key := "claude:permission_required:sess-1:system"

	allow, err := store.ReserveSend(key, 60*time.Second, now)
	if err != nil || !allow {
		t.Fatalf("first ReserveSend() = (%v, %v), want (true, nil)", allow, err)
	}

	allow, err = store.ReserveSend(key, 60*time.Second, now.Add(time.Second))
	if err != nil || allow {
		t.Fatalf("second ReserveSend() = (%v, %v), want (false, nil)", allow, err)
	}

	if err := store.ClearReservation(key); err != nil {
		t.Fatalf("ClearReservation() error = %v", err)
	}

	allow, err = store.ReserveSend(key, 60*time.Second, now.Add(2*time.Second))
	if err != nil || !allow {
		t.Fatalf("third ReserveSend() = (%v, %v), want (true, nil)", allow, err)
	}
}
