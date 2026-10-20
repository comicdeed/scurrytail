package journal

import (
	"testing"
)

func TestCleanKernelMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "standard journalctl kernel message",
			input:    "Oct 20 09:11:31.579630 dino-nb100637 kernel: usb 5-2.1: Failed to query UVC control: -71",
			expected: "usb 5-2.1: Failed to query UVC control: -71",
		},
		{
			name:     "syslog pid format",
			input:    "Oct 20 09:11:31 host systemd[1]: Unit failed",
			expected: "Unit failed",
		},
		{
			name:     "no prefix",
			input:    "Raw kernel alert without standard prefix",
			expected: "Raw kernel alert without standard prefix",
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CleanKernelMessage(tc.input)
			if got != tc.expected {
				t.Errorf("CleanKernelMessage(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}
