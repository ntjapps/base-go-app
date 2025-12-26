package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"base-go-app/internal/config"
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"base-go-app/internal/queue"
)

// healthHandler returns an http.HandlerFunc for /healthcheck
func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := http.StatusOK
		body := map[string]interface{}{"status": "ok"}

		// Check database
		if ok, err := database.Ping(ctx); !ok {
			body["database"] = false
			if err != nil {
				body["database_error"] = err.Error()
			}
			// DB is optional, so we don't fail the healthcheck
			// status = http.StatusInternalServerError
			body["status"] = "degraded"
		} else {
			body["database"] = true
		}

		// Check rabbit
		rabbitOk := queue.RabbitConnected()
		body["rabbitmq"] = rabbitOk
		if !rabbitOk {
			// prefer to mark degraded
			if status == http.StatusOK {
				// RabbitMQ is critical? If so, keep 500. If optional, use 200.
				// User asked for resilience when RabbitMQ is not up too.
				// Let's keep it 200 but degraded for now to prevent restart loops.
				// status = http.StatusInternalServerError
				body["status"] = "degraded"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
}

func startHealthServer() {
	port := os.Getenv("HEALTH_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/healthcheck", healthHandler())

	addr := fmt.Sprintf(":%s", port)
	go func() {
		log.Printf("Health server listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Health server failed: %v", err)
		}
	}()
}

func main() {
	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create a context that is canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start Health Server early so readiness is visible
	startHealthServer()

	// Start Queue Consumer (prioritized) and get the done channel
	done := queue.StartConsumer(ctx, cfg)

	// Connect to Database in background and run AutoMigrate once DB becomes available.
	// The DB is optional for startup; this goroutine will perform a single
	// migration when a connection is established and will exit on ctx cancellation.
	go func() {
		if err := database.Connect(cfg); err != nil {
			log.Printf("Failed to start database connection: %v", err)
			// Connect starts its own reconnect loop; continue and wait for connection
		}
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if database.Connected() && database.DB != nil {
				if err := database.DB.AutoMigrate(&models.ServerLog{}); err != nil {
					log.Printf("Failed to migrate database: %v", err)
				} else {
					log.Println("AutoMigrate completed")
				}
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	// Wait for termination signal
	<-ctx.Done()
	log.Println("Shutting down...")

	// Attempt graceful shutdown: wait for consumer to stop with timeout
	select {
	case <-done:
		log.Println("Consumer stopped")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for consumer shutdown")
	}

	// Close DB connection
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Shutdown complete")
}
