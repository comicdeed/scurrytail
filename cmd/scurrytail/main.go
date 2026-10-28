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
	"github.com/comicdeed/scurrytail/internal/logging"
	"github.com/comicdeed/scurrytail/internal/notify"
	"github.com/comicdeed/scurrytail/internal/tracker"
	"github.com/comicdeed/scurrytail/internal/tray"
)

const (
	projectURL = "https://github.com/comicdeed/scurrytail"
)

var (
	// Version can be populated at build time via -ldflags "-X main.version=..."
	version = "v0.3.1"
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

	logger, logErr := logging.Setup(cfg.LogTarget, cfg.LogLevel, "scurrytail")
	if logErr != nil {
		log.Fatalf("Fatal: configuring logger: %v", logErr)
	}
	defer logger.Close()

	if loadedPath != "" {
		logger.Infof("Loaded configuration from %s", loadedPath)
	}

	selectedBackend, defaultViewer, err := journal.ResolveBackend(cfg.Backend)
	if err != nil {
		logger.Fatalf("Fatal: %v", err)
	}

	logger.Infof("Starting ScurryTail %s (backend: %s, threshold: %d, window: %s, cooldown: %s, log_target: %s, log_level: %s)",
		version, selectedBackend, cfg.Threshold, cfg.Window, cfg.Cooldown, cfg.LogTarget, cfg.LogLevel)

	// Resolve viewer command arguments: CLI flag / env / config file > backend default
	var activeViewerArgs []string
	if len(cfg.ViewerCmd) > 0 {
		activeViewerArgs = cfg.ViewerCmd
	} else {
		activeViewerArgs = strings.Fields(defaultViewer)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mt, err := tracker.NewMultiTracker(cfg)
	if err != nil {
		logger.Fatalf("Fatal: %v", err)
	}

	notifier := notify.New("ScurryTail")

	catNames := mt.Categories()

	var uiTray *tray.Tray
	uiTray = tray.New(
		version,
		selectedBackend,
		cfg.Threshold,
		catNames,
		func() {
			if err := launcher.LaunchArgs(activeViewerArgs); err != nil {
				logger.Errorf("Failed to open kernel log viewer: %v", err)
			}
		},
		func() {
			if err := launcher.OpenURL(projectURL); err != nil {
				logger.Errorf("Failed to open project website: %v", err)
			}
		},
		func() {
			cancel()
		},
	)

	// OS signal handling (independent listener guarantees immediate shutdown even if D-Bus tray is delayed)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-sigCh:
			logger.Infof("Received signal %s, shutting down...", sig)
			cancel()
			uiTray.Quit()
		case <-ctx.Done():
		}
	}()

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
		var wakeCategory, wakeLatest string

		handleTrayUpdate := func(prune tracker.PruneResult) {
			oldStatus, newStatus, changed := uiTray.UpdateWithCategories(
				prune.TotalCount,
				prune.Latest,
				prune.CategoryCounts,
				prune.AnyAlerting,
				prune.AlertCategory,
				prune.AlertUrgency,
			)
			if !changed {
				return
			}

			switch newStatus {
			case tray.StatusPassive:
				logger.Infof("[STATE] %s -> Passive (rolling window clean, 0 errors)", oldStatus)
				wakeCategory = ""
				wakeLatest = ""
			case tray.StatusActive:
				if oldStatus == tray.StatusPassive {
					catLabel := wakeCategory
					if catLabel == "" {
						catLabel = "unclassified"
					}
					truncated := tray.TruncateRune(wakeLatest, 60)
					logger.Infof("[STATE] %s -> Active (1 active error in %s window, category: %s, latest: %q)",
						oldStatus, cfg.Window, catLabel, truncated)
				} else {
					logger.Infof("[STATE] %s -> Active (%d active error(s) in %s window)", oldStatus, prune.TotalCount, cfg.Window)
				}
			case tray.StatusNeedsAttention:
				cat := prune.AlertCategory
				if cat == "" {
					cat = "global"
				}
				urgency := prune.AlertUrgency
				if urgency == "" {
					urgency = "normal"
				}
				logger.Warningf("[STATE] %s -> NeedsAttention (%d active error(s) in %s window, category: %s, urgency: %s)",
					oldStatus, prune.TotalCount, cfg.Window, cat, urgency)
			}
		}

		for {
			select {
			case <-ctx.Done():
				return

			case err, ok := <-errCh:
				if ok && err != nil {
					logger.Errorf("Journal stream error: %v", err)
				}

			case msg, ok := <-msgCh:
				if !ok {
					if ctx.Err() == nil {
						logger.Errorf("Kernel log stream terminated unexpectedly; shutting down")
						cancel()
						uiTray.Quit()
					}
					return
				}
				now := time.Now()
				res := mt.Record(msg, now)
				if res.Ignored {
					continue
				}

				if wakeCategory == "" && wakeLatest == "" {
					wakeCategory = res.Category
					wakeLatest = res.Latest
				}

				if res.ShouldAlert {
					uiTray.RequestAttention()
					alertCat := res.Category
					if alertCat == "" {
						alertCat = "global"
					}

					if res.SupersededBy != "" {
						logger.Warningf("[ALERT] Threshold reached [%s]: %d errors in %s (threshold %d, urgency %s) -> dispatched notification; tray display remains latched to [%s] (%s)",
							alertCat, res.Count, res.Window, res.Threshold, res.Urgency, res.SupersededBy, res.SupersededRank)
					} else {
						logger.Warningf("[ALERT] Threshold reached [%s]: %d errors in %s (threshold %d, urgency %s) -> dispatching notification",
							alertCat, res.Count, res.Window, res.Threshold, res.Urgency)
					}

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

					// Asynchronous notification delivery to eliminate D-Bus head-of-line blocking
					urgency := res.Urgency
					go func(s, b, u string) {
						if err := notifier.SendWithUrgency(s, b, u); err != nil {
							logger.Errorf("Failed to deliver notification: %v", err)
						}
					}(summary, body, urgency)
				}

				pendingTrayUpdate = true

			case now := <-trayCoalesceTicker.C:
				if pendingTrayUpdate {
					handleTrayUpdate(mt.Prune(now))
					pendingTrayUpdate = false
				}

			case now := <-ticker.C:
				for _, exp := range mt.CheckCooldownExpirations(now) {
					label := exp.Category
					if label == "" {
						label = "global"
					}
					logger.Infof("[COOLDOWN] Cooldown expired [%s]: resumed monitoring (suppressed %d error(s) during %s cooldown)",
						label, exp.Suppressed, exp.Cooldown)
				}

				handleTrayUpdate(mt.Prune(now))
				pendingTrayUpdate = false
			}
		}
	}()

	// Starts main tray event loop (blocking)
	uiTray.Run(func() {
		cancel()
	})
}
