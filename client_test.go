package jep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEventUsesV07(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		var req CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { t.Fatal(err) }
		resp := EventResponse{
			Event: JEPEvent{
				JEP: JEPWireVersion, ID: "urn:uuid:00000000-0000-7000-8000-000000000001",
				Verb: req.Verb, Who: req.Who, When: 123, What: req.What, Sig: "header..sig",
			},
			EventHash: "sha256:abc",
			Validation: ValidationResult{
				Status: "valid", Mode: "archival", Profile: JEPCoreProfile,
				EventIdentity: map[string]string{"who": req.Who, "id": "urn:uuid:00000000-0000-7000-8000-000000000001"},
				Checks: map[string]string{"syntax":"pass","cryptographic":"pass","event_identity":"pass"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "")
	resp, err := client.CreateEvent(&CreateEventRequest{
		Verb: VerbJudgment, Who: "did:example:agent", What: map[string]interface{}{"claim":"approve"},
	})
	if err != nil { t.Fatal(err) }
	if seenPath != "/v0.7/events/create" { t.Fatalf("path=%s", seenPath) }
	if !resp.Validation.Valid() || resp.Event.ID == "" { t.Fatalf("unexpected response: %+v", resp) }
}

func TestVerifyEventUsesV07Result(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0.7/events/verify" { t.Fatalf("path=%s", r.URL.Path) }
		_ = json.NewEncoder(w).Encode(ValidationResult{
			Status:"valid", Mode:"acceptance", Profile:JEPCoreProfile,
			Checks:map[string]string{"syntax":"pass","cryptographic":"pass","event_identity":"pass"},
			Acceptance:map[string]interface{}{"outcome":"accepted","effect_applied":true},
		})
	}))
	defer server.Close()

	client := NewClientWithURL(server.URL, "")
	result, err := client.VerifyEvent(&VerifyEventRequest{
		Event:JEPEvent{JEP:"1", ID:"urn:example:1", Verb:VerbJudgment, Who:"did:example:a", When:1, What:map[string]interface{}{"claim":"x"}, Sig:"h..s"},
		Mode:"acceptance",
	})
	if err != nil { t.Fatal(err) }
	if !result.Valid() || result.Profile != JEPCoreProfile { t.Fatalf("%+v", result) }
	if result.Acceptance["outcome"] != "accepted" { t.Fatalf("%+v", result.Acceptance) }
}

func TestVerbMinimums(t *testing.T) {
	client := NewClient("")
	if _, err := client.CreateEvent(&CreateEventRequest{Verb:VerbDelegation, What:map[string]interface{}{"delegatee":"b"}}); err == nil {
		t.Fatal("expected D scope error")
	}
	if _, err := client.CreateEvent(&CreateEventRequest{Verb:VerbTermination, What:map[string]interface{}{"termination_scope":"future"}}); err == nil {
		t.Fatal("expected T ref error")
	}
	if _, err := client.CreateEvent(&CreateEventRequest{Verb:VerbVerification, Ref:"x", What:map[string]interface{}{"verification_scope":"syntax"}}); err == nil {
		t.Fatal("expected V result error")
	}
}

func TestTypedReference(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		_ = json.NewEncoder(w).Encode(EventResponse{
			Event:JEPEvent{JEP:"1", ID:"urn:example:1", Verb:VerbTermination, Who:"a", When:1, What:captured["what"], Ref:captured["ref"], Sig:"h..s"},
			Validation:ValidationResult{Status:"valid",Profile:JEPCoreProfile},
		})
	}))
	defer server.Close()
	client := NewClientWithURL(server.URL, "")
	ref := map[string]interface{}{"type":"jep:event","value":map[string]interface{}{"who":"a","id":"urn:example:0"}}
	_, err := client.Termination("a", map[string]interface{}{"termination_scope":"future"}, ref)
	if err != nil { t.Fatal(err) }
	if captured["ref"] == nil { t.Fatal("typed ref lost") }
}

func TestHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(HealthResponse{OK:true,Profile:JEPCoreProfile})
	}))
	defer server.Close()
	client := NewClientWithURL(server.URL, "")
	h, err := client.Health()
	if err != nil || h.Profile != "jep-core-0.7" { t.Fatalf("%+v %v", h, err) }
}
