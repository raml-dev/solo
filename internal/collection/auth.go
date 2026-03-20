package collection

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// AuthConfiguration defines how to obtain a token for a collection or request.
type AuthConfiguration struct {
	Enabled   bool              `json:"enabled"`
	TokenURL  string            `json:"tokenUrl"`
	Template  map[string]string `json:"template"`
	TokenPath string            `json:"tokenPath"`
}

// Hash generates a unique identifier for an AuthConfiguration.
func (c *AuthConfiguration) Hash() string {
	h := sha256.New()
	h.Write([]byte(c.TokenURL))
	// Simple marshal for stable hash of the template
	templateBytes, _ := json.Marshal(c.Template)
	h.Write(templateBytes)
	return hex.EncodeToString(h.Sum(nil))
}
