package logging

import (
	"testing"

	golog "github.com/codemedic/go-log"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected golog.Level
		wantErr  bool
	}{
		{"debug", golog.Debug, false},
		{"DEBUG", golog.Debug, false},
		{"info", golog.Info, false},
		{"", golog.Info, false},
		{"warn", golog.Warning, false},
		{"warning", golog.Warning, false},
		{"WARNING", golog.Warning, false},
		{"error", golog.Error, false},
		{"invalid", 0, true},
	}

	for _, tc := range tests {
		lvl, err := ParseLevel(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("ParseLevel(%q) expected error, got nil", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", tc.input, err)
		}
		if !tc.wantErr && lvl != tc.expected {
			t.Errorf("ParseLevel(%q) = %v, expected %v", tc.input, lvl, tc.expected)
		}
	}
}

func TestSetup_Targets(t *testing.T) {
	// Stderr target
	l, err := Setup(TargetStderr, "info", "testtag")
	if err != nil {
		t.Fatalf("Setup(stderr) failed: %v", err)
	}
	if l.Level() != golog.Info {
		t.Errorf("expected level Info, got %v", l.Level())
	}
	l.Close()

	// Auto target (should succeed either via syslog or fallback to stderr)
	lAuto, err := Setup(TargetAuto, "debug", "testtag")
	if err != nil {
		t.Fatalf("Setup(auto) failed: %v", err)
	}
	if lAuto.Level() != golog.Debug {
		t.Errorf("expected level Debug, got %v", lAuto.Level())
	}
	lAuto.Close()

	// Invalid target
	_, err = Setup("invalid_target", "info", "testtag")
	if err == nil {
		t.Errorf("expected error for invalid target, got nil")
	}

	// Invalid level
	_, err = Setup(TargetStderr, "invalid_level", "testtag")
	if err == nil {
		t.Errorf("expected error for invalid level, got nil")
	}
}
