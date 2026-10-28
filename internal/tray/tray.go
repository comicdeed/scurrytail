package tray

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"sync"

	"fyne.io/systray"
)

//go:embed assets/icon_normal.png
var iconNormal []byte

//go:embed assets/icon_warning.png
var iconWarning []byte

// Status represents the operational alerting state of ScurryTail.
type Status int

const (
	StatusPassive Status = iota
	StatusActive
	StatusNeedsAttention
)

func (s Status) String() string {
	switch s {
	case StatusPassive:
		return "Passive"
	case StatusActive:
		return "Active"
	case StatusNeedsAttention:
		return "NeedsAttention"
	default:
		return "Unknown"
	}
}

// Tray manages the system tray icon, tooltips, and context menu items.
type Tray struct {
	mu            sync.Mutex
	version       string
	backend       string
	threshold     int
	categoryNames []string
	status        Status
	mStats        *systray.MenuItem
	mLatest       *systray.MenuItem
	mCategories   *systray.MenuItem
	mCatItems     map[string]*systray.MenuItem
	mViewLogs     *systray.MenuItem
	mAbout        *systray.MenuItem
	mWebsite      *systray.MenuItem
	mQuit         *systray.MenuItem
	onViewLogs    func()
	onWebsite     func()
	onQuit        func()
	readyChan     chan struct{}
}

// New initializes a Tray instance with callbacks for log viewing, Website URL opening, and user-triggered exit.
func New(version string, backend string, threshold int, categoryNames []string, onViewLogs func(), onWebsite func(), onQuit func()) *Tray {
	sortedCats := make([]string, len(categoryNames))
	copy(sortedCats, categoryNames)
	sort.Strings(sortedCats)

	return &Tray{
		version:       version,
		backend:       backend,
		threshold:     threshold,
		categoryNames: sortedCats,
		mCatItems:     make(map[string]*systray.MenuItem),
		onViewLogs:    onViewLogs,
		onWebsite:     onWebsite,
		onQuit:        onQuit,
		readyChan:     make(chan struct{}),
	}
}

// Ready returns a channel that closes when the tray interface has been initialized.
func (t *Tray) Ready() <-chan struct{} {
	return t.readyChan
}

// Init sets up the StatusNotifierItem properties and context menu structure.
// This must be called inside the systray.Run onReady callback.
func (t *Tray) Init() {
	systray.SetIcon(iconNormal)
	systray.SetTitle("ScurryTail")
	systray.SetTooltip("ScurryTail\nNimble kernel log watcher & error monitor")
	systray.Hide()

	t.mu.Lock()
	t.status = StatusPassive

	t.mStats = systray.AddMenuItem("Errors (Recent): 0", "Count of kernel errors in the rolling window")
	t.mStats.Disable()

	t.mLatest = systray.AddMenuItem("Latest: None", "Most recent kernel error snippet")
	t.mLatest.Disable()

	// Pre-allocate category menu items if categories are configured
	if len(t.categoryNames) > 0 {
		t.mCategories = systray.AddMenuItem("Categories", "Error breakdown by category")
		t.mCategories.Hide()

		for _, cat := range t.categoryNames {
			item := t.mCategories.AddSubMenuItem(fmt.Sprintf("%s: 0", cat), fmt.Sprintf("Active %s errors", cat))
			item.Disable()
			item.Hide()
			t.mCatItems[cat] = item
		}
	}

	systray.AddSeparator()

	viewLabel := "View Kernel Logs"
	if t.backend != "" {
		viewLabel = fmt.Sprintf("View Kernel Logs (%s)", t.backend)
	}
	tooltip := fmt.Sprintf("Open kernel logs in default terminal via %s", t.backend)
	if t.backend == "journalctl" {
		tooltip = "Open correlated kernel errors and audit events in default terminal via journalctl"
	}
	t.mViewLogs = systray.AddMenuItem(viewLabel, tooltip)
	if t.onViewLogs == nil {
		t.mViewLogs.Disable()
	}

	systray.AddSeparator()

	t.mAbout = systray.AddMenuItem("About", "Show version and project information")
	mVersion := t.mAbout.AddSubMenuItem(fmt.Sprintf("ScurryTail %s", t.version), "Application version")
	mVersion.Disable()

	t.mAbout.AddSeparator()

	t.mWebsite = t.mAbout.AddSubMenuItem("Website", "Open project repository in default browser")
	if t.onWebsite == nil {
		t.mWebsite.Disable()
	}

	t.mQuit = systray.AddMenuItem("Quit", "Terminate background monitor")
	t.mu.Unlock()

	close(t.readyChan)

	if t.onViewLogs != nil {
		go func() {
			for range t.mViewLogs.ClickedCh {
				t.onViewLogs()
			}
		}()
	}

	if t.onWebsite != nil {
		go func() {
			for range t.mWebsite.ClickedCh {
				t.onWebsite()
			}
		}()
	}

	go func() {
		for range t.mQuit.ClickedCh {
			if t.onQuit != nil {
				t.onQuit()
			}
			systray.Quit()
			return
		}
	}()
}

