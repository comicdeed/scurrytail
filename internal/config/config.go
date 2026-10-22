package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/conf/v3/yaml"
	goyaml "gopkg.in/yaml.v3"
)

// StringList supports unmarshaling from either a YAML list of strings or a single command string,
// and supports parsing from CLI flags / env vars via conf.Setter.
type StringList []string

// Set implements conf.Setter for ardanlabs/conf/v3 CLI flags and environment variables.
func (s *StringList) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		*s = nil
		return nil
	}
	*s = strings.Fields(trimmed)
	return nil
}

// UnmarshalYAML implements custom unmarshaling for YAML.
func (s *StringList) UnmarshalYAML(value *goyaml.Node) error {
	if value.Kind == goyaml.ScalarNode {
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
	if value.Kind == goyaml.SequenceNode {
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
	conf.Version
	ConfigFile string        `conf:"flag:config,short:c,env:CONFIG,help:Path to YAML configuration file"`
	Backend    string        `conf:"default:auto,help:Kernel log streaming backend: auto, journalctl, or dmesg" yaml:"backend"`
	Threshold  int           `conf:"default:5,help:Number of kernel errors per window to trigger an alert" yaml:"threshold"`
	Window     time.Duration `conf:"default:1m,help:Rolling evaluation window duration" yaml:"window"`
	Cooldown   time.Duration `conf:"default:45s,help:Notification suppression cooldown period" yaml:"cooldown"`
	ViewerCmd  StringList    `conf:"flag:viewer-cmd,help:Command to view kernel logs in terminal" yaml:"viewer_cmd"`
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

func findExplicitConfigFile(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-c" || arg == "--config" || arg == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "-c=") || strings.HasPrefix(arg, "--config=") || strings.HasPrefix(arg, "-config=") {
			parts := strings.SplitN(arg, "=", 2)
			return parts[1]
		}
	}
	return os.Getenv("DMESG_WATCH_CONFIG")
}

// Load loads application configuration using ardanlabs/conf/v3.
// Precedence: defaults -> YAML config file -> environment variables -> command line flags.
func Load(versionStr string, args []string) (Config, string, string, error) {
	var cfg Config
	cfg.Version = conf.Version{
		Build: versionStr,
		Desc:  "Lightweight Linux kernel error monitor",
	}

	explicitPath := findExplicitConfigFile(args)
	configPath := FindConfigFile(explicitPath)

	var parsers []conf.Parsers
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			if explicitPath != "" {
				return cfg, "", "", fmt.Errorf("reading config file %q: %w", configPath, err)
			}
		} else {
			parsers = append(parsers, yaml.WithData(data))
		}
	}

	origArgs := os.Args
	if args != nil {
		os.Args = append([]string{origArgs[0]}, args...)
		defer func() { os.Args = origArgs }()
	}

	help, err := conf.Parse("DMESG_WATCH", &cfg, parsers...)
	if err != nil {
		return cfg, configPath, help, err
	}

	return cfg, configPath, "", nil
}
