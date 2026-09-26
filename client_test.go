package jep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEvent07(t *testing.T) {
	var seenPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		var req CreateEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil { t.Fatal(err) }
		_ = json.NewEncoder(w).Encode(EventResponse{
			Event:JEPEvent{JEP:"1",ID:"urn:uuid:1",Verb:req.Verb,Who:req.Who,What:req.What,Sig:"h..s"},
			EventHash:"sha256:abc",
			Validation:ValidationResult{Status:"valid",Mode:"archival",Profile:JEPCoreProfile,Checks:map[string]string{"syntax":"pass"}},
		})
	}))
	defer server.Close()

	client:=NewClientWithURL(server.URL,"")
	resp,err:=client.CreateEvent(&CreateEventRequest{Verb:VerbJudgment,Who:"did:example:a",What:map[string]interface{}{"claim":"approve"}})
	if err!=nil{t.Fatal(err)}
	if seenPath!="/v0.7/events/create"{t.Fatalf("path=%s",seenPath)}
	if resp.Event.ID==""||!resp.Validation.Valid(){t.Fatalf("unexpected response: %+v",resp)}
}

func TestVerifyEvent07(t *testing.T) {
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.URL.Path!="/v0.7/events/verify"{t.Fatalf("path=%s",r.URL.Path)}
		_ = json.NewEncoder(w).Encode(ValidationResult{Status:"valid",Mode:"archival",Profile:JEPCoreProfile,Checks:map[string]string{"syntax":"pass"}})
	}))
	defer server.Close()
	client:=NewClientWithURL(server.URL,"")
	result,err:=client.VerifyEvent(&VerifyEventRequest{Event:JEPEvent{JEP:"1",ID:"urn:uuid:1",Verb:VerbJudgment,Who:"a",When:1,What:map[string]interface{}{"claim":"x"},Sig:"h..s"}})
	if err!=nil{t.Fatal(err)}
	if !result.Valid()||result.Profile!="jep-core-0.7"{t.Fatalf("unexpected %+v",result)}
}

func TestLegacyVerifyIsExplicit(t *testing.T) {
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if r.URL.Path!="/events/verify"{t.Fatalf("path=%s",r.URL.Path)}
		_ = json.NewEncoder(w).Encode(LegacyValidationResult{Valid:true,Level:1,Mode:"archival",Profile:"jep-core-0.6"})
	}))
	defer server.Close()
	client:=NewClientWithURL(server.URL,"")
	result,err:=client.VerifyLegacyEvent(&LegacyVerifyEventRequest{Event:LegacyJEPEvent{JEP:"1",Verb:VerbJudgment,Who:"a",When:1,What:"x",Nonce:"n",Sig:"h..s"}})
	if err!=nil{t.Fatal(err)}
	if !result.Valid||result.Profile!="jep-core-0.6"{t.Fatalf("unexpected %+v",result)}
}

func TestValidation(t *testing.T) {
	client:=NewClient("")
	if _,err:=client.CreateEvent(nil);err==nil{t.Fatal("expected nil create error")}
	if _,err:=client.CreateEvent(&CreateEventRequest{Verb:Verb("X"),What:"x"});err==nil{t.Fatal("expected verb error")}
	if _,err:=client.VerifyEvent(nil);err==nil{t.Fatal("expected nil verify error")}
}

func TestHealth(t *testing.T) {
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){_ = json.NewEncoder(w).Encode(HealthResponse{OK:true,Profile:JEPCoreProfile})}))
	defer server.Close()
	client:=NewClientWithURL(server.URL,"")
	h,err:=client.Health();if err!=nil{t.Fatal(err)}
	if !h.OK||h.Profile!="jep-core-0.7"{t.Fatalf("unexpected %+v",h)}
}

func TestConvenienceHelpers(t *testing.T) {
	var verbs []Verb
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		var req CreateEventRequest;_ = json.NewDecoder(r.Body).Decode(&req);verbs=append(verbs,req.Verb)
		_ = json.NewEncoder(w).Encode(EventResponse{Event:JEPEvent{JEP:"1",ID:"urn:uuid:x",Verb:req.Verb,Who:req.Who,What:req.What,Sig:"h..s"},Validation:ValidationResult{Status:"valid",Profile:JEPCoreProfile,Checks:map[string]string{}}})
	}))
	defer server.Close()
	client:=NewClientWithURL(server.URL,"")
	ref:=map[string]interface{}{"type":"jep:event","value":map[string]interface{}{"who":"a","id":"urn:uuid:x"}}
	_,_=client.Judgment("a",map[string]interface{}{"claim":"x"})
	_,_=client.Delegation("a",map[string]interface{}{"delegatee":"b","scope":map[string]interface{}{"task":"x"}})
	_,_=client.Termination("a",map[string]interface{}{"termination_scope":"future"},ref)
	_,_=client.Verification("a",map[string]interface{}{"verification_scope":[]string{"cryptographic"},"result":"pass"},ref)
	want:=[]Verb{VerbJudgment,VerbDelegation,VerbTermination,VerbVerification}
	for i,v:=range want{if verbs[i]!=v{t.Fatalf("%d got %s want %s",i,verbs[i],v)}}
}
