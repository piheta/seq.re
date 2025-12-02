package commands

import (
	"fmt"
	"os"

	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/config"
)

// ConfigSet sets the server URL in configuration
func ConfigSet(serverURL, cliVersion string) error {
	// Try to check server version
	apiClient := client.New(serverURL)
	serverVersion, err := apiClient.GetVersion()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Warning: Could not verify server version: %v\n", err)
	} else if serverVersion.Version != cliVersion {
		_, _ = fmt.Fprint(os.Stdout, "Warning: Version mismatch detected!\n")
		_, _ = fmt.Fprintf(os.Stdout, "  CLI version:    %s\n", cliVersion)
		_, _ = fmt.Fprintf(os.Stdout, "  Server version: %s\n", serverVersion.Version)
		_, _ = fmt.Fprint(os.Stdout, "This may cause compatibility issues.\n\n")
	}

	cfg, _ := config.Load()
	cfg.Server = serverURL
	err = config.Save(cfg)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Server URL set to: %s\n", serverURL)

	return nil
}

// ConfigGet displays the current configuration
func ConfigGet() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Server == "" {
		_, _ = fmt.Fprint(os.Stdout, "No server URL configured\n")
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "Server URL: %s\n", cfg.Server)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Auto-copy clipboard: %v\n", cfg.AutoCopyClipboard)

	return nil
}

// ConfigEnableClipboard enables auto-copy to clipboard
func ConfigEnableClipboard() error {
	return setClipboard(true)
}

// ConfigDisableClipboard disables auto-copy to clipboard
func ConfigDisableClipboard() error {
	return setClipboard(false)
}

//nolint:revive // flag parameter is acceptable for private helper
func setClipboard(enabled bool) error {
	cfg, _ := config.Load()
	cfg.AutoCopyClipboard = enabled

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if enabled {
		_, _ = fmt.Fprint(os.Stdout, "Auto-copy to clipboard enabled\n")
	} else {
		_, _ = fmt.Fprint(os.Stdout, "Auto-copy to clipboard disabled\n")
	}

	return nil
}
