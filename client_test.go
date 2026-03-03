package hjs

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key")
	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got %s", client.apiKey)
	}
	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected baseURL %s, got %s", DefaultBaseURL, client.baseURL)
	}
}

func TestJudgmentValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing entity
	_, err := client.Judgment(&JudgmentRequest{
		Action: "test",
	})
	if err == nil {
		t.Error("Expected error for missing entity, got nil")
	}

	// Test missing action
	_, err = client.Judgment(&JudgmentRequest{
		Entity: "test",
	})
	if err == nil {
		t.Error("Expected error for missing action, got nil")
	}
}

func TestDelegationValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing delegator
	_, err := client.Delegation(&DelegationRequest{
		Delegatee: "test",
	})
	if err == nil {
		t.Error("Expected error for missing delegator, got nil")
	}

	// Test missing delegatee
	_, err = client.Delegation(&DelegationRequest{
		Delegator: "test",
	})
	if err == nil {
		t.Error("Expected error for missing delegatee, got nil")
	}
}

func TestTerminationValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing terminator
	_, err := client.Termination(&TerminationRequest{
		TargetID:   "test",
		TargetType: "judgment",
	})
	if err == nil {
		t.Error("Expected error for missing terminator, got nil")
	}

	// Test invalid target type
	_, err = client.Termination(&TerminationRequest{
		Terminator: "test",
		TargetID:   "test",
		TargetType: "invalid",
	})
	if err == nil {
		t.Error("Expected error for invalid target type, got nil")
	}
}

func TestVerificationValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing verifier
	_, err := client.Verification(&VerificationRequest{
		TargetID:   "test",
		TargetType: "judgment",
	})
	if err == nil {
		t.Error("Expected error for missing verifier, got nil")
	}
}

func TestVerifyValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing id
	_, err := client.Verify("")
	if err == nil {
		t.Error("Expected error for missing id, got nil")
	}
}
