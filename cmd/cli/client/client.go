package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/piheta/seq.re/cmd/cli/models"
)

// Client handles API requests to the seqre server
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New creates a new API client
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
	}
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

// CreateImage uploads a raw image file
//nolint:revive // onetime flag is acceptable for control flow
func (c *Client) CreateImage(imageData []byte, filename string, onetime bool) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field with actual filename
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return "", fmt.Errorf("failed to write image data: %w", err)
	}

	// Add onetime flag if true
	if onetime {
		if err := writer.WriteField("onetime", "true"); err != nil {
			return "", fmt.Errorf("failed to write onetime field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/images", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var imageURL string
	if err := json.Unmarshal(respBody, &imageURL); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return imageURL, nil
}

// CreateEncryptedImage uploads an encrypted image blob
//nolint:revive // onetime flag is acceptable for control flow
func (c *Client) CreateEncryptedImage(encryptedData []byte, onetime bool) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", "encrypted.bin")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(encryptedData); err != nil {
		return "", fmt.Errorf("failed to write encrypted data: %w", err)
	}

	// Add encrypted flag
	if err := writer.WriteField("encrypted", "true"); err != nil {
		return "", fmt.Errorf("failed to write encrypted field: %w", err)
	}

	// Add onetime flag if true
	if onetime {
		if err := writer.WriteField("onetime", "true"); err != nil {
			return "", fmt.Errorf("failed to write onetime field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/images", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var imageURL string
	if err := json.Unmarshal(respBody, &imageURL); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return imageURL, nil
}

// GetImageRaw retrieves a raw (unencrypted) image by short code
func (c *Client) GetImageRaw(short string) ([]byte, error) {
	resp, err := http.Get(c.BaseURL + "/i/" + short)
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

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return imageData, nil
}

// GetImage retrieves an encrypted image by short code (returns base64 encoded data)
func (c *Client) GetImage(short string) (string, error) {
	resp, err := http.Get(c.BaseURL + "/i/" + short)
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

	// For encrypted images, server returns JSON with base64 data
	var imageResp struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &imageResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return imageResp.Data, nil
}
