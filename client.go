// Package jep provides a Go SDK for the JEP Core 0.7 reference API.
//
// Default endpoints:
//
//	POST /v0.7/events/create
//	POST /v0.7/events/verify
//
// Historical pre-0.7 compatibility is explicit and never selected by fallback.
package jep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseURL       = "http://127.0.0.1:8000"
	DefaultTimeout       = 30 * time.Second
	JEPWireVersion       = "1"
	JEPCoreProfile       = "jep-core-0.7"
	LegacyJEPCoreProfile = "jep-core-0.6"
)

type Verb string

const (
	VerbJudgment     Verb = "J"
	VerbDelegation   Verb = "D"
	VerbTermination  Verb = "T"
	VerbVerification Verb = "V"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{baseURL: DefaultBaseURL, apiKey: apiKey, httpClient: &http.Client{Timeout: DefaultTimeout}}
}

func NewClientWithURL(baseURL, apiKey string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, httpClient: &http.Client{Timeout: DefaultTimeout}}
}

func (c *Client) SetTimeout(timeout time.Duration) { c.httpClient.Timeout = timeout }
func (c *Client) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

type JEPEvent struct {
	JEP     string                 `json:"jep"`
	ID      string                 `json:"id"`
	Verb    Verb                   `json:"verb"`
	Who     string                 `json:"who"`
	When    int64                  `json:"when"`
	What    interface{}            `json:"what"`
	Aud     string                 `json:"aud,omitempty"`
	Ref     interface{}            `json:"ref,omitempty"`
	Ext     map[string]interface{} `json:"ext,omitempty"`
	ExtCrit []string               `json:"ext_crit,omitempty"`
	Sig     interface{}            `json:"sig,omitempty"`
	wire    map[string]json.RawMessage
}

// UnmarshalJSON preserves number tokens and member presence in imported artifacts.
// Verification is delegated to the API; unknown members must reach it unchanged
// so an invalid event cannot become valid through client-side field removal.
func (e *JEPEvent) UnmarshalJSON(data []byte) error {
	type eventFields JEPEvent
	var decoded eventFields
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	*e = JEPEvent(decoded)
	e.wire = wire
	return nil
}

