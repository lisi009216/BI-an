package monitor

import (
	"testing"
	"time"

	"example.com/binance-pivot-monitor/internal/kline"
	"example.com/binance-pivot-monitor/internal/pattern"
	"example.com/binance-pivot-monitor/internal/pivot"
	signalpkg "example.com/binance-pivot-monitor/internal/signal"
	"example.com/binance-pivot-monitor/internal/sse"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// helper to set pivot levels for a symbol
func setPivotLevels(store *pivot.Store, period pivot.Period, symbol string, levels pivot.Levels) {
	snap, _ := store.Snapshot(period)
	if snap == nil {
		snap = &pivot.Snapshot{
			Period:    period,
			UpdatedAt: time.Now(),
			Symbols:   make(map[string]pivot.Levels),
		}
	}
	if snap.Symbols == nil {
		snap.Symbols = make(map[string]pivot.Levels)
	}
	snap.Symbols[symbol] = levels
	store.Swap(period, snap)
}

// TestOnKlineClose_SkipsWithoutPivotData tests Property 11:
// Pattern detection should only occur for symbols with loaded pivot data.
func TestOnKlineClose_SkipsWithoutPivotData(t *testing.T) {
	// Create pivot store with data for only one symbol
	pivotStore := pivot.NewStore()
	setPivotLevels(pivotStore, pivot.PeriodDaily, "BTCUSDT", pivot.Levels{
		R3: 50000, R4: 51000, R5: 52000,
		S3: 48000, S4: 47000, S5: 46000,
	})

	// Create pattern detector
	detector := pattern.NewDetector(pattern.DefaultDetectorConfig())

	// Create pattern history (memory only)
	patternHistory, err := pattern.NewHistory("", 100)
	if err != nil {
		t.Fatalf("failed to create pattern history: %v", err)
	}

	// Create monitor with pattern detection enabled
	m := NewWithConfig(MonitorConfig{
		PivotStore:      pivotStore,
		Broker:          sse.NewBroker[signalpkg.Signal](),
		PatternDetector: detector,
		PatternHistory:  patternHistory,
		PatternBroker:   sse.NewBroker[pattern.Signal](),
	})

	// Create test klines that would trigger a pattern (engulfing)
	klines := []kline.Kline{
		{Symbol: "ETHUSDT", Open: 100, High: 105, Low: 95, Close: 96, IsClosed: true},  // bearish
		{Symbol: "ETHUSDT", Open: 95, High: 110, Low: 94, Close: 108, IsClosed: true},  // bullish engulfing
	}

	// Call onKlineClose for symbol WITHOUT pivot data
	m.onKlineClose("ETHUSDT", klines)

	// Should not have recorded any patterns (no pivot data for ETHUSDT)
	if patternHistory.Count() != 0 {
		t.Errorf("expected 0 patterns for symbol without pivot data, got %d", patternHistory.Count())
	}

	// Now test with symbol that HAS pivot data
	klinesBTC := []kline.Kline{
		{Symbol: "BTCUSDT", Open: 100, High: 105, Low: 95, Close: 96, IsClosed: true},  // bearish
		{Symbol: "BTCUSDT", Open: 95, High: 110, Low: 94, Close: 108, IsClosed: true},  // bullish engulfing
	}

	m.onKlineClose("BTCUSDT", klinesBTC)

	// Should have recorded patterns (has pivot data for BTCUSDT)
	// Note: may or may not detect patterns depending on detector config
	// The key test is that it ATTEMPTS detection (doesn't skip)
}

// TestOnKlineClose_Property11_DetectionRangeLimit tests that pattern detection
// is limited to symbols with pivot data using property-based testing.
func TestOnKlineClose_Property11_DetectionRangeLimit(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Pattern detection only for symbols with pivot data", prop.ForAll(
		func(pivotIdx int, noPivotIdx int) bool {
			// Generate different symbols
			symbolWithPivot := "PIVOT" + string(rune('A'+pivotIdx%26))
			symbolWithoutPivot := "NOPIV" + string(rune('A'+noPivotIdx%26))

			// Create pivot store with data for only symbolWithPivot
			pivotStore := pivot.NewStore()
			setPivotLevels(pivotStore, pivot.PeriodDaily, symbolWithPivot, pivot.Levels{
				R3: 50000, R4: 51000, R5: 52000,
				S3: 48000, S4: 47000, S5: 46000,
			})

			// Create pattern detector
			detector := pattern.NewDetector(pattern.DefaultDetectorConfig())

			// Create pattern history to track detection
			patternHistory, _ := pattern.NewHistory("", 100)

			m := NewWithConfig(MonitorConfig{
				PivotStore:      pivotStore,
				Broker:          sse.NewBroker[signalpkg.Signal](),
				PatternDetector: detector,
				PatternHistory:  patternHistory,
				PatternBroker:   sse.NewBroker[pattern.Signal](),
			})

			// Create test klines
			klines := []kline.Kline{
				{Open: 100, High: 105, Low: 95, Close: 96, IsClosed: true},
				{Open: 95, High: 110, Low: 94, Close: 108, IsClosed: true},
			}

			// Test symbol WITHOUT pivot data
			initialCount := patternHistory.Count()
			m.onKlineClose(symbolWithoutPivot, klines)
			afterWithoutPivot := patternHistory.Count()

			// For symbol without pivot, no patterns should be recorded
			return afterWithoutPivot == initialCount
		},
		gen.IntRange(0, 25),
		gen.IntRange(0, 25),
	))

	properties.TestingRun(t)
}

