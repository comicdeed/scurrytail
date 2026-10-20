package tracker

import (
	"testing"
	"time"
)

func TestTracker_Prune(t *testing.T) {
	cfg := Config{
		Window:    60 * time.Second,
		Threshold: 3,
		Cooldown:  45 * time.Second,
	}
	tr := New(cfg)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tr.Record("err 1", baseTime)
	tr.Record("err 2", baseTime.Add(10*time.Second))
	tr.Record("err 3", baseTime.Add(20*time.Second))

	count, latest := tr.Prune(baseTime.Add(25 * time.Second))
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
	if latest != "err 3" {
		t.Fatalf("expected latest 'err 3', got %q", latest)
	}

	// Advance time past 60s from baseTime: err 1 (at 0s) should expire
	count, _ = tr.Prune(baseTime.Add(65 * time.Second))
	if count != 2 {
		t.Fatalf("expected count 2 after pruning err 1, got %d", count)
	}

	// Advance past 85s: all should expire
	count, _ = tr.Prune(baseTime.Add(90 * time.Second))
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestTracker_ShouldAlert_And_Cooldown(t *testing.T) {
	cfg := Config{
		Window:    60 * time.Second,
		Threshold: 3,
		Cooldown:  45 * time.Second,
	}
	tr := New(cfg)

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// Less than threshold (2 < 3)
	tr.Record("err 1", baseTime)
	tr.Record("err 2", baseTime.Add(5*time.Second))
	shouldAlert, count, _ := tr.ShouldAlert(baseTime.Add(6 * time.Second))
	if shouldAlert {
		t.Fatalf("expected false alert with count %d < threshold %d", count, cfg.Threshold)
	}

	// Hit threshold (3rd error)
	tr.Record("err 3", baseTime.Add(10*time.Second))
	shouldAlert, count, latest := tr.ShouldAlert(baseTime.Add(10 * time.Second))
	if !shouldAlert {
		t.Fatalf("expected alert when hitting threshold, got false (count %d)", count)
	}
	if count != 3 || latest != "err 3" {
		t.Fatalf("unexpected stats: count=%d, latest=%s", count, latest)
	}

	// 4th error inside cooldown period (10s + 20s = 30s, cooldown is 45s)
	tr.Record("err 4", baseTime.Add(30*time.Second))
	shouldAlert, count, _ = tr.ShouldAlert(baseTime.Add(30 * time.Second))
	if shouldAlert {
		t.Fatalf("expected alert to be suppressed by cooldown (count %d)", count)
	}

	// 5th error after cooldown expired (10s + 50s = 60s, which is 50s after alert at 10s > 45s cooldown)
	// But err 1 was at 0s, at 60s it is still within window
	tr.Record("err 5", baseTime.Add(60*time.Second))
	shouldAlert, count, _ = tr.ShouldAlert(baseTime.Add(60 * time.Second))
	if !shouldAlert {
		t.Fatalf("expected alert to fire after cooldown elapsed (count %d)", count)
	}
}
