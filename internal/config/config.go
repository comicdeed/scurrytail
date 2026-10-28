package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

// CategoryConfig defines alert rate thresholds and patterns for a specific error category.
type CategoryConfig struct {
	Threshold int           `yaml:"threshold"`
	Window    time.Duration `yaml:"window"`
	Cooldown  time.Duration `yaml:"cooldown"`
	Urgency   string        `yaml:"urgency"` // "low", "normal", or "critical"
	Ignore    bool          `yaml:"ignore"`
	Patterns  []string      `yaml:"patterns"`

	ThresholdSet bool `conf:"-" yaml:"-"`
	WindowSet    bool `conf:"-" yaml:"-"`
	CooldownSet  bool `conf:"-" yaml:"-"`
}

// PatternRule defines an individual classification or suppression rule.
type PatternRule struct {
	Pattern  string `yaml:"pattern"`
	Category string `yaml:"category,omitempty"`
	Ignore   bool   `yaml:"ignore,omitempty"`
}

// Config represents the application configuration settings.
type Config struct {
	conf.Version
	ConfigFile  string                    `conf:"flag:config,short:c,env:CONFIG,help:Path to YAML configuration file"`
	Backend     string                    `conf:"default:auto,help:Kernel log streaming backend: auto, journalctl, or dmesg" yaml:"backend"`
	Threshold   int                       `conf:"default:5,help:Number of kernel errors per window to trigger an alert" yaml:"threshold"`
	Window      time.Duration             `conf:"default:1m,help:Rolling evaluation window duration" yaml:"window"`
	Cooldown    time.Duration             `conf:"default:45s,help:Notification suppression cooldown period" yaml:"cooldown"`
	ViewerCmd   StringList                `conf:"flag:viewer-cmd,help:Command to view kernel logs in terminal" yaml:"viewer_cmd"`
	Ignore      []string                  `conf:"-" yaml:"ignore"`
	Categories  map[string]CategoryConfig `conf:"-" yaml:"categories"`
	Patterns    []PatternRule             `conf:"-" yaml:"patterns"`
	LoadedFiles []string                  `conf:"-" yaml:"-"`
}

// Default returns a Config populated with standard defaults.
func Default() Config {
	return Config{
		Backend:    "auto",
		Threshold:  5,
		Window:     60 * time.Second,
		Cooldown:   45 * time.Second,
		ViewerCmd:  nil,
		Categories: make(map[string]CategoryConfig),
	}
}

// categoryLayer holds pointer fields for lossless merging across YAML layers.
type categoryLayer struct {
	Threshold *int           `yaml:"threshold"`
	Window    *time.Duration `yaml:"window"`
	Cooldown  *time.Duration `yaml:"cooldown"`
	Urgency   *string        `yaml:"urgency"`
	Ignore    *bool          `yaml:"ignore"`
	Patterns  []string       `yaml:"patterns"`
}

// layerConfig represents an individual configuration layer before merging.
type layerConfig struct {
	Backend    *string                  `yaml:"backend"`
	Threshold  *int                     `yaml:"threshold"`
	Window     *time.Duration           `yaml:"window"`
	Cooldown   *time.Duration           `yaml:"cooldown"`
	ViewerCmd  *StringList              `yaml:"viewer_cmd"`
	Ignore     []string                 `yaml:"ignore"`
	Categories map[string]categoryLayer `yaml:"categories"`
	Patterns   []PatternRule            `yaml:"patterns"`
}

// DiscoverConfigFiles locates all configuration files and drop-ins in hierarchical order.
func DiscoverConfigFiles(explicitPath string) ([]string, error) {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err != nil {
			return nil, fmt.Errorf("explicit config file not found: %w", err)
		}
		var files []string
		files = append(files, explicitPath)

		// Discover adjacent conf.d if present (e.g. /path/to/conf.d/*.yaml)
		adjacentConfD := filepath.Join(filepath.Dir(explicitPath), "conf.d")
		dropIns, err := readDropInDir(adjacentConfD)
		if err == nil {
			files = append(files, dropIns...)
		}
		return files, nil
	}

	var files []string

	// 1. System Base Config: /etc/scurrytail/config.yaml (or legacy /etc/dmesg-watch/config.yaml)
	for _, sysPath := range []string{
		"/etc/scurrytail/config.yaml", "/etc/scurrytail/config.yml",
		"/etc/dmesg-watch/config.yaml", "/etc/dmesg-watch/config.yml",
	} {
		if _, err := os.Stat(sysPath); err == nil {
			files = append(files, sysPath)
			break
		}
	}

	// 2. System Drop-ins: /etc/scurrytail/conf.d/*.yaml (or legacy /etc/dmesg-watch/conf.d/*.yaml)
	for _, sysDir := range []string{"/etc/scurrytail/conf.d", "/etc/dmesg-watch/conf.d"} {
		dropIns, err := readDropInDir(sysDir)
		if err == nil && len(dropIns) > 0 {
			files = append(files, dropIns...)
			break
		}
	}

	// 3. User Base Config: ~/.config/scurrytail/config.yaml
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			xdgHome = filepath.Join(home, ".config")
		}
	}

	if xdgHome != "" {
		for _, dir := range []string{"scurrytail", "dmesg-watch"} {
			var found bool
			for _, name := range []string{"config.yaml", "config.yml"} {
				p := filepath.Join(xdgHome, dir, name)
				if _, err := os.Stat(p); err == nil {
					files = append(files, p)
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// 4. User Drop-ins: ~/.config/scurrytail/conf.d/*.yaml
		for _, dir := range []string{"scurrytail", "dmesg-watch"} {
			dropIns, err := readDropInDir(filepath.Join(xdgHome, dir, "conf.d"))
			if err == nil && len(dropIns) > 0 {
				files = append(files, dropIns...)
				break
			}
		}
	}

	return files, nil
}

// readDropInDir scans a drop-in directory for *.yaml and *.yml files, returning sorted paths.
func readDropInDir(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, filepath.Join(dirPath, name))
		}
	}

	sort.Strings(files)
	return files, nil
}

