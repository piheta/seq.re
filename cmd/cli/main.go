package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type IPResponse struct {
	IP string `json:"ip"`
}

type LinkRequest struct {
	URL string `json:"url"`
}

type Config struct {
	Server string `yaml:"server"`
}

func getIP(serverURL string) (string, error) {
	resp, err := http.Get(serverURL + "/api/ip")
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ipResp IPResponse
	if err := json.Unmarshal(body, &ipResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return ipResp.IP, nil
}

func getShortenedURL(serverURL, url string) (string, error) {
	linkReq := LinkRequest{URL: url}
	reqBody, err := json.Marshal(linkReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/api/link", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var shortURL string
	if err := json.Unmarshal(body, &shortURL); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return shortURL, nil
}

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprint(os.Stdout, "Usage: seqre <command> [args]\n")
		_, _ = fmt.Fprint(os.Stdout, "Commands:\n")
		_, _ = fmt.Fprint(os.Stdout, "  url <URL>              Create a shortened URL\n")
		_, _ = fmt.Fprint(os.Stdout, "  ip                     Get your IP address\n")
		_, _ = fmt.Fprint(os.Stdout, "  config set <server>    Set the server URL\n")
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle config command specially (doesn't need server connection)
	if command == "config" {
		if len(os.Args) < 4 || os.Args[2] != "set" {
			_, _ = fmt.Fprint(os.Stdout, "Usage: seqre config set <server>\n")
			os.Exit(1)
		}
		err := saveConfig(os.Args[3])
		if err != nil {
			slog.Error("Failed to save config", slog.String("error", err.Error()))
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stdout, "Server URL set to: %s\n", os.Args[3])
		return
	}

	serverURL := getServerURL()

	switch command {
	case "url":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprint(os.Stdout, "Usage: seqre url <URL>\n")
			os.Exit(1)
		}
		url := normalizeURL(os.Args[2])
		shortURL, err := getShortenedURL(serverURL, url)
		if err != nil {
			slog.Error("Failed to shorten URL", slog.String("error", err.Error()))
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, shortURL)

	case "ip":
		ip, err := getIP(serverURL)
		if err != nil {
			slog.Error("Failed to get IP", slog.String("error", err.Error()))
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stdout, ip)

	default:
		slog.Error("Unknown command", slog.String("command", command))
		os.Exit(1)
	}
}

func normalizeURL(input string) string {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "https://" + input
	}
	return input
}

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "seqre", "config")
}

func loadConfig() (Config, error) {
	configPath := getConfigPath()
	if configPath == "" {
		return Config{}, nil
	}

	data, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func saveConfig(serverURL string) error {
	configPath := getConfigPath()
	if configPath == "" {
		return errors.New("could not determine config path")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := Config{Server: serverURL}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func getServerURL() string {
	if url := os.Getenv("SEQRE_SERVER"); url != "" {
		return url
	}

	cfg, err := loadConfig()
	if err != nil {
		slog.Warn("Failed to load config", slog.String("error", err.Error()))
	}
	if cfg.Server != "" {
		return cfg.Server
	}

	return "http://localhost:8081"
}
