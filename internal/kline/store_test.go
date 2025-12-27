package kline

import (
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestGetKlineOpenTime(t *testing.T) {
	tests := []struct {
		name     string
		ts       time.Time
		interval time.Duration
		expected int // expected minute
	}{
		{"minute 0", time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC), 5 * time.Minute, 0},
		{"minute 3", time.Date(2024, 1, 1, 10, 3, 30, 0, time.UTC), 5 * time.Minute, 0},
		{"minute 5", time.Date(2024, 1, 1, 10, 5, 30, 0, time.UTC), 5 * time.Minute, 5},
		{"minute 7", time.Date(2024, 1, 1, 10, 7, 30, 0, time.UTC), 5 * time.Minute, 5},
		{"minute 59", time.Date(2024, 1, 1, 10, 59, 30, 0, time.UTC), 5 * time.Minute, 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getKlineOpenTime(tt.ts, tt.interval)
			if result.Minute() != tt.expected {
				t.Errorf("getKlineOpenTime() minute = %v, want %v", result.Minute(), tt.expected)
			}
			if result.Second() != 0 {
				t.Errorf("getKlineOpenTime() second = %v, want 0", result.Second())
			}
			if result.Nanosecond() != 0 {
				t.Errorf("getKlineOpenTime() nanosecond = %v, want 0", result.Nanosecond())
			}
		})
	}
}

func TestStore_Update_NewSymbol(t *testing.T) {
	store := NewStore(5*time.Minute, 12)
	ts := time.Date(2024, 1, 1, 10, 2, 30, 0, time.UTC)

	closed := store.Update("BTCUSDT", 50000.0, ts)

	if closed {
		t.Error("Expected no kline close on first update")
	}

	current, ok := store.GetCurrentKline("BTCUSDT")
	if !ok {
		t.Fatal("Expected current kline to exist")
	}

	if current.Open != 50000.0 {
		t.Errorf("Open = %v, want 50000.0", current.Open)
	}
	if current.High != 50000.0 {
		t.Errorf("High = %v, want 50000.0", current.High)
	}
	if current.Low != 50000.0 {
		t.Errorf("Low = %v, want 50000.0", current.Low)
	}
	if current.Close != 50000.0 {
		t.Errorf("Close = %v, want 50000.0", current.Close)
	}
}

func TestStore_Update_OHLC(t *testing.T) {
	store := NewStore(5*time.Minute, 12)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// First price
	store.Update("BTCUSDT", 50000.0, baseTime)

	// Higher price
	store.Update("BTCUSDT", 51000.0, baseTime.Add(1*time.Minute))

	// Lower price
	store.Update("BTCUSDT", 49000.0, baseTime.Add(2*time.Minute))

	// Final price
	store.Update("BTCUSDT", 50500.0, baseTime.Add(3*time.Minute))

	current, _ := store.GetCurrentKline("BTCUSDT")

	if current.Open != 50000.0 {
		t.Errorf("Open = %v, want 50000.0", current.Open)
	}
	if current.High != 51000.0 {
		t.Errorf("High = %v, want 51000.0", current.High)
	}
	if current.Low != 49000.0 {
		t.Errorf("Low = %v, want 49000.0", current.Low)
	}
	if current.Close != 50500.0 {
		t.Errorf("Close = %v, want 50500.0", current.Close)
	}
}

func TestStore_Update_KlineClose(t *testing.T) {
	store := NewStore(5*time.Minute, 12)

	var closedSymbol string
	var closedKlines []Kline
	var wg sync.WaitGroup
	wg.Add(1)

	store.SetOnClose(func(symbol string, klines []Kline) {
		closedSymbol = symbol
		closedKlines = klines
		wg.Done()
	})

	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Update within first kline period
	store.Update("BTCUSDT", 50000.0, baseTime)
	store.Update("BTCUSDT", 51000.0, baseTime.Add(2*time.Minute))

	// Cross 5-minute boundary - should close kline
	closed := store.Update("BTCUSDT", 52000.0, baseTime.Add(5*time.Minute))

	if !closed {
		t.Error("Expected kline to close")
	}

	// Wait for callback
	wg.Wait()

	if closedSymbol != "BTCUSDT" {
		t.Errorf("Callback symbol = %v, want BTCUSDT", closedSymbol)
	}

	if len(closedKlines) != 1 {
		t.Fatalf("Callback klines length = %v, want 1", len(closedKlines))
	}

	if closedKlines[0].Open != 50000.0 {
		t.Errorf("Closed kline Open = %v, want 50000.0", closedKlines[0].Open)
	}
	if closedKlines[0].High != 51000.0 {
		t.Errorf("Closed kline High = %v, want 51000.0", closedKlines[0].High)
	}
}

