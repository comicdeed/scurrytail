package journal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ResolveBackend validates and resolves the streaming backend ("auto", "journalctl", or "dmesg").
// It returns the normalized backend name, default viewer command, or an error if prerequisites are missing.
func ResolveBackend(requested string) (backend string, defaultViewer string, err error) {
	req := strings.ToLower(strings.TrimSpace(requested))
	switch req {
	case "", "auto":
		if isSystemdBooted() && hasCommand("journalctl") {
			return "journalctl", "journalctl -k -f -n 50", nil
		}
		if hasCommand("dmesg") {
			return "dmesg", "dmesg -w -l err,crit,alert,emerg -T", nil
		}
		if hasCommand("journalctl") {
			return "journalctl", "journalctl -k -f -n 50", nil
		}
		return "", "", fmt.Errorf("neither 'journalctl' nor 'dmesg' could be found on PATH")

	case "journalctl", "journald":
		if !hasCommand("journalctl") {
			return "", "", fmt.Errorf("'journalctl' not found on PATH")
		}
		return "journalctl", "journalctl -k -f -n 50", nil

	case "dmesg":
		if !hasCommand("dmesg") {
			return "", "", fmt.Errorf("'dmesg' not found on PATH")
		}
		return "dmesg", "dmesg -w -l err,crit,alert,emerg -T", nil

	default:
		return "", "", fmt.Errorf("unsupported backend %q (valid options: auto, journalctl, dmesg)", requested)
	}
}

func isSystemdBooted() bool {
	info, err := os.Stat("/run/systemd/system")
	return err == nil && info.IsDir()
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// CleanKernelMessage extracts the core error message from a log line,
// stripping timestamps (ISO or uptime seconds), hostname, and kernel/syslog prefixes.
func CleanKernelMessage(raw string) string {
	line := strings.TrimSpace(raw)
	if line == "" {
		return ""
	}

	// Case 1: Standard journalctl kernel line ("kernel: <msg>")
	if idx := strings.Index(line, "kernel: "); idx != -1 {
		return strings.TrimSpace(line[idx+len("kernel: "):])
	}

	// Case 2: Process/syslog format ("process[pid]: <msg>")
	if idx := strings.Index(line, "]: "); idx != -1 {
		return strings.TrimSpace(line[idx+len("]: "):])
	}

	// Case 3: dmesg uptime timestamp format ("[12345.678901] <subsystem>: <msg>")
	if strings.HasPrefix(line, "[") {
		if closeIdx := strings.Index(line, "] "); closeIdx != -1 {
			return strings.TrimSpace(line[closeIdx+2:])
		}
	}

	// Case 4: dmesg ISO timestamp format ("2026-10-21T01:38:37,848625+01:00 <subsystem>: <msg>")
	if len(line) > 25 && line[10] == 'T' {
		if spIdx := strings.Index(line, " "); spIdx != -1 {
			return strings.TrimSpace(line[spIdx+1:])
		}
	}

	return line
}

// Stream launches the requested backend in the background,
// emitting cleaned log lines to the returned channel until ctx is cancelled.
func Stream(ctx context.Context, backend string) (<-chan string, <-chan error) {
	msgCh := make(chan string, 32)
	errCh := make(chan error, 1)

	var cmd *exec.Cmd
	switch backend {
	case "dmesg":
		cmd = exec.CommandContext(ctx, "dmesg", "-W", "-l", "err,crit,alert,emerg", "--time-format=iso")
	default:
		cmd = exec.CommandContext(ctx, "journalctl", "-k", "-f", "-n", "0", "-p", "3", "--output", "short-precise")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- err
		close(msgCh)
		close(errCh)
		return msgCh, errCh
	}

	if err := cmd.Start(); err != nil {
		errCh <- err
		close(msgCh)
		close(errCh)
		return msgCh, errCh
	}

	go func() {
		defer close(msgCh)
		defer close(errCh)

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			cleaned := CleanKernelMessage(scanner.Text())
			if cleaned != "" {
				select {
				case msgCh <- cleaned:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			errCh <- err
		}
		_ = cmd.Wait()
	}()

	return msgCh, errCh
}
