package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeSpecForBinary(t *testing.T) {
	tests := []struct {
		path         string
		expectedArgs []string
	}{
		{"/usr/bin/ghostty", []string{"-e"}},
		{"/usr/bin/konsole", []string{"-e"}},
		{"/usr/bin/gnome-terminal", []string{"--"}},
		{"/usr/bin/ptyxis", []string{"--"}},
		{"/usr/bin/wezterm", []string{"start", "--"}},
		{"/usr/bin/alacritty", []string{"-e"}},
		{"/usr/bin/x-terminal-emulator", []string{"-e"}},
	}

	for _, tt := range tests {
		spec := makeSpecForBinary(tt.path)
		if spec.Binary != tt.path {
			t.Errorf("expected binary %s, got %s", tt.path, spec.Binary)
		}
		if len(spec.ArgsPrefix) != len(tt.expectedArgs) {
			t.Fatalf("expected prefix %v, got %v for %s", tt.expectedArgs, spec.ArgsPrefix, tt.path)
		}
		for i := range spec.ArgsPrefix {
			if spec.ArgsPrefix[i] != tt.expectedArgs[i] {
				t.Errorf("arg mismatch at %d: expected %s, got %s", i, tt.expectedArgs[i], spec.ArgsPrefix[i])
			}
		}
	}
}

func TestBuildCommand(t *testing.T) {
	// Create a dummy fake terminal in a temporary directory
	tmpDir := t.TempDir()
	fakeTerm := filepath.Join(tmpDir, "fake-terminal")
	if err := os.WriteFile(fakeTerm, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("failed to create fake terminal: %v", err)
	}

	// Override TERMINAL env var to point to our fake terminal
	t.Setenv("TERMINAL", fakeTerm)

	t.Run("default command", func(t *testing.T) {
		cmd, err := BuildCommand("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd.Path != fakeTerm {
			t.Errorf("expected cmd.Path %s, got %s", fakeTerm, cmd.Path)
		}
		// Expect: fake-terminal -e journalctl -k -f -n 50
		expected := []string{fakeTerm, "-e", "journalctl", "-k", "-f", "-n", "50"}
		if strings.Join(cmd.Args, " ") != strings.Join(expected, " ") {
			t.Errorf("expected args %v, got %v", expected, cmd.Args)
		}
		if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setsid {
			t.Errorf("expected Setsid to be true")
		}
	})

	t.Run("custom command with pipes", func(t *testing.T) {
		cmd, err := BuildCommand("journalctl -k -e | grep -i error")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{fakeTerm, "-e", "sh", "-c", "journalctl -k -e | grep -i error"}
		if strings.Join(cmd.Args, " ") != strings.Join(expected, " ") {
			t.Errorf("expected args %v, got %v", expected, cmd.Args)
		}
	})
}

func TestOpenURL_Empty(t *testing.T) {
	if err := OpenURL(""); err == nil {
		t.Errorf("expected error for empty URL, got nil")
	}
	if err := OpenURL("   "); err == nil {
		t.Errorf("expected error for whitespace URL, got nil")
	}
}
