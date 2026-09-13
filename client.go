// Package jep provides a Go SDK for the JEP v0.6 API seed.
//
// It is aligned with the JEP v0.6 event API shape:
//
//	POST /events/create
//	POST /events/verify
//
// This SDK is an implementation seed. It does not define new JEP-Core
// semantics and does not perform legal, factual, or compliance validation.
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
	DefaultBaseURL = "http://127.0.0.1:8000"
	DefaultTimeout = 30 * time.Second
	JEPWireVersion = "1"
	JEPCoreProfile = "jep-core-0.6"
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
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

func NewClientWithURL(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

func (c *Client) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

type JEPEvent struct {
	// Reference supports structured v0.6 references while Ref retains the
	// existing string-pointer API. Ref takes precedence when set.
	Reference       json.RawMessage `json:"-"`
	SignatureObject json.RawMessage `json:"-"`
	wire            map[string]json.RawMessage
	baseline        map[string]json.RawMessage
	JEP             string                 `json:"jep"`
	Verb            Verb                   `json:"verb"`
	Who             string                 `json:"who"`
	When            int64                  `json:"when"`
	What            interface{}            `json:"what,omitempty"`
	Nonce           string                 `json:"nonce"`
	Aud             string                 `json:"aud,omitempty"`
	Ref             *string                `json:"ref,omitempty"`
	Ext             map[string]interface{} `json:"ext,omitempty"`
	ExtCrit         []string               `json:"ext_crit,omitempty"`
	Sig             string                 `json:"sig,omitempty"`
}

// encodedFields avoids calling MarshalJSON recursively.
func (e JEPEvent) encodedFields() (map[string]json.RawMessage, error) {
	type plain JEPEvent
	raw, err := json.Marshal(plain(e))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	if e.Ref == nil && e.Reference != nil {
		fields["ref"] = e.Reference
	}
	if e.Sig == "" && e.SignatureObject != nil {
		fields["sig"] = e.SignatureObject
	}
	return fields, nil
}

func (e *JEPEvent) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if fields == nil {
		return fmt.Errorf("event must be a JSON object")
	}
	typed := make(map[string]json.RawMessage, len(fields))
	for k, v := range fields {
		typed[k] = v
	}
	type plain JEPEvent
	var decoded plain
	if v := bytes.TrimSpace(fields["ref"]); len(v) > 0 && v[0] == '{' {
		decoded.Reference = append(json.RawMessage(nil), v...)
		delete(typed, "ref")
	}
	if v := bytes.TrimSpace(fields["sig"]); len(v) > 0 && v[0] == '{' {
		decoded.SignatureObject = append(json.RawMessage(nil), v...)
		delete(typed, "sig")
	}
	normalized, err := json.Marshal(typed)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	event := JEPEvent(decoded)
	event.wire = fields
	event.baseline, err = event.encodedFields()
	if err != nil {
		return err
	}
	*e = event
	return nil
}

func (e JEPEvent) MarshalJSON() ([]byte, error) {
	fields, err := e.encodedFields()
	if err != nil {
		return nil, err
	}
	// Preserve unchanged wire values, including explicit nulls, empty members,
	// unknown fields and exact number literals. Edits are still serialized.
	for k, original := range e.wire {
		if bytes.Equal(fields[k], e.baseline[k]) {
			fields[k] = original
		}
	}
	return json.Marshal(fields)
}

type CreateEventRequest struct {
	Verb          Verb                   `json:"verb"`
	Who           string                 `json:"who,omitempty"`
	What          interface{}            `json:"what"`
	Aud           string                 `json:"aud,omitempty"`
	Ref           *string                `json:"ref,omitempty"`
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
	ConsumeNonce     bool     `json:"consume_nonce,omitempty"`
	ExpectedAudience string   `json:"expected_audience,omitempty"`
}

type ValidationResult struct {
	ConformanceClass string                   `json:"conformance_class,omitempty"`
	Valid            bool                     `json:"valid"`
	Level            int                      `json:"level"`
	Mode             string                   `json:"mode"`
	Profile          string                   `json:"profile"`
	Scopes           []string                 `json:"scopes,omitempty"`
	EventHash        string                   `json:"event_hash,omitempty"`
	Warnings         []map[string]interface{} `json:"warnings,omitempty"`
	Errors           []map[string]interface{} `json:"errors,omitempty"`
}

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
	if req.What == nil && req.Verb != VerbJudgment {
		return nil, &ValidationError{Message: "what is required"}
	}

	var result EventResponse
	if err := c.doJSON(http.MethodPost, "/events/create", req, &result); err != nil {
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
	if err := c.doJSON(http.MethodPost, "/events/verify", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Health() (*HealthResponse, error) {
	var result HealthResponse
	if err := c.doJSON(http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Convenience helpers for JEP primitives.

func (c *Client) Judgment(who string, what interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{
		Verb: VerbJudgment,
		Who:  who,
		What: what,
	})
}

func (c *Client) Delegation(who string, what interface{}) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{
		Verb: VerbDelegation,
		Who:  who,
		What: what,
	})
}

func (c *Client) Termination(who string, what interface{}, ref *string) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{
		Verb: VerbTermination,
		Who:  who,
		What: what,
		Ref:  ref,
	})
}

func (c *Client) Verification(who string, what interface{}, ref string) (*EventResponse, error) {
	return c.CreateEvent(&CreateEventRequest{
		Verb: VerbVerification,
		Who:  who,
		What: what,
		Ref:  &ref,
	})
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
	req.Header.Set("User-Agent", "JEP-Go-SDK/0.6.0")
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

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("JEP validation error: %s", e.Message)
}
