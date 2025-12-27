package signal

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"example.com/binance-pivot-monitor/internal/pattern"
)

func TestCombiner_AddPivotSignal(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add pattern signal first
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
	c.AddPatternSignal(patSig)

	// Add pivot signal - should correlate
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "up",
		TriggeredAt: now.Add(5 * time.Minute),
	}

	combined := c.AddPivotSignal(pivSig)

	if len(combined) != 1 {
		t.Fatalf("Expected 1 combined signal, got %d", len(combined))
	}

	if combined[0].Correlation != CorrelationStrong {
		t.Errorf("Expected strong correlation, got %s", combined[0].Correlation)
	}
}

func TestCombiner_AddPatternSignal(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add pivot signal first
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "down",
		TriggeredAt: now,
	}
	c.AddPivotSignal(pivSig)

	// Add pattern signal - should correlate
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternShootingStar, pattern.DirectionBearish, 75, now.Add(5*time.Minute))

	combined := c.AddPatternSignal(patSig)

	if len(combined) != 1 {
		t.Fatalf("Expected 1 combined signal, got %d", len(combined))
	}

	if combined[0].Correlation != CorrelationStrong {
		t.Errorf("Expected strong correlation, got %s", combined[0].Correlation)
	}
}

func TestCombiner_NoCorrelation_DifferentSymbol(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add pattern for BTCUSDT
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
	c.AddPatternSignal(patSig)

	// Add pivot for ETHUSDT - should not correlate
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "ETHUSDT",
		Direction:   "up",
		TriggeredAt: now.Add(5 * time.Minute),
	}

	combined := c.AddPivotSignal(pivSig)

	if len(combined) != 0 {
		t.Errorf("Expected no combined signals, got %d", len(combined))
	}
}

func TestCombiner_NoCorrelation_OutsideWindow(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add pattern signal
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
	c.AddPatternSignal(patSig)

	// Add pivot signal outside window
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "up",
		TriggeredAt: now.Add(20 * time.Minute), // Outside 15 min window
	}

	combined := c.AddPivotSignal(pivSig)

	if len(combined) != 0 {
		t.Errorf("Expected no combined signals, got %d", len(combined))
	}
}

func TestCombiner_WeakCorrelation(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add bullish pattern
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
	c.AddPatternSignal(patSig)

	// Add down pivot - direction conflict
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "down",
		TriggeredAt: now.Add(5 * time.Minute),
	}

	combined := c.AddPivotSignal(pivSig)

	if len(combined) != 1 {
		t.Fatalf("Expected 1 combined signal, got %d", len(combined))
	}

	if combined[0].Correlation != CorrelationWeak {
		t.Errorf("Expected weak correlation, got %s", combined[0].Correlation)
	}
}

func TestCombiner_ModerateCorrelation_NeutralPattern(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	now := time.Now()

	// Add neutral pattern (Doji)
	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternDoji, pattern.DirectionNeutral, 75, now)
	c.AddPatternSignal(patSig)

	// Add pivot
	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "up",
		TriggeredAt: now.Add(5 * time.Minute),
	}

	combined := c.AddPivotSignal(pivSig)

	if len(combined) != 1 {
		t.Fatalf("Expected 1 combined signal, got %d", len(combined))
	}

	if combined[0].Correlation != CorrelationModerate {
		t.Errorf("Expected moderate correlation, got %s", combined[0].Correlation)
	}
}

func TestCombiner_Callback(t *testing.T) {
	c := NewCombiner(15 * time.Minute)

	var callbackCalled bool
	var receivedCombined CombinedSignal

	c.SetOnCombined(func(cs CombinedSignal) {
		callbackCalled = true
		receivedCombined = cs
	})

	now := time.Now()

	patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
	c.AddPatternSignal(patSig)

	pivSig := Signal{
		ID:          "test-1",
		Symbol:      "BTCUSDT",
		Direction:   "up",
		TriggeredAt: now.Add(5 * time.Minute),
	}
	c.AddPivotSignal(pivSig)

	if !callbackCalled {
		t.Error("Callback was not called")
	}

	if receivedCombined.Correlation != CorrelationStrong {
		t.Errorf("Callback received wrong correlation: %s", receivedCombined.Correlation)
	}
}

