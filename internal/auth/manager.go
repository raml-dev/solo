// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"solo/internal/collection"
	"solo/internal/tools"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Token represents a stored access token with its metadata.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthStore is the structure of the encrypted JSON file.
type AuthStore struct {
	Tokens       map[string]Token  `json:"tokens"`
	BearerTokens map[string]string `json:"bearerTokens,omitempty"`
}

type AuthManager struct {
	configDir string
	fileName  string
	ctx       context.Context
	mu        sync.Mutex
}

func NewAuthManager(configDir string) *AuthManager {
	return &AuthManager{
		configDir: configDir,
		fileName:  "auth_store.json",
	}
}

func (m *AuthManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// GetOrRefreshToken returns a valid token, refreshing it if necessary.
// Returns (token, refreshed, error)
func (m *AuthManager) GetOrRefreshToken(ctx context.Context, config collection.AuthConfiguration) (string, bool, error) {
	if config.EffectiveMode() != collection.AuthModeOAuth2 {
		return "", false, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	hash := config.Hash()
	store, err := m.loadStore()
	if err != nil {
		return "", false, err
	}

	token, exists := store.Tokens[hash]
	// Proactive refresh: if token expires in less than 30 seconds
	if exists && time.Until(token.ExpiresAt) > 30*time.Second {
		return token.AccessToken, false, nil
	}

	// Fetch new token
	newToken, err := m.fetchToken(ctx, config)
	if err != nil {
		return "", false, err
	}

	// Update store
	if store.Tokens == nil {
		store.Tokens = make(map[string]Token)
	}
	store.Tokens[hash] = *newToken
	if err := m.saveStore(store); err != nil {
		return "", false, fmt.Errorf("failed to save token: %w", err)
	}

	return newToken.AccessToken, true, nil
}

// StoreBearerToken encrypts and persists a bearer token, returning its stable identifier.
func (m *AuthManager) StoreBearerToken(tokenID, token string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	store, err := m.loadStore()
	if err != nil {
		return "", err
	}
	if tokenID == "" {
		tokenID = uuid.NewString()
	}
	if store.BearerTokens == nil {
		store.BearerTokens = make(map[string]string)
	}
	store.BearerTokens[tokenID] = token
	if err := m.saveStore(store); err != nil {
		return "", fmt.Errorf("failed to save bearer token: %w", err)
	}
	return tokenID, nil
}

// GetBearerToken decrypts and returns a bearer token by identifier.
func (m *AuthManager) GetBearerToken(tokenID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(tokenID) == "" {
		return "", fmt.Errorf("bearer token is not configured")
	}
	store, err := m.loadStore()
	if err != nil {
		return "", err
	}
	token, ok := store.BearerTokens[tokenID]
	if !ok {
		return "", fmt.Errorf("bearer token %q was not found on this machine", tokenID)
	}
	return token, nil
}

// DeleteBearerToken removes a bearer token from the encrypted store.
func (m *AuthManager) DeleteBearerToken(tokenID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(tokenID) == "" {
		return nil
	}
	store, err := m.loadStore()
	if err != nil {
		return err
	}
	if _, ok := store.BearerTokens[tokenID]; !ok {
		return nil
	}
	delete(store.BearerTokens, tokenID)
	return m.saveStore(store)
}

func (m *AuthManager) fetchToken(ctx context.Context, config collection.AuthConfiguration) (*Token, error) {
	data := url.Values{}
	for k, v := range config.Template {
		data.Set(k, v)
	}

	slog.Debug("Attempting OAuth2 token acquisition",
		"url", config.TokenURL,
		"params_count", len(config.Template))

	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("OAuth2 request network error", "url", config.TokenURL, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Prepare redacted template for event emission
	var eventData map[string]interface{}
	if m.ctx != nil {
		templateCopy := make(map[string]string)
		for k, v := range config.Template {
			if k == "client_secret" || k == "password" {
				templateCopy[k] = "[REDACTED]"
			} else {
				templateCopy[k] = v
			}
		}
		eventData = map[string]interface{}{
			"url":    config.TokenURL,
			"method": "POST",
			"params": templateCopy,
			"status": resp.StatusCode,
			"body":   "[REDACTED]",
		}
	}

	if resp.StatusCode >= 400 {
		slog.Error("OAuth2 provider returned error",
			"status", resp.StatusCode)

		if m.ctx != nil {
			eventData["error"] = true
			runtime.EventsEmit(m.ctx, "auth:token-refreshed", eventData)
		}

		return nil, fmt.Errorf("auth request failed with status %d", resp.StatusCode)
	}

	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "auth:token-refreshed", eventData)
	}

	slog.Debug("OAuth2 response received", "body_length", len(body))

	tokenPath := config.TokenPath
	if tokenPath == "" {
		tokenPath = "access_token"
	}

	accessToken := gjson.GetBytes(body, tokenPath).String()
	if accessToken == "" {
		return nil, fmt.Errorf("could not find token at path %q in response", tokenPath)
	}

	expiresIn := gjson.GetBytes(body, "expires_in").Int()
	if expiresIn == 0 {
		expiresIn = 3600 // Default 1 hour if not provided
	}

	return &Token{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}

func (m *AuthManager) loadStore() (*AuthStore, error) {
	data, err := tools.ReadConfigFile(m.configDir, m.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return newAuthStore(), nil
		}
		return nil, fmt.Errorf("failed to read auth store: %w", err)
	}

	decrypted, err := Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt auth store: %w", err)
	}

	var store AuthStore
	if err := json.Unmarshal(decrypted, &store); err != nil {
		return nil, fmt.Errorf("failed to parse auth store: %w", err)
	}
	if store.Tokens == nil {
		store.Tokens = make(map[string]Token)
	}
	if store.BearerTokens == nil {
		store.BearerTokens = make(map[string]string)
	}

	return &store, nil
}

func newAuthStore() *AuthStore {
	return &AuthStore{
		Tokens:       make(map[string]Token),
		BearerTokens: make(map[string]string),
	}
}

func (m *AuthManager) saveStore(store *AuthStore) error {
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	encrypted, err := Encrypt(data)
	if err != nil {
		return err
	}

	return tools.CreateConfigFile(m.configDir, m.fileName, encrypted)
}
