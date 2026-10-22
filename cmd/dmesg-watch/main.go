package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/comicdeed/dmesg-watch/internal/config"
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
	version = "v0.3.0-beta.7"
)

func main() {
	cfg, loadedPath, help, err := config.Load(version, os.Args[1:])
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		if errors.Is(err, conf.ErrVersionWanted) {
			fmt.Printf("dmesg-watch %s\n", version)
			os.Exit(0)
		}
		log.Fatalf("Fatal: %v", err)
	}

	if loadedPath != "" {
		log.Printf("Loaded configuration from %s", loadedPath)
	}

	selectedBackend, defaultViewer, err := journal.ResolveBackend(cfg.Backend)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	// Resolve viewer command arguments: CLI flag / env / config file > backend default
	var activeViewerArgs []string
	if len(cfg.ViewerCmd) > 0 {
		activeViewerArgs = cfg.ViewerCmd
	} else {
		activeViewerArgs = strings.Fields(defaultViewer)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// OS signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	track := tracker.New(tracker.Config{
		Window:    cfg.Window,
		Threshold: cfg.Threshold,
		Cooldown:  cfg.Cooldown,
	})

	notifier := notify.New("dmesg-watch")

	var uiTray *tray.Tray
	uiTray = tray.New(
		version,
		selectedBackend,
		func() {
			if err := launcher.LaunchArgs(activeViewerArgs); err != nil {
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

		msgCh, errCh := journal.Stream(ctx, selectedBackend)
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
						count, cfg.Window.String(), latest)
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
