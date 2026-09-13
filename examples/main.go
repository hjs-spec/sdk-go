package main

import (
	"fmt"
	"log"

	jep "github.com/hjs-spec/sdk-go"
)

func main() {
	client := jep.NewClientWithURL("http://127.0.0.1:8000", "")

	resp, err := client.CreateEvent(&jep.CreateEventRequest{
		Verb: jep.VerbJudgment,
		Who:  "did:example:agent-789",
		What: map[string]interface{}{
			"claim":   "approve",
			"subject": "demo",
		},
		Aud: "https://api.example.org",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Event hash: %s\n", resp.EventHash)
	fmt.Printf("Valid: %v\n", resp.Validation.Valid)

	result, err := client.VerifyEvent(&jep.VerifyEventRequest{
		Event: resp.Event,
		Mode:  "archival",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Verification profile: %s\n", result.Profile)
}
