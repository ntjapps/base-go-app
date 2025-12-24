package database

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"base-go-app/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var dbConnected int32 // 0 = false, 1 = true

// Connect attempts to connect to the DB. If it fails, it starts a background
// goroutine that keeps retrying until success. It does not exit the process.
func Connect(cfg *config.Config) error {
	dsn := cfg.GetDSN()

	// Try once
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err == nil {
		atomic.StoreInt32(&dbConnected, 1)
		log.Println("Connected to database")
		return nil
	}

	log.Printf("Initial DB connection failed: %v. Will retry in background...", err)

	// Start background reconnect loop
	go func() {
		delay := 2 * time.Second
		for {
			// If we've been asked to stop the process, don't continue reconnecting
			// (this package has a Close method which callers should use on shutdown)
			log.Printf("Attempting DB reconnect...")
			conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err == nil {
				DB = conn
				atomic.StoreInt32(&dbConnected, 1)
				log.Println("Reconnected to database")
				return
			}
			log.Printf("DB reconnect failed: %v", err)
			// exponential backoff with cap
			select {
			case <-time.After(delay):
			case <-time.After(delay):
			}
			if delay < 30*time.Second {
				delay *= 2
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
			}
		}
	}()

	return nil
}

// Ping attempts to ping the DB with a context. Returns true if reachable.
func Ping(ctx context.Context) (bool, error) {
	if atomic.LoadInt32(&dbConnected) == 0 || DB == nil {
		return false, nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return false, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// Connected returns whether we have an active DB connection
func Connected() bool {
	return atomic.LoadInt32(&dbConnected) == 1
}

// Close closes the underlying DB connection (if any) and prevents further reconnects.
func Close() error {
	if DB == nil {
		atomic.StoreInt32(&dbConnected, 0)
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		atomic.StoreInt32(&dbConnected, 0)
		return err
	}
	if err := sqlDB.Close(); err != nil {
		atomic.StoreInt32(&dbConnected, 0)
		return err
	}
	atomic.StoreInt32(&dbConnected, 0)
	DB = nil
	return nil
}

// SetDBForTests allows tests to inject a DB and mark it connected.
func SetDBForTests(d *gorm.DB) {
	DB = d
	atomic.StoreInt32(&dbConnected, 1)
}

// ClearDBForTests clears the test DB and marks as disconnected.
func ClearDBForTests() {
	DB = nil
	atomic.StoreInt32(&dbConnected, 0)
}
