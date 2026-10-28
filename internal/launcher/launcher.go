package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ErrNoTerminalFound is returned when no supported terminal emulator can be located.
var ErrNoTerminalFound = errors.New("no suitable terminal emulator found")

// TerminalSpec describes how to invoke a specific terminal emulator.
type TerminalSpec struct {
	Binary     string
	ArgsPrefix []string
}

// Candidates defines known terminal emulators and their command execution prefixes.
var candidates = []TerminalSpec{
	{Binary: "ghostty", ArgsPrefix: []string{"-e"}},
	{Binary: "konsole", ArgsPrefix: []string{"-e"}},
	{Binary: "gnome-terminal", ArgsPrefix: []string{"--"}},
	{Binary: "ptyxis", ArgsPrefix: []string{"--"}},
	{Binary: "alacritty", ArgsPrefix: []string{"-e"}},
	{Binary: "xfce4-terminal", ArgsPrefix: []string{"-e"}},
	{Binary: "kitty", ArgsPrefix: []string{"-e"}},
	{Binary: "foot", ArgsPrefix: []string{"-e"}},
	{Binary: "wezterm", ArgsPrefix: []string{"start", "--"}},
	{Binary: "x-terminal-emulator", ArgsPrefix: []string{"-e"}},
	{Binary: "xterm", ArgsPrefix: []string{"-e"}},
}

// DetectTerminal determines the appropriate terminal emulator to use.
// It checks in priority order:
// 1. $TERMINAL environment variable
// 2. KDE Plasma TerminalApplication configuration (if on KDE)
// 3. xdg-terminal-exec (FreeDesktop standard)
// 4. Candidate terminal list on PATH
func DetectTerminal() (*TerminalSpec, error) {
	// 1. Environment variable override
	if termEnv := strings.TrimSpace(os.Getenv("TERMINAL")); termEnv != "" {
		if path, err := exec.LookPath(termEnv); err == nil {
			return makeSpecForBinary(path), nil
		}
	}

	// 2. KDE Desktop setting
	if isKDE() {
		if spec := detectKDEDefaultTerminal(); spec != nil {
			return spec, nil
		}
	}

	// 3. FreeDesktop xdg-terminal-exec
	if path, err := exec.LookPath("xdg-terminal-exec"); err == nil {
		return &TerminalSpec{
			Binary:     path,
			ArgsPrefix: nil,
		}, nil
	}

	// 4. Prioritized candidates list
	for _, c := range candidates {
		if path, err := exec.LookPath(c.Binary); err == nil {
			return &TerminalSpec{
				Binary:     path,
				ArgsPrefix: c.ArgsPrefix,
			}, nil
		}
	}

	return nil, ErrNoTerminalFound
}

func isKDE() bool {
	desktop := strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP"))
	return strings.Contains(desktop, "KDE")
}

func detectKDEDefaultTerminal() *TerminalSpec {
	for _, tool := range []string{"kreadconfig6", "kreadconfig5"} {
		toolPath, err := exec.LookPath(tool)
		if err != nil {
			continue
		}
		out, err := exec.Command(toolPath, "--group", "General", "--key", "TerminalApplication").Output()
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(string(out))
		if raw == "" {
			continue
		}

		parts := strings.Fields(raw)
		if len(parts) == 0 {
			continue
		}

		// Find the actual binary within the configured command (skipping env vars like FOO=BAR)
		binIndex := -1
		for i, p := range parts {
			if !strings.Contains(p, "=") && p != "env" {
				binIndex = i
				break
			}
		}

		if binIndex == -1 {
			continue
		}

		binCandidate := parts[binIndex]
		fullPath, err := exec.LookPath(binCandidate)
		if err != nil {
			// Binary configured in KDE does not exist on disk, fall through
			continue
		}

		// If env was used or additional flags were present, preserve them up to the binary
		spec := makeSpecForBinary(fullPath)
		if binIndex > 0 {
			prefix := append([]string{}, parts[:binIndex]...)
			prefix = append(prefix, fullPath)
			if len(parts) > binIndex+1 {
				prefix = append(prefix, parts[binIndex+1:]...)
			}
			spec.Binary = prefix[0]
			spec.ArgsPrefix = append(prefix[1:], spec.ArgsPrefix...)
		}
		return spec
	}
	return nil
}

func makeSpecForBinary(path string) *TerminalSpec {
	base := strings.ToLower(path)
	switch {
	case strings.Contains(base, "gnome-terminal") || strings.Contains(base, "ptyxis"):
		return &TerminalSpec{Binary: path, ArgsPrefix: []string{"--"}}
	case strings.Contains(base, "wezterm"):
		return &TerminalSpec{Binary: path, ArgsPrefix: []string{"start", "--"}}
	default:
		return &TerminalSpec{Binary: path, ArgsPrefix: []string{"-e"}}
	}
}

// BuildCommandArgs constructs the exec.Cmd for the requested argument slice using the resolved terminal.
func BuildCommandArgs(viewerArgs []string) (*exec.Cmd, error) {
	spec, err := DetectTerminal()
	if err != nil {
		return nil, err
	}

	if len(viewerArgs) == 0 {
		viewerArgs = []string{"journalctl", "-b", "_TRANSPORT=kernel", "-p", "3", "+", "-b", "_SYSTEMD_USER_UNIT=scurrytail.service", "+", "-b", "SYSLOG_IDENTIFIER=scurrytail", "-n", "50", "-f"}
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, spec.ArgsPrefix...)

	// If a single argument contains shell metacharacters, wrap in sh -c
	if len(viewerArgs) == 1 && strings.ContainsAny(viewerArgs[0], "|;&><$") {
		cmdArgs = append(cmdArgs, "sh", "-c", viewerArgs[0])
	} else {
		cmdArgs = append(cmdArgs, viewerArgs...)
	}

	cmd := exec.Command(spec.Binary, cmdArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	return cmd, nil
}

// BuildCommand constructs the exec.Cmd for the requested command string using the resolved terminal.
func BuildCommand(viewerCmd string) (*exec.Cmd, error) {
	viewerCmd = strings.TrimSpace(viewerCmd)
	if viewerCmd == "" {
		return BuildCommandArgs(nil)
	}
	if strings.ContainsAny(viewerCmd, "|;&><$") {
		return BuildCommandArgs([]string{viewerCmd})
	}
	return BuildCommandArgs(strings.Fields(viewerCmd))
}

// LaunchArgs resolves the terminal emulator and asynchronously runs the command args in a detached session.
func LaunchArgs(viewerArgs []string) error {
	cmd, err := BuildCommandArgs(viewerArgs)
	if err != nil {
		return fmt.Errorf("failed to build viewer command: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terminal process (%s): %w", cmd.Path, err)
	}

	return nil
}

// Launch resolves the terminal emulator and asynchronously runs the viewer command in a detached session.
func Launch(viewerCmd string) error {
	cmd, err := BuildCommand(viewerCmd)
	if err != nil {
		return fmt.Errorf("failed to build viewer command: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start terminal process (%s): %w", cmd.Path, err)
	}

	return nil
}

// OpenURL launches the specified URL in the user's default web browser using xdg-open.
func OpenURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("cannot open empty URL")
	}

	cmd := exec.Command("xdg-open", rawURL)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open URL via xdg-open: %w", err)
	}

	return nil
}