func TestStore_RollingWindow(t *testing.T) {
	maxCount := 3
	store := NewStore(5*time.Minute, maxCount)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Create 5 klines (exceeds maxCount of 3)
	// Each iteration: create kline at time i*5min, then close it at (i+1)*5min
	for i := 0; i < 5; i++ {
		ts := baseTime.Add(time.Duration(i*5) * time.Minute)
		store.Update("BTCUSDT", float64(50000+i*100), ts)
		// Trigger close by moving to next period
		closeTs := baseTime.Add(time.Duration((i+1)*5) * time.Minute)
		store.Update("BTCUSDT", float64(50000+(i+1)*100), closeTs)
	}

	klines, ok := store.GetKlines("BTCUSDT")
	if !ok {
		t.Fatal("Expected klines to exist")
	}

	if len(klines) != maxCount {
		t.Errorf("Klines count = %v, want %v", len(klines), maxCount)
	}

	// After 5 klines created and maxCount=3, we should have klines 2, 3, 4 (0-indexed)
	// Kline 2 was created at i=2, with Open = 50000 + 2*100 = 50200
	// But the close triggers create new klines, so let's verify the actual values
	// The first remaining kline should be from iteration 2
	expectedFirstOpen := 50200.0
	if klines[0].Open != expectedFirstOpen {
		t.Errorf("First kline Open = %v, want %v", klines[0].Open, expectedFirstOpen)
	}
}

func TestStore_InvalidPrice(t *testing.T) {
	store := NewStore(5*time.Minute, 12)
	ts := time.Now()

	// Zero price should be ignored
	closed := store.Update("BTCUSDT", 0, ts)
	if closed {
		t.Error("Zero price should not close kline")
	}

	// Negative price should be ignored
	closed = store.Update("BTCUSDT", -100, ts)
	if closed {
		t.Error("Negative price should not close kline")
	}

	_, ok := store.GetCurrentKline("BTCUSDT")
	if ok {
		t.Error("No kline should exist after invalid prices")
	}
}

func TestStore_CleanupStale(t *testing.T) {
	store := NewStore(5*time.Minute, 12)
	now := time.Now()

	// Add some symbols
	store.Update("BTCUSDT", 50000, now)
	store.Update("ETHUSDT", 3000, now.Add(-2*time.Hour)) // Stale

	// Cleanup with 1 hour threshold
	removed := store.CleanupStale(1 * time.Hour)

	if removed != 1 {
		t.Errorf("Removed = %v, want 1", removed)
	}

	if store.SymbolCount() != 1 {
		t.Errorf("SymbolCount = %v, want 1", store.SymbolCount())
	}

	_, ok := store.GetCurrentKline("ETHUSDT")
	if ok {
		t.Error("ETHUSDT should have been removed")
	}
}

// Property Tests

func TestProperty_KlineTimeBoundaryAlignment(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Kline open time is aligned to 5-minute boundary", prop.ForAll(
		func(year, month, day, hour, minute, second int) bool {
			ts := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
			interval := 5 * time.Minute

			result := getKlineOpenTime(ts, interval)

			// Minute should be one of: 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55
			validMinutes := map[int]bool{0: true, 5: true, 10: true, 15: true, 20: true, 25: true, 30: true, 35: true, 40: true, 45: true, 50: true, 55: true}
			if !validMinutes[result.Minute()] {
				return false
			}

			// Second and nanosecond should be zero
			if result.Second() != 0 || result.Nanosecond() != 0 {
				return false
			}

			return true
		},
		gen.IntRange(2020, 2030),
		gen.IntRange(1, 12),
		gen.IntRange(1, 28),
		gen.IntRange(0, 23),
		gen.IntRange(0, 59),
		gen.IntRange(0, 59),
	))

	properties.TestingRun(t)
}