// MergeYAMLLayers parses multiple configuration files and merges them in sequential hierarchical order.
func MergeYAMLLayers(paths []string) (Config, []byte, error) {
	merged := Default()
	catExplicitlySet := make(map[string]map[string]bool)

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return merged, nil, fmt.Errorf("reading config file %q: %w", p, err)
		}

		var layer layerConfig
		if err := goyaml.Unmarshal(data, &layer); err != nil {
			return merged, nil, fmt.Errorf("parsing YAML in %q: %w", p, err)
		}

		if layer.Backend != nil {
			merged.Backend = *layer.Backend
		}
		if layer.Threshold != nil {
			merged.Threshold = *layer.Threshold
		}
		if layer.Window != nil {
			merged.Window = *layer.Window
		}
		if layer.Cooldown != nil {
			merged.Cooldown = *layer.Cooldown
		}
		if layer.ViewerCmd != nil {
			merged.ViewerCmd = *layer.ViewerCmd
		}

		// Append ignore patterns (with deduplication)
		for _, pat := range layer.Ignore {
			pat = strings.TrimSpace(pat)
			if pat != "" && !containsString(merged.Ignore, pat) {
				merged.Ignore = append(merged.Ignore, pat)
			}
		}

		// Merge categories
		for catName, catL := range layer.Categories {
			catName = strings.TrimSpace(catName)
			if catName == "" {
				continue
			}

			cat := merged.Categories[catName]
			if catExplicitlySet[catName] == nil {
				catExplicitlySet[catName] = make(map[string]bool)
			}

			if catL.Threshold != nil {
				cat.Threshold = *catL.Threshold
				catExplicitlySet[catName]["threshold"] = true
			}
			if catL.Window != nil {
				cat.Window = *catL.Window
				catExplicitlySet[catName]["window"] = true
			}
			if catL.Cooldown != nil {
				cat.Cooldown = *catL.Cooldown
				catExplicitlySet[catName]["cooldown"] = true
			}
			if catL.Urgency != nil {
				cat.Urgency = strings.ToLower(strings.TrimSpace(*catL.Urgency))
				catExplicitlySet[catName]["urgency"] = true
			}
			if catL.Ignore != nil {
				cat.Ignore = *catL.Ignore
				catExplicitlySet[catName]["ignore"] = true
			}

			for _, pat := range catL.Patterns {
				pat = strings.TrimSpace(pat)
				if pat != "" && !containsString(cat.Patterns, pat) {
					cat.Patterns = append(cat.Patterns, pat)
				}
			}

			merged.Categories[catName] = cat
		}

		// Append flat patterns
		for _, rule := range layer.Patterns {
			rule.Pattern = strings.TrimSpace(rule.Pattern)
			if rule.Pattern != "" {
				merged.Patterns = append(merged.Patterns, rule)
			}
		}
	}

	// Apply default inheritance for categories where properties were not explicitly set
	for catName, cat := range merged.Categories {
		if !catExplicitlySet[catName]["threshold"] || cat.Threshold <= 0 {
			cat.Threshold = merged.Threshold
			cat.ThresholdSet = false
		} else {
			cat.ThresholdSet = true
		}
		if !catExplicitlySet[catName]["window"] || cat.Window <= 0 {
			cat.Window = merged.Window
			cat.WindowSet = false
		} else {
			cat.WindowSet = true
		}
		if !catExplicitlySet[catName]["cooldown"] || cat.Cooldown <= 0 {
			cat.Cooldown = merged.Cooldown
			cat.CooldownSet = false
		} else {
			cat.CooldownSet = true
		}
		if !catExplicitlySet[catName]["urgency"] || cat.Urgency == "" {
			cat.Urgency = "normal"
		}
		merged.Categories[catName] = cat
	}

	merged.LoadedFiles = paths

	// Re-encode merged scalars for ardanlabs/conf/v3
	type scalarOverlay struct {
		Backend   string     `yaml:"backend"`
		Threshold int        `yaml:"threshold"`
		Window    string     `yaml:"window"`
		Cooldown  string     `yaml:"cooldown"`
		ViewerCmd StringList `yaml:"viewer_cmd,omitempty"`
	}

	overlay := scalarOverlay{
		Backend:   merged.Backend,
		Threshold: merged.Threshold,
		Window:    merged.Window.String(),
		Cooldown:  merged.Cooldown.String(),
		ViewerCmd: merged.ViewerCmd,
	}

	outYAML, err := goyaml.Marshal(overlay)
	if err != nil {
		return merged, nil, fmt.Errorf("encoding merged YAML: %w", err)
	}

	return merged, outYAML, nil
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ValidatePatterns verifies that all regex patterns in the config compile cleanly.
func ValidatePatterns(cfg Config) error {
	for _, pat := range cfg.Ignore {
		if _, err := regexp.Compile(pat); err != nil {
			return fmt.Errorf("invalid regex in ignore pattern %q: %w", pat, err)
		}
	}

	for catName, cat := range cfg.Categories {
		for _, pat := range cat.Patterns {
			if _, err := regexp.Compile(pat); err != nil {
				return fmt.Errorf("invalid regex in category %q pattern %q: %w", catName, pat, err)
			}
		}
	}

	for _, rule := range cfg.Patterns {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid regex in pattern rule %q: %w", rule.Pattern, err)
		}
	}

	return nil
}

