package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
)

// URLShorten creates a shortened URL
func URLShorten(apiClient *client.Client, url string) error {
	normalizedURL := normalizeURL(url)
	shortURL, err := apiClient.CreateLink(normalizedURL)
	if err != nil {
		return fmt.Errorf("failed to shorten URL: %w", err)
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
func URLExpand(apiClient *client.Client, short string) error {
	linkResp, err := apiClient.GetLink(short)
	if err != nil {
		return fmt.Errorf("failed to expand URL: %w", err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "URL: %s\n", linkResp.URL)
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
