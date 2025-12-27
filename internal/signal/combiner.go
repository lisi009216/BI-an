package signal

import (
	"sync"
	"time"

	"example.com/binance-pivot-monitor/internal/pattern"
)

// CorrelationStrength represents the strength of correlation between signals.
type CorrelationStrength string

const (
	CorrelationStrong   CorrelationStrength = "strong"   // Direction match
	CorrelationModerate CorrelationStrength = "moderate" // Neutral pattern
	CorrelationWeak     CorrelationStrength = "weak"     // Direction conflict
)

// CombinedSignal represents a correlated pivot and pattern signal.
type CombinedSignal struct {
	PivotSignal   *Signal          `json:"pivot_signal"`
	PatternSignal *pattern.Signal  `json:"pattern_signal"`
	Correlation   CorrelationStrength `json:"correlation"`
	CombinedAt    time.Time        `json:"combined_at"`
}

// Combiner correlates pivot signals with pattern signals.
type Combiner struct {
	mu             sync.RWMutex
	recentPivots   map[string][]Signal         // symbol -> recent pivot signals
	recentPatterns map[string][]pattern.Signal // symbol -> recent pattern signals
	window         time.Duration               // Correlation time window
	onCombined     func(CombinedSignal)
}

// NewCombiner creates a new signal combiner.
// window: time window for correlating signals (default 15 minutes).
func NewCombiner(window time.Duration) *Combiner {
	return &Combiner{
		recentPivots:   make(map[string][]Signal),
		recentPatterns: make(map[string][]pattern.Signal),
		window:         window,
	}
}

// SetOnCombined sets the callback for combined signals.
func (c *Combiner) SetOnCombined(fn func(CombinedSignal)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onCombined = fn
}

// AddPivotSignal adds a pivot signal and checks for correlations.
func (c *Combiner) AddPivotSignal(sig Signal) []CombinedSignal {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add to recent pivots
	c.recentPivots[sig.Symbol] = append(c.recentPivots[sig.Symbol], sig)
	c.cleanupOld()

	// Check for correlations with recent patterns
	var combined []CombinedSignal
	patterns := c.recentPatterns[sig.Symbol]
	for i := range patterns {
		pat := &patterns[i]
		if c.isWithinWindow(sig.TriggeredAt, pat.DetectedAt) {
			corr := c.checkCorrelation(sig, *pat)
			cs := CombinedSignal{
				PivotSignal:   &sig,
				PatternSignal: pat,
				Correlation:   corr,
				CombinedAt:    time.Now().UTC(),
			}
			combined = append(combined, cs)

			if c.onCombined != nil {
				c.onCombined(cs)
			}
		}
	}

	return combined
}

// AddPatternSignal adds a pattern signal and checks for correlations.
func (c *Combiner) AddPatternSignal(sig pattern.Signal) []CombinedSignal {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add to recent patterns
	c.recentPatterns[sig.Symbol] = append(c.recentPatterns[sig.Symbol], sig)
	c.cleanupOld()

	// Check for correlations with recent pivots
	var combined []CombinedSignal
	pivots := c.recentPivots[sig.Symbol]
	for i := range pivots {
		piv := &pivots[i]
		if c.isWithinWindow(piv.TriggeredAt, sig.DetectedAt) {
			corr := c.checkCorrelation(*piv, sig)
			cs := CombinedSignal{
				PivotSignal:   piv,
				PatternSignal: &sig,
				Correlation:   corr,
				CombinedAt:    time.Now().UTC(),
			}
			combined = append(combined, cs)

			if c.onCombined != nil {
				c.onCombined(cs)
			}
		}
	}

	return combined
}

// isWithinWindow checks if two times are within the correlation window.
func (c *Combiner) isWithinWindow(t1, t2 time.Time) bool {
	diff := t1.Sub(t2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= c.window
}

// checkCorrelation determines the correlation strength between signals.
func (c *Combiner) checkCorrelation(pivot Signal, pat pattern.Signal) CorrelationStrength {
	// Neutral patterns are always moderate
	if pat.Direction == pattern.DirectionNeutral {
		return CorrelationModerate
	}

	// Check direction match
	pivotUp := pivot.Direction == "up"
	patternBullish := pat.Direction == pattern.DirectionBullish

	if pivotUp && patternBullish {
		return CorrelationStrong
	}
	if !pivotUp && !patternBullish {
		return CorrelationStrong
	}

	// Direction conflict
	return CorrelationWeak
}

// cleanupOld removes signals outside the time window.
func (c *Combiner) cleanupOld() {
	now := time.Now()
	cutoff := now.Add(-c.window * 2) // Keep 2x window for safety

	for symbol := range c.recentPivots {
		var kept []Signal
		for _, sig := range c.recentPivots[symbol] {
			if sig.TriggeredAt.After(cutoff) {
				kept = append(kept, sig)
			}
		}
		if len(kept) > 0 {
			c.recentPivots[symbol] = kept
		} else {
			delete(c.recentPivots, symbol)
		}
	}

	for symbol := range c.recentPatterns {
		var kept []pattern.Signal
		for _, sig := range c.recentPatterns[symbol] {
			if sig.DetectedAt.After(cutoff) {
				kept = append(kept, sig)
			}
		}
		if len(kept) > 0 {
			c.recentPatterns[symbol] = kept
		} else {
			delete(c.recentPatterns, symbol)
		}
	}
}

// GetRecentPivots returns recent pivot signals for a symbol.
func (c *Combiner) GetRecentPivots(symbol string) []Signal {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pivots := c.recentPivots[symbol]
	result := make([]Signal, len(pivots))
	copy(result, pivots)
	return result
}

// GetRecentPatterns returns recent pattern signals for a symbol.
func (c *Combiner) GetRecentPatterns(symbol string) []pattern.Signal {
	c.mu.RLock()
	defer c.mu.RUnlock()

	patterns := c.recentPatterns[symbol]
	result := make([]pattern.Signal, len(patterns))
	copy(result, patterns)
	return result
}
