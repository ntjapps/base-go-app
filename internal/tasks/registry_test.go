package tasks

import (
	"context"
	"encoding/json"
	"testing"
)

type mockHandler struct {
	called bool
}

func (m *mockHandler) Handle(ctx context.Context, payload json.RawMessage) error {
	m.called = true
	return nil
}

func TestRegistry(t *testing.T) {
	ClearRegistry()

	name := "test_task"
	handler := &mockHandler{}

	RegisterTask(name, handler)

	h, ok := LookupTask(name)
	if !ok {
		t.Fatalf("expected task to be found")
	}

	if h != handler {
		t.Fatalf("expected handler to match")
	}

	_, ok = LookupTask("unknown")
	if ok {
		t.Fatalf("expected unknown task to not be found")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	ClearRegistry()
	name := "dup_task"
	RegisterTask(name, &mockHandler{})

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic on duplicate registration")
		}
	}()

	RegisterTask(name, &mockHandler{})
}
