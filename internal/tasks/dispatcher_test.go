package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"base-go-app/internal/broadcast"
	"base-go-app/internal/webhook"
)

type mockBroadcaster struct {
	lastChannel string
	lastEvent   string
	lastPayload interface{}
}

func (m *mockBroadcaster) Broadcast(ctx context.Context, channel, event string, payload interface{}) error {
	m.lastChannel = channel
	m.lastEvent = event
	m.lastPayload = payload
	return nil
}

func TestDispatcherSuccess(t *testing.T) {
	ClearRegistry()
	RegisterTask("test_task", &mockHandler{})

	d := NewDispatcher(&broadcast.NoOpBroadcaster{}, &webhook.NoOpClient{})

	payload := TaskPayload{
		Task:        "test_task",
		ID:          "123",
		MaxAttempts: 1,
		Payload:     json.RawMessage(`{}`),
	}
	body, _ := json.Marshal(payload)

	res := d.Dispatch(context.Background(), body)
	if !res.Success {
		t.Fatalf("expected success, got error: %v", res.Error)
	}
	if res.Retry {
		t.Fatalf("expected no retry")
	}
}

func TestDispatcherRetry(t *testing.T) {
	ClearRegistry()
	// Register a handler that always fails
	RegisterTask("fail_task", &failHandler{})

	d := NewDispatcher(&broadcast.NoOpBroadcaster{}, &webhook.NoOpClient{})

	payload := TaskPayload{
		Task:        "fail_task",
		ID:          "123",
		Attempt:     0,
		MaxAttempts: 3,
		Payload:     json.RawMessage(`{}`),
	}
	body, _ := json.Marshal(payload)

	res := d.Dispatch(context.Background(), body)
	if res.Success {
		t.Fatalf("expected failure")
	}
	if !res.Retry {
		t.Fatalf("expected retry")
	}
	if res.RetryAttempt != 1 {
		t.Fatalf("expected retry attempt 1, got %d", res.RetryAttempt)
	}
}

func TestDispatcherExhausted(t *testing.T) {
	ClearRegistry()
	RegisterTask("fail_task", &failHandler{})

	d := NewDispatcher(&broadcast.NoOpBroadcaster{}, &webhook.NoOpClient{})

	payload := TaskPayload{
		Task:        "fail_task",
		ID:          "123",
		Attempt:     2,
		MaxAttempts: 3,
		Payload:     json.RawMessage(`{}`),
	}
	body, _ := json.Marshal(payload)

	res := d.Dispatch(context.Background(), body)
	if res.Success {
		t.Fatalf("expected failure")
	}
	if res.Retry {
		t.Fatalf("expected no retry (exhausted)")
	}
}

type failHandler struct{}

func (f *failHandler) Handle(ctx context.Context, payload json.RawMessage) error {
	return errors.New("always fail")
}
