// Package jep provides a Go SDK for JEP Core 0.7.
//
// Current methods target POST /v0.7/events/create and /v0.7/events/verify.
// Explicit Legacy methods preserve pre-0.7 API compatibility. No current
// failure triggers automatic legacy fallback.
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
	JEPCoreProfile = "jep-core-0.7"
)

type Verb string

const (
	VerbJudgment     Verb = "J"
	VerbDelegation   Verb = "D"
	VerbTermination  Verb = "T"
	VerbVerification Verb = "V"
)

type Client struct {
	baseURL string
	apiKey string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{baseURL: DefaultBaseURL, apiKey: apiKey, httpClient: &http.Client{Timeout: DefaultTimeout}}
}
func NewClientWithURL(baseURL, apiKey string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, httpClient: &http.Client{Timeout: DefaultTimeout}}
}
func (c *Client) SetTimeout(timeout time.Duration) { c.httpClient.Timeout = timeout }
func (c *Client) SetHTTPClient(client *http.Client) { if client != nil { c.httpClient = client } }

// JEPEvent is the current JEP Core 0.7 event shape.
type JEPEvent struct {
	Reference json.RawMessage `json:"-"`
	SignatureObject json.RawMessage `json:"-"`
	wire map[string]json.RawMessage
	baseline map[string]json.RawMessage
	JEP string `json:"jep"`
	ID string `json:"id"`
	Verb Verb `json:"verb"`
	Who string `json:"who"`
	When int64 `json:"when"`
	What interface{} `json:"what"`
	Aud string `json:"aud,omitempty"`
	Ref *string `json:"ref,omitempty"`
	Ext map[string]interface{} `json:"ext,omitempty"`
	ExtCrit []string `json:"ext_crit,omitempty"`
	Sig string `json:"sig,omitempty"`
}

func (e JEPEvent) encodedFields() (map[string]json.RawMessage, error) {
	type plain JEPEvent
	raw, err := json.Marshal(plain(e))
	if err != nil { return nil, err }
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil { return nil, err }
	if e.Ref == nil && e.Reference != nil { fields["ref"] = e.Reference }
	if e.Sig == "" && e.SignatureObject != nil { fields["sig"] = e.SignatureObject }
	return fields, nil
}

func (e *JEPEvent) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil { return err }
	if fields == nil { return fmt.Errorf("event must be a JSON object") }
	typed := make(map[string]json.RawMessage, len(fields))
	for k,v := range fields { typed[k]=v }
	type plain JEPEvent
	var decoded plain
	if v := bytes.TrimSpace(fields["ref"]); len(v)>0 && v[0]=='{' {
		decoded.Reference=append(json.RawMessage(nil),v...); delete(typed,"ref")
	}
	if v := bytes.TrimSpace(fields["sig"]); len(v)>0 && v[0]=='{' {
		decoded.SignatureObject=append(json.RawMessage(nil),v...); delete(typed,"sig")
	}
	normalized,err:=json.Marshal(typed); if err!=nil{return err}
	decoder:=json.NewDecoder(bytes.NewReader(normalized)); decoder.UseNumber()
	if err:=decoder.Decode(&decoded);err!=nil{return err}
	event:=JEPEvent(decoded); event.wire=fields; event.baseline,err=event.encodedFields(); if err!=nil{return err}
	*e=event; return nil
}

func (e JEPEvent) MarshalJSON() ([]byte,error) {
	fields,err:=e.encodedFields(); if err!=nil{return nil,err}
	for k,original:=range e.wire {
		if bytes.Equal(fields[k],e.baseline[k]) { fields[k]=original }
	}
	return json.Marshal(fields)
}

// LegacyJEPEvent is the explicit pre-0.7 event shape.
type LegacyJEPEvent struct {
	JEP string `json:"jep"`
	Verb Verb `json:"verb"`
	Who string `json:"who"`
	When int64 `json:"when"`
	What interface{} `json:"what,omitempty"`
	Nonce string `json:"nonce"`
	Aud string `json:"aud,omitempty"`
	Ref interface{} `json:"ref,omitempty"`
	Ext map[string]interface{} `json:"ext,omitempty"`
	ExtCrit []string `json:"ext_crit,omitempty"`
	Sig string `json:"sig,omitempty"`
}

