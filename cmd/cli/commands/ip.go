package commands

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
)

// IP retrieves and displays the public IP address
func IP(apiClient *client.Client) error {
	ip, err := apiClient.GetIP()
	if err != nil {
		return fmt.Errorf("failed to get IP: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stdout, ip)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(ip); err == nil {
			_, _ = fmt.Fprintln(os.Stdout, "Copied to clipboard")
		}
	}

	return nil
}