// Property tests

func TestProperty_TimeWindowCorrelation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Signals within window are correlated, outside are not", prop.ForAll(
		func(minutesDiff int) bool {
			c := NewCombiner(15 * time.Minute)
			now := time.Now()

			// Use a fixed base time for pattern signal to avoid cleanup issues
			patternTime := now
			patSig := pattern.Signal{
				ID:         "test-pattern",
				Symbol:     "BTCUSDT",
				Pattern:    pattern.PatternHammer,
				Direction:  pattern.DirectionBullish,
				Confidence: 75,
				KlineTime:  patternTime,
				DetectedAt: patternTime, // Use same time to avoid cleanup
			}
			c.AddPatternSignal(patSig)

			pivSig := Signal{
				ID:          "test",
				Symbol:      "BTCUSDT",
				Direction:   "up",
				TriggeredAt: patternTime.Add(time.Duration(minutesDiff) * time.Minute),
			}

			combined := c.AddPivotSignal(pivSig)

			// Within window means absolute difference <= 15
			absDiff := minutesDiff
			if absDiff < 0 {
				absDiff = -absDiff
			}
			withinWindow := absDiff <= 15
			hasCorrelation := len(combined) > 0

			return withinWindow == hasCorrelation
		},
		gen.IntRange(-30, 30),
	))

	properties.TestingRun(t)
}

func TestProperty_DirectionMatchStrong(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Matching directions produce strong correlation", prop.ForAll(
		func(isUp bool) bool {
			c := NewCombiner(15 * time.Minute)
			now := time.Now()

			var patDir pattern.Direction
			var pivDir string
			if isUp {
				patDir = pattern.DirectionBullish
				pivDir = "up"
			} else {
				patDir = pattern.DirectionBearish
				pivDir = "down"
			}

			patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, patDir, 75, now)
			c.AddPatternSignal(patSig)

			pivSig := Signal{
				ID:          "test",
				Symbol:      "BTCUSDT",
				Direction:   pivDir,
				TriggeredAt: now.Add(5 * time.Minute),
			}

			combined := c.AddPivotSignal(pivSig)

			if len(combined) != 1 {
				return false
			}

			return combined[0].Correlation == CorrelationStrong
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

func TestProperty_DirectionConflictWeak(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Conflicting directions produce weak correlation", prop.ForAll(
		func(isUp bool) bool {
			c := NewCombiner(15 * time.Minute)
			now := time.Now()

			var patDir pattern.Direction
			var pivDir string
			if isUp {
				patDir = pattern.DirectionBullish
				pivDir = "down" // Conflict
			} else {
				patDir = pattern.DirectionBearish
				pivDir = "up" // Conflict
			}

			patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, patDir, 75, now)
			c.AddPatternSignal(patSig)

			pivSig := Signal{
				ID:          "test",
				Symbol:      "BTCUSDT",
				Direction:   pivDir,
				TriggeredAt: now.Add(5 * time.Minute),
			}

			combined := c.AddPivotSignal(pivSig)

			if len(combined) != 1 {
				return false
			}

			return combined[0].Correlation == CorrelationWeak
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

func TestProperty_CombinedSignalCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Combined signals contain both pivot and pattern", prop.ForAll(
		func(dummy int) bool {
			c := NewCombiner(15 * time.Minute)
			now := time.Now()

			patSig := pattern.NewSignal("BTCUSDT", pattern.PatternHammer, pattern.DirectionBullish, 75, now)
			c.AddPatternSignal(patSig)

			pivSig := Signal{
				ID:          "test",
				Symbol:      "BTCUSDT",
				Direction:   "up",
				TriggeredAt: now.Add(5 * time.Minute),
			}

			combined := c.AddPivotSignal(pivSig)

			if len(combined) != 1 {
				return false
			}

			cs := combined[0]

			// Check completeness
			if cs.PivotSignal == nil {
				return false
			}
			if cs.PatternSignal == nil {
				return false
			}
			if cs.Correlation == "" {
				return false
			}
			if cs.CombinedAt.IsZero() {
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}
