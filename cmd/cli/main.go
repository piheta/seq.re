package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

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
		fmt.Println("Usage: seqre <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  url <URL>              Create a shortened URL")
		fmt.Println("  ip                     Get your IP address")
		fmt.Println("  config set <server>    Set the server URL")
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle config command specially (doesn't need server connection)
	if command == "config" {
		if len(os.Args) < 4 || os.Args[2] != "set" {
			fmt.Println("Usage: seqre config set <server>")
			os.Exit(1)
		}
		err := saveConfig(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Server URL set to: %s\n", os.Args[3])
		return
	}

	serverURL := getServerURL()

	switch command {
	case "url":
		if len(os.Args) < 3 {
			fmt.Println("Usage: seqre url <URL>")
			os.Exit(1)
		}
		url := normalizeURL(os.Args[2])
		shortURL, err := getShortenedURL(serverURL, url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(shortURL)

	case "ip":
		ip, err := getIP(serverURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(ip)

	default:
		fmt.Printf("Unknown command: %s\n", command)
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

	data, err := os.ReadFile(configPath)
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
		return fmt.Errorf("could not determine config path")
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := Config{Server: serverURL}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
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
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
	if cfg.Server != "" {
		return cfg.Server
	}

	return "http://localhost:8081"
}
