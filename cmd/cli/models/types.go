//nolint:revive // models package can have many structs
package models

import "time"

// IPResponse represents the response from the IP endpoint
type IPResponse struct {
	IP string `json:"ip"`
}

// LinkRequest represents a request to create a shortened URL
type LinkRequest struct {
	URL       string `json:"url"`
	Encrypted bool   `json:"encrypted"`
	OneTime   bool   `json:"onetime"`
}

// LinkResponse represents link information
type LinkResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// VersionResponse represents version information
type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// SecretRequest represents a request to create a secret
type SecretRequest struct {
	Data string `json:"data"`
}

// SecretResponse represents the response from creating a secret
type SecretResponse struct {
	Short string `json:"short"`
}

// Config represents the CLI configuration
type Config struct {
	Server            string `yaml:"server"`
	AutoCopyClipboard bool   `yaml:"auto_copy_clipboard"`
}

// PasteRequest represents a request to create a paste
type PasteRequest struct {
	Content   string `json:"content"`
	Language  string `json:"language,omitempty"`
	Encrypted bool   `json:"encrypted"`
	OneTime   bool   `json:"onetime"`
}

// PasteResponse represents the response from getting a paste
type PasteResponse struct {
	Data string `json:"data"`
}
