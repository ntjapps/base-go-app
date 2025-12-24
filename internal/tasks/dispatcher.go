package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"base-go-app/internal/broadcast"
	"base-go-app/internal/webhook"
)

const DefaultMaxAttempts = 5

// Dispatcher handles task execution, retries, and notifications.
type Dispatcher struct {
	Broadcaster   broadcast.Broadcaster
	WebhookClient webhook.Client
}

// NewDispatcher creates a new dispatcher with dependencies.
func NewDispatcher(b broadcast.Broadcaster, w webhook.Client) *Dispatcher {
	if b == nil {
		b = &broadcast.NoOpBroadcaster{}
	}
	if w == nil {
		w = &webhook.NoOpClient{}
	}
	return &Dispatcher{
		Broadcaster:   b,
		WebhookClient: w,
	}
}

// DispatchResult represents the outcome of a dispatch.
type DispatchResult struct {
	Success      bool
	Retry        bool
	RetryAttempt int
	Error        error
}

// Dispatch processes a raw message body.
func (d *Dispatcher) Dispatch(ctx context.Context, body []byte) DispatchResult {
	var envelope TaskPayload
	if err := json.Unmarshal(body, &envelope); err != nil {
		// If we can't parse it, we can't retry it safely (poison message).
		// However, for migration, we might want to check if it's a legacy Celery message.
		// For now, we assume new format or fail.
		log.Printf("Error unmarshaling task envelope: %v", err)
		return DispatchResult{Success: false, Error: err}
	}

	// Validate task
	handler, ok := LookupTask(envelope.Task)
	if !ok {
		err := fmt.Errorf("unknown task: %s", envelope.Task)
		log.Printf("%v", err)
		return DispatchResult{Success: false, Error: err}
	}

	// Set defaults
	if envelope.MaxAttempts <= 0 {
		envelope.MaxAttempts = DefaultMaxAttempts
	}

	// Create context with timeout if specified
	taskCtx := ctx
	if envelope.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx, time.Duration(envelope.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// Execute handler
	start := time.Now()
	err := handler.Handle(taskCtx, envelope.Payload)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Task %s (id=%s) failed: %v", envelope.Task, envelope.ID, err)

		// Check retries
		if envelope.Attempt < envelope.MaxAttempts-1 {
			// Retry
			return DispatchResult{
				Success:      false,
				Retry:        true,
				RetryAttempt: envelope.Attempt + 1,
				Error:        err,
			}
		}

		// Exhausted retries
		d.notify(ctx, &envelope, "error", nil, err)
		return DispatchResult{Success: false, Error: err}
	}

	log.Printf("Task %s (id=%s) succeeded in %v", envelope.Task, envelope.ID, duration)
	d.notify(ctx, &envelope, "success", nil, nil) // Payload result not yet supported in Handle return
	return DispatchResult{Success: true}
}

func (d *Dispatcher) notify(ctx context.Context, envelope *TaskPayload, status string, result interface{}, err error) {
	if envelope.Notify == nil {
		return
	}

	// Prepare notification payload
	notifyPayload := map[string]interface{}{
		"id":         envelope.ID,
		"task":       envelope.Task,
		"status":     status,
		"attempt":    envelope.Attempt,
		"created_at": envelope.CreatedAt,
		"finished_at": time.Now().Format(time.RFC3339),
	}
	if err != nil {
		notifyPayload["error"] = err.Error()
	}
	if result != nil {
		notifyPayload["result"] = result
	}

	// Sockudo
	if s := envelope.Notify.Sockudo; s != nil {
		payloadToSend := notifyPayload
		if !s.IncludePayload {
			// Create a copy without result/payload if needed
			// For now result is separate, but if we added envelope.Payload we'd strip it here
		}
		
		go func() {
			// Use a detached context for notifications to ensure they run even if task ctx is canceled
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := d.Broadcaster.Broadcast(notifyCtx, s.Channel, s.Event, payloadToSend); err != nil {
				log.Printf("Failed to broadcast to Sockudo: %v", err)
			}
		}()
	}

	// Webhook
	if w := envelope.Notify.Webhook; w != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := d.WebhookClient.Send(notifyCtx, w.URL, notifyPayload, w.OAuthClientID, w.OAuthScope); err != nil {
				log.Printf("Failed to send webhook: %v", err)
			}
		}()
	}
}

// Helper to check if backoff is enabled
func BackoffEnabled() bool {
	return os.Getenv("BACKOFF_ENABLED") == "true"
}

func GetBackoffDuration(attempt int) time.Duration {
	initial := 2
	if s := os.Getenv("BACKOFF_INITIAL_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			initial = v
		}
	}
	
	// Simple exponential: initial * 2^attempt
	delay := time.Duration(initial * (1 << attempt)) * time.Second
	
	max := 30 * time.Second
	if s := os.Getenv("BACKOFF_MAX_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			max = time.Duration(v) * time.Second
		}
	}

	if delay > max {
		delay = max
	}
	return delay
}
