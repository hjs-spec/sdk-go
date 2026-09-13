package jep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEvent(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var req CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Verb != VerbJudgment {
			t.Fatalf("expected J, got %s", req.Verb)
		}

		resp := EventResponse{
			Event: JEPEvent{
				JEP:   JEPWireVersion,
				Verb:  req.Verb,
				Who:   req.Who,
				What:  req.What,
				Nonce: "nonce-1",
				Sig:   "header..sig",
			},
			EventHash: "sha256:abc",
			Validation: ValidationResult{
				Valid:     true,
				Level:     1,
				Mode:      "archival",
				Profile:   JEPCoreProfile,
				EventHash: "sha256:abc",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "test-key")
	resp, err := client.CreateEvent(&CreateEventRequest{
		Verb: VerbJudgment,
		Who:  "did:example:agent",
		What: map[string]interface{}{"claim": "approve"},
	})
	if err != nil {
		t.Fatalf("CreateEvent error: %v", err)
	}
	if seenPath != "/events/create" {
		t.Fatalf("expected /events/create, got %s", seenPath)
	}
	if resp.EventHash != "sha256:abc" {
		t.Fatalf("unexpected event hash: %s", resp.EventHash)
	}
	if !resp.Validation.Valid {
		t.Fatalf("expected valid response")
	}
}

func TestVerifyEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/verify" {
			t.Fatalf("expected /events/verify, got %s", r.URL.Path)
		}
		resp := ValidationResult{
			Valid:     true,
			Level:     1,
			Mode:      "archival",
			Profile:   JEPCoreProfile,
			EventHash: "sha256:def",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "")
	result, err := client.VerifyEvent(&VerifyEventRequest{
		Event: JEPEvent{
			JEP:   JEPWireVersion,
			Verb:  VerbJudgment,
			Who:   "did:example:agent",
			When:  123,
			What:  "sha256:abc",
			Nonce: "nonce-1",
			Sig:   "header..sig",
		},
		Mode: "archival",
	})
	if err != nil {
		t.Fatalf("VerifyEvent error: %v", err)
	}
	if !result.Valid || result.Profile != JEPCoreProfile {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestValidation(t *testing.T) {
	client := NewClient("")

	if _, err := client.CreateEvent(nil); err == nil {
		t.Fatal("expected error for nil request")
	}
	if _, err := client.CreateEvent(&CreateEventRequest{Verb: Verb("X"), What: "x"}); err == nil {
		t.Fatal("expected error for invalid verb")
	}
	if _, err := client.CreateEvent(&CreateEventRequest{Verb: VerbDelegation}); err == nil {
		t.Fatal("expected error for missing what")
	}
	if _, err := client.VerifyEvent(nil); err == nil {
		t.Fatal("expected error for nil verify request")
	}
}

func TestHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("expected /health, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(HealthResponse{OK: true, Profile: JEPCoreProfile})
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "")
	health, err := client.Health()
	if err != nil {
		t.Fatalf("Health error: %v", err)
	}
	if !health.OK || health.Profile != JEPCoreProfile {
		t.Fatalf("unexpected health: %+v", health)
	}
}

func TestConvenienceHelpers(t *testing.T) {
	var verbs []Verb
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateEventRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		verbs = append(verbs, req.Verb)
		_ = json.NewEncoder(w).Encode(EventResponse{
			Event:      JEPEvent{JEP: JEPWireVersion, Verb: req.Verb, Who: req.Who, What: req.What, Nonce: "n", Sig: "h..s"},
			EventHash:  "sha256:x",
			Validation: ValidationResult{Valid: true, Profile: JEPCoreProfile},
		})
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "")
	ref := "sha256:parent"
	_, _ = client.Judgment("agent", "judge")
	_, _ = client.Delegation("agent", "delegate")
	_, _ = client.Termination("agent", "terminate", &ref)
	_, _ = client.Verification("agent", "verify", ref)

	expected := []Verb{VerbJudgment, VerbDelegation, VerbTermination, VerbVerification}
	for i, want := range expected {
		if verbs[i] != want {
			t.Fatalf("verb[%d] = %s, want %s", i, verbs[i], want)
		}
	}
}

func TestValidationResultConformanceAndLegacyResponse(t *testing.T) {
	for _, payload := range []string{
		`{"valid":true,"level":1,"mode":"archival","profile":"jep-core-0.6","conformance_class":"JEP-Baseline-Ed25519-JWS-JCS-0.6","warnings":[{"code":"ACCEPTANCE_NOT_CHECKED","message":"archival","level":1,"recoverable":false}]}`,
		`{"valid":true,"level":1,"mode":"archival","profile":"jep-core-0.6"}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(payload))
		}))
		client := NewClientWithURL(server.URL, "")
		result, err := client.VerifyEvent(&VerifyEventRequest{Event: JEPEvent{JEP: "1", Verb: VerbJudgment}})
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &raw); err != nil {
			t.Fatal(err)
		}
		expected, _ := raw["conformance_class"].(string)
		if !result.Valid || result.ConformanceClass != expected {
			t.Fatalf("lost result metadata: %+v", result)
		}
		if expected != "" && (len(result.Warnings) != 1 || result.Warnings[0]["recoverable"] != false) {
			t.Fatalf("lost diagnostics: %+v", result)
		}
	}
}
