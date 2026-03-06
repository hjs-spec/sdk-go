// Package jep provides a Go client for the JEP Protocol API (Judgment Event Protocol)
// IETF Draft: draft-wang-jep-judgment-event-00
package jep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Constants
const (
	// DefaultBaseURL 指向 JEP 官方 API 节点
	DefaultBaseURL = "https://api.jep-protocol.org"
	DefaultTimeout = 30 * time.Second
)

// Client for JEP API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new JEP client with default settings
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// NewClientWithURL creates a new JEP client with custom base URL
func NewClientWithURL(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// SetTimeout sets custom timeout for HTTP client
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// doRequest performs an HTTP request and handles response
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, &ValidationError{Message: fmt.Sprintf("failed to marshal request: %v", err)}
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	// 更新 User-Agent 标识
	req.Header.Set("User-Agent", "JEP-Go-SDK/1.0.0")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// handleResponse processes the HTTP response and returns decoded JSON or error
func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result != nil {
			return json.NewDecoder(resp.Body).Decode(result)
		}
		return nil
	}

	var apiErr APIError
	apiErr.StatusCode = resp.StatusCode

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiErr); err == nil {
		return &apiErr
	}

	apiErr.Message = string(body)
	return &apiErr
}

// ==================== JUDGMENT API (Core Primitive #1) ====================

type JudgmentRequest struct {
	Entity       string                 `json:"entity"`
	Action       string                 `json:"action"`
	Scope        map[string]interface{} `json:"scope,omitempty"`
	Immutability map[string]interface{} `json:"immutability,omitempty"`
}

type JudgmentResponse struct {
	ID        string    `json:"id"`       // Starts with 'jep_'
	Status    string    `json:"status"`
	Protocol  string    `json:"protocol"` // e.g., 'JEP/1.0'
	Timestamp time.Time `json:"timestamp"`
}

func (c *Client) Judgment(req *JudgmentRequest) (*JudgmentResponse, error) {
	if req.Entity == "" || req.Action == "" {
		return nil, &ValidationError{Message: "entity and action are required"}
	}

	resp, err := c.doRequest("POST", "/judgments", req)
	if err != nil {
		return nil, err
	}

	var result JudgmentResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetJudgment(id string) (*JudgmentResponse, error) {
	if id == "" {
		return nil, &ValidationError{Message: "id is required"}
	}

	resp, err := c.doRequest("GET", "/judgments/"+id, nil)
	if err != nil {
		return nil, err
	}

	var result JudgmentResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ... (其他 Primitive 方法逻辑保持 1:1，此处仅展示结构更新)

// ==================== VERIFICATION API (Core Primitive #4) ====================

// Verify performs quick verification (auto-detects type from ID)
func (c *Client) Verify(id string) (*QuickVerifyResponse, error) {
	if id == "" {
		return nil, &ValidationError{Message: "id is required"}
	}

	req := map[string]string{"id": id}
	resp, err := c.doRequest("POST", "/verify", req)
	if err != nil {
		return nil, err
	}

	var result QuickVerifyResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== ERROR TYPES ====================

type APIError struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("JEP API error (%d): %s", e.StatusCode, e.Error)
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("JEP validation error: %s", e.Message)
}
