package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Backend != "auto" {
		t.Errorf("expected backend auto, got %s", cfg.Backend)
	}
	if cfg.Threshold != 5 {
		t.Errorf("expected threshold 5, got %d", cfg.Threshold)
	}
	if cfg.Window != 60*time.Second {
		t.Errorf("expected window 60s, got %v", cfg.Window)
	}
	if cfg.Cooldown != 45*time.Second {
		t.Errorf("expected cooldown 45s, got %v", cfg.Cooldown)
	}
	if cfg.ViewerCmd != nil {
		t.Errorf("expected nil ViewerCmd, got %v", cfg.ViewerCmd)
	}
}

func TestLoad_ArrayViewerCmd(t *testing.T) {
	yamlContent := `
backend: dmesg
threshold: 10
window: 2m
cooldown: 30s
viewer_cmd:
  - ghostty
  - -e
  - dmesg
  - -w
`
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(confPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, path, err := Load(confPath)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if path != confPath {
		t.Errorf("expected path %s, got %s", confPath, path)
	}
	if cfg.Backend != "dmesg" {
		t.Errorf("expected backend dmesg, got %s", cfg.Backend)
	}
	if cfg.Threshold != 10 {
		t.Errorf("expected threshold 10, got %d", cfg.Threshold)
	}
	if cfg.Window != 2*time.Minute {
		t.Errorf("expected window 2m, got %v", cfg.Window)
	}
	if cfg.Cooldown != 30*time.Second {
		t.Errorf("expected cooldown 30s, got %v", cfg.Cooldown)
	}
	expectedCmd := StringList{"ghostty", "-e", "dmesg", "-w"}
	if !reflect.DeepEqual(cfg.ViewerCmd, expectedCmd) {
		t.Errorf("expected viewer_cmd %v, got %v", expectedCmd, cfg.ViewerCmd)
	}
}

func TestLoad_StringViewerCmd(t *testing.T) {
	yamlContent := `
viewer_cmd: "journalctl -k -f"
`
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(confPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, _, err := Load(confPath)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	expectedCmd := StringList{"journalctl", "-k", "-f"}
	if !reflect.DeepEqual(cfg.ViewerCmd, expectedCmd) {
		t.Errorf("expected viewer_cmd %v, got %v", expectedCmd, cfg.ViewerCmd)
	}
}

func TestLoad_ExplicitMissing(t *testing.T) {
	_, _, err := Load("/non/existent/config.yaml")
	if err == nil {
		t.Errorf("expected error for missing explicit file, got nil")
	}
}
