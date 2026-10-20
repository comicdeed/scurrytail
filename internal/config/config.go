package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// StringList supports unmarshaling from either a YAML list of strings or a single command string.
type StringList []string

// UnmarshalYAML implements custom unmarshaling for StringList.
func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		var str string
		if err := value.Decode(&str); err != nil {
			return err
		}
		trimmed := strings.TrimSpace(str)
		if trimmed == "" {
			*s = nil
			return nil
		}
		*s = strings.Fields(trimmed)
		return nil
	}
	if value.Kind == yaml.SequenceNode {
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*s = list
		return nil
	}
	return fmt.Errorf("viewer_cmd must be a list of strings or a single string")
}

// Config represents the application configuration settings.
type Config struct {
	Backend   string        `yaml:"backend"`
	Threshold int           `yaml:"threshold"`
	Window    time.Duration `yaml:"window"`
	Cooldown  time.Duration `yaml:"cooldown"`
	ViewerCmd StringList    `yaml:"viewer_cmd"`
}

// Default returns a Config populated with standard defaults.
func Default() Config {
	return Config{
		Backend:   "auto",
		Threshold: 5,
		Window:    60 * time.Second,
		Cooldown:  45 * time.Second,
		ViewerCmd: nil,
	}
}

// FindConfigFile locates the configuration file in standard XDG / system locations.
func FindConfigFile(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}

	// 1. User config: $XDG_CONFIG_HOME/dmesg-watch/config.yaml or ~/.config/dmesg-watch/config.yaml
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			xdgHome = filepath.Join(home, ".config")
		}
	}

	if xdgHome != "" {
		for _, name := range []string{"config.yaml", "config.yml"} {
			p := filepath.Join(xdgHome, "dmesg-watch", name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// 2. System-wide config: /etc/dmesg-watch/config.yaml
	for _, sysPath := range []string{"/etc/dmesg-watch/config.yaml", "/etc/dmesg-watch/config.yml"} {
		if _, err := os.Stat(sysPath); err == nil {
			return sysPath
		}
	}

	return ""
}

// Load loads the configuration from disk, overlaying any provided values on top of Default().
// If explicitPath is empty and no configuration file is found, it returns Default() without error.
func Load(explicitPath string) (Config, string, error) {
	cfg := Default()
	path := FindConfigFile(explicitPath)
	if path == "" {
		return cfg, "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if explicitPath != "" {
			return cfg, "", fmt.Errorf("failed to read config file %q: %w", path, err)
		}
		return cfg, "", nil
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, path, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	return cfg, path, nil
}
