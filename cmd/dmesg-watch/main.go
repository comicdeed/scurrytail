package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	version = "v0.3.0-beta.6"
)

func main() {
	defaultCfg := config.Default()

	var (
		configFile  = flag.String("config", "", "Path to YAML configuration file")
		threshold   = flag.Int("threshold", defaultCfg.Threshold, "Number of kernel errors per window to trigger an alert")
		window      = flag.Duration("window", defaultCfg.Window, "Rolling evaluation window duration")
		cooldown    = flag.Duration("cooldown", defaultCfg.Cooldown, "Notification suppression cooldown period")
		backend     = flag.String("backend", defaultCfg.Backend, "Kernel log streaming backend: auto, journalctl, or dmesg")
		viewerCmd   = flag.String("viewer-cmd", "", "Command to view kernel logs in terminal (defaults to backend follow command)")
		showVersion = flag.Bool("version", false, "Print application version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("dmesg-watch %s\n", version)
		os.Exit(0)
	}

	// Load optional configuration file (defaults -> config file -> CLI flags)
	fileCfg, loadedPath, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}
	if loadedPath != "" {
		log.Printf("Loaded configuration from %s", loadedPath)
	}

	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

	activeThreshold := *threshold
	if !explicitFlags["threshold"] && fileCfg.Threshold > 0 {
		activeThreshold = fileCfg.Threshold
	}

	activeWindow := *window
	if !explicitFlags["window"] && fileCfg.Window > 0 {
		activeWindow = fileCfg.Window
	}

	activeCooldown := *cooldown
	if !explicitFlags["cooldown"] && fileCfg.Cooldown > 0 {
		activeCooldown = fileCfg.Cooldown
	}

	activeBackend := *backend
	if !explicitFlags["backend"] && fileCfg.Backend != "" {
		activeBackend = fileCfg.Backend
	}

	selectedBackend, defaultViewer, err := journal.ResolveBackend(activeBackend)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	// Resolve viewer command arguments: CLI flag > config file > backend default
	var activeViewerArgs []string
	if explicitFlags["viewer-cmd"] && strings.TrimSpace(*viewerCmd) != "" {
		activeViewerArgs = strings.Fields(strings.TrimSpace(*viewerCmd))
	} else if len(fileCfg.ViewerCmd) > 0 {
		activeViewerArgs = fileCfg.ViewerCmd
	} else {
		activeViewerArgs = strings.Fields(defaultViewer)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// OS signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	track := tracker.New(tracker.Config{
		Window:    activeWindow,
		Threshold: activeThreshold,
		Cooldown:  activeCooldown,
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
						count, activeWindow.String(), latest)
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
