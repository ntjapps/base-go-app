package broadcast

import (
	"context"
)

type MockBroadcaster struct {
	LastChannel string
	LastEvent   string
	LastPayload interface{}
}

func (m *MockBroadcaster) Broadcast(ctx context.Context, channel, event string, payload interface{}) error {
	m.LastChannel = channel
	m.LastEvent = event
	m.LastPayload = payload
	return nil
}
