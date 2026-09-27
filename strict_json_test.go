package jep

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRejectHostileEventJSON(t *testing.T) {
	cases := map[string][]byte{
		"duplicate-top":   []byte(`{"who":"a","who":"b"}`),
		"duplicate-nested": []byte(`{"what":{"x":1,"\u0078":2}}`),
		"duplicate-array": []byte(`{"what":[{"x":1,"x":2}]}`),
		"surrogate-value": []byte(`{"what":"\ud800"}`),
		"surrogate-key": []byte(`{"what":{"\udfff":1}}`),
		"wrong-pair": []byte(`{"what":"\ud800\u0041"}`),
		"reversed-pair": []byte(`{"what":"\udc00\ud800"}`),
		"invalid-utf8": append(append([]byte(`{"what":"`), byte(0xff)), []byte(`"}`)...),
		"trailing-json": []byte(`{"what":1} {}`),
		"nonfinite": []byte(`{"what":NaN}`),
		"null": []byte(`null`),
		"array": []byte(`[]`),
		"case-alias": []byte(`{"id":"a","ID":"b"}`),
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			event := JEPEvent{ID: "unchanged"}
			if err := json.Unmarshal(raw, &event); err == nil {
				t.Fatal("invalid original JSON was silently imported")
			}
			if event.ID != "unchanged" {
				t.Fatal("failed import changed receiver")
			}
		})
	}
}

func TestStrictInputPreservesSignedMembers(t *testing.T) {
	raw := []byte(`{"jep":"1","id":"urn:test:a","who":"did:example:a","verb":"J","when":1,"what":{"large":1000000000000000100,"exact":1000000000000000128,"exp":1e20,"emoji":"\ud83d\ude00","literal":"\\ud800","__proto__":{"x":1}},"ext":{},"ext_crit":[],"sig":"protected..sig","future":false}`)
	var event JEPEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var before, after map[string]json.RawMessage
	_ = json.Unmarshal(raw, &before)
	_ = json.Unmarshal(out, &after)
	// Check exact numeric tokens separately; other strings may use equivalent
	// escapes. Signed-artifact JCS equality is exercised by the Core interop gate.
	what := event.What.(map[string]interface{})
	if what["large"].(json.Number).String() != "1000000000000000100" {
		t.Fatal("number token changed")
	}
	for _, key := range []string{"jep", "id", "who", "verb", "when", "ext", "ext_crit", "sig", "future"} {
		if !reflect.DeepEqual(before[key], after[key]) {
			t.Fatalf("member %s changed", key)
		}
	}
	if err := validateEventJSON(out); err != nil {
		t.Fatal(err)
	}
}

func FuzzStrictEventImport(f *testing.F) {
	for _, s := range []string{`{}`, `{"what":"\ud800"}`, `{"what":"\\ud800"}`, `{"what":"\ud83d\ude00"}`, `{"x":1,"x":2}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		var event JEPEvent
		_ = json.Unmarshal(raw, &event)
	})
}
