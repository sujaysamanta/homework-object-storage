package utils

import (
	"testing"
)

func TestIsValidObjectID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		// Valid IDs
		{"Valid alphanumeric ID", "abc123", true},
		{"Valid numeric ID", "123456", true},
		{"Valid alphabetic ID", "abcDEF", true},
		{"Valid ID with max length", "12345678901234567890123456789012", true}, // 32 chars

		// Invalid IDs
		{"Empty ID", "", false},
		{"ID too long", "123456789012345678901234567890123", false}, // 33 chars
		{"ID with special chars", "abc-123", false},
		{"ID with space", "abc 123", false},
		{"ID with underscore", "abc_123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidObjectID(tt.id)
			if result != tt.expected {
				t.Errorf("IsValidObjectID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}
