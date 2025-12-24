package broadcast

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSockudoBroadcaster_Broadcast(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/broadcast", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "test-key", r.Header.Get("X-App-Key"))

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "test-channel", body["channel"])
		assert.Equal(t, "test-event", body["event"])
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b := &SockudoBroadcaster{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		HTTPClient: server.Client(),
	}

	err := b.Broadcast(context.Background(), "test-channel", "test-event", map[string]string{"foo": "bar"})
	assert.NoError(t, err)
}

func TestSockudoBroadcaster_Broadcast_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	b := &SockudoBroadcaster{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		HTTPClient: server.Client(),
	}

	err := b.Broadcast(context.Background(), "test-channel", "test-event", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sockudo broadcast failed with status: 500")
}

func TestSockudoBroadcaster_Broadcast_NoConfig(t *testing.T) {
	b := &SockudoBroadcaster{
		BaseURL: "",
	}
	err := b.Broadcast(context.Background(), "ch", "ev", nil)
	assert.NoError(t, err)
}
