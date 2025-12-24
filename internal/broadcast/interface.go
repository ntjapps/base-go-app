package broadcast

import (
	"context"
)

// Broadcaster defines the interface for sending real-time notifications.
type Broadcaster interface {
	Broadcast(ctx context.Context, channel, event string, payload interface{}) error
}

// NoOpBroadcaster is a no-op implementation.
type NoOpBroadcaster struct{}

func (n *NoOpBroadcaster) Broadcast(ctx context.Context, channel, event string, payload interface{}) error {
	return nil
}
