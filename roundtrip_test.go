package jep

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestV07EventRoundTrip(t *testing.T) {
	refs := []interface{}{
		nil,
		map[string]interface{}{
			"type": "jep:event",
			"value": map[string]interface{}{
				"who": "did:example:b",
				"id":  "urn:uuid:00000000-0000-7000-8000-000000000002",
			},
			"hash": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	for _, ref := range refs {
		event := JEPEvent{
			JEP:  "1",
			ID:   "urn:uuid:00000000-0000-7000-8000-000000000001",
			Verb: VerbJudgment,
			Who:  "did:example:a",
			When: 123,
			What: map[string]interface{}{"claim": "approve"},
			Ref:  ref,
			Sig:  "h..s",
		}
		raw, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var decoded JEPEvent
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		roundtrip, err := json.Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		var originalValue, returnedValue interface{}
		// Compare transported JSON, not the model's private presence bookkeeping.
		_ = json.Unmarshal(raw, &originalValue)
		_ = json.Unmarshal(roundtrip, &returnedValue)
		left, _ := json.Marshal(originalValue)
		right, _ := json.Marshal(returnedValue)
		if !bytes.Equal(left, right) {
			t.Fatalf("roundtrip changed event: %s != %s", left, right)
		}
	}
}

func TestValidationResultValid(t *testing.T) {
	if !(ValidationResult{Status: "valid"}).Valid() {
		t.Fatal("valid status should return true")
	}
	if (ValidationResult{Status: "invalid"}).Valid() {
		t.Fatal("invalid status should return false")
	}
}
