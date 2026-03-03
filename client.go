// Package hjs provides a Go client for the HJS Protocol API
package hjs

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
	DefaultBaseURL = "https://api.hjs.sh"
	DefaultTimeout = 30 * time.Second
)

// Client for HJS API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new HJS client with default settings
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// NewClientWithURL creates a new HJS client with custom base URL
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
	req.Header.Set("User-Agent", "HJS-Go-SDK/0.1.0")
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

	// Parse error response
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

// JudgmentRequest represents a judgment creation request
type JudgmentRequest struct {
	Entity       string                 `json:"entity"`
	Action       string                 `json:"action"`
	Scope        map[string]interface{} `json:"scope,omitempty"`
	Immutability map[string]interface{} `json:"immutability,omitempty"`
}

// JudgmentResponse represents a judgment creation response
type JudgmentResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Protocol  string    `json:"protocol"`
	Timestamp time.Time `json:"timestamp"`
}

// Judgment creates a new judgment record
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

// GetJudgment retrieves a judgment by ID
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

// ListJudgmentsParams represents parameters for listing judgments
type ListJudgmentsParams struct {
	Entity string `json:"entity,omitempty"`
	Page   int    `json:"page,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// ListJudgmentsResponse represents a paginated list of judgments
type ListJudgmentsResponse struct {
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int                `json:"total"`
	Data    []JudgmentResponse `json:"data"`
}

// ListJudgments lists judgments with optional filters
func (c *Client) ListJudgments(params *ListJudgmentsParams) (*ListJudgmentsResponse, error) {
	resp, err := c.doRequest("GET", "/judgments", params)
	if err != nil {
		return nil, err
	}

	var result ListJudgmentsResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== DELEGATION API (Core Primitive #2) ====================

// DelegationRequest represents a delegation creation request
type DelegationRequest struct {
	Delegator  string                 `json:"delegator"`
	Delegatee  string                 `json:"delegatee"`
	JudgmentID string                 `json:"judgment_id,omitempty"`
	Scope      map[string]interface{} `json:"scope,omitempty"`
	Expiry     string                 `json:"expiry,omitempty"`
}

// DelegationResponse represents a delegation creation response
type DelegationResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Delegator string    `json:"delegator"`
	Delegatee string    `json:"delegatee"`
	Scope     interface{} `json:"scope"`
	Expiry    string    `json:"expiry,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Delegation creates a new delegation
func (c *Client) Delegation(req *DelegationRequest) (*DelegationResponse, error) {
	if req.Delegator == "" || req.Delegatee == "" {
		return nil, &ValidationError{Message: "delegator and delegatee are required"}
	}

	resp, err := c.doRequest("POST", "/delegations", req)
	if err != nil {
		return nil, err
	}

	var result DelegationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDelegation retrieves a delegation by ID
func (c *Client) GetDelegation(id string) (*DelegationResponse, error) {
	if id == "" {
		return nil, &ValidationError{Message: "id is required"}
	}

	resp, err := c.doRequest("GET", "/delegations/"+id, nil)
	if err != nil {
		return nil, err
	}

	var result DelegationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListDelegationsParams represents parameters for listing delegations
type ListDelegationsParams struct {
	Delegator string `json:"delegator,omitempty"`
	Delegatee string `json:"delegatee,omitempty"`
	Status    string `json:"status,omitempty"`
	Page      int    `json:"page,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// ListDelegationsResponse represents a paginated list of delegations
type ListDelegationsResponse struct {
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
	Total int                 `json:"total"`
	Data  []DelegationResponse `json:"data"`
}

// ListDelegations lists delegations with optional filters
func (c *Client) ListDelegations(params *ListDelegationsParams) (*ListDelegationsResponse, error) {
	resp, err := c.doRequest("GET", "/delegations", params)
	if err != nil {
		return nil, err
	}

	var result ListDelegationsResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== TERMINATION API (Core Primitive #3) ====================

// TerminationRequest represents a termination creation request
type TerminationRequest struct {
	Terminator string `json:"terminator"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
	Reason     string `json:"reason,omitempty"`
}

// TerminationResponse represents a termination creation response
type TerminationResponse struct {
	ID         string    `json:"id"`
	Terminator string    `json:"terminator"`
	TargetID   string    `json:"target_id"`
	TargetType string    `json:"target_type"`
	Reason     string    `json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Termination creates a new termination
func (c *Client) Termination(req *TerminationRequest) (*TerminationResponse, error) {
	if req.Terminator == "" || req.TargetID == "" || req.TargetType == "" {
		return nil, &ValidationError{Message: "terminator, target_id, and target_type are required"}
	}
	if req.TargetType != "judgment" && req.TargetType != "delegation" {
		return nil, &ValidationError{Message: "target_type must be 'judgment' or 'delegation'"}
	}

	resp, err := c.doRequest("POST", "/terminations", req)
	if err != nil {
		return nil, err
	}

	var result TerminationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTermination retrieves a termination by ID
func (c *Client) GetTermination(id string) (*TerminationResponse, error) {
	if id == "" {
		return nil, &ValidationError{Message: "id is required"}
	}

	resp, err := c.doRequest("GET", "/terminations/"+id, nil)
	if err != nil {
		return nil, err
	}

	var result TerminationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== VERIFICATION API (Core Primitive #4) ====================

// VerificationRequest represents a verification request
type VerificationRequest struct {
	Verifier   string `json:"verifier"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
}

// VerificationResponse represents a verification response
type VerificationResponse struct {
	ID         string                 `json:"id"`
	Result     string                 `json:"result"`
	Details    map[string]interface{} `json:"details"`
	VerifiedAt time.Time              `json:"verified_at"`
}

// QuickVerifyResponse represents a quick verification response
type QuickVerifyResponse struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// Verification performs detailed verification
func (c *Client) Verification(req *VerificationRequest) (*VerificationResponse, error) {
	if req.Verifier == "" || req.TargetID == "" || req.TargetType == "" {
		return nil, &ValidationError{Message: "verifier, target_id, and target_type are required"}
	}

	resp, err := c.doRequest("POST", "/verifications", req)
	if err != nil {
		return nil, err
	}

	var result VerificationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

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

// ==================== UTILITY METHODS ====================

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// Health checks API health
func (c *Client) Health() (*HealthResponse, error) {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var result HealthResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DocsResponse represents API documentation response
type DocsResponse struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Endpoints   map[string]interface{} `json:"endpoints"`
}

// Docs gets API documentation
func (c *Client) Docs() (*DocsResponse, error) {
	resp, err := c.doRequest("GET", "/api/docs", nil)
	if err != nil {
		return nil, err
	}

	var result DocsResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateKeyResponse represents API key generation response
type GenerateKeyResponse struct {
	Key     string `json:"key"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Created string `json:"created"`
}

// GenerateKey creates a new API key
func (c *Client) GenerateKey(email, name string) (*GenerateKeyResponse, error) {
	if email == "" {
		return nil, &ValidationError{Message: "email is required"}
	}

	req := map[string]string{
		"email": email,
		"name":  name,
	}

	resp, err := c.doRequest("POST", "/developer/keys", req)
	if err != nil {
		return nil, err
	}

	var result GenerateKeyResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ==================== ERROR TYPES ====================

// APIError represents an API error response
type APIError struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}
	if e.Error != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Error)
	}
	return fmt.Sprintf("API error: status %d", e.StatusCode)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}
