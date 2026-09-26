package jep

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestImportedEventPreservesSignedMembers(t *testing.T) {
	source := `{"jep":"1","id":"urn:example:signed","verb":"J","who":"did:example:alice","when":123,"what":{"amount":9007199254740993,"small":1e-7},"ext":{},"ext_crit":[],"sig":"protected..signature","unexpected":{"kept":true}}`
	var event JEPEvent
	if err := json.Unmarshal([]byte(source), &event); err != nil {
		t.Fatal(err)
	}
	output, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`"ext":{}`, `"ext_crit":[]`, `9007199254740993`, `"unexpected":{"kept":true}`, `"sig":"protected..signature"`} {
		if !strings.Contains(string(output), fragment) {
			t.Fatalf("signed member lost or changed: %s in %s", fragment, output)
		}
	}
	// Editing the typed model must reach the verifier, not secretly replay old raw bytes.
	event.What = map[string]interface{}{"claim": "changed"}
	output, err = json.Marshal(event)
	if err != nil || !strings.Contains(string(output), `"claim":"changed"`) {
		t.Fatalf("edit hidden: %s %v", output, err)
	}
}

func TestImportedEventDoesNotAddAbsentOptionalMembers(t *testing.T) {
	var event JEPEvent
	if err := json.Unmarshal([]byte(`{"jep":"1","id":"i","verb":"J","who":"a","when":1,"what":{"claim":"j"},"sig":"h..s"}`), &event); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{`"ext":`, `"ext_crit":`, `"aud":`, `"ref":`} {
		if strings.Contains(string(raw), name) {
			t.Fatalf("absent member added: %s", raw)
		}
	}
}

func TestProfileDefinedNullValuesRemainPresent(t *testing.T) {
	for _, req := range []*CreateEventRequest{
		{Verb: VerbDelegation, What: map[string]interface{}{"delegatee": "b", "scope": nil}},
		{Verb: VerbVerification, Ref: "sha256:" + strings.Repeat("a", 64), What: map[string]interface{}{"verification_scope": []string{"custom"}, "result": nil}},
	} {
		if err := validateVerbShape(req); err != nil {
			t.Fatalf("Core requires member presence; profile owns its value: %v", err)
		}
	}
}
