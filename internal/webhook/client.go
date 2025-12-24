package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type OAuthClient struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	
	token     string
	expiresAt time.Time
	mu        sync.RWMutex
	
	HTTPClient *http.Client
}

func NewOAuthClient(tokenURL, clientID, clientSecret, scope string) *OAuthClient {
	return &OAuthClient{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *OAuthClient) Send(ctx context.Context, targetURL string, payload interface{}, overrideClientID, overrideScope string) error {
	token, err := c.getToken(ctx, overrideClientID, overrideScope)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *OAuthClient) getToken(ctx context.Context, overrideID, overrideScope string) (string, error) {
	// If overrides are present, we can't use the shared cache easily without a map.
	// For simplicity, if overrides are used, fetch a new token every time (or implement map cache later).
	if overrideID != "" || overrideScope != "" {
		return c.fetchToken(ctx, overrideID, overrideScope)
	}

	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.expiresAt) {
		defer c.mu.RUnlock()
		return c.token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check
	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	token, err := c.fetchToken(ctx, "", "")
	if err != nil {
		return "", err
	}

	c.token = token
	// Assume 1 hour expiry if not parsed, but usually response has expires_in
	c.expiresAt = time.Now().Add(55 * time.Minute) 
	return token, nil
}

func (c *OAuthClient) fetchToken(ctx context.Context, overrideID, overrideScope string) (string, error) {
	clientID := c.ClientID
	if overrideID != "" {
		clientID = overrideID
	}
	scope := c.Scope
	if overrideScope != "" {
		scope = overrideScope
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("scope", scope)

	req, err := http.NewRequestWithContext(ctx, "POST", c.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token fetch failed: %d", resp.StatusCode)
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.AccessToken, nil
}
