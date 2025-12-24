package tasks

import (
	"context"
	"encoding/json"
)

// TaskHandler is the interface that all task handlers must implement.
type TaskHandler interface {
	// Handle processes the task payload.
	// It returns an error if the task failed.
	Handle(ctx context.Context, payload json.RawMessage) error
}

// TaskPayload represents the standard envelope for tasks.
type TaskPayload struct {
	Version        string          `json:"version"`
	ID             string          `json:"id"`
	Task           string          `json:"task"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      string          `json:"created_at"`
	Attempt        int             `json:"attempt"`
	MaxAttempts    int             `json:"max_attempts"`
	TimeoutSeconds int             `json:"timeout_seconds,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	Meta           json.RawMessage `json:"meta,omitempty"`
	Notify         *NotifyConfig   `json:"notify,omitempty"`
}

// NotifyConfig defines notification preferences for task completion.
type NotifyConfig struct {
	Sockudo *SockudoConfig `json:"sockudo,omitempty"`
	Webhook *WebhookConfig `json:"webhook,omitempty"`
}

type SockudoConfig struct {
	Channel        string `json:"channel"`
	Event          string `json:"event"`
	IncludePayload bool   `json:"include_payload"`
}

type WebhookConfig struct {
	URL            string `json:"url"`
	OAuthClientID  string `json:"oauth_client_id,omitempty"`
	OAuthScope     string `json:"oauth_scope,omitempty"`
}
