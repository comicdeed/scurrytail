package tracker

import (
	"testing"
	"time"

	"github.com/comicdeed/scurrytail/internal/config"
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

func TestMultiTracker_ClassifiedAndIgnored(t *testing.T) {
	cfg := config.Config{
		Threshold: 5,
		Window:    60 * time.Second,
		Cooldown:  45 * time.Second,
		Ignore: []string{
			"ACPI:.*Method _OSI failed",
		},
		Categories: map[string]config.CategoryConfig{
			"disk": {
				Threshold: 1,
				Window:    5 * time.Minute,
				Cooldown:  1 * time.Minute,
				Urgency:   "critical",
				Patterns: []string{
					"EXT4-fs error",
					"nvme.*status: 0x",
				},
			},
			"wifi": {
				Threshold: 3,
				Window:    2 * time.Minute,
				Cooldown:  30 * time.Second,
				Urgency:   "low",
				Patterns: []string{
					"iwlwifi.*",
				},
			},
		},
	}

	mt, err := NewMultiTracker(cfg)
	if err != nil {
		t.Fatalf("failed to create MultiTracker: %v", err)
	}

	now := time.Now()

	// 1. Ignored error: should be dropped immediately
	res := mt.Record("ACPI: Method _OSI failed on device", now)
	if !res.Ignored {
		t.Fatalf("expected error to be ignored")
	}
	prune := mt.Prune(now)
	if prune.TotalCount != 0 {
		t.Fatalf("expected TotalCount 0 after ignored error, got %d", prune.TotalCount)
	}

	// 2. Classified error (disk): threshold is 1, should alert immediately
	res = mt.Record("EXT4-fs error (device sda1): bad block", now)
	if res.Ignored {
		t.Fatalf("unexpected ignore for disk error")
	}
	if res.Category != "disk" {
		t.Fatalf("expected category 'disk', got %q", res.Category)
	}
	if !res.ShouldAlert {
		t.Fatalf("expected disk error to alert with threshold 1")
	}
	if res.Urgency != "critical" {
		t.Fatalf("expected critical urgency, got %s", res.Urgency)
	}
	if res.Count != 1 {
		t.Fatalf("unexpected disk count: got %d, expected 1", res.Count)
	}
	prune = mt.Prune(now)
	if prune.TotalCount != 1 || prune.CategoryCounts["disk"] != 1 {
		t.Fatalf("unexpected counts: total=%d, disk=%d", prune.TotalCount, prune.CategoryCounts["disk"])
	}

	// 3. Unclassified error: should route to unclassified (not disk)
	res = mt.Record("usb 1-1: device descriptor read/64, error -110", now)
	if res.Category != "" {
		t.Fatalf("expected empty category for unclassified error, got %q", res.Category)
	}
	if res.ShouldAlert {
		t.Fatalf("expected unclassified error not to alert (count 1 < threshold 5)")
	}
	prune = mt.Prune(now)
	if prune.TotalCount != 2 {
		t.Fatalf("expected TotalCount 2, got %d", prune.TotalCount)
	}
	// Disk count should still be 1, not 2
	if prune.CategoryCounts["disk"] != 1 {
		t.Fatalf("disk category count corrupted by unclassified error: got %d", prune.CategoryCounts["disk"])
	}
}

func TestMultiTracker_BoundedMemory(t *testing.T) {
	cfg := config.Config{
		Threshold: 5,
		Window:    60 * time.Second,
		Cooldown:  45 * time.Second,
	}

	mt, err := NewMultiTracker(cfg)
	if err != nil {
		t.Fatalf("failed to create MultiTracker: %v", err)
	}

	now := time.Now()
	// Simulate flood of 10,000 errors
	for i := 0; i < 10000; i++ {
		mt.Record("flood error", now)
	}

	prune := mt.Prune(now)
	// Bounded capacity clamps retained timestamps to maxCapacity (128 for threshold 5)
	if prune.TotalCount > 128 {
		t.Fatalf("expected bounded count <= 128, got %d", prune.TotalCount)
	}
}

func TestMultiTracker_PriorityLatching(t *testing.T) {
	cfg := config.Config{
		Threshold: 5,
		Window:    60 * time.Second,
		Cooldown:  45 * time.Second,
		Categories: map[string]config.CategoryConfig{
			"storage": {
				Threshold: 1,
				Window:    5 * time.Minute,
				Cooldown:  1 * time.Minute,
				Urgency:   "critical",
				Patterns:  []string{"EXT4-fs.*"},
			},
			"dock": {
				Threshold: 2,
				Window:    30 * time.Second,
				Cooldown:  10 * time.Second,
				Urgency:   "low",
				Patterns:  []string{"hub.*"},
			},
		},
	}

	mt, err := NewMultiTracker(cfg)
	if err != nil {
		t.Fatalf("failed to create MultiTracker: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// 1. Critical storage error occurs at t = 0
	mt.Record("EXT4-fs (sda1): I/O error while writing superblock", baseTime)
	prune := mt.Prune(baseTime)

	if !prune.AnyAlerting {
		t.Fatalf("expected AnyAlerting true for critical storage error")
	}
	if prune.AlertCategory != "storage" {
		t.Errorf("expected AlertCategory 'storage', got %q", prune.AlertCategory)
	}
	if prune.AlertUrgency != "critical" {
		t.Errorf("expected AlertUrgency 'critical', got %q", prune.AlertUrgency)
	}
	if prune.Latest != "EXT4-fs (sda1): I/O error while writing superblock" {
		t.Errorf("unexpected Latest: %s", prune.Latest)
	}

	// 2. Low-priority dock burst occurs at t = 10s (exceeds dock threshold 2)
	mt.Record("hub 7-1: hub_ext_port_status failed", baseTime.Add(10*time.Second))
	res := mt.Record("hub 7-1: hub_ext_port_status failed", baseTime.Add(11*time.Second))

	if !res.ShouldAlert {
		t.Fatalf("expected dock error to trigger ShouldAlert")
	}
	if res.SupersededBy != "storage" {
		t.Errorf("expected dock alert to be superseded by 'storage', got %q", res.SupersededBy)
	}
	if res.SupersededRank != "critical" {
		t.Errorf("expected SupersededRank 'critical', got %q", res.SupersededRank)
	}

	prune = mt.Prune(baseTime.Add(12 * time.Second))
	if !prune.AnyAlerting {
		t.Fatalf("expected AnyAlerting true")
	}
	// Crucial check: Storage is critical, dock is low. Storage must retain priority!
	if prune.AlertCategory != "storage" {
		t.Errorf("expected AlertCategory to stay 'storage', got %q", prune.AlertCategory)
	}
	if prune.AlertUrgency != "critical" {
		t.Errorf("expected AlertUrgency to stay 'critical', got %q", prune.AlertUrgency)
	}
	if prune.Latest != "EXT4-fs (sda1): I/O error while writing superblock" {
		t.Errorf("expected Latest to stay latched to critical error, got %q", prune.Latest)
	}

	// 3. At t = 45s: dock 30s window has expired, but storage 5m is still within window
	prune = mt.Prune(baseTime.Add(45 * time.Second))
	if !prune.AnyAlerting {
		t.Fatalf("expected AnyAlerting to remain true for active storage window")
	}
	if prune.CategoryCounts["dock"] != 0 {
		t.Errorf("expected dock count 0 after window expiry, got %d", prune.CategoryCounts["dock"])
	}
	if prune.CategoryCounts["storage"] != 1 {
		t.Errorf("expected storage count 1, got %d", prune.CategoryCounts["storage"])
	}

	// 4. At t = 6m: storage 5m window has also expired
	prune = mt.Prune(baseTime.Add(6 * time.Minute))
	if prune.AnyAlerting {
		t.Errorf("expected AnyAlerting false after all windows expired")
	}
	if prune.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", prune.TotalCount)
	}
}

func TestMultiTracker_CooldownExpirations(t *testing.T) {
	cfg := config.Config{
		Threshold: 2,
		Window:    60 * time.Second,
		Cooldown:  10 * time.Second,
		Categories: map[string]config.CategoryConfig{
			"wifi": {
				Threshold: 2,
				Window:    30 * time.Second,
				Cooldown:  10 * time.Second,
				Urgency:   "normal",
				Patterns:  []string{"iwlwifi.*"},
			},
		},
	}

	mt, err := NewMultiTracker(cfg)
	if err != nil {
		t.Fatalf("failed to create MultiTracker: %v", err)
	}

	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// First error (below threshold)
	r1 := mt.Record("iwlwifi: link down", baseTime)
	if r1.ShouldAlert {
		t.Errorf("did not expect alert on 1st error")
	}

	// Second error (reaches threshold -> alerts)
	r2 := mt.Record("iwlwifi: link down", baseTime.Add(1*time.Second))
	if !r2.ShouldAlert {
		t.Fatalf("expected alert on 2nd error")
	}

	// 5 subsequent errors within the 10s cooldown window (should be debounced/suppressed)
	for i := 2; i <= 6; i++ {
		r := mt.Record("iwlwifi: link down", baseTime.Add(time.Duration(i)*time.Second))
		if r.ShouldAlert {
			t.Errorf("error %d should have been debounced by cooldown", i)
		}
	}

	// Check at t = 5s (cooldown still active) -> no expirations
	expMid := mt.CheckCooldownExpirations(baseTime.Add(5 * time.Second))
	if len(expMid) != 0 {
		t.Errorf("expected 0 expirations during cooldown, got %d", len(expMid))
	}

	// Check at t = 12s (cooldown of 10s elapsed since t = 1s) -> 1 expiration with 5 suppressed errors
	expEnd := mt.CheckCooldownExpirations(baseTime.Add(12 * time.Second))
	if len(expEnd) != 1 {
		t.Fatalf("expected 1 cooldown expiration, got %d", len(expEnd))
	}
	if expEnd[0].Category != "wifi" {
		t.Errorf("expected category 'wifi', got %q", expEnd[0].Category)
	}
	if expEnd[0].Suppressed != 5 {
		t.Errorf("expected 5 suppressed errors, got %d", expEnd[0].Suppressed)
	}

	// Immediate subsequent check should return 0 (counter was reset)
	expAfter := mt.CheckCooldownExpirations(baseTime.Add(13 * time.Second))
	if len(expAfter) != 0 {
		t.Errorf("expected 0 expirations after reset, got %d", len(expAfter))
	}
}

func TestMultiTracker_Categories(t *testing.T) {
	cfg := config.Config{
		Window:    10 * time.Second,
		Threshold: 2,
		Cooldown:  10 * time.Second,
		Categories: map[string]config.CategoryConfig{
			"storage": {
				Patterns: []string{"nvme"},
			},
			"ignored_cat": {
				Ignore:   true,
				Patterns: []string{"benign"},
			},
		},
		Patterns: []config.PatternRule{
			{Pattern: "amdgpu", Category: "gpu"},
		},
	}

	mt, err := NewMultiTracker(cfg)
	if err != nil {
		t.Fatalf("NewMultiTracker failed: %v", err)
	}

	cats := mt.Categories()
	expected := []string{"gpu", "storage"}
	if len(cats) != len(expected) {
		t.Fatalf("expected categories %v, got %v", expected, cats)
	}
	for i := range expected {
		if cats[i] != expected[i] {
			t.Errorf("category mismatch at index %d: expected %s, got %s", i, expected[i], cats[i])
		}
	}
}