type CreateEventRequest struct {
	ID string `json:"id,omitempty"`
	Verb Verb `json:"verb"`
	Who string `json:"who,omitempty"`
	What interface{} `json:"what"`
	Aud string `json:"aud,omitempty"`
	Ref interface{} `json:"ref,omitempty"`
	TTLMinutes *int `json:"ttl_minutes,omitempty"`
	DigestOnlyWho bool `json:"digest_only_who,omitempty"`
	Ext map[string]interface{} `json:"ext,omitempty"`
	ExtCrit []string `json:"ext_crit,omitempty"`
}

type EventIdentity struct {
	Who string `json:"who"`
	ID string `json:"id"`
}
type AcceptanceResult struct {
	Outcome string `json:"outcome"`
	EffectApplied bool `json:"effect_applied"`
}
type ValidationResult struct {
	Status string `json:"status"`
	Mode string `json:"mode"`
	Profile string `json:"profile"`
	ConformanceClass string `json:"conformance_class,omitempty"`
	EventIdentity *EventIdentity `json:"event_identity,omitempty"`
	EventHash string `json:"event_hash,omitempty"`
	Checks map[string]string `json:"checks"`
	Acceptance *AcceptanceResult `json:"acceptance,omitempty"`
	Warnings []map[string]interface{} `json:"warnings,omitempty"`
	Errors []map[string]interface{} `json:"errors,omitempty"`
}
func (r ValidationResult) Valid() bool { return r.Status=="valid" }

type LegacyValidationResult struct {
	ConformanceClass string `json:"conformance_class,omitempty"`
	Valid bool `json:"valid"`
	Level int `json:"level"`
	Mode string `json:"mode"`
	Profile string `json:"profile"`
	Scopes []string `json:"scopes,omitempty"`
	EventHash string `json:"event_hash,omitempty"`
	Warnings []map[string]interface{} `json:"warnings,omitempty"`
	Errors []map[string]interface{} `json:"errors,omitempty"`
}

type EventResponse struct {
	Event JEPEvent `json:"event"`
	EventHash string `json:"event_hash"`
	Validation ValidationResult `json:"validation"`
}
type LegacyEventResponse struct {
	Event LegacyJEPEvent `json:"event"`
	EventHash string `json:"event_hash"`
	Validation LegacyValidationResult `json:"validation"`
}

type VerifyEventRequest struct {
	Event JEPEvent `json:"event"`
	Mode string `json:"mode,omitempty"`
	ExpectedAudience string `json:"expected_audience,omitempty"`
	MaxAgeSeconds *int `json:"max_age_seconds,omitempty"`
}
type LegacyVerifyEventRequest struct {
	Event LegacyJEPEvent `json:"event"`
	Mode string `json:"mode,omitempty"`
	ConsumeNonce bool `json:"consume_nonce,omitempty"`
	ExpectedAudience string `json:"expected_audience,omitempty"`
}
type HealthResponse struct { OK bool `json:"ok"`; Profile string `json:"profile"` }

