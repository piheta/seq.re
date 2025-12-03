package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
	"github.com/piheta/seq.re/cmd/cli/crypto"
)

// PasteCreate reads a file and creates a paste
//nolint:revive // encrypted and onetime flags are acceptable for control flow
func PasteCreate(apiClient *client.Client, filePath string, language string, encrypted bool, onetime bool) error {
	// Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 -- User-provided file path is intentional
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Auto-detect language from file extension if not provided
	if language == "" {
		language = detectLanguage(filePath)
	}

	var pasteURL string

	if encrypted {
		// Generate random AES-128 key
		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}

		// Encrypt the content
		encryptedData, err := crypto.Encrypt(content, key)
		if err != nil {
			return fmt.Errorf("failed to encrypt content: %w", err)
		}

		// Send encrypted data to server
		pasteURL, err = apiClient.CreatePaste(encryptedData, language, true, onetime)
		if err != nil {
			return fmt.Errorf("failed to create paste: %w", err)
		}

		// Encode key for URL fragment
		keyFragment := crypto.EncodeKey(key)

		// Append key fragment to server URL
		pasteURL = fmt.Sprintf("%s#%s", pasteURL, keyFragment)
	} else {
		// Send plain text to server
		pasteURL, err = apiClient.CreatePaste(string(content), language, false, onetime)
		if err != nil {
			return fmt.Errorf("failed to create paste: %w", err)
		}
	}

	_, _ = fmt.Fprint(os.Stdout, pasteURL)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(pasteURL); err == nil {
			_, _ = fmt.Fprint(os.Stdout, "     \033[90m\033[2m ✓ copied\033[0m")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
}

// PasteGet retrieves a paste by URL or short code
func PasteGet(apiClient *client.Client, urlOrShort string, keyFragment string) error {
	// Extract short code from URL if provided
	short := extractShortFromURL(urlOrShort)

	// Check if URL contains key fragment
	if keyFragment == "" && strings.Contains(urlOrShort, "#") {
		parts := strings.Split(urlOrShort, "#")
		if len(parts) == 2 {
			keyFragment = parts[1]
		}
	}

	var content string
	var err error

	if keyFragment != "" {
		// Encrypted paste
		key, err := crypto.DecodeKey(keyFragment)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}

		// Get encrypted data from server
		encryptedData, err := apiClient.GetPaste(short)
		if err != nil {
			return fmt.Errorf("failed to get paste: %w", err)
		}

		// Decrypt the content
		plaintext, err := crypto.Decrypt(encryptedData, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt paste: %w", err)
		}

		content = string(plaintext)
	} else {
		// Plain text paste
		content, err = apiClient.GetPasteRaw(short)
		if err != nil {
			return fmt.Errorf("failed to get paste: %w", err)
		}
	}

	_, _ = fmt.Fprint(os.Stdout, content)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(content); err == nil {
			_, _ = fmt.Fprint(os.Stdout, "\n     \033[90m\033[2m ✓ copied\033[0m")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
}

// detectLanguage attempts to detect the language from file extension
func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	languageMap := map[string]string{
		".go":         "go",
		".py":         "python",
		".js":         "javascript",
		".ts":         "typescript",
		".jsx":        "javascript",
		".tsx":        "typescript",
		".rs":         "rust",
		".java":       "java",
		".c":          "c",
		".cpp":        "cpp",
		".cc":         "cpp",
		".cxx":        "cpp",
		".h":          "c",
		".hpp":        "cpp",
		".rb":         "ruby",
		".php":        "php",
		".swift":      "swift",
		".kt":         "kotlin",
		".sh":         "bash",
		".bash":       "bash",
		".zsh":        "bash",
		".sql":        "sql",
		".json":       "json",
		".yaml":       "yaml",
		".yml":        "yaml",
		".md":         "markdown",
		".html":       "html",
		".css":        "css",
		".xml":        "xml",
		".dockerfile": "dockerfile",
		".txt":        "plain",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return "plain"
}

// extractShortFromURL extracts the short code from a URL or returns the input if it's already a short code
func extractShortFromURL(urlOrShort string) string {
	// Remove fragment if present
	urlOrShort = strings.Split(urlOrShort, "#")[0]

	// If it's already a 6-character short code, return it
	if len(urlOrShort) == 6 && !strings.Contains(urlOrShort, "/") {
		return urlOrShort
	}

	// Extract short code from URL like http://localhost:8080/p/abc123
	parts := strings.Split(urlOrShort, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return urlOrShort
}
