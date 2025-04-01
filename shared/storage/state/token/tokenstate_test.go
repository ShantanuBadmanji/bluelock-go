package token

import "testing"

func TestValidateTokenStatus(t *testing.T) {
	// Test valid statuses
	validStatuses := []TokenStatus{TokenActive, TokenExhausted, TokenUnauthorized}
	for _, status := range validStatuses {
		if !IsTokenStatusValid(status) {
			t.Errorf("Expected status %s to be valid", status)
		}
	}

	// Test invalid statuses
	invalidStatuses := []TokenStatus{"inactive", "blocked", "unknown"}
	for _, status := range invalidStatuses {
		if IsTokenStatusValid(status) {
			t.Errorf("Expected status %s to be invalid", status)
		}
	}
}
