package database

import (
	"context"
	"testing"
	"time"

	"base-go-app/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnectRetriesDoesNotCrash(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "9999",
		DBUser:     "dummy",
		DBPassword: "dummy",
		DBDatabase: "nosuchdb",
	}
	// Should not return an error, but not be connected
	if err := Connect(cfg); err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	// Ensure Connected() is false immediately.
	if Connected() {
		t.Fatalf("expected not connected immediately")
	}
	// Now inject a sqlite in-memory DB to simulate a successful reconnect
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	SetDBForTests(sqliteDB)
	// Wait a moment and confirm Connected() is true
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		if Connected() {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("db did not become connected in time")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	ClearDBForTests()
}
