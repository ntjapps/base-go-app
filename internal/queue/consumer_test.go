package queue

import (
	"context"
	"testing"
	"time"

	"base-go-app/internal/config"
)

func TestStartConsumerStopsOnContextCancel(t *testing.T) {
	cfg := &config.Config{
		RabbitMQHost: "127.0.0.1",
		RabbitMQPort: "9999", // assuming nothing is there
		RabbitMQUser: "guest",
		RabbitMQPassword: "guest",
		RabbitMQVHost: "/",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := StartConsumer(ctx, cfg)
	// Wait for the done channel or timeout
	select {
	case <-done:
		// done early (possible if connection error quickly and ctx timed)
	case <-time.After(3 * time.Second):
		// now cancel and wait
		cancel()
		select {
		case <-done:
			// passed
		case <-time.After(2 * time.Second):
			t.Fatalf("consumer did not stop after cancel")
		}
	}
}
