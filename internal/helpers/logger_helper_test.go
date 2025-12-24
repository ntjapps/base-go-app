package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaravelLogPayload(t *testing.T) {
	payload := LaravelLogPayload("test message", "error", nil, nil)

	assert.Equal(t, "test message", payload.Message)
	assert.Equal(t, "400", payload.Level)
	assert.Equal(t, "ERROR", payload.LevelName)
	assert.Equal(t, "celery", payload.Channel)
	assert.NotEmpty(t, payload.ID)
	assert.NotEmpty(t, payload.Datetime)
}

func TestLaravelLogPayload_Levels(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "100"},
		{"info", "200"},
		{"warning", "300"},
		{"error", "400"},
		{"critical", "500"},
	}

	for _, tt := range tests {
		p := LaravelLogPayload("msg", tt.level, nil, nil)
		assert.Equal(t, tt.expected, p.Level)
	}
}
