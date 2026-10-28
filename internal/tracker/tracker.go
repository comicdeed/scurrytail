package tracker

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/comicdeed/scurrytail/internal/config"
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
	maxCapacity      int
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

	maxCap := cfg.Threshold * 2
	if maxCap < 128 {
		maxCap = 128
	}

	return &Tracker{
		cfg:         cfg,
		maxCapacity: maxCap,
		latestMsg:   "None",
	}
}

// Record registers a new error message and its occurrence timestamp.
func (t *Tracker) Record(msg string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked(now)

	if len(t.timestamps) >= t.maxCapacity {
		// Bounded circular shift to prevent unbounded slice allocation during floods
		copy(t.timestamps, t.timestamps[1:])
		t.timestamps[len(t.timestamps)-1] = now
	} else {
		t.timestamps = append(t.timestamps, now)
	}
	t.latestMsg = msg
}

// Prune evicts timestamps older than cfg.Window and returns current count and latest message.
func (t *Tracker) Prune(now time.Time) (int, string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked(now)
	return len(t.timestamps), t.latestMsg
}

// pruneLocked evicts expired timestamps and copies retained elements to avoid capacity leaks. Caller must hold t.mu.
func (t *Tracker) pruneLocked(now time.Time) {
	cutoff := now.Add(-t.cfg.Window)
	i := 0
	for ; i < len(t.timestamps); i++ {
		if !t.timestamps[i].Before(cutoff) {
			break
		}
	}
	if i > 0 {
		copy(t.timestamps, t.timestamps[i:])
		t.timestamps = t.timestamps[:len(t.timestamps)-i]
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

// CompiledRule encapsulates a compiled pattern with a two-tier literal/regex matching strategy.
type CompiledRule struct {
	Pattern   string
	IsLiteral bool
	Regex     *regexp.Regexp
	Category  string // Empty string denotes unclassified fallback
	IsIgnore  bool
}

// Matches checks if the message satisfies the rule using literal fast-path or regex.
func (r *CompiledRule) Matches(msg string) bool {
	if r.IsLiteral {
		return strings.Contains(msg, r.Pattern)
	}
	return r.Regex.MatchString(msg)
}

// RecordResult summarizes the outcome of processing a log message through MultiTracker.
type RecordResult struct {
	Ignored        bool
	Category       string // Empty string denotes unclassified
	ShouldAlert    bool
	Count          int
	Threshold      int
	Window         time.Duration
	Urgency        string
	Latest         string
	TotalCount     int
	CategoryCounts map[string]int
}

// UrgencyRank represents the priority order for desktop notifications and tray status latching.
type UrgencyRank int

const (
	UrgencyRankNone UrgencyRank = iota
	UrgencyRankLow
	UrgencyRankNormal
	UrgencyRankCritical
)

// ParseUrgencyRank maps an urgency string to its numerical priority rank.
func ParseUrgencyRank(u string) UrgencyRank {
	switch strings.ToLower(strings.TrimSpace(u)) {
	case "critical":
		return UrgencyRankCritical
	case "low":
		return UrgencyRankLow
	default:
		return UrgencyRankNormal
	}
}

func (u UrgencyRank) String() string {
	switch u {
	case UrgencyRankCritical:
		return "critical"
	case UrgencyRankNormal:
		return "normal"
	case UrgencyRankLow:
		return "low"
	default:
		return ""
	}
}

// PruneResult contains aggregate statistics across all trackers after a periodic prune.
type PruneResult struct {
	TotalCount     int
	Latest         string
	CategoryCounts map[string]int
	AnyAlerting    bool
	AlertCategory  string
	AlertUrgency   string
}

// MultiTracker coordinates category classification, unclassified tracking, and benign error suppression.
type MultiTracker struct {
	mu              sync.RWMutex
	rules           []CompiledRule
	unclassified    *Tracker
	categories      map[string]*Tracker
	urgencies       map[string]string
	windows         map[string]time.Duration
	thresholds      map[string]int
	globalWindow    time.Duration
	globalThresh    int
	latestMsg       string
	activeAlertMsg  string
	activeAlertCat  string
	activeAlertRank UrgencyRank
}

// NewMultiTracker builds a MultiTracker from the resolved application configuration.
func NewMultiTracker(cfg config.Config) (*MultiTracker, error) {
	unclassTracker := New(Config{
		Window:    cfg.Window,
		Threshold: cfg.Threshold,
		Cooldown:  cfg.Cooldown,
	})

	var rules []CompiledRule

	// 1. Top-level ignore patterns (evaluated first)
	for _, pat := range cfg.Ignore {
		rule, err := compileRule(pat, "", true)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	// Sort category names alphabetically to guarantee deterministic rule evaluation
	catNames := make([]string, 0, len(cfg.Categories))
	for name := range cfg.Categories {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	// 2. Category-level ignore: true rules (evaluated early)
	categoryTrackers := make(map[string]*Tracker)
	urgencies := make(map[string]string)
	windows := make(map[string]time.Duration)
	thresholds := make(map[string]int)

	for _, name := range catNames {
		cat := cfg.Categories[name]
		if cat.Ignore {
			for _, pat := range cat.Patterns {
				rule, err := compileRule(pat, name, true)
				if err != nil {
					return nil, err
				}
				rules = append(rules, rule)
			}
		}
	}

	// 3. Flat pattern rules in explicit sequence
	for _, r := range cfg.Patterns {
		// Auto-register undeclared categories so they inherit global defaults
		if r.Category != "" && !r.Ignore && categoryTrackers[r.Category] == nil {
			categoryTrackers[r.Category] = New(Config{
				Window:    cfg.Window,
				Threshold: cfg.Threshold,
				Cooldown:  cfg.Cooldown,
			})
			urgencies[r.Category] = "normal"
			windows[r.Category] = cfg.Window
			thresholds[r.Category] = cfg.Threshold
		}

		rule, err := compileRule(r.Pattern, r.Category, r.Ignore)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	// 4. Classified category patterns (non-ignored)
	for _, name := range catNames {
		cat := cfg.Categories[name]
		if cat.Ignore {
			continue
		}

		categoryTrackers[name] = New(Config{
			Window:    cat.Window,
			Threshold: cat.Threshold,
			Cooldown:  cat.Cooldown,
		})

		urgencies[name] = cat.Urgency
		windows[name] = cat.Window
		thresholds[name] = cat.Threshold

		for _, pat := range cat.Patterns {
			rule, err := compileRule(pat, name, false)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rule)
		}
	}

	return &MultiTracker{
		rules:        rules,
		unclassified: unclassTracker,
		categories:   categoryTrackers,
		urgencies:    urgencies,
		windows:      windows,
		thresholds:   thresholds,
		globalWindow: cfg.Window,
		globalThresh: cfg.Threshold,
		latestMsg:    "None",
	}, nil
}

func compileRule(pattern, category string, isIgnore bool) (CompiledRule, error) {
	rule := CompiledRule{
		Pattern:  pattern,
		Category: category,
		IsIgnore: isIgnore,
	}

	// If pattern contains no regex meta-characters, use zero-allocation literal fast-path
	if !strings.ContainsAny(pattern, `.*+?^${}()|[]\`) {
		rule.IsLiteral = true
		return rule, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return rule, err
	}
	rule.Regex = re
	return rule, nil
}

// Record evaluates a message against patterns (lock-free) and atomically records the event.
func (m *MultiTracker) Record(msg string, now time.Time) RecordResult {
	// Lock-free pattern matching pass
	var matchedRule *CompiledRule
	for i := range m.rules {
		if m.rules[i].Matches(msg) {
			matchedRule = &m.rules[i]
			break
		}
	}

	// If benign ignore matched: drop immediately with 0 locks acquired
	if matchedRule != nil && matchedRule.IsIgnore {
		return RecordResult{Ignored: true}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.latestMsg = msg

	if matchedRule != nil && matchedRule.Category != "" {
		catName := matchedRule.Category
		tr, exists := m.categories[catName]
		if exists {
			tr.Record(msg, now)
			shouldAlert, count, latest := tr.ShouldAlert(now)

			if shouldAlert {
				rank := ParseUrgencyRank(m.urgencies[catName])
				if rank >= m.activeAlertRank {
					m.activeAlertRank = rank
					m.activeAlertCat = catName
					m.activeAlertMsg = msg
				}
			}

			return RecordResult{
				Category:    catName,
				ShouldAlert: shouldAlert,
				Count:       count,
				Threshold:   m.thresholds[catName],
				Window:      m.windows[catName],
				Urgency:     m.urgencies[catName],
				Latest:      latest,
			}
		}
	}

	// Unclassified fallback
	m.unclassified.Record(msg, now)
	shouldAlert, count, latest := m.unclassified.ShouldAlert(now)

	if shouldAlert {
		rank := UrgencyRankNormal
		if rank >= m.activeAlertRank {
			m.activeAlertRank = rank
			m.activeAlertCat = ""
			m.activeAlertMsg = msg
		}
	}

	return RecordResult{
		Category:    "",
		ShouldAlert: shouldAlert,
		Count:       count,
		Threshold:   m.globalThresh,
		Window:      m.globalWindow,
		Urgency:     "normal",
		Latest:      latest,
	}
}

// Prune evicts expired timestamps across all trackers and returns an aggregate status snapshot.
func (m *MultiTracker) Prune(now time.Time) PruneResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	tot, catCounts, anyAlert, alertCat, alertUrgency := m.snapshotLocked(now)

	displayLatest := m.latestMsg
	if anyAlert && m.activeAlertMsg != "" {
		displayLatest = m.activeAlertMsg
	}

	return PruneResult{
		TotalCount:     tot,
		Latest:         displayLatest,
		CategoryCounts: catCounts,
		AnyAlerting:    anyAlert,
		AlertCategory:  alertCat,
		AlertUrgency:   alertUrgency,
	}
}

// snapshotLocked gathers active counts across all trackers and determines the highest active alert. Caller must hold m.mu.
func (m *MultiTracker) snapshotLocked(now time.Time) (int, map[string]int, bool, string, string) {
	unclassCount, unclassLatest := m.unclassified.Prune(now)
	total := unclassCount
	unclassAlerting := unclassCount >= m.globalThresh

	catCounts := make(map[string]int)
	var maxRank UrgencyRank
	var maxCat string
	var maxLatest string

	if unclassAlerting {
		maxRank = UrgencyRankNormal
		maxCat = ""
		maxLatest = unclassLatest
	}

	for name, tr := range m.categories {
		cnt, catLatest := tr.Prune(now)
		if cnt > 0 {
			catCounts[name] = cnt
			total += cnt
			if cnt >= m.thresholds[name] {
				rank := ParseUrgencyRank(m.urgencies[name])
				if rank >= maxRank {
					maxRank = rank
					maxCat = name
					maxLatest = catLatest
				}
			}
		}
	}

	anyAlert := maxRank > UrgencyRankNone

	if anyAlert {
		if m.activeAlertCat != "" {
			if catCounts[m.activeAlertCat] >= m.thresholds[m.activeAlertCat] {
				maxCat = m.activeAlertCat
				maxRank = m.activeAlertRank
			} else {
				m.activeAlertRank = maxRank
				m.activeAlertCat = maxCat
				m.activeAlertMsg = maxLatest
			}
		} else if unclassAlerting {
			maxCat = ""
			maxRank = UrgencyRankNormal
		} else {
			m.activeAlertRank = maxRank
			m.activeAlertCat = maxCat
			m.activeAlertMsg = maxLatest
		}
	} else {
		m.activeAlertRank = UrgencyRankNone
		m.activeAlertCat = ""
		m.activeAlertMsg = ""
	}

	return total, catCounts, anyAlert, maxCat, maxRank.String()
}
