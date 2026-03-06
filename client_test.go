package jep // 已同步更名为 jep

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
		Action: "jep_compliance_test", // 语义化动作名
	})
	if err == nil {
		t.Error("Expected error for missing entity, got nil")
	}

	// Test missing action
	_, err = client.Judgment(&JudgmentRequest{
		Entity: "tester@jep-protocol.org",
	})
	if err == nil {
		t.Error("Expected error for missing action, got nil")
	}
}

func TestDelegationValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing delegator
	_, err := client.Delegation(&DelegationRequest{
		Delegatee: "delegatee@example.com",
	})
	if err == nil {
		t.Error("Expected error for missing delegator, got nil")
	}

	// Test missing delegatee
	_, err = client.Delegation(&DelegationRequest{
		Delegator: "delegator@example.com",
	})
	if err == nil {
		t.Error("Expected error for missing delegatee, got nil")
	}
}

func TestTerminationValidation(t *testing.T) {
	client := NewClient("test-key")

	// Test missing terminator
	_, err := client.Termination(&TerminationRequest{
		TargetID:   "jep_12345", // 使用 jep_ 前缀示例
		TargetType: "judgment",
	})
	if err == nil {
		t.Error("Expected error for missing terminator, got nil")
	}

	// Test invalid target type
	_, err = client.Termination(&TerminationRequest{
		Terminator: "admin@jep-protocol.org",
		TargetID:   "jep_12345",
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
		TargetID:   "jep_12345",
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
