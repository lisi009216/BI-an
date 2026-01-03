package signal

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// =============================================================================
// Task 2.2: Property Test - Signal History Capacity
// Validates: Requirements 2.1, 2.2
// =============================================================================

// TestProperty_SignalHistoryCapacity tests that history respects max capacity
// and Query respects the 4000 limit.
func TestProperty_SignalHistoryCapacity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("History respects max capacity per period bucket", prop.ForAll(
		func(maxCap int, numSignals int) bool {
			if maxCap < 100 {
				maxCap = 100
			}
			if maxCap > 500 {
				maxCap = 500
			}
			if numSignals < 0 {
				numSignals = 0
			}
			if numSignals > 1000 {
				numSignals = 1000
			}

			h := NewHistory(maxCap)

			// Get the daily bucket capacity (80% of total)
			dailyBucketMax := int(float64(maxCap) * dailyRatio)
			if dailyBucketMax < 100 {
				dailyBucketMax = 100
			}

			// Add signals (all daily period)
			for i := 0; i < numSignals; i++ {
				h.Add(Signal{
					ID:          string(rune('A' + i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: time.Now(),
				})
			}

			// Count should not exceed daily bucket max
			count := h.Count()
			if numSignals <= dailyBucketMax {
				return count == numSignals
			}
			return count == dailyBucketMax
		},
		gen.IntRange(100, 500),
		gen.IntRange(0, 1000),
	))

	properties.Property("Query limit capped at 4000", prop.ForAll(
		func(requestedLimit int) bool {
			h := NewHistory(5000)

			// Add 4500 signals
			for i := 0; i < 4500; i++ {
				h.Add(Signal{
					ID:          string(rune('A' + i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: time.Now(),
				})
			}

			results := h.Query("", "", "", "", "", requestedLimit)

			// If requested > 4000, should be capped at 4000
			if requestedLimit > 4000 {
				return len(results) == 4000
			}
			// If requested <= 0, default to 200
			if requestedLimit <= 0 {
				return len(results) == 200
			}
			// Otherwise should return requested amount
			return len(results) == requestedLimit
		},
		gen.IntRange(-100, 5000),
	))

	properties.TestingRun(t)
}

// TestHistory_QueryLimit4000 tests that Query can return up to 4000 signals.
func TestHistory_QueryLimit4000(t *testing.T) {
	h := NewHistory(5000)

	// Add 4500 signals
	for i := 0; i < 4500; i++ {
		h.Add(Signal{
			ID:          string(rune('A' + i%26)),
			Symbol:      "TESTUSDT",
			Period:      "1d",
			Level:       "R1",
			Price:       float64(i),
			Direction:   "up",
			TriggeredAt: time.Now(),
		})
	}

	// Query with limit 4000
	results := h.Query("", "", "", "", "", 4000)
	if len(results) != 4000 {
		t.Errorf("expected 4000 results, got %d", len(results))
	}

	// Query with limit 5000 should be capped at 4000
	results = h.Query("", "", "", "", "", 5000)
	if len(results) != 4000 {
		t.Errorf("expected 4000 results (capped), got %d", len(results))
	}

	// Query with limit 0 should default to 200
	results = h.Query("", "", "", "", "", 0)
	if len(results) != 200 {
		t.Errorf("expected 200 results (default), got %d", len(results))
	}
}


// =============================================================================
// Property Tests for Signal History Separation
// Feature: signal-history-separation
// =============================================================================

// TestProperty_PeriodSpecificStorage tests that signals are stored in the correct period bucket.
// **Feature: signal-history-separation, Property 1: Period-specific storage**
// **Validates: Requirements 1.1**
func TestProperty_PeriodSpecificStorage(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for period values
	periodGen := gen.OneConstOf("1d", "1w", "d", "w", "daily", "weekly", "4h", "1h", "")

	properties.Property("Signals are stored in correct period bucket", prop.ForAll(
		func(period string, numSignals int) bool {
			if numSignals < 1 {
				numSignals = 1
			}
			if numSignals > 50 {
				numSignals = 50
			}

			h := NewHistory(1000)

			// Add signals with the given period
			for i := 0; i < numSignals; i++ {
				h.Add(Signal{
					ID:          string(rune('A' + i%26)),
					Symbol:      "TESTUSDT",
					Period:      period,
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: time.Now().Add(time.Duration(i) * time.Second),
				})
			}

			// Determine expected bucket key
			expectedBucket := normalizePeriod(period)

			// Verify signals are in the correct bucket
			h.bucketsMu.RLock()
			bucket, ok := h.buckets[expectedBucket]
			h.bucketsMu.RUnlock()

			if !ok {
				t.Logf("Bucket %s not found for period %s", expectedBucket, period)
				return false
			}

			bucket.mu.RLock()
			bucketCount := len(bucket.signals)
			bucket.mu.RUnlock()

			// All signals should be in this bucket
			if bucketCount != numSignals {
				t.Logf("Expected %d signals in bucket %s, got %d", numSignals, expectedBucket, bucketCount)
				return false
			}

			// Other buckets should be empty (for this test)
			h.bucketsMu.RLock()
			for key, b := range h.buckets {
				if key != expectedBucket {
					b.mu.RLock()
					otherCount := len(b.signals)
					b.mu.RUnlock()
					if otherCount != 0 {
						h.bucketsMu.RUnlock()
						t.Logf("Expected 0 signals in bucket %s, got %d", key, otherCount)
						return false
					}
				}
			}
			h.bucketsMu.RUnlock()

			return true
		},
		periodGen,
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// TestProperty_CrossPeriodIsolation tests that eviction only affects the same period bucket.
// **Feature: signal-history-separation, Property 2: Cross-period isolation on eviction**
// **Validates: Requirements 1.2**
func TestProperty_CrossPeriodIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Eviction only affects same period bucket", prop.ForAll(
		func(dailyCount, weeklyCount int) bool {
			// Use small capacity to trigger eviction
			h := NewHistory(100) // This gives ~80 daily, ~15 weekly

			// Override bucket capacities for testing
			h.bucketsMu.Lock()
			h.buckets[PeriodDaily].max = 10
			h.buckets[PeriodWeekly].max = 10
			h.bucketsMu.Unlock()

			if dailyCount < 1 {
				dailyCount = 1
			}
			if dailyCount > 30 {
				dailyCount = 30
			}
			if weeklyCount < 1 {
				weeklyCount = 1
			}
			if weeklyCount > 30 {
				weeklyCount = 30
			}

			// Add daily signals
			for i := 0; i < dailyCount; i++ {
				h.Add(Signal{
					ID:          "D" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: time.Now().Add(time.Duration(i) * time.Second),
				})
			}

			// Add weekly signals
			for i := 0; i < weeklyCount; i++ {
				h.Add(Signal{
					ID:          "W" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1w",
					Level:       "S1",
					Price:       float64(i + 1000),
					Direction:   "down",
					TriggeredAt: time.Now().Add(time.Duration(i+dailyCount) * time.Second),
				})
			}

			// Check bucket counts
			h.bucketsMu.RLock()
			dailyBucket := h.buckets[PeriodDaily]
			weeklyBucket := h.buckets[PeriodWeekly]
			h.bucketsMu.RUnlock()

			dailyBucket.mu.RLock()
			actualDaily := len(dailyBucket.signals)
			dailyBucket.mu.RUnlock()

			weeklyBucket.mu.RLock()
			actualWeekly := len(weeklyBucket.signals)
			weeklyBucket.mu.RUnlock()

			// Daily bucket should have min(dailyCount, 10)
			expectedDaily := dailyCount
			if expectedDaily > 10 {
				expectedDaily = 10
			}

			// Weekly bucket should have min(weeklyCount, 10)
			expectedWeekly := weeklyCount
			if expectedWeekly > 10 {
				expectedWeekly = 10
			}

			if actualDaily != expectedDaily {
				t.Logf("Expected %d daily signals, got %d", expectedDaily, actualDaily)
				return false
			}

			if actualWeekly != expectedWeekly {
				t.Logf("Expected %d weekly signals, got %d", expectedWeekly, actualWeekly)
				return false
			}

			return true
		},
		gen.IntRange(1, 30),
		gen.IntRange(1, 30),
	))

	properties.TestingRun(t)
}


// TestProperty_MergeAndSort tests that queries without period filter merge and sort correctly.
// **Feature: signal-history-separation, Property 3: Merge and chronological sort**
// **Validates: Requirements 1.4, 4.5, 4.6**
func TestProperty_MergeAndSort(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Query without period filter returns sorted results", prop.ForAll(
		func(dailyCount, weeklyCount int) bool {
			if dailyCount < 1 {
				dailyCount = 1
			}
			if dailyCount > 20 {
				dailyCount = 20
			}
			if weeklyCount < 1 {
				weeklyCount = 1
			}
			if weeklyCount > 20 {
				weeklyCount = 20
			}

			h := NewHistory(1000)
			baseTime := time.Now()

			// Add daily signals with even timestamps
			for i := 0; i < dailyCount; i++ {
				h.Add(Signal{
					ID:          "D" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: baseTime.Add(time.Duration(i*2) * time.Second),
				})
			}

			// Add weekly signals with odd timestamps (interleaved)
			for i := 0; i < weeklyCount; i++ {
				h.Add(Signal{
					ID:          "W" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1w",
					Level:       "S1",
					Price:       float64(i + 1000),
					Direction:   "down",
					TriggeredAt: baseTime.Add(time.Duration(i*2+1) * time.Second),
				})
			}

			// Query without period filter
			results := h.Query("", "", "", "", "", 1000)

			// Should have all signals
			expectedTotal := dailyCount + weeklyCount
			if len(results) != expectedTotal {
				t.Logf("Expected %d results, got %d", expectedTotal, len(results))
				return false
			}

			// Verify sorted by triggered_at descending
			for i := 1; i < len(results); i++ {
				if results[i-1].TriggeredAt.Before(results[i].TriggeredAt) {
					t.Logf("Results not sorted: %v before %v at index %d",
						results[i-1].TriggeredAt, results[i].TriggeredAt, i)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}

// TestProperty_PeriodFilter tests that period filter only returns signals from that period.
// **Feature: signal-history-separation, Property 4: Period filter queries correct bucket**
// **Validates: Requirements 4.4**
func TestProperty_PeriodFilter(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Period filter returns only matching signals", prop.ForAll(
		func(dailyCount, weeklyCount int, queryPeriod string) bool {
			if dailyCount < 1 {
				dailyCount = 1
			}
			if dailyCount > 20 {
				dailyCount = 20
			}
			if weeklyCount < 1 {
				weeklyCount = 1
			}
			if weeklyCount > 20 {
				weeklyCount = 20
			}

			h := NewHistory(1000)
			baseTime := time.Now()

			// Add daily signals
			for i := 0; i < dailyCount; i++ {
				h.Add(Signal{
					ID:          "D" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: baseTime.Add(time.Duration(i) * time.Second),
				})
			}

			// Add weekly signals
			for i := 0; i < weeklyCount; i++ {
				h.Add(Signal{
					ID:          "W" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1w",
					Level:       "S1",
					Price:       float64(i + 1000),
					Direction:   "down",
					TriggeredAt: baseTime.Add(time.Duration(i+dailyCount) * time.Second),
				})
			}

			// Query with period filter
			results := h.Query("", queryPeriod, "", "", "", 1000)

			// Determine expected count based on query period
			var expectedCount int
			normalizedQuery := normalizePeriod(queryPeriod)
			switch normalizedQuery {
			case PeriodDaily:
				expectedCount = dailyCount
			case PeriodWeekly:
				expectedCount = weeklyCount
			default:
				expectedCount = 0 // No signals in "other" bucket
			}

			if len(results) != expectedCount {
				t.Logf("Query period=%s (normalized=%s): expected %d results, got %d",
					queryPeriod, normalizedQuery, expectedCount, len(results))
				return false
			}

			// Verify all results have the correct period
			for _, s := range results {
				if normalizePeriod(s.Period) != normalizedQuery {
					t.Logf("Signal period %s doesn't match query %s", s.Period, queryPeriod)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
		gen.OneConstOf("1d", "1w", "d", "w", "daily", "weekly"),
	))

	properties.TestingRun(t)
}


// TestProperty_PersistenceRoundTrip tests that signals survive persistence reload.
// **Feature: signal-history-separation, Property 5: Persistence round-trip**
// **Validates: Requirements 3.2, 5.1, 5.3**
func TestProperty_PersistenceRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("Signals survive persistence round-trip", prop.ForAll(
		func(dailyCount, weeklyCount int) bool {
			if dailyCount < 1 {
				dailyCount = 1
			}
			if dailyCount > 20 {
				dailyCount = 20
			}
			if weeklyCount < 1 {
				weeklyCount = 1
			}
			if weeklyCount > 20 {
				weeklyCount = 20
			}

			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "history_test_*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tmpDir)

			filePath := tmpDir + "/history.jsonl"
			baseTime := time.Now().Truncate(time.Second) // Truncate for comparison

			// Create history and add signals
			h1 := NewHistory(1000)
			if err := h1.EnablePersistence(filePath); err != nil {
				t.Logf("Failed to enable persistence: %v", err)
				return false
			}

			// Track added signals
			var addedSignals []Signal

			// Add daily signals
			for i := 0; i < dailyCount; i++ {
				s := Signal{
					ID:          "D" + string(rune('A'+i%26)),
					Symbol:      "TESTUSDT",
					Period:      "1d",
					Level:       "R1",
					Price:       float64(i),
					Direction:   "up",
					TriggeredAt: baseTime.Add(time.Duration(i) * time.Second),
					Source:      "test",
				}
				h1.Add(s)
				addedSignals = append(addedSignals, s)
			}

			// Add weekly signals
			for i := 0; i < weeklyCount; i++ {
				s := Signal{
					ID:          "W" + string(rune('A'+i%26)),
					Symbol:      "BTCUSDT",
					Period:      "1w",
					Level:       "S1",
					Price:       float64(i + 1000),
					Direction:   "down",
					TriggeredAt: baseTime.Add(time.Duration(i+dailyCount) * time.Second),
					Source:      "test",
				}
				h1.Add(s)
				addedSignals = append(addedSignals, s)
			}

			// Create new history and load from persistence
			h2 := NewHistory(1000)
			if err := h2.EnablePersistence(filePath); err != nil {
				t.Logf("Failed to enable persistence on reload: %v", err)
				return false
			}

			// Verify count matches
			if h2.Count() != len(addedSignals) {
				t.Logf("Count mismatch: expected %d, got %d", len(addedSignals), h2.Count())
				return false
			}

			// Query all signals and verify
			results := h2.Query("", "", "", "", "", 1000)
			if len(results) != len(addedSignals) {
				t.Logf("Query results mismatch: expected %d, got %d", len(addedSignals), len(results))
				return false
			}

			// Verify each signal is present with correct fields
			resultMap := make(map[string]Signal)
			for _, s := range results {
				resultMap[s.ID] = s
			}

			for _, expected := range addedSignals {
				actual, ok := resultMap[expected.ID]
				if !ok {
					t.Logf("Signal %s not found after reload", expected.ID)
					return false
				}
				if actual.Symbol != expected.Symbol ||
					actual.Period != expected.Period ||
					actual.Level != expected.Level ||
					actual.Price != expected.Price ||
					actual.Direction != expected.Direction ||
					actual.Source != expected.Source ||
					!actual.TriggeredAt.Equal(expected.TriggeredAt) {
					t.Logf("Signal %s fields mismatch: expected %+v, got %+v", expected.ID, expected, actual)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}

// TestMigrationFromUnified tests migration from old unified file to separated storage.
func TestMigrationFromUnified(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "migration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	unifiedPath := tmpDir + "/history.jsonl"
	baseTime := time.Now().Truncate(time.Second)

	// Create old unified file with mixed signals
	f, err := os.Create(unifiedPath)
	if err != nil {
		t.Fatalf("Failed to create unified file: %v", err)
	}
	enc := json.NewEncoder(f)

	// Write daily signals
	for i := 0; i < 5; i++ {
		enc.Encode(Signal{
			ID:          "D" + string(rune('A'+i)),
			Symbol:      "TESTUSDT",
			Period:      "1d",
			Level:       "R1",
			Price:       float64(i),
			Direction:   "up",
			TriggeredAt: baseTime.Add(time.Duration(i*2) * time.Second),
		})
	}

	// Write weekly signals
	for i := 0; i < 3; i++ {
		enc.Encode(Signal{
			ID:          "W" + string(rune('A'+i)),
			Symbol:      "BTCUSDT",
			Period:      "1w",
			Level:       "S1",
			Price:       float64(i + 1000),
			Direction:   "down",
			TriggeredAt: baseTime.Add(time.Duration(i*2+1) * time.Second),
		})
	}
	f.Close()

	// Create history and enable persistence (should trigger migration)
	h := NewHistory(1000)
	if err := h.EnablePersistence(unifiedPath); err != nil {
		t.Fatalf("EnablePersistence failed: %v", err)
	}

	// Verify migration happened
	if !h.migrated {
		t.Error("Expected migration to be marked as complete")
	}

	// Verify old file was renamed
	if _, err := os.Stat(unifiedPath + ".migrated"); os.IsNotExist(err) {
		t.Error("Expected old file to be renamed to .migrated")
	}

	// Verify period files were created
	dailyFile := tmpDir + "/history_1d.jsonl"
	weeklyFile := tmpDir + "/history_1w.jsonl"

	if _, err := os.Stat(dailyFile); os.IsNotExist(err) {
		t.Error("Expected daily file to be created")
	}
	if _, err := os.Stat(weeklyFile); os.IsNotExist(err) {
		t.Error("Expected weekly file to be created")
	}

	// Verify counts
	if h.Count() != 8 {
		t.Errorf("Expected 8 total signals, got %d", h.Count())
	}

	// Verify daily signals
	dailyResults := h.Query("", "1d", "", "", "", 100)
	if len(dailyResults) != 5 {
		t.Errorf("Expected 5 daily signals, got %d", len(dailyResults))
	}

	// Verify weekly signals
	weeklyResults := h.Query("", "1w", "", "", "", 100)
	if len(weeklyResults) != 3 {
		t.Errorf("Expected 3 weekly signals, got %d", len(weeklyResults))
	}
}