// FindConfigFile locates the primary configuration file (backward compatibility helper).
func FindConfigFile(explicitPath string) string {
	files, err := DiscoverConfigFiles(explicitPath)
	if err != nil || len(files) == 0 {
		return ""
	}
	return files[0]
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
	if v := os.Getenv("SCURRYTAIL_CONFIG"); v != "" {
		return v
	}
	return os.Getenv("DMESG_WATCH_CONFIG")
}

// Load loads application configuration using ardanlabs/conf/v3 with layered YAML merging.
// Precedence: defaults -> YAML layers (base + conf.d) -> environment variables -> command line flags.
func Load(versionStr string, args []string) (Config, string, string, error) {
	var cfg Config
	cfg.Version = conf.Version{
		Build: versionStr,
		Desc:  "A nimble Linux system tray applet for tailing fast-moving kernel logs and catching errors in real time",
	}

	explicitPath := findExplicitConfigFile(args)
	files, err := DiscoverConfigFiles(explicitPath)
	if err != nil {
		return cfg, "", "", err
	}

	var parsers []conf.Parsers
	var mergedYAML []byte
	var mergedCfg Config

	if len(files) > 0 {
		var mergeErr error
		mergedCfg, mergedYAML, mergeErr = MergeYAMLLayers(files)
		if mergeErr != nil {
			return cfg, "", "", mergeErr
		}
		parsers = append(parsers, yaml.WithData(mergedYAML))
	} else {
		mergedCfg = Default()
	}

	origArgs := os.Args
	if args != nil {
		os.Args = append([]string{origArgs[0]}, args...)
		defer func() { os.Args = origArgs }()
	}

	help, err := conf.Parse("SCURRYTAIL", &cfg, parsers...)
	if err != nil {
		var loadedSummary string
		if len(files) > 0 {
			loadedSummary = strings.Join(files, ", ")
		}
		return cfg, loadedSummary, help, err
	}

	// Carry over complex structures from merged YAML
	cfg.Ignore = mergedCfg.Ignore
	cfg.Categories = mergedCfg.Categories
	cfg.Patterns = mergedCfg.Patterns
	cfg.LoadedFiles = files

	// If CLI flags / env vars overrode global settings, update categories that inherited global defaults
	for catName, cat := range cfg.Categories {
		if !cat.ThresholdSet && cfg.Threshold != mergedCfg.Threshold {
			cat.Threshold = cfg.Threshold
		}
		if !cat.WindowSet && cfg.Window != mergedCfg.Window {
			cat.Window = cfg.Window
		}
		if !cat.CooldownSet && cfg.Cooldown != mergedCfg.Cooldown {
			cat.Cooldown = cfg.Cooldown
		}
		cfg.Categories[catName] = cat
	}

	// Validate compiled regexes early
	if err := ValidatePatterns(cfg); err != nil {
		return cfg, strings.Join(files, ", "), "", err
	}

	var loadedSummary string
	if len(files) > 0 {
		loadedSummary = strings.Join(files, ", ")
	}

	return cfg, loadedSummary, "", nil
}
