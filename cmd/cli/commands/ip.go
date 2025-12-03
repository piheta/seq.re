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
	_, _ = fmt.Fprint(os.Stdout, ip)

	cfg, _ := config.Load()
	if cfg.AutoCopyClipboard {
		if err := clipboard.WriteAll(ip); err == nil {
			_, _ = fmt.Fprint(os.Stdout, "     \033[90m\033[2m ✓ copied\033[0m")
		}
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
}
