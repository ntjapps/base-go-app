package main

import (
	"log"
	"sync"
	"time"

	"base-go-app/internal/config"
	"base-go-app/internal/publisher"
)

// Example demonstrating sending tasks to multiple queues for parallel processing
// Different task types go to different queues, allowing multiple workers to process them in parallel

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create publisher (reuse for all tasks)
	pub, err := publisher.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	log.Println("=== Sending tasks to multiple queues for parallel processing ===")

	// Example 1: Send logger tasks to dedicated logger queue
	log.Println("1. Sending logger task to 'go.logger' queue...")
	taskID, err := pub.SendGoTask(
		"logger",
		map[string]interface{}{
			"message":    "User logged in",
			"channel":    "auth",
			"level":      "200",
			"level_name": "INFO",
			"datetime":   time.Now().Format("2006-01-02 15:04:05"),
		},
		"go.logger", // Queue name for logger tasks
		nil,
	)
	if err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		log.Printf("   ✓ Task ID: %s (queue: go.logger)\n", taskID)
	}

	// Example 2: Send data update task to a different queue
	log.Println("2. Sending data update task to 'go.data_processor' queue...")
	taskID, err = pub.SendGoTask(
		"update_user_data",
		map[string]interface{}{
			"user_id": 12345,
			"fields": map[string]interface{}{
				"last_login": time.Now().Format(time.RFC3339),
				"status":     "active",
			},
		},
		"go.data_processor", // Queue name for data processing tasks
		nil,
	)
	if err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		log.Printf("   ✓ Task ID: %s (queue: go.data_processor)\n", taskID)
	}

	// Example 3: Send email notification to email queue
	log.Println("3. Sending email task to 'go.email' queue...")
	timeout := 120
	taskID, err = pub.SendGoTask(
		"send_email",
		map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Welcome!",
			"body":    "Thank you for signing up",
		},
		"go.email", // Queue name for email tasks
		&publisher.TaskOptions{
			TimeoutSeconds: &timeout,
		},
	)
	if err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		log.Printf("   ✓ Task ID: %s (queue: go.email)\n", taskID)
	}

	// Example 4: Send image processing task to dedicated queue
	log.Println("4. Sending image processing task to 'go.image_processor' queue...")
	timeout = 600 // Longer timeout for heavy processing
	taskID, err = pub.SendGoTask(
		"resize_image",
		map[string]interface{}{
			"image_id": "img_12345",
			"sizes":    []string{"thumbnail", "medium", "large"},
		},
		"go.image_processor", // Queue name for image processing
		&publisher.TaskOptions{
			TimeoutSeconds: &timeout,
		},
	)
	if err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		log.Printf("   ✓ Task ID: %s (queue: go.image_processor)\n", taskID)
	}

	// Example 5: Send to Python Celery worker on custom queue
	log.Println("5. Sending Celery task to 'python.analytics' queue...")
	taskID, err = pub.SendCeleryTask(
		"generate_analytics_report",
		[]interface{}{
			"2025-12-01",
			"2025-12-30",
			map[string]interface{}{"format": "pdf"},
		},
		"python.analytics", // Queue name for Python analytics tasks
	)
	if err != nil {
		log.Printf("   ✗ Failed: %v", err)
	} else {
		log.Printf("   ✓ Task ID: %s (queue: python.analytics)\n", taskID)
	}

	// Example 6: Batch processing - send multiple tasks to different queues concurrently
	log.Println("\n6. Batch processing - sending tasks to multiple queues concurrently...")
	
	var wg sync.WaitGroup
	tasks := []struct {
		taskName string
		queue    string
		payload  map[string]interface{}
	}{
		{
			taskName: "logger",
			queue:    "go.logger",
			payload: map[string]interface{}{
				"message": "Batch task 1",
				"level":   "200",
			},
		},
		{
			taskName: "process_order",
			queue:    "go.orders",
			payload: map[string]interface{}{
				"order_id": "ORD-001",
				"action":   "fulfill",
			},
		},
		{
			taskName: "send_notification",
			queue:    "go.notifications",
			payload: map[string]interface{}{
				"user_id": 123,
				"message": "Your order is ready",
			},
		},
		{
			taskName: "update_inventory",
			queue:    "go.inventory",
			payload: map[string]interface{}{
				"product_id": "PROD-456",
				"quantity":   -1,
			},
		},
	}

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t struct {
			taskName string
			queue    string
			payload  map[string]interface{}
		}) {
			defer wg.Done()
			
			id, err := pub.SendGoTask(t.taskName, t.payload, t.queue, nil)
			if err != nil {
				log.Printf("   [%d] ✗ Failed to send %s to %s: %v", idx+1, t.taskName, t.queue, err)
			} else {
				log.Printf("   [%d] ✓ Sent %s to %s (ID: %s)", idx+1, t.taskName, t.queue, id)
			}
		}(i, task)
	}
	
	wg.Wait()

	log.Println("\n=== All tasks sent successfully! ===")
	log.Println("\nKey Points:")
	log.Println("• Each queue can have dedicated workers for parallel processing")
	log.Println("• go.logger - Fast, high-throughput logging")
	log.Println("• go.data_processor - Database updates")
	log.Println("• go.email - Email sending")
	log.Println("• go.image_processor - CPU-intensive processing")
	log.Println("• python.analytics - Python-specific tasks")
	log.Println("\nWorkers listening to different queues will process tasks in parallel!")
}