func (c *Client) CreateEvent(req *CreateEventRequest) (*EventResponse,error) {
	if req==nil{return nil,&ValidationError{Message:"request is required"}}
	if err:=validateVerb(req.Verb);err!=nil{return nil,err}
	if req.What==nil{return nil,&ValidationError{Message:"what is required"}}
	var result EventResponse
	if err:=c.doJSON(http.MethodPost,"/v0.7/events/create",req,&result);err!=nil{return nil,err}
	return &result,nil
}
func (c *Client) VerifyEvent(req *VerifyEventRequest) (*ValidationResult,error) {
	if req==nil{return nil,&ValidationError{Message:"request is required"}}
	if req.Event.JEP==""||req.Event.ID==""{return nil,&ValidationError{Message:"event is required"}}
	var result ValidationResult
	if err:=c.doJSON(http.MethodPost,"/v0.7/events/verify",req,&result);err!=nil{return nil,err}
	return &result,nil
}
func (c *Client) CreateLegacyEvent(req *CreateEventRequest) (*LegacyEventResponse,error) {
	if req==nil{return nil,&ValidationError{Message:"request is required"}}
	if err:=validateVerb(req.Verb);err!=nil{return nil,err}
	if req.What==nil{return nil,&ValidationError{Message:"what is required"}}
	var result LegacyEventResponse
	if err:=c.doJSON(http.MethodPost,"/events/create",req,&result);err!=nil{return nil,err}
	return &result,nil
}
func (c *Client) VerifyLegacyEvent(req *LegacyVerifyEventRequest) (*LegacyValidationResult,error) {
	if req==nil||req.Event.JEP==""{return nil,&ValidationError{Message:"event is required"}}
	var result LegacyValidationResult
	if err:=c.doJSON(http.MethodPost,"/events/verify",req,&result);err!=nil{return nil,err}
	return &result,nil
}
func (c *Client) Health() (*HealthResponse,error) {
	var result HealthResponse
	if err:=c.doJSON(http.MethodGet,"/health",nil,&result);err!=nil{return nil,err}
	return &result,nil
}

func (c *Client) Judgment(who string,what interface{})(*EventResponse,error){
	return c.CreateEvent(&CreateEventRequest{Verb:VerbJudgment,Who:who,What:what})
}
func (c *Client) Delegation(who string,what interface{})(*EventResponse,error){
	return c.CreateEvent(&CreateEventRequest{Verb:VerbDelegation,Who:who,What:what})
}
func (c *Client) Termination(who string,what interface{},ref interface{})(*EventResponse,error){
	return c.CreateEvent(&CreateEventRequest{Verb:VerbTermination,Who:who,What:what,Ref:ref})
}
func (c *Client) Verification(who string,what interface{},ref interface{})(*EventResponse,error){
	return c.CreateEvent(&CreateEventRequest{Verb:VerbVerification,Who:who,What:what,Ref:ref})
}

func (c *Client) doJSON(method,path string,body interface{},out interface{}) error {
	url:=strings.TrimRight(c.baseURL,"/")+path
	var reqBody io.Reader
	if body!=nil { data,err:=json.Marshal(body);if err!=nil{return &ValidationError{Message:fmt.Sprintf("failed to marshal request: %v",err)}};reqBody=bytes.NewReader(data) }
	req,err:=http.NewRequest(method,url,reqBody);if err!=nil{return err}
	req.Header.Set("Content-Type","application/json");req.Header.Set("User-Agent","JEP-Go-SDK/0.7.0")
	if c.apiKey!=""{req.Header.Set("Authorization","Bearer "+c.apiKey);req.Header.Set("X-API-Key",c.apiKey)}
	resp,err:=c.httpClient.Do(req);if err!=nil{return err};defer resp.Body.Close()
	payload,err:=io.ReadAll(resp.Body);if err!=nil{return err}
	if resp.StatusCode<200||resp.StatusCode>=300 {
		apiErr:=&APIError{StatusCode:resp.StatusCode}
		if err:=json.Unmarshal(payload,apiErr);err!=nil||(apiErr.Code==""&&apiErr.Message==""){apiErr.Message=string(payload)}
		return apiErr
	}
	if out==nil||len(payload)==0{return nil}
	if err:=json.Unmarshal(payload,out);err!=nil{return &ValidationError{Message:fmt.Sprintf("failed to decode response: %v",err)}}
	return nil
}
func validateVerb(verb Verb) error {
	switch verb{case VerbJudgment,VerbDelegation,VerbTermination,VerbVerification:return nil;default:return &ValidationError{Message:"verb must be J, D, T, or V"}}
}
type APIError struct { Code string `json:"error,omitempty"`; Message string `json:"message,omitempty"`; StatusCode int `json:"-"` }
func (e *APIError) Error() string {msg:=e.Message;if msg==""{msg=e.Code};return fmt.Sprintf("JEP API error (%d): %s",e.StatusCode,msg)}
type ValidationError struct{Message string}
func (e *ValidationError) Error()string{return fmt.Sprintf("JEP validation error: %s",e.Message)}
