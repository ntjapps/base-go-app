package main

import (
	"log"
	"os"

	"base-go-app/internal/config"
	"base-go-app/internal/publisher"
)

// Example demonstrating how to publish tasks to RabbitMQ
// This example shows both Celery (Python) and Go worker task formats

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Create publisher
	pub, err := publisher.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
		os.Exit(1)
	}
	defer pub.Close()

	log.Println("Publisher connected successfully!")

	// Example 1: Send a Celery task (for Python Celery workers)
	log.Println("\n=== Example 1: Celery Task ===")
	taskID, err := pub.SendCeleryTask(
		"celery_test_task",
		[]interface{}{"arg1", "arg2"},
		"celery",
	)
	if err != nil {
		log.Printf("Failed to send Celery task: %v", err)
	} else {
		log.Printf("✓ Celery task submitted with ID: %s", taskID)
	}

	// Example 2: Send a Celery task with body arguments
	log.Println("\n=== Example 2: Celery Task with Arguments ===")
	taskID, err = pub.SendCeleryTask(
		"celery_test_body_task",
		[]interface{}{"value1", "value2", "value3"},
		"celery",
	)
	if err != nil {
		log.Printf("Failed to send Celery body task: %v", err)
	} else {
		log.Printf("✓ Celery body task submitted with ID: %s", taskID)
	}

	// Example 3: Send a Go worker task (logger)
	log.Println("\n=== Example 3: Go Worker Task (Logger) ===")
	logPayload := map[string]interface{}{
		"message":    "Test log from publisher example",
		"channel":    "example",
		"level":      "200",
		"level_name": "INFO",
		"datetime":   "2025-12-30 10:00:00",
		"context": map[string]interface{}{
			"user_id": 123,
			"action":  "test_publish",
		},
		"extra": map[string]interface{}{
			"request_id": "example-123",
		},
	}

	taskID, err = pub.SendGoTask("logger", logPayload, "logger", nil)
	if err != nil {
		log.Printf("Failed to send logger task: %v", err)
	} else {
		log.Printf("✓ Logger task submitted with ID: %s", taskID)
	}

	// Example 4: Send a Go worker task with options
	log.Println("\n=== Example 4: Go Worker Task with Options ===")
	timeout := 300
	maxAttempts := 3
	options := &publisher.TaskOptions{
		TimeoutSeconds: &timeout,
		Notify: map[string]string{
			"webhook": "http://example.com/webhook/callback",
		},
		MaxAttempts: &maxAttempts,
	}

	notificationPayload := map[string]interface{}{
		"message":    "Important notification with custom options",
		"channel":    "notifications",
		"level":      "300",
		"level_name": "WARNING",
	}

	taskID, err = pub.SendGoTask("notification", notificationPayload, "notifications", options)
	if err != nil {
		log.Printf("Failed to send notification task: %v", err)
	} else {
		log.Printf("✓ Notification task submitted with ID: %s", taskID)
		log.Printf("  - Timeout: %d seconds", timeout)
		log.Printf("  - Max Attempts: %d", maxAttempts)
		log.Printf("  - Webhook: %s", options.Notify["webhook"])
	}

	log.Println("\n=== All examples completed! ===")
}