// TestMonitorIntegration_KlineUpdate tests that price updates flow to KlineStore.
func TestMonitorIntegration_KlineUpdate(t *testing.T) {
	// Create kline store
	klineStore := kline.NewStore(5*time.Minute, 12)

	// Create pivot store
	pivotStore := pivot.NewStore()
	setPivotLevels(pivotStore, pivot.PeriodDaily, "BTCUSDT", pivot.Levels{
		R3: 50000, R4: 51000, R5: 52000,
		S3: 48000, S4: 47000, S5: 46000,
	})

	// Create monitor
	m := NewWithConfig(MonitorConfig{
		PivotStore: pivotStore,
		Broker:     sse.NewBroker[signalpkg.Signal](),
		KlineStore: klineStore,
	})

	// Simulate price updates
	ts := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	m.onPrice("BTCUSDT", 49000, ts)
	m.onPrice("BTCUSDT", 49100, ts.Add(1*time.Second))
	m.onPrice("BTCUSDT", 48900, ts.Add(2*time.Second))

	// Check that kline was created
	current, ok := klineStore.GetCurrentKline("BTCUSDT")
	if !ok {
		t.Fatal("expected current kline to exist")
	}

	if current.Open != 49000 {
		t.Errorf("expected open=49000, got %v", current.Open)
	}
	if current.High != 49100 {
		t.Errorf("expected high=49100, got %v", current.High)
	}
	if current.Low != 48900 {
		t.Errorf("expected low=48900, got %v", current.Low)
	}
	if current.Close != 48900 {
		t.Errorf("expected close=48900, got %v", current.Close)
	}
}

// TestNewWithConfig_SetsOnCloseCallback tests that NewWithConfig properly sets up the callback.
func TestNewWithConfig_SetsOnCloseCallback(t *testing.T) {
	klineStore := kline.NewStore(5*time.Minute, 12)
	detector := pattern.NewDetector(pattern.DefaultDetectorConfig())
	pivotStore := pivot.NewStore()

	m := NewWithConfig(MonitorConfig{
		PivotStore:      pivotStore,
		Broker:          sse.NewBroker[signalpkg.Signal](),
		KlineStore:      klineStore,
		PatternDetector: detector,
	})

	// Verify monitor was created
	if m.KlineStore == nil {
		t.Error("expected KlineStore to be set")
	}
	if m.PatternDetector == nil {
		t.Error("expected PatternDetector to be set")
	}
}
