package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
	"github.com/piheta/seq.re/cmd/cli/crypto"
)

// URLShorten creates a shortened URL
//
//nolint:revive // encrypted and onetime flags are acceptable for control flow
func URLShorten(apiClient *client.Client, url string, encrypted bool, onetime bool) error {
	normalizedURL := normalizeURL(url)

	var shortURL string
	var err error

	if encrypted {
		// Generate random AES-128 key
		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}

		// Encrypt the URL
		encryptedURL, err := crypto.Encrypt([]byte(normalizedURL), key)
		if err != nil {
			return fmt.Errorf("failed to encrypt URL: %w", err)
		}

		// Send encrypted URL to server
		shortURL, err = apiClient.CreateLink(encryptedURL, true, onetime)
		if err != nil {
			return fmt.Errorf("failed to shorten URL: %w", err)
		}

		// Encode key for URL fragment
		keyFragment := crypto.EncodeKey(key)

		// Append key fragment to server URL
		shortURL = fmt.Sprintf("%s#%s", shortURL, keyFragment)
	} else {
		// Send plain URL to server
		shortURL, err = apiClient.CreateLink(normalizedURL, false, onetime)
		if err != nil {
			return fmt.Errorf("failed to shorten URL: %w", err)
		}
	}

	_, _ = fmt.Fprint(os.Stdout, shortURL)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(shortURL); err == nil {
			_, _ = fmt.Fprint(os.Stdout, "     \033[90m\033[2m ✓ copied\033[0m")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
}

// URLExpand expands a shortened URL and displays its information
func URLExpand(apiClient *client.Client, short string, keyFragment string) error {
	linkResp, err := apiClient.GetLink(short)
	if err != nil {
		return fmt.Errorf("failed to expand URL: %w", err)
	}

	url := linkResp.URL

	// If key fragment is provided, decrypt the URL
	if keyFragment != "" {
		key, err := crypto.DecodeKey(keyFragment)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}

		// Decrypt the URL
		plaintext, err := crypto.Decrypt(url, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt URL: %w", err)
		}

		url = string(plaintext)
	}

	_, _ = fmt.Fprintf(os.Stdout, "URL: %s\n", url)
	_, _ = fmt.Fprintf(os.Stdout, "Expires: %s\n", linkResp.ExpiresAt.Format(time.RFC3339))

	return nil
}

// normalizeURL ensures URL has a protocol scheme
func normalizeURL(input string) string {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "https://" + input
	}
	return input
}