// MarshalJSON retains explicitly present optional fields, while reflecting
// edits to exported fields. It does not canonicalize or verify signatures.
func (e JEPEvent) MarshalJSON() ([]byte, error) {
	type eventFields JEPEvent
	raw, err := json.Marshal(eventFields(e))
	if err != nil {
		return nil, err
	}
	if e.wire == nil {
		return raw, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	optional := map[string]interface{}{
		"aud": e.Aud, "ref": e.Ref, "ext": e.Ext,
		"ext_crit": e.ExtCrit, "sig": e.Sig,
	}
	for name, original := range e.wire {
		if _, present := fields[name]; present {
			continue
		}
		if value, known := optional[name]; known {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			fields[name] = encoded
		} else {
			fields[name] = original
		}
	}
	return json.Marshal(fields)
}

type CreateEventRequest struct {
	ID            string                 `json:"id,omitempty"`
	Verb          Verb                   `json:"verb"`
	Who           string                 `json:"who,omitempty"`
	What          interface{}            `json:"what"`
	Aud           string                 `json:"aud,omitempty"`
	Ref           interface{}            `json:"ref,omitempty"`
	TTLMinutes    *int                   `json:"ttl_minutes,omitempty"`
	DigestOnlyWho bool                   `json:"digest_only_who,omitempty"`
	Ext           map[string]interface{} `json:"ext,omitempty"`
	ExtCrit       []string               `json:"ext_crit,omitempty"`
}

type EventResponse struct {
	Event      JEPEvent         `json:"event"`
	EventHash  string           `json:"event_hash"`
	Validation ValidationResult `json:"validation"`
}

type VerifyEventRequest struct {
	Event            JEPEvent `json:"event"`
	Mode             string   `json:"mode,omitempty"`
	ExpectedAudience string   `json:"expected_audience,omitempty"`
	MaxAgeSeconds    *int     `json:"max_age_seconds,omitempty"`
}

type ValidationResult struct {
	ConformanceClass string                   `json:"conformance_class,omitempty"`
	Status           string                   `json:"status"`
	Mode             string                   `json:"mode"`
	Profile          string                   `json:"profile"`
	EventIdentity    map[string]string        `json:"event_identity,omitempty"`
	Checks           map[string]string        `json:"checks,omitempty"`
	Acceptance       map[string]interface{}   `json:"acceptance,omitempty"`
	EventHash        string                   `json:"event_hash,omitempty"`
	Warnings         []map[string]interface{} `json:"warnings,omitempty"`
	Errors           []map[string]interface{} `json:"errors,omitempty"`
}

func (v ValidationResult) Valid() bool { return v.Status == "valid" }

type HealthResponse struct {
	OK      bool   `json:"ok"`
	Profile string `json:"profile"`
}

func (c *Client) CreateEvent(req *CreateEventRequest) (*EventResponse, error) {
	if req == nil {
		return nil, &ValidationError{Message: "request is required"}
	}
	if err := validateVerb(req.Verb); err != nil {
		return nil, err
	}
	if req.What == nil {
		return nil, &ValidationError{Message: "what is required"}
	}
	if err := validateVerbShape(req); err != nil {
		return nil, err
	}
	var result EventResponse
	if err := c.doJSON(http.MethodPost, "/v0.7/events/create", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) VerifyEvent(req *VerifyEventRequest) (*ValidationResult, error) {
	if req == nil {
		return nil, &ValidationError{Message: "request is required"}
	}
	if req.Event.JEP == "" {
		return nil, &ValidationError{Message: "event is required"}
	}
	var result ValidationResult
	if err := c.doJSON(http.MethodPost, "/v0.7/events/verify", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) VerifyEventLegacy(payload interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.doJSON(http.MethodPost, "/events/verify-legacy", payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Health() (*HealthResponse, error) {
	var result HealthResponse
	if err := c.doJSON(http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Judgment(who string, what interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{Verb: VerbJudgment, Who: who, What: what})
}

func (c *Client) Delegation(who string, what interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{Verb: VerbDelegation, Who: who, What: what})
}

func (c *Client) Termination(who string, what interface{}, ref interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{Verb: VerbTermination, Who: who, What: what, Ref: ref})
}

func (c *Client) Verification(who string, what interface{}, ref interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{Verb: VerbVerification, Who: who, What: what, Ref: ref})
}

func (c *Client) doJSON(method, path string, body interface{}, out interface{}) error {
	url := strings.TrimRight(c.baseURL, "/") + path
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return &ValidationError{Message: fmt.Sprintf("failed to marshal request: %v", err)}
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "JEP-Go-SDK/0.7.1")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("X-API-Key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(payload, apiErr); err != nil || (apiErr.Code == "" && apiErr.Message == "") {
			apiErr.Message = string(payload)
		}
		return apiErr
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &ValidationError{Message: fmt.Sprintf("failed to decode response: %v", err)}
	}
	return nil
}

func validateVerb(verb Verb) error {
	switch verb {
	case VerbJudgment, VerbDelegation, VerbTermination, VerbVerification:
		return nil
	default:
		return &ValidationError{Message: "verb must be J, D, T, or V"}
	}
}

func validateVerbShape(req *CreateEventRequest) error {
	obj, _ := req.What.(map[string]interface{})
	switch req.Verb {
	case VerbDelegation:
		_, hasScope := obj["scope"]
		if obj == nil || obj["delegatee"] == nil || !hasScope {
			return &ValidationError{Message: "D requires what.delegatee and what.scope"}
		}
	case VerbTermination:
		if req.Ref == nil {
			return &ValidationError{Message: "T requires ref"}
		}
		if obj == nil || obj["termination_scope"] == nil {
			return &ValidationError{Message: "T requires what.termination_scope"}
		}
	case VerbVerification:
		if req.Ref == nil {
			return &ValidationError{Message: "V requires ref"}
		}
		_, hasResult := obj["result"]
		if obj == nil || obj["verification_scope"] == nil || !hasResult {
			return &ValidationError{Message: "V requires what.verification_scope and what.result"}
		}
	}
	return nil
}

type APIError struct {
	Code       string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.Code
	}
	return fmt.Sprintf("JEP API error (%d): %s", e.StatusCode, msg)
}

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return fmt.Sprintf("JEP validation error: %s", e.Message) }
