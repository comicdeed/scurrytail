package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/comicdeed/dmesg-watch/internal/journal"
	"github.com/comicdeed/dmesg-watch/internal/launcher"
	"github.com/comicdeed/dmesg-watch/internal/notify"
	"github.com/comicdeed/dmesg-watch/internal/tracker"
	"github.com/comicdeed/dmesg-watch/internal/tray"
)

const (
	projectURL = "https://github.com/comicdeed/dmesg-watch"
)

var (
	// Version can be populated at build time via -ldflags "-X main.version=..."
	version = "v0.3.0-beta.4"
)

func main() {
	var (
		threshold   = flag.Int("threshold", 5, "Number of kernel errors per window to trigger an alert")
		window      = flag.Duration("window", 60*time.Second, "Rolling evaluation window duration")
		cooldown    = flag.Duration("cooldown", 45*time.Second, "Notification suppression cooldown period")
		viewerCmd   = flag.String("viewer-cmd", "journalctl -k -f -n 50", "Command to view kernel logs in terminal")
		showVersion = flag.Bool("version", false, "Print application version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("dmesg-watch %s\n", version)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// OS signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	track := tracker.New(tracker.Config{
		Window:    *window,
		Threshold: *threshold,
		Cooldown:  *cooldown,
	})

	notifier := notify.New("dmesg-watch")

	var uiTray *tray.Tray
	uiTray = tray.New(
		version,
		func() {
			if err := launcher.Launch(*viewerCmd); err != nil {
				log.Printf("Failed to open kernel log viewer: %v", err)
			}
		},
		func() {
			if err := launcher.OpenURL(projectURL); err != nil {
				log.Printf("Failed to open project website: %v", err)
			}
		},
		func() {
			cancel()
		},
	)

	// Background worker coordinates log stream and metrics
	go func() {
		<-uiTray.Ready()

		msgCh, errCh := journal.Stream(ctx)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case sig := <-sigCh:
				log.Printf("Received signal %s, shutting down...", sig)
				cancel()
				uiTray.Quit()
				return

			case err, ok := <-errCh:
				if ok && err != nil {
					log.Printf("Journal stream error: %v", err)
				}

			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				now := time.Now()
				track.Record(msg, now)

				shouldAlert, count, latest := track.ShouldAlert(now)
				uiTray.Update(count, latest)

				if shouldAlert {
					summary := "⚠️ High Kernel Error Rate"
					body := fmt.Sprintf("Detected %d kernel errors in the last %s!\nLatest: %s",
						count, window.String(), latest)
					if err := notifier.Send(summary, body); err != nil {
						log.Printf("Failed to deliver notification: %v", err)
					}
				}

			case now := <-ticker.C:
				count, latest := track.Prune(now)
				uiTray.Update(count, latest)
			}
		}
	}()

	// Starts main tray event loop (blocking)
	uiTray.Run(func() {
		cancel()
	})
}
