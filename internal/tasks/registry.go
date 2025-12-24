package tasks

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]TaskHandler)
	mu       sync.RWMutex
)

// RegisterTask registers a handler for a given task name.
// It panics if a handler is already registered for the name.
func RegisterTask(name string, h TaskHandler) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("task handler already registered for %s", name))
	}
	registry[name] = h
}

// LookupTask returns the handler for the given task name.
func LookupTask(name string) (TaskHandler, bool) {
	mu.RLock()
	defer mu.RUnlock()

	h, ok := registry[name]
	return h, ok
}

// ClearRegistry clears all registered tasks (useful for tests).
func ClearRegistry() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]TaskHandler)
}
