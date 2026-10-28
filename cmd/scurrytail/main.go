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
	"github.com/comicdeed/scurrytail/internal/config"
	"github.com/comicdeed/scurrytail/internal/journal"
	"github.com/comicdeed/scurrytail/internal/launcher"
	"github.com/comicdeed/scurrytail/internal/notify"
	"github.com/comicdeed/scurrytail/internal/tracker"
	"github.com/comicdeed/scurrytail/internal/tray"
)

const (
	projectURL = "https://github.com/comicdeed/scurrytail"
)

var (
	// Version can be populated at build time via -ldflags "-X main.version=..."
	version = "v0.3.0-beta.13"
)

func main() {
	cfg, loadedPath, help, err := config.Load(version, os.Args[1:])
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		if errors.Is(err, conf.ErrVersionWanted) {
			fmt.Printf("ScurryTail %s\n", version)
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

	mt, err := tracker.NewMultiTracker(cfg)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	notifier := notify.New("ScurryTail")

	catNames := make([]string, 0, len(cfg.Categories))
	for name, cat := range cfg.Categories {
		if cat.Ignore {
			continue
		}
		catNames = append(catNames, name)
	}

	var uiTray *tray.Tray
	uiTray = tray.New(
		version,
		selectedBackend,
		cfg.Threshold,
		catNames,
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

		// Throttle tray D-Bus updates to at most 4Hz (every 250ms) during log storms
		trayCoalesceTicker := time.NewTicker(250 * time.Millisecond)
		defer trayCoalesceTicker.Stop()
		var pendingTrayUpdate bool

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
				res := mt.Record(msg, now)
				if res.Ignored {
					continue
				}

				if res.ShouldAlert {
					uiTray.RequestAttention()
					var summary, body string
					if res.Category != "" {
						summary = fmt.Sprintf("⚠️ High Kernel Error Rate [%s]", res.Category)
						body = fmt.Sprintf("Detected %d %s errors in the last %s!\nLatest: %s",
							res.Count, res.Category, res.Window.String(), res.Latest)
					} else {
						summary = "⚠️ High Kernel Error Rate"
						body = fmt.Sprintf("Detected %d kernel errors in the last %s!\nLatest: %s",
							res.Count, res.Window.String(), res.Latest)
					}
					if err := notifier.SendWithUrgency(summary, body, res.Urgency); err != nil {
						log.Printf("Failed to deliver notification: %v", err)
					}
				}

				pendingTrayUpdate = true

			case now := <-trayCoalesceTicker.C:
				if pendingTrayUpdate {
					prune := mt.Prune(now)
					uiTray.UpdateWithCategories(prune.TotalCount, prune.Latest, prune.CategoryCounts, prune.AnyAlerting, prune.AlertCategory, prune.AlertUrgency)
					pendingTrayUpdate = false
				}

			case now := <-ticker.C:
				prune := mt.Prune(now)
				uiTray.UpdateWithCategories(prune.TotalCount, prune.Latest, prune.CategoryCounts, prune.AnyAlerting, prune.AlertCategory, prune.AlertUrgency)
				pendingTrayUpdate = false
			}
		}
	}()

	// Starts main tray event loop (blocking)
	uiTray.Run(func() {
		cancel()
	})
}
