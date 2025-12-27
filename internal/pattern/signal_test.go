package pattern

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestNewSignal(t *testing.T) {
	klineTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)
	signal := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime)

	if signal.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %v, want BTCUSDT", signal.Symbol)
	}
	if signal.Pattern != PatternHammer {
		t.Errorf("Pattern = %v, want %v", signal.Pattern, PatternHammer)
	}
	if signal.PatternCN != "锤子线" {
		t.Errorf("PatternCN = %v, want 锤子线", signal.PatternCN)
	}
	if signal.Direction != DirectionBullish {
		t.Errorf("Direction = %v, want bullish", signal.Direction)
	}
	if signal.Confidence != 75 {
		t.Errorf("Confidence = %v, want 75", signal.Confidence)
	}
	if signal.ID == "" {
		t.Error("ID should not be empty")
	}
	if signal.DetectedAt.IsZero() {
		t.Error("DetectedAt should not be zero")
	}

	// Check stats are populated
	if signal.UpPercent != 60 {
		t.Errorf("UpPercent = %v, want 60", signal.UpPercent)
	}
	if signal.DownPercent != 40 {
		t.Errorf("DownPercent = %v, want 40", signal.DownPercent)
	}
	if signal.EfficiencyRank != "B+" {
		t.Errorf("EfficiencyRank = %v, want B+", signal.EfficiencyRank)
	}
	if signal.Source != "custom" {
		t.Errorf("Source = %v, want custom", signal.Source)
	}
}

func TestGenerateID(t *testing.T) {
	klineTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)

	id1 := generateID("BTCUSDT", PatternHammer, klineTime)
	id2 := generateID("BTCUSDT", PatternEngulfing, klineTime)
	id3 := generateID("ETHUSDT", PatternHammer, klineTime)

	// Same symbol + pattern + time should produce same ID
	id1Again := generateID("BTCUSDT", PatternHammer, klineTime)
	if id1 != id1Again {
		t.Error("Same inputs should produce same ID")
	}

	// Different pattern should produce different ID
	if id1 == id2 {
		t.Error("Different patterns should produce different IDs")
	}

	// Different symbol should produce different ID
	if id1 == id3 {
		t.Error("Different symbols should produce different IDs")
	}
}

func TestSignal_IsValid(t *testing.T) {
	klineTime := time.Date(2024, 1, 1, 10, 5, 0, 0, time.UTC)

	tests := []struct {
		name     string
		signal   Signal
		expected bool
	}{
		{
			name:     "valid signal",
			signal:   NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime),
			expected: true,
		},
		{
			name: "empty ID",
			signal: Signal{
				Symbol:     "BTCUSDT",
				Pattern:    PatternHammer,
				Direction:  DirectionBullish,
				Confidence: 75,
				KlineTime:  klineTime,
				DetectedAt: time.Now(),
			},
			expected: false,
		},
		{
			name: "empty symbol",
			signal: Signal{
				ID:         "test-id",
				Pattern:    PatternHammer,
				Direction:  DirectionBullish,
				Confidence: 75,
				KlineTime:  klineTime,
				DetectedAt: time.Now(),
			},
			expected: false,
		},
		{
			name: "invalid direction",
			signal: Signal{
				ID:         "test-id",
				Symbol:     "BTCUSDT",
				Pattern:    PatternHammer,
				Direction:  "invalid",
				Confidence: 75,
				KlineTime:  klineTime,
				DetectedAt: time.Now(),
			},
			expected: false,
		},
		{
			name: "confidence out of range",
			signal: Signal{
				ID:         "test-id",
				Symbol:     "BTCUSDT",
				Pattern:    PatternHammer,
				Direction:  DirectionBullish,
				Confidence: 150,
				KlineTime:  klineTime,
				DetectedAt: time.Now(),
			},
			expected: false,
		},
		{
			name: "zero kline time",
			signal: Signal{
				ID:         "test-id",
				Symbol:     "BTCUSDT",
				Pattern:    PatternHammer,
				Direction:  DirectionBullish,
				Confidence: 75,
				DetectedAt: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.signal.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Property test: Signal completeness
func TestProperty_SignalCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Get all pattern types
	patternTypes := make([]PatternType, 0, len(PatternStatsMap))
	for pt := range PatternStatsMap {
		patternTypes = append(patternTypes, pt)
	}

	directions := []Direction{DirectionBullish, DirectionBearish, DirectionNeutral}

	properties.Property("NewSignal creates valid signals with all required fields", prop.ForAll(
		func(symbolIdx, patternIdx, directionIdx, confidence int, year, month, day, hour, minute int) bool {
			symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}
			symbol := symbols[symbolIdx%len(symbols)]
			pattern := patternTypes[patternIdx%len(patternTypes)]
			direction := directions[directionIdx%len(directions)]
			conf := confidence % 101 // 0-100

			klineTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
			signal := NewSignal(symbol, pattern, direction, conf, klineTime)

			// Check all required fields
			if signal.ID == "" {
				return false
			}
			if signal.Symbol != symbol {
				return false
			}
			if signal.Pattern != pattern {
				return false
			}
			if signal.Direction != direction {
				return false
			}
			if signal.Confidence != conf {
				return false
			}
			if signal.KlineTime.IsZero() {
				return false
			}
			if signal.DetectedAt.IsZero() {
				return false
			}

			// Check stats are populated
			stats, ok := PatternStatsMap[pattern]
			if ok {
				if signal.UpPercent != stats.UpPercent {
					return false
				}
				if signal.DownPercent != stats.DownPercent {
					return false
				}
				if signal.EfficiencyRank != stats.EfficiencyRank {
					return false
				}
			}

			return signal.IsValid()
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(2020, 2030),
		gen.IntRange(1, 12),
		gen.IntRange(1, 28),
		gen.IntRange(0, 23),
		gen.IntRange(0, 59),
	))

	properties.TestingRun(t)
}