func TestProperty_OHLCInvariants(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("OHLC invariants hold after price updates", prop.ForAll(
		func(prices []float64) bool {
			if len(prices) == 0 {
				return true
			}

			// Filter out invalid prices
			validPrices := make([]float64, 0, len(prices))
			for _, p := range prices {
				if p > 0 {
					validPrices = append(validPrices, p)
				}
			}
			if len(validPrices) == 0 {
				return true
			}

			store := NewStore(5*time.Minute, 12)
			baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

			for i, price := range validPrices {
				store.Update("TEST", price, baseTime.Add(time.Duration(i)*time.Second))
			}

			current, ok := store.GetCurrentKline("TEST")
			if !ok {
				return true // No kline created
			}

			// Open equals first price
			if current.Open != validPrices[0] {
				return false
			}

			// High >= all prices
			for _, p := range validPrices {
				if current.High < p {
					return false
				}
			}

			// Low <= all prices
			for _, p := range validPrices {
				if current.Low > p {
					return false
				}
			}

			// Close equals last price
			if current.Close != validPrices[len(validPrices)-1] {
				return false
			}

			// High >= Low
			if current.High < current.Low {
				return false
			}

			// High >= Open and High >= Close
			if current.High < current.Open || current.High < current.Close {
				return false
			}

			// Low <= Open and Low <= Close
			if current.Low > current.Open || current.Low > current.Close {
				return false
			}

			return true
		},
		gen.SliceOf(gen.Float64Range(0.01, 100000)),
	))

	properties.TestingRun(t)
}

func TestProperty_RollingWindowSizeLimit(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Rolling window never exceeds maxCount", prop.ForAll(
		func(maxCount, numKlines int) bool {
			if maxCount < 1 {
				maxCount = 1
			}
			if numKlines < 0 {
				numKlines = 0
			}

			store := NewStore(5*time.Minute, maxCount)
			baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

			// Create numKlines klines
			for i := 0; i < numKlines; i++ {
				ts := baseTime.Add(time.Duration(i*5) * time.Minute)
				store.Update("TEST", float64(50000+i), ts)
				// Trigger close
				store.Update("TEST", float64(50000+i), ts.Add(5*time.Minute))
			}

			count := store.KlineCount("TEST")
			return count <= maxCount
		},
		gen.IntRange(1, 100),
		gen.IntRange(0, 200),
	))

	properties.TestingRun(t)
}

// Property 3: 参数校验防护
// *For any* negative or zero value passed to NewStore, the system should not panic
// and should use a positive default value.
// **Validates: Requirements 2.2**

func TestNewStore_NegativeMaxCount(t *testing.T) {
	// 测试负数不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewStore panicked with negative maxCount: %v", r)
		}
	}()

	store := NewStore(5*time.Minute, -10)
	if store == nil {
		t.Error("NewStore returned nil")
	}

	// 验证使用了默认值
	if store.maxCount != DefaultKlineCount {
		t.Errorf("maxCount = %d, want default %d", store.maxCount, DefaultKlineCount)
	}
}

func TestNewStore_ZeroMaxCount(t *testing.T) {
	// 测试零不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewStore panicked with zero maxCount: %v", r)
		}
	}()

	store := NewStore(5*time.Minute, 0)
	if store == nil {
		t.Error("NewStore returned nil")
	}

	// 验证使用了默认值
	if store.maxCount != DefaultKlineCount {
		t.Errorf("maxCount = %d, want default %d", store.maxCount, DefaultKlineCount)
	}
}

func TestNewStore_ValidMaxCount(t *testing.T) {
	store := NewStore(5*time.Minute, 20)
	if store == nil {
		t.Error("NewStore returned nil")
	}

	// 验证使用了传入的值
	if store.maxCount != 20 {
		t.Errorf("maxCount = %d, want 20", store.maxCount)
	}
}

func TestProperty_NewStoreNeverPanics(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("NewStore never panics with any maxCount value", prop.ForAll(
		func(maxCount int) bool {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewStore panicked with maxCount=%d: %v", maxCount, r)
				}
			}()

			store := NewStore(5*time.Minute, maxCount)
			if store == nil {
				return false
			}

			// maxCount 应该总是正数
			if store.maxCount <= 0 {
				return false
			}

			return true
		},
		gen.IntRange(-1000, 1000),
	))

	properties.TestingRun(t)
}
