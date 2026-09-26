package jep

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSignedRoundTripPreserves07WireMembers(t *testing.T) {
	for _,ref:=range []string{`null`,`{"type":"jep:event","value":{"who":"a","id":"urn:uuid:b"}}`} {
		raw:=[]byte(`{"jep":"1","id":"urn:uuid:a","verb":"J","who":"agent","when":123,"what":{"large":9007199254740993},"ref":`+ref+`,"ext":{},"ext_crit":[],"sig":"h..s","future":null}`)
		var event JEPEvent
		if err:=json.Unmarshal(raw,&event);err!=nil{t.Fatal(err)}
		encoded,err:=json.Marshal(VerifyEventRequest{Event:event});if err!=nil{t.Fatal(err)}
		var request map[string]json.RawMessage
		if err=json.Unmarshal(encoded,&request);err!=nil{t.Fatal(err)}
		decode:=func(data []byte)interface{}{d:=json.NewDecoder(bytes.NewReader(data));d.UseNumber();var out interface{};if err:=d.Decode(&out);err!=nil{t.Fatal(err)};return out}
		if !reflect.DeepEqual(decode(raw),decode(request["event"])){t.Fatalf("signed payload changed: %s",encoded)}
		event.Who="changed";encoded,err=json.Marshal(event)
		if err!=nil||!bytes.Contains(encoded,[]byte(`"who":"changed"`)){t.Fatalf("edit lost: %s %v",encoded,err)}
	}
}

func TestCurrentWireDoesNotAddNonce(t *testing.T) {
	event:=JEPEvent{JEP:"1",ID:"urn:uuid:a",Verb:VerbJudgment,Who:"a",When:1,What:map[string]interface{}{"claim":"x"},Sig:"h..s"}
	raw,err:=json.Marshal(event);if err!=nil{t.Fatal(err)}
	if bytes.Contains(raw,[]byte(`"nonce"`)){t.Fatalf("current event gained legacy nonce: %s",raw)}
}
