package journal

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
)

// CleanKernelMessage extracts the core error message from a journalctl log line,
// stripping timestamp, hostname, and kernel/syslog prefixes.
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

	return line
}

// Stream launches journalctl -k -f -p 3 --output short-precise in the background,
// emitting cleaned log lines to the returned channel until ctx is cancelled.
func Stream(ctx context.Context) (<-chan string, <-chan error) {
	msgCh := make(chan string, 32)
	errCh := make(chan error, 1)

	cmd := exec.CommandContext(ctx, "journalctl", "-k", "-f", "-p", "3", "--output", "short-precise")
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
