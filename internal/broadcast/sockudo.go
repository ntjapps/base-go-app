package broadcast

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SockudoBroadcaster struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewSockudoBroadcaster() *SockudoBroadcaster {
	return &SockudoBroadcaster{
		BaseURL: os.Getenv("SOCKUDO_URL"),
		APIKey:  os.Getenv("SOCKUDO_KEY"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *SockudoBroadcaster) Broadcast(ctx context.Context, channel, event string, payload interface{}) error {
	if s.BaseURL == "" {
		return nil // Not configured
	}

	url := fmt.Sprintf("%s/api/v1/broadcast", s.BaseURL)
	
	body := map[string]interface{}{
		"channel": channel,
		"event":   event,
		"data":    payload,
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))
	req.Header.Set("X-App-Key", s.APIKey) // Support both styles

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sockudo broadcast failed with status: %d", resp.StatusCode)
	}

	return nil
}
