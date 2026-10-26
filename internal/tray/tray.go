package tray

import (
	_ "embed"
	"fmt"
	"sync"

	"fyne.io/systray"
)

//go:embed assets/icon_normal.png
var iconNormal []byte

//go:embed assets/icon_warning.png
var iconWarning []byte

// Tray manages the system tray icon, tooltips, and context menu items.
type Tray struct {
	mu             sync.Mutex
	version        string
	backend        string
	threshold      int
	visible        bool
	attention      bool
	isWarningState bool
	mStats         *systray.MenuItem
	mLatest        *systray.MenuItem
	mViewLogs      *systray.MenuItem
	mAbout         *systray.MenuItem
	mWebsite       *systray.MenuItem
	mQuit          *systray.MenuItem
	onViewLogs     func()
	onWebsite      func()
	onQuit         func()
	readyChan      chan struct{}
}

// New initializes a Tray instance with callbacks for log viewing, Website URL opening, and user-triggered exit.
func New(version string, backend string, threshold int, onViewLogs func(), onWebsite func(), onQuit func()) *Tray {
	return &Tray{
		version:    version,
		backend:    backend,
		threshold:  threshold,
		onViewLogs: onViewLogs,
		onWebsite:  onWebsite,
		onQuit:     onQuit,
		readyChan:  make(chan struct{}),
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
	systray.SetTitle("dmesg watch")
	systray.SetTooltip("dmesg watch\nMonitors kernel error logs in real time")
	systray.Hide()

	t.mu.Lock()
	t.isWarningState = false
	t.visible = false
	t.attention = false

	t.mStats = systray.AddMenuItem("Errors (Last 60s): 0", "Count of kernel errors in the rolling window")
	t.mStats.Disable()

	t.mLatest = systray.AddMenuItem("Latest: None", "Most recent kernel error snippet")
	t.mLatest.Disable()

	systray.AddSeparator()

	viewLabel := "View Kernel Logs"
	if t.backend != "" {
		viewLabel = fmt.Sprintf("View Kernel Logs (%s)", t.backend)
	}
	t.mViewLogs = systray.AddMenuItem(viewLabel, fmt.Sprintf("Open kernel logs in default terminal via %s", t.backend))
	if t.onViewLogs == nil {
		t.mViewLogs.Disable()
	}

	systray.AddSeparator()

	t.mAbout = systray.AddMenuItem("About", "Show version and project information")
	mVersion := t.mAbout.AddSubMenuItem(fmt.Sprintf("dmesg-watch %s", t.version), "Application version")
	mVersion.Disable()

	mDesc := t.mAbout.AddSubMenuItem("Real-time kernel error monitor", "Application summary")
	mDesc.Disable()

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

// Show makes the tray icon visible in the notification area.
func (t *Tray) Show() {
	t.mu.Lock()
	defer t.mu.Unlock()
	systray.Show()
	t.visible = true
}

// Hide hides the tray icon from the notification area.
func (t *Tray) Hide() {
	t.mu.Lock()
	defer t.mu.Unlock()
	systray.Hide()
	t.visible = false
	t.attention = false
}

// RequestAttention asks the user to look at the systray icon (NeedsAttention status).
func (t *Tray) RequestAttention() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.visible {
		systray.Show()
		t.visible = true
	}
	systray.RequestAttention()
	t.attention = true
}

// Update dynamically refreshes the menu items and controls tray visibility,
// warning icons, and attention status based on error count and threshold.
func (t *Tray) Update(count int, latest string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.mStats == nil || t.mLatest == nil {
		return
	}

	truncated := latest
	if len(truncated) > 40 {
		truncated = truncated[:37] + "..."
	}

	t.mStats.SetTitle(fmt.Sprintf("Errors (Last 60s): %d", count))
	t.mLatest.SetTitle(fmt.Sprintf("Latest: %s", truncated))

	// When there are no errors, hide the icon
	if count == 0 {
		if t.visible {
			systray.Hide()
			t.visible = false
			t.attention = false
			t.isWarningState = false
			systray.SetIcon(iconNormal)
		}
		return
	}

	// count > 0: show the icon if currently hidden
	if !t.visible {
		systray.Show()
		t.visible = true
	}

	if !t.isWarningState {
		systray.SetIcon(iconWarning)
		t.isWarningState = true
	}

	// Request attention when errors go beyond/reach threshold
	if t.threshold > 0 && count >= t.threshold {
		if !t.attention {
			systray.RequestAttention()
			t.attention = true
		}
	} else if t.attention {
		t.attention = false
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
