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
		{
			name:     "dmesg uptime timestamp format",
			input:    "[100028.848625] kvm: ignored rdmsr: 0x492 data 0x0",
			expected: "kvm: ignored rdmsr: 0x492 data 0x0",
		},
		{
			name:     "dmesg ISO timestamp format",
			input:    "2026-10-21T01:38:37,848625+01:00 kvm: ignored rdmsr: 0x492 data 0x0",
			expected: "kvm: ignored rdmsr: 0x492 data 0x0",
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

func TestResolveBackend_Invalid(t *testing.T) {
	_, _, err := ResolveBackend("invalid-backend")
	if err == nil {
		t.Fatalf("expected error for invalid backend, got nil")
	}
}
