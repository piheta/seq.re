package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/piheta/seq.re/cmd/cli/models"
)

// Client handles API requests to the seqre server
type Client struct {
	BaseURL string
}

// New creates a new API client
func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// GetIP retrieves the public IP address
func (c *Client) GetIP() (string, error) {
	resp, err := http.Get(c.BaseURL + "/api/ip")
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

	var ipResp models.IPResponse
	if err := json.Unmarshal(body, &ipResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return ipResp.IP, nil
}

// GetVersion retrieves the server version information
func (c *Client) GetVersion() (*models.VersionResponse, error) {
	resp, err := http.Get(c.BaseURL + "/api/version")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var versionResp models.VersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &versionResp, nil
}

// GetLink retrieves link information by short code
func (c *Client) GetLink(short string) (*models.LinkResponse, error) {
	resp, err := http.Get(c.BaseURL + "/api/links/" + short)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var linkResp models.LinkResponse
	if err := json.Unmarshal(body, &linkResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &linkResp, nil
}

// CreateLink creates a shortened URL
func (c *Client) CreateLink(url string) (string, error) {
	linkReq := models.LinkRequest{URL: url}
	reqBody, err := json.Marshal(linkReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/api/links", "application/json", bytes.NewBuffer(reqBody))
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

// CreateSecret creates a new secret and returns the full URL
func (c *Client) CreateSecret(encryptedData string) (string, error) {
	secretReq := models.SecretRequest{Data: encryptedData}
	reqBody, err := json.Marshal(secretReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/api/secrets", "application/json", bytes.NewBuffer(reqBody))
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

	// Server returns full URL as a string
	var fullURL string
	if err := json.Unmarshal(body, &fullURL); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return fullURL, nil
}

// GetSecret retrieves a secret by short code
func (c *Client) GetSecret(short string) (string, error) {
	resp, err := http.Get(c.BaseURL + "/api/secrets/" + short)
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

	var secretReq models.SecretRequest
	if err := json.Unmarshal(body, &secretReq); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return secretReq.Data, nil
}
