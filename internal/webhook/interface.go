package webhook

import (
	"context"
)

// Client defines the interface for sending webhook callbacks.
type Client interface {
	Send(ctx context.Context, url string, payload interface{}, oauthClientID, oauthScope string) error
}

// NoOpClient is a no-op implementation.
type NoOpClient struct{}

func (n *NoOpClient) Send(ctx context.Context, url string, payload interface{}, oauthClientID, oauthScope string) error {
	return nil
}