// setStatusLocked handles transitions between SNI statuses and manages icons.
// Caller must hold t.mu.
func (t *Tray) setStatusLocked(target Status) {
	if t.status == target {
		return
	}

	switch target {
	case StatusPassive:
		systray.Hide()
		systray.SetIcon(iconNormal)

	case StatusActive:
		if t.status == StatusPassive {
			systray.SetIcon(iconWarning)
		}
		// Calling Show() sets SNI Status to Active and clears NeedsAttention in fyne-io-systray
		systray.Show()

	case StatusNeedsAttention:
		if t.status == StatusPassive {
			systray.SetIcon(iconWarning)
			systray.Show()
		}
		systray.RequestAttention()
	}

	t.status = target
}

// RequestAttention asks the user to look at the systray icon (NeedsAttention status).
// If the icon was hidden, it is displayed in warning state first.
// Always emits RequestAttention to handle cases where the user cleared attention via click.
func (t *Tray) RequestAttention() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == StatusPassive {
		systray.SetIcon(iconWarning)
		systray.Show()
	}
	systray.RequestAttention()
	t.status = StatusNeedsAttention
}

// Update dynamically refreshes standard metrics (backward-compatibility wrapper).
func (t *Tray) Update(count int, latest string) {
	t.UpdateWithCategories(count, latest, nil, false, "", "")
}

// UpdateWithCategories dynamically refreshes menu items, category breakdowns,
// dynamic priority tooltips, and tray visibility without reallocating D-Bus menu items.
func (t *Tray) UpdateWithCategories(count int, latest string, categoryCounts map[string]int, anyAlerting bool, alertCat string, alertUrgency string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.mStats == nil || t.mLatest == nil {
		return
	}

	truncated := latest
	if runes := []rune(truncated); len(runes) > 40 {
		truncated = string(runes[:37]) + "..."
	}

	t.mStats.SetTitle(fmt.Sprintf("Errors (Recent): %d", count))
	t.mLatest.SetTitle(fmt.Sprintf("Latest: %s", truncated))

	// Refresh category items in place
	if t.mCategories != nil {
		var hasActiveCat bool
		for cat, item := range t.mCatItems {
			cnt := categoryCounts[cat]
			if cnt > 0 {
				item.SetTitle(fmt.Sprintf("%s: %d", cat, cnt))
				item.Show()
				hasActiveCat = true
			} else {
				item.Hide()
			}
		}

		if hasActiveCat {
			t.mCategories.Show()
		} else {
			t.mCategories.Hide()
		}
	}

	// Update dynamic tooltip with active priority and category
	switch {
	case anyAlerting:
		label := alertCat
		if label == "" {
			label = "global"
		}
		urgencyLabel := strings.ToUpper(alertUrgency)
		if urgencyLabel == "" {
			urgencyLabel = "ALERT"
		}
		systray.SetTooltip(fmt.Sprintf("ScurryTail: %s [%s]\n%d active errors in rolling window\nLatest: %s",
			urgencyLabel, label, count, truncated))

	case count > 0:
		systray.SetTooltip(fmt.Sprintf("ScurryTail: ACTIVE\n%d active errors in rolling window\nLatest: %s",
			count, truncated))

	default:
		systray.SetTooltip("ScurryTail\nNimble kernel log watcher & error monitor")
	}

	switch {
	case count == 0:
		t.setStatusLocked(StatusPassive)

	case anyAlerting || (t.threshold > 0 && count >= t.threshold):
		t.setStatusLocked(StatusNeedsAttention)

	default:
		t.setStatusLocked(StatusActive)
	}
}

// Run blocks and starts the tray event loop.
func (t *Tray) Run(onExit func()) {
	systray.Run(t.Init, onExit)
}

// Quit requests the systray event loop to terminate.
func (t *Tray) Quit() {
	systray.Quit()
}
