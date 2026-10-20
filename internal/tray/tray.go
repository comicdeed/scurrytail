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
	isWarningState bool
	mStats         *systray.MenuItem
	mLatest        *systray.MenuItem
	mAbout         *systray.MenuItem
	mQuit          *systray.MenuItem
	onAbout        func()
	onQuit         func()
	readyChan      chan struct{}
}

// New initializes a Tray instance with callbacks for About and user-triggered exit.
func New(version string, onAbout func(), onQuit func()) *Tray {
	return &Tray{
		version:   version,
		onAbout:   onAbout,
		onQuit:    onQuit,
		readyChan: make(chan struct{}),
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

	t.mu.Lock()
	t.isWarningState = false

	// Unclickable title header showing application and version
	mTitle := systray.AddMenuItem(fmt.Sprintf("dmesg-watch %s", t.version), "Application version")
	mTitle.Disable()

	systray.AddSeparator()

	t.mStats = systray.AddMenuItem("Errors (Last 60s): 0", "Count of kernel errors in the rolling window")
	t.mStats.Disable()

	t.mLatest = systray.AddMenuItem("Latest: None", "Most recent kernel error snippet")
	t.mLatest.Disable()

	systray.AddSeparator()

	t.mAbout = systray.AddMenuItem("About dmesg-watch", "Show version and project information")
	t.mQuit = systray.AddMenuItem("Quit dmesg watch", "Terminate dmesg-watch background monitor")
	t.mu.Unlock()

	close(t.readyChan)

	go func() {
		for range t.mAbout.ClickedCh {
			if t.onAbout != nil {
				t.onAbout()
			}
		}
	}()

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

// Update dynamically refreshes the menu items and switches the tray icon
// between the calm normal icon (0 errors) and the amber warning icon (> 0 errors).
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

	// Dynamic icon switching: switch to warning when errors occur, back to calm when clear
	shouldWarn := count > 0
	if shouldWarn != t.isWarningState {
		t.isWarningState = shouldWarn
		if shouldWarn {
			systray.SetIcon(iconWarning)
		} else {
			systray.SetIcon(iconNormal)
		}
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
