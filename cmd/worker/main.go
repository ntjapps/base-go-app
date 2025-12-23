package main

import (
	"base-go-app/internal/config"
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"base-go-app/internal/queue"
	"log"
)

func main() {
	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Database
	err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate the schema
	err = database.DB.AutoMigrate(&models.ServerLog{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Start Queue Consumer
	queue.StartConsumer(cfg)
}
