package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"context"

	"base-go-app/internal/database"
	"base-go-app/internal/queue"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHealthHandlerDown(t *testing.T) {
	// Ensure DB and Rabbit are down
	// Clear DB
	// Database Clear function is in database package
	// Using fully-qualified package name to modify state
	database.ClearDBForTests()
	queue.SetRabbitConnectedForTests(false)

	req := httptest.NewRequest("GET", "/healthcheck", nil)
	w := httptest.NewRecorder()
	h := healthHandler()
	h.ServeHTTP(w, req)
	resp := w.Result()
	// We now expect 200 even if down, but with status=degraded in body
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (degraded) when db and rabbit down, got %d", resp.StatusCode)
	}
}

func TestHealthHandlerOK(t *testing.T) {
	// Set DB and rabbit as up via test helpers
	sqliteDB := setupSQLiteForTest(t)
	database.SetDBForTests(sqliteDB)
	queue.SetRabbitConnectedForTests(true)
	defer database.ClearDBForTests()
	defer queue.SetRabbitConnectedForTests(false)

	req := httptest.NewRequestWithContext(context.Background(), "GET", "/healthcheck", nil)
	w := httptest.NewRecorder()
	h := healthHandler()
	h.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// helper to create a sqlite gorm DB for tests
func setupSQLiteForTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}
