package commands

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
	"github.com/piheta/seq.re/cmd/cli/crypto"
)

// SecretCreate encrypts and creates a secret, returning a URL with fragment
func SecretCreate(apiClient *client.Client, secret string) error {
	// Generate random AES-128 key
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Encrypt the secret
	encryptedData, err := crypto.Encrypt([]byte(secret), key)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Send encrypted data to server - returns full URL without fragment
	secretURL, err := apiClient.CreateSecret(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Encode key for URL fragment
	keyFragment := crypto.EncodeKey(key)

	// Append key fragment to server URL
	fullURL := fmt.Sprintf("%s#%s", secretURL, keyFragment)
	_, _ = fmt.Fprintln(os.Stdout, fullURL)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(fullURL); err == nil {
			_, _ = fmt.Fprintln(os.Stdout, "Copied to clipboard")
		}
	}

	return nil
}

// SecretGet retrieves and decrypts a secret using the short code and key fragment
func SecretGet(apiClient *client.Client, short string, keyFragment string) error {
	// Decode key from fragment
	key, err := crypto.DecodeKey(keyFragment)
	if err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	// Get encrypted data from server
	encryptedData, err := apiClient.GetSecret(short)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	// Decrypt the secret
	plaintext, err := crypto.Decrypt(encryptedData, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, string(plaintext))

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(string(plaintext)); err == nil {
			_, _ = fmt.Fprintln(os.Stdout, "Copied to clipboard")
		}
	}

	return nil
}
