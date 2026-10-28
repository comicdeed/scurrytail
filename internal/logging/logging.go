package logging

import (
	"fmt"
	"os"
	"strings"

	golog "github.com/codemedic/go-log"
)

// Supported logging targets.
const (
	TargetAuto   = "auto"
	TargetSyslog = "syslog"
	TargetStderr = "stderr"
)

// ParseLevel converts a level string to a golog.Level, supporting debug, info, warn/warning, error.
func ParseLevel(str string) (golog.Level, error) {
	trimmed := strings.ToLower(strings.TrimSpace(str))
	if trimmed == "warn" {
		trimmed = "warning"
	}
	if trimmed == "" {
		return golog.Info, nil
	}
	switch trimmed {
	case "debug":
		return golog.Debug, nil
	case "info":
		return golog.Info, nil
	case "warning":
		return golog.Warning, nil
	case "error":
		return golog.Error, nil
	default:
		return 0, fmt.Errorf("invalid log level %q: must be one of debug, info, warning, error", str)
	}
}

// Setup initializes and returns a configured golog.Log instance based on target, level, and syslog tag.
func Setup(target, levelStr, tag string) (golog.Log, error) {
	lvl, err := ParseLevel(levelStr)
	if err != nil {
		return golog.Log{}, err
	}

	normTarget := strings.ToLower(strings.TrimSpace(target))
	if normTarget == "" {
		normTarget = TargetAuto
	}

	switch normTarget {
	case TargetAuto:
		// When running interactively in a terminal TTY, emit directly to stderr for instant feedback.
		if isInteractiveTerminal() {
			return golog.NewStderr(golog.WithLevel(lvl))
		}
		// In non-interactive mode (daemon, background job, systemd unit), attempt syslog first.
		sysLog, err := golog.NewSyslog(
			golog.WithSyslogTag(tag),
			golog.WithLevel(lvl),
		)
		if err == nil {
			return sysLog, nil
		}
		// Gracefully fall back to stderr if syslog is unavailable (e.g. in containers or minimal environments)
		return golog.NewStderr(golog.WithLevel(lvl))

	case TargetSyslog:
		return golog.NewSyslog(
			golog.WithSyslogTag(tag),
			golog.WithLevel(lvl),
		)

	case TargetStderr:
		return golog.NewStderr(golog.WithLevel(lvl))

	default:
		return golog.Log{}, fmt.Errorf("invalid log target %q: must be one of auto, syslog, stderr", target)
	}
}

func isInteractiveTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
