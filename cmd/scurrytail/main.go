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
	version = "v0.3.0-beta.14"
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

	// OS signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	mt, err := tracker.NewMultiTracker(cfg)
	if err != nil {
		logger.Fatalf("Fatal: %v", err)
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
		currentStatus := tray.StatusPassive
		var wakeCategory, wakeLatest string

		evaluateState := func(prune tracker.PruneResult) {
			var targetStatus tray.Status
			switch {
			case prune.TotalCount == 0:
				targetStatus = tray.StatusPassive
			case prune.AnyAlerting || (cfg.Threshold > 0 && prune.TotalCount >= cfg.Threshold):
				targetStatus = tray.StatusNeedsAttention
			default:
				targetStatus = tray.StatusActive
			}

			if targetStatus != currentStatus {
				switch targetStatus {
				case tray.StatusPassive:
					logger.Infof("[STATE] %s -> Passive (rolling window clean, 0 errors)", currentStatus)
					wakeCategory = ""
					wakeLatest = ""
				case tray.StatusActive:
					if currentStatus == tray.StatusPassive {
						catLabel := wakeCategory
						if catLabel == "" {
							catLabel = "unclassified"
						}
						truncated := wakeLatest
						if len(truncated) > 60 {
							truncated = truncated[:57] + "..."
						}
						logger.Infof("[STATE] %s -> Active (1 active error in %s window, category: %s, latest: %q)",
							currentStatus, cfg.Window, catLabel, truncated)
					} else {
						logger.Infof("[STATE] %s -> Active (%d active error(s) in %s window)", currentStatus, prune.TotalCount, cfg.Window)
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
						currentStatus, prune.TotalCount, cfg.Window, cat, urgency)
				}
				currentStatus = targetStatus
			}
		}

		for {
			select {
			case <-ctx.Done():
				return

			case sig := <-sigCh:
				logger.Infof("Received signal %s, shutting down...", sig)
				cancel()
				uiTray.Quit()
				return

			case err, ok := <-errCh:
				if ok && err != nil {
					logger.Errorf("Journal stream error: %v", err)
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

				if currentStatus == tray.StatusPassive {
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
					prune := mt.Prune(now)
					evaluateState(prune)
					uiTray.UpdateWithCategories(prune.TotalCount, prune.Latest, prune.CategoryCounts, prune.AnyAlerting, prune.AlertCategory, prune.AlertUrgency)
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

				prune := mt.Prune(now)
				evaluateState(prune)
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
