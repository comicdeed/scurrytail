package tray

import (
	"testing"
)

func TestCalculateStatus(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		anyAlerting bool
		threshold   int
		want        Status
	}{
		{"zero errors is passive", 0, false, 5, StatusPassive},
		{"zero errors with alert flag is still passive", 0, true, 5, StatusPassive},
		{"errors below threshold is active", 3, false, 5, StatusActive},
		{"errors reaching threshold is needs attention", 5, false, 5, StatusNeedsAttention},
		{"errors exceeding threshold is needs attention", 10, false, 5, StatusNeedsAttention},
		{"errors below threshold with anyAlerting is needs attention", 2, true, 5, StatusNeedsAttention},
		{"zero threshold with errors is active", 3, false, 0, StatusActive},
		{"zero threshold with anyAlerting is needs attention", 3, true, 0, StatusNeedsAttention},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateStatus(tt.count, tt.anyAlerting, tt.threshold)
			if got != tt.want {
				t.Errorf("CalculateStatus(%d, %v, %d) = %v, want %v",
					tt.count, tt.anyAlerting, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestTruncateRune(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		expected string
	}{
		{"short ascii", "hello", 10, "hello"},
		{"exact ascii", "hello", 5, "hello"},
		{"truncate ascii", "hello world", 8, "hello..."},
		{"short unicode", "こんにちは", 10, "こんにちは"},
		{"truncate unicode", "こんにちは世界", 5, "こん..."},
		{"very short max", "abcdef", 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateRune(tt.input, tt.maxRunes)
			if got != tt.expected {
				t.Errorf("TruncateRune(%q, %d) = %q, want %q",
					tt.input, tt.maxRunes, got, tt.expected)
			}
		})
	}
}
