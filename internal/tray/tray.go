package tray

import (
	_ "embed"
	"fmt"
	"sync"

	"fyne.io/systray"
)

//go:embed assets/icon.png
var defaultIcon []byte

// Tray manages the system tray icon, tooltips, and context menu items.
type Tray struct {
	mu        sync.Mutex
	mStats    *systray.MenuItem
	mLatest   *systray.MenuItem
	mQuit     *systray.MenuItem
	onQuit    func()
	readyChan chan struct{}
}

// New initializes a Tray instance with a callback for user-triggered exit.
func New(onQuit func()) *Tray {
	return &Tray{
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
	systray.SetIcon(defaultIcon)
	systray.SetTitle("dmesg-watch")
	systray.SetTooltip("dmesg watch - Monitoring Kernel")

	t.mu.Lock()
	t.mStats = systray.AddMenuItem("Errors (Last 60s): 0", "Count of kernel errors in the rolling window")
	t.mStats.Disable()

	t.mLatest = systray.AddMenuItem("Latest: None", "Most recent kernel error snippet")
	t.mLatest.Disable()

	systray.AddSeparator()

	t.mQuit = systray.AddMenuItem("Quit dmesg watch", "Terminate dmesg-watch background monitor")
	t.mu.Unlock()

	close(t.readyChan)

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

// Update dynamically refreshes the menu items with updated error metrics.
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
}

// Run blocks and starts the tray event loop.
func (t *Tray) Run(onExit func()) {
	systray.Run(t.Init, onExit)
}

// Quit requests the systray event loop to terminate.
func (t *Tray) Quit() {
	systray.Quit()
}
