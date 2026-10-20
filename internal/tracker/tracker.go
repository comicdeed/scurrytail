package tracker

import (
	"sync"
	"time"
)

// Config holds operational parameters for the sliding-window error tracker.
type Config struct {
	Window    time.Duration // Duration to retain error timestamps (default: 60s)
	Threshold int           // Number of errors within Window to trigger an alert (default: 5)
	Cooldown  time.Duration // Minimum duration between notifications (default: 45s)
}

// DefaultConfig provides recommended defaults matching the original behavior.
func DefaultConfig() Config {
	return Config{
		Window:    60 * time.Second,
		Threshold: 5,
		Cooldown:  45 * time.Second,
	}
}

// Tracker tracks error events across a rolling time window and evaluates alert thresholds.
type Tracker struct {
	mu               sync.Mutex
	cfg              Config
	timestamps       []time.Time
	latestMsg        string
	lastNotification time.Time
}

// New creates a new Tracker with the specified configuration.
func New(cfg Config) *Tracker {
	if cfg.Window <= 0 {
		cfg.Window = 60 * time.Second
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 5
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 45 * time.Second
	}

	return &Tracker{
		cfg:       cfg,
		latestMsg: "None",
	}
}

// Record registers a new error message and its occurrence timestamp.
func (t *Tracker) Record(msg string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.timestamps = append(t.timestamps, now)
	t.latestMsg = msg
	t.pruneLocked(now)
}

// Prune evicts timestamps older than cfg.Window and returns current count and latest message.
func (t *Tracker) Prune(now time.Time) (int, string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked(now)
	return len(t.timestamps), t.latestMsg
}

// pruneLocked evicts expired timestamps. Caller must hold t.mu.
func (t *Tracker) pruneLocked(now time.Time) {
	cutoff := now.Add(-t.cfg.Window)
	i := 0
	for ; i < len(t.timestamps); i++ {
		if !t.timestamps[i].Before(cutoff) {
			break
		}
	}
	if i > 0 {
		t.timestamps = t.timestamps[i:]
	}
}

// ShouldAlert checks if the current error count satisfies the alert threshold
// and the cooldown period has elapsed. If true, updates lastNotification and returns count and latest msg.
func (t *Tracker) ShouldAlert(now time.Time) (bool, int, string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked(now)
	count := len(t.timestamps)
	if count < t.cfg.Threshold {
		return false, count, t.latestMsg
	}

	if !t.lastNotification.IsZero() && now.Sub(t.lastNotification) < t.cfg.Cooldown {
		return false, count, t.latestMsg
	}

	t.lastNotification = now
	return true, count, t.latestMsg
}

// Stats returns the current count and latest message without pruning.
func (t *Tracker) Stats() (int, string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.timestamps), t.latestMsg
}
