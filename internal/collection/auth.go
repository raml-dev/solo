// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const (
	AuthModeNone   = "none"
	AuthModeOAuth2 = "oauth2"
	AuthModeBearer = "bearer"
)

// AuthSecretStore persists bearer credentials outside collection JSON.
type AuthSecretStore interface {
	StoreBearerToken(tokenID, token string) (string, error)
	GetBearerToken(tokenID string) (string, error)
	DeleteBearerToken(tokenID string) error
}

// AuthConfiguration defines the mutually exclusive authentication mode for a request.
type AuthConfiguration struct {
	Enabled       bool              `json:"enabled,omitempty"`
	Mode          string            `json:"mode,omitempty"`
	BearerToken   string            `json:"bearerToken,omitempty"`
	BearerTokenID string            `json:"bearerTokenId,omitempty"`
	TokenURL      string            `json:"tokenUrl,omitempty"`
	Template      map[string]string `json:"template,omitempty"`
	TokenPath     string            `json:"tokenPath,omitempty"`
}

// EffectiveMode returns the normalized mode, including legacy OAuth2 configurations.
func (c *AuthConfiguration) EffectiveMode() string {
	if c == nil {
		return AuthModeNone
	}

	switch strings.ToLower(strings.TrimSpace(c.Mode)) {
	case AuthModeBearer:
		return AuthModeBearer
	case AuthModeOAuth2:
		return AuthModeOAuth2
	case AuthModeNone:
		return AuthModeNone
	default:
		if c.Enabled {
			return AuthModeOAuth2
		}
		return AuthModeNone
	}
}

// Normalize updates legacy and incomplete authentication data to the canonical shape.
func (c *AuthConfiguration) Normalize() {
	if c == nil {
		return
	}

	c.Mode = c.EffectiveMode()
	c.Enabled = c.Mode == AuthModeOAuth2
	if c.TokenPath == "" {
		c.TokenPath = "access_token"
	}
	if c.Template == nil {
		c.Template = map[string]string{}
	}
}

// Hash generates a unique identifier for an AuthConfiguration.
func (c *AuthConfiguration) Hash() string {
	h := sha256.New()
	h.Write([]byte(c.EffectiveMode()))
	h.Write([]byte(c.TokenURL))
	h.Write([]byte(c.TokenPath))
	// Simple marshal for stable hash of the template
	templateBytes, _ := json.Marshal(c.Template)
	h.Write(templateBytes)
	return hex.EncodeToString(h.Sum(nil))
}
