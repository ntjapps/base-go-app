package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthClient_Send(t *testing.T) {
	// Mock Token Server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token", r.URL.Path)
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-token",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer tokenServer.Close()

	// Mock Target Server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook", r.URL.Path)
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer targetServer.Close()

	client := NewOAuthClient(tokenServer.URL+"/token", "id", "secret", "scope")
	client.HTTPClient = tokenServer.Client() // Use the test client

	err := client.Send(context.Background(), targetServer.URL+"/webhook", map[string]string{"foo": "bar"}, "", "")
	assert.NoError(t, err)
}

func TestOAuthClient_Send_TokenError(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tokenServer.Close()

	client := NewOAuthClient(tokenServer.URL, "id", "secret", "scope")
	client.HTTPClient = tokenServer.Client()

	err := client.Send(context.Background(), "http://example.com", nil, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get token")
}
