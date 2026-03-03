package main

import (
	"fmt"
	"log"
	"github.com/hjs-protocol/sdk-go"
)

func main() {
	// Create client
	client := hjs.NewClient("your-api-key")

	// 1. Record a judgment
	fmt.Println("📝 Recording judgment...")
	judgment, err := client.Judgment(&hjs.JudgmentRequest{
		Entity: "alice@bank.com",
		Action: "loan_approved",
		Scope: map[string]interface{}{
			"amount":   100000,
			"currency": "USD",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Judgment recorded: %s\n", judgment.ID)

	// 2. Create a delegation
	fmt.Println("\n📝 Creating delegation...")
	delegation, err := client.Delegation(&hjs.DelegationRequest{
		Delegator: "manager@company.com",
		Delegatee: "employee@company.com",
		JudgmentID: judgment.ID,
		Scope: map[string]interface{}{
			"permissions": []string{"approve_under_1000", "read"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Delegation created: %s\n", delegation.ID)

	// 3. Verify the delegation
	fmt.Println("\n🔍 Verifying delegation...")
	verification, err := client.Verify(delegation.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Verification result: %s\n", verification.Status)

	// 4. List judgments
	fmt.Println("\n📋 Listing judgments...")
	list, err := client.ListJudgments(&hjs.ListJudgmentsParams{
		Limit: 10,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d judgments\n", list.Total)

	// 5. Check health
	fmt.Println("\n🏥 Checking health...")
	health, err := client.Health()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("API health: %s\n", health.Status)
}
