package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ardanlabs/conf/v3"
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

func TestLoad_Defaults(t *testing.T) {
	cfg, path, _, err := Load("v0.1.0", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path, got %s", path)
	}
	if cfg.Backend != "auto" || cfg.Threshold != 5 || cfg.Window != time.Minute || cfg.Cooldown != 45*time.Second {
		t.Errorf("unexpected default config values: %+v", cfg)
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

	cfg, path, _, err := Load("v0.1.0", []string{"--config", confPath})
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

	cfg, _, _, err := Load("v0.1.0", []string{"-c", confPath})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	expectedCmd := StringList{"journalctl", "-k", "-f"}
	if !reflect.DeepEqual(cfg.ViewerCmd, expectedCmd) {
		t.Errorf("expected viewer_cmd %v, got %v", expectedCmd, cfg.ViewerCmd)
	}
}

func TestLoad_EnvVarOverride(t *testing.T) {
	t.Setenv("DMESG_WATCH_BACKEND", "dmesg")
	t.Setenv("DMESG_WATCH_THRESHOLD", "15")
	t.Setenv("DMESG_WATCH_COOLDOWN", "10s")

	cfg, _, _, err := Load("v0.1.0", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend != "dmesg" {
		t.Errorf("expected backend dmesg from env, got %s", cfg.Backend)
	}
	if cfg.Threshold != 15 {
		t.Errorf("expected threshold 15 from env, got %d", cfg.Threshold)
	}
	if cfg.Cooldown != 10*time.Second {
		t.Errorf("expected cooldown 10s from env, got %v", cfg.Cooldown)
	}
}

func TestLoad_FlagOverridesEnvAndYaml(t *testing.T) {
	yamlContent := `
threshold: 10
backend: journalctl
`
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(confPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	t.Setenv("DMESG_WATCH_THRESHOLD", "20")

	// Flag should win over both YAML (10) and Env (20)
	cfg, _, _, err := Load("v0.1.0", []string{"-c", confPath, "--threshold=30", "--backend=dmesg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Threshold != 30 {
		t.Errorf("expected threshold 30 from flag, got %d", cfg.Threshold)
	}
	if cfg.Backend != "dmesg" {
		t.Errorf("expected backend dmesg from flag, got %s", cfg.Backend)
	}
}

func TestLoad_HelpWanted(t *testing.T) {
	_, _, help, err := Load("v0.1.0", []string{"--help"})
	if !errors.Is(err, conf.ErrHelpWanted) {
		t.Fatalf("expected ErrHelpWanted, got: %v", err)
	}
	if help == "" {
		t.Errorf("expected non-empty help string")
	}
}

func TestLoad_VersionWanted(t *testing.T) {
	_, _, version, err := Load("v0.1.0", []string{"--version"})
	if !errors.Is(err, conf.ErrVersionWanted) {
		t.Fatalf("expected ErrVersionWanted, got: %v", err)
	}
	if version == "" {
		t.Errorf("expected non-empty version string")
	}
}

func TestLoad_ExplicitMissing(t *testing.T) {
	_, _, _, err := Load("v0.1.0", []string{"--config", "/non/existent/config.yaml"})
	if err == nil {
		t.Errorf("expected error for missing explicit file, got nil")
	}
}
