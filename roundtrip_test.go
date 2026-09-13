package jep

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSignedRoundTripPreservesWireMembers(t *testing.T) {
	for _, ref := range []string{`null`, `{"type":"event","value":"sha256:aa"}`} {
		raw := []byte(`{"jep":"1","verb":"J","who":"agent","when":123,"nonce":"n","what":{"large":9007199254740993},"ref":` + ref + `,"ext":{},"ext_crit":[],"sig":"h..s","future":null}`)
		var event JEPEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(VerifyEventRequest{Event: event})
		if err != nil {
			t.Fatal(err)
		}
		var request map[string]json.RawMessage
		if err = json.Unmarshal(encoded, &request); err != nil {
			t.Fatal(err)
		}
		decode := func(data []byte) interface{} {
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.UseNumber()
			var out interface{}
			if err := decoder.Decode(&out); err != nil {
				t.Fatal(err)
			}
			return out
		}
		if !reflect.DeepEqual(decode(raw), decode(request["event"])) {
			t.Fatalf("signed payload changed: %s", encoded)
		}
		event.Who = "changed"
		encoded, err = json.Marshal(event)
		if err != nil || !bytes.Contains(encoded, []byte(`"who":"changed"`)) {
			t.Fatalf("edit was lost: %s %v", encoded, err)
		}
	}
}
