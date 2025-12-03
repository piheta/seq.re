package commands

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
	"github.com/piheta/seq.re/cmd/cli/crypto"
)

// ImageUpload uploads an image, optionally encrypting it and/or making it one-time
//
//nolint:revive // encrypted and onetime flags are acceptable for control flow
func ImageUpload(apiClient *client.Client, imagePath string, encrypted bool, onetime bool) error {
	// Read image file
	imageData, err := os.ReadFile(imagePath) //nolint:gosec // User-provided path is intentional
	if err != nil {
		return fmt.Errorf("failed to read image file: %w", err)
	}

	var imageURL string
	var keyFragment string

	if encrypted {
		// Generate random AES-128 key
		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}

		// Encrypt the image data (returns base64 string)
		encryptedDataB64, err := crypto.Encrypt(imageData, key)
		if err != nil {
			return fmt.Errorf("failed to encrypt image: %w", err)
		}

		// Decode base64 to get raw encrypted bytes for storage
		encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedDataB64)
		if err != nil {
			return fmt.Errorf("failed to decode encrypted data: %w", err)
		}

		// Send encrypted bytes to server
		imageURL, err = apiClient.CreateEncryptedImage(encryptedBytes, onetime)
		if err != nil {
			return fmt.Errorf("failed to upload encrypted image: %w", err)
		}

		// Encode key for URL fragment
		keyFragment = crypto.EncodeKey(key)
	} else {
		// Send raw image data to server
		imageURL, err = apiClient.CreateImage(imageData, imagePath, onetime)
		if err != nil {
			return fmt.Errorf("failed to upload image: %w", err)
		}
	}

	// Build final URL with fragment if encrypted
	var fullURL string
	if encrypted {
		fullURL = fmt.Sprintf("%s#%s", imageURL, keyFragment)
	} else {
		fullURL = imageURL
	}

	_, _ = fmt.Fprint(os.Stdout, fullURL)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(fullURL); err == nil {
			_, _ = fmt.Fprint(os.Stdout, "     \033[90m\033[2m ✓ copied\033[0m")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
}

// ImageGet retrieves and optionally decrypts an image
func ImageGet(apiClient *client.Client, short string, keyFragment string) error {
	encrypted := keyFragment != ""

	if encrypted {
		// Decode key from fragment
		key, err := crypto.DecodeKey(keyFragment)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}

		// Get encrypted data from server
		encryptedData, err := apiClient.GetImage(short)
		if err != nil {
			return fmt.Errorf("failed to get image: %w", err)
		}

		// Decrypt the image
		plaintext, err := crypto.Decrypt(encryptedData, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt image: %w", err)
		}

		// Write decrypted image to stdout or file
		_, _ = os.Stdout.Write(plaintext)
	} else {
		// Get raw image from server
		imageData, err := apiClient.GetImageRaw(short)
		if err != nil {
			return fmt.Errorf("failed to get image: %w", err)
		}

		// Write raw image to stdout
		_, _ = os.Stdout.Write(imageData)
	}

	return nil
}
