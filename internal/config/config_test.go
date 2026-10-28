package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
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
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("SCURRYTAIL_BACKEND", "dmesg")
	t.Setenv("SCURRYTAIL_THRESHOLD", "15")
	t.Setenv("SCURRYTAIL_COOLDOWN", "10s")

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

	t.Setenv("SCURRYTAIL_THRESHOLD", "20")

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

func TestLoad_LayeredConfD(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")
	baseYAML := `
threshold: 5
window: 60s
cooldown: 45s
ignore:
  - "ACPI:.*Method _OSI failed"
categories:
  disk:
    threshold: 1
    patterns:
      - "EXT4-fs error"
`
	if err := os.WriteFile(confPath, []byte(baseYAML), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	confD := filepath.Join(tmpDir, "conf.d")
	if err := os.MkdirAll(confD, 0755); err != nil {
		t.Fatalf("failed to create conf.d: %v", err)
	}

	wifiYAML := `
categories:
  wifi:
    threshold: 8
    window: 2m
    patterns:
      - "iwlwifi.*"
`
	if err := os.WriteFile(filepath.Join(confD, "10-wifi.yaml"), []byte(wifiYAML), 0644); err != nil {
		t.Fatalf("failed to write 10-wifi.yaml: %v", err)
	}

	diskOverrideYAML := `
threshold: 7
categories:
  disk:
    urgency: critical
    patterns:
      - "nvme.*status: 0x"
ignore:
  - "thermal thermal_zone.*"
`
	if err := os.WriteFile(filepath.Join(confD, "20-disk-override.yaml"), []byte(diskOverrideYAML), 0644); err != nil {
		t.Fatalf("failed to write 20-disk-override.yaml: %v", err)
	}

	cfg, loadedSummary, _, err := Load("v0.1.0", []string{"--config", confPath})
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if cfg.Threshold != 7 {
		t.Errorf("expected global threshold 7, got %d", cfg.Threshold)
	}

	expectedIgnore := []string{"ACPI:.*Method _OSI failed", "thermal thermal_zone.*"}
	if !reflect.DeepEqual(cfg.Ignore, expectedIgnore) {
		t.Errorf("expected ignore %v, got %v", expectedIgnore, cfg.Ignore)
	}

	diskCat, ok := cfg.Categories["disk"]
	if !ok {
		t.Fatalf("missing disk category")
	}
	if diskCat.Threshold != 1 {
		t.Errorf("expected disk threshold 1, got %d", diskCat.Threshold)
	}
	if diskCat.Urgency != "critical" {
		t.Errorf("expected disk urgency critical, got %s", diskCat.Urgency)
	}
	expectedDiskPatterns := []string{"EXT4-fs error", "nvme.*status: 0x"}
	if !reflect.DeepEqual(diskCat.Patterns, expectedDiskPatterns) {
		t.Errorf("expected disk patterns %v, got %v", expectedDiskPatterns, diskCat.Patterns)
	}

	wifiCat, ok := cfg.Categories["wifi"]
	if !ok {
		t.Fatalf("missing wifi category")
	}
	if wifiCat.Threshold != 8 {
		t.Errorf("expected wifi threshold 8, got %d", wifiCat.Threshold)
	}
	if wifiCat.Window != 2*time.Minute {
		t.Errorf("expected wifi window 2m, got %v", wifiCat.Window)
	}
	// Wifi cooldown wasn't set, so it should inherit global (45s)
	if wifiCat.Cooldown != 45*time.Second {
		t.Errorf("expected wifi cooldown 45s, got %v", wifiCat.Cooldown)
	}

	if !strings.Contains(loadedSummary, "10-wifi.yaml") || !strings.Contains(loadedSummary, "20-disk-override.yaml") {
		t.Errorf("loadedSummary does not contain conf.d files: %s", loadedSummary)
	}
}

func TestLoad_InvalidRegex(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")
	invalidYAML := `
categories:
  bad:
    patterns:
      - "[unclosed regex"
`
	if err := os.WriteFile(confPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, _, _, err := Load("v0.1.0", []string{"--config", confPath})
	if err == nil {
		t.Fatalf("expected error on invalid regex, got nil")
	}
	if !strings.Contains(err.Error(), "invalid regex") {
		t.Errorf("expected error to mention invalid regex, got: %v", err)
	}
}
