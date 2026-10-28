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

	// Override TERMINAL and XDG_RUNTIME_DIR env vars to point to temporary directory
	t.Setenv("TERMINAL", fakeTerm)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	expectedScript := filepath.Join(tmpDir, "scurrytail", "view-logs.sh")

	t.Run("default command", func(t *testing.T) {
		cmd, err := BuildCommand("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd.Path != fakeTerm {
			t.Errorf("expected cmd.Path %s, got %s", fakeTerm, cmd.Path)
		}
		expected := []string{fakeTerm, "-e", expectedScript}
		if strings.Join(cmd.Args, " ") != strings.Join(expected, " ") {
			t.Errorf("expected args %v, got %v", expected, cmd.Args)
		}
		if cmd.SysProcAttr == nil || !cmd.SysProcAttr.Setsid {
			t.Errorf("expected Setsid to be true")
		}

		data, err := os.ReadFile(expectedScript)
		if err != nil {
			t.Fatalf("failed to read generated runner script: %v", err)
		}
		if !strings.Contains(string(data), DefaultJournalctlViewer) {
			t.Errorf("expected runner script to contain default viewer, got:\n%s", string(data))
		}
	})

	t.Run("custom command with pipes", func(t *testing.T) {
		cmd, err := BuildCommand("journalctl -k -e | grep -i error")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{fakeTerm, "-e", expectedScript}
		if strings.Join(cmd.Args, " ") != strings.Join(expected, " ") {
			t.Errorf("expected args %v, got %v", expected, cmd.Args)
		}

		data, err := os.ReadFile(expectedScript)
		if err != nil {
			t.Fatalf("failed to read generated runner script: %v", err)
		}
		if !strings.Contains(string(data), "journalctl -k -e | grep -i error") {
			t.Errorf("expected runner script to contain custom command, got:\n%s", string(data))
		}
	})
}

func TestOpenURL_Validation(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"", true},
		{"   ", true},
		{"file:///etc/passwd", true},
		{"--help", true},
		{"javascript:alert(1)", true},
		{"ftp://example.com", true},
	}

	for _, tt := range tests {
		err := OpenURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("OpenURL(%q) err = %v, wantErr = %v", tt.url, err, tt.wantErr)
		}
	}
}

func TestPrepareRunnerScript_Escaping(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	scriptPath, err := PrepareRunnerScript([]string{"journalctl", "-k", "--grep=foo bar", "semi;colon"})
	if err != nil {
		t.Fatalf("PrepareRunnerScript failed: %v", err)
	}

	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read generated script: %v", err)
	}

	expectedExec := "exec journalctl -k '--grep=foo bar' 'semi;colon'\n"
	if !strings.Contains(string(data), expectedExec) {
		t.Errorf("expected script to contain %q, got:\n%s", expectedExec, string(data))
	}
}
