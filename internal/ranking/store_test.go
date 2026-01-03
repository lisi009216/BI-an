package ranking

import (
	"testing"
	"testing/quick"
	"time"
)

func TestStoreAddAndLatest(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	// Empty store
	if store.Latest() != nil {
		t.Error("Expected nil for empty store")
	}

	// Add first snapshot
	snap1 := &Snapshot{
		Timestamp: time.Now().Add(-10 * time.Minute),
		Items:     map[string]*SnapshotItem{"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1}},
	}
	store.Add(snap1)

	if store.Latest() != snap1 {
		t.Error("Latest should return the added snapshot")
	}
	if store.Count() != 1 {
		t.Errorf("Count = %d, want 1", store.Count())
	}

	// Add second snapshot
	snap2 := &Snapshot{
		Timestamp: time.Now().Add(-5 * time.Minute),
		Items:     map[string]*SnapshotItem{"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 2}},
	}
	store.Add(snap2)

	if store.Latest() != snap2 {
		t.Error("Latest should return the most recent snapshot")
	}
	if store.Previous() != snap1 {
		t.Error("Previous should return the second most recent snapshot")
	}
	if store.Count() != 2 {
		t.Errorf("Count = %d, want 2", store.Count())
	}
}

func TestStoreFindSnapshotByTime(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	now := time.Now()
	snap1 := &Snapshot{Timestamp: now.Add(-60 * time.Minute), Items: map[string]*SnapshotItem{}}
	snap2 := &Snapshot{Timestamp: now.Add(-30 * time.Minute), Items: map[string]*SnapshotItem{}}
	snap3 := &Snapshot{Timestamp: now.Add(-10 * time.Minute), Items: map[string]*SnapshotItem{}}

	store.Add(snap1)
	store.Add(snap2)
	store.Add(snap3)

	// Find snapshot at 45 minutes ago (should return snap1, the one at 60 min ago)
	target := now.Add(-45 * time.Minute)
	found := store.FindSnapshotByTime(target)
	if found != snap1 {
		t.Errorf("Expected snap1 (60 min ago), got timestamp %v", found.Timestamp)
	}

	// Find snapshot at 25 minutes ago (should return snap2, the one at 30 min ago)
	target = now.Add(-25 * time.Minute)
	found = store.FindSnapshotByTime(target)
	if found != snap2 {
		t.Errorf("Expected snap2 (30 min ago), got timestamp %v", found.Timestamp)
	}

	// Find snapshot at 5 minutes ago (should return snap3, the one at 10 min ago)
	target = now.Add(-5 * time.Minute)
	found = store.FindSnapshotByTime(target)
	if found != snap3 {
		t.Errorf("Expected snap3 (10 min ago), got timestamp %v", found.Timestamp)
	}

	// Find snapshot at 2 hours ago (no snapshot that old, should return oldest)
	target = now.Add(-2 * time.Hour)
	found = store.FindSnapshotByTime(target)
	if found != snap1 {
		t.Errorf("Expected oldest snapshot (snap1), got timestamp %v", found.Timestamp)
	}
}

func TestStoreCleanup(t *testing.T) {
	// Use 1 hour max age for testing
	store := NewStore("", 1*time.Hour)

	now := time.Now()

	// Add old snapshot (2 hours ago - should be cleaned up)
	oldSnap := &Snapshot{
		Timestamp: now.Add(-2 * time.Hour),
		Items:     map[string]*SnapshotItem{"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1}},
	}
	store.Add(oldSnap)

	// Add recent snapshot (30 minutes ago - should be kept)
	recentSnap := &Snapshot{
		Timestamp: now.Add(-30 * time.Minute),
		Items:     map[string]*SnapshotItem{"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 2}},
	}
	store.Add(recentSnap)

	// Old snapshot should have been cleaned up
	if store.Count() != 1 {
		t.Errorf("Count = %d, want 1 (old snapshot should be cleaned up)", store.Count())
	}

	if store.Latest() != recentSnap {
		t.Error("Latest should be the recent snapshot")
	}
}

// TestRetentionWindowProperty tests the 24-hour retention window property.
// Property 4: 24-Hour Retention Window
// Validates: Requirements 1.4, 1.5
// Note: In production, snapshots are added chronologically (oldest first).
// This test simulates that behavior.
func TestRetentionWindowProperty(t *testing.T) {
	property := func(count uint8) bool {
		if count == 0 {
			return true
		}

		// Limit count to reasonable number
		n := int(count%50) + 1

		maxAge := 1 * time.Hour // Use 1 hour for faster testing
		store := NewStore("", maxAge)
		baseTime := time.Now()

		// Add snapshots in chronological order (oldest first, like real sampling)
		// This simulates snapshots being added every 5 minutes over time
		for i := n - 1; i >= 0; i-- {
			age := time.Duration(i*5) * time.Minute
			snapTime := baseTime.Add(-age)
			snap := &Snapshot{
				Timestamp: snapTime,
				Items:     map[string]*SnapshotItem{},
			}
			store.Add(snap)
		}

		// Verify all remaining snapshots are within maxAge from now
		now := time.Now()
		cutoff := now.Add(-maxAge)
		for _, snap := range store.All() {
			// Allow tolerance for test execution time (up to 5 seconds)
			if snap.Timestamp.Before(cutoff.Add(-5 * time.Second)) {
				return false // Found a snapshot older than maxAge
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Retention window property failed: %v", err)
	}
}

func TestStoreEmpty(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	if store.Latest() != nil {
		t.Error("Latest should be nil for empty store")
	}
	if store.Previous() != nil {
		t.Error("Previous should be nil for empty store")
	}
	if store.FindSnapshotByTime(time.Now()) != nil {
		t.Error("FindSnapshotByTime should return nil for empty store")
	}
	if store.Count() != 0 {
		t.Errorf("Count = %d, want 0", store.Count())
	}
}

func TestStoreAddNil(t *testing.T) {
	store := NewStore("", 24*time.Hour)
	store.Add(nil)

	if store.Count() != 0 {
		t.Errorf("Count = %d, want 0 (nil should not be added)", store.Count())
	}
}

func TestStoreAutoTimestamp(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	// Add snapshot without timestamp
	snap := &Snapshot{
		Items: map[string]*SnapshotItem{"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1}},
	}
	store.Add(snap)

	// Timestamp should be set automatically
	if snap.Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}
}


// TestGetCurrentBasic tests basic GetCurrent functionality.
func TestGetCurrentBasic(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	// Empty store
	resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})
	if len(resp.Items) != 0 {
		t.Errorf("Expected empty items for empty store, got %d", len(resp.Items))
	}

	// Add first snapshot
	now := time.Now()
	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, TradesRank: 2, Price: 100.0, Volume: 1000, TradeCount: 500},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 2, TradesRank: 1, Price: 50.0, Volume: 800, TradeCount: 600},
		},
	}
	store.Add(snap1)

	// Add second snapshot with rank changes
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 2, TradesRank: 1, Price: 105.0, Volume: 900, TradeCount: 550},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 1, TradesRank: 2, Price: 48.0, Volume: 1100, TradeCount: 580},
		},
	}
	store.Add(snap2)

	// Test volume ranking
	resp = store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})
	if len(resp.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(resp.Items))
	}

	// First item should be ETHUSDT (rank 1 in volume)
	if resp.Items[0].Symbol != "ETHUSDT" {
		t.Errorf("Expected ETHUSDT at rank 1, got %s", resp.Items[0].Symbol)
	}
	if resp.Items[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", resp.Items[0].Rank)
	}
	// ETHUSDT went from rank 2 to rank 1, so change = 2 - 1 = 1 (improved)
	if resp.Items[0].RankChange == nil || *resp.Items[0].RankChange != 1 {
		t.Errorf("Expected rank change 1, got %v", resp.Items[0].RankChange)
	}

	// Second item should be BTCUSDT (rank 2 in volume)
	if resp.Items[1].Symbol != "BTCUSDT" {
		t.Errorf("Expected BTCUSDT at rank 2, got %s", resp.Items[1].Symbol)
	}
	// BTCUSDT went from rank 1 to rank 2, so change = 1 - 2 = -1 (dropped)
	if resp.Items[1].RankChange == nil || *resp.Items[1].RankChange != -1 {
		t.Errorf("Expected rank change -1, got %v", resp.Items[1].RankChange)
	}
}

// TestGetCurrentWithCompare tests GetCurrent with compare duration.
func TestGetCurrentWithCompare(t *testing.T) {
	store := NewStore("", 24*time.Hour)
	now := time.Now()

	// Add snapshots at different times
	snap1 := &Snapshot{
		Timestamp: now.Add(-60 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 100.0},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-30 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 2, Price: 110.0},
		},
	}
	snap3 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 3, Price: 120.0},
		},
	}

	store.Add(snap1)
	store.Add(snap2)
	store.Add(snap3)

	// Compare with 1 hour ago (should use snap1)
	resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume, Compare: 1 * time.Hour})
	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}

	// Current rank is 3, previous was 1, change = 1 - 3 = -2
	if resp.Items[0].RankChange == nil || *resp.Items[0].RankChange != -2 {
		t.Errorf("Expected rank change -2, got %v", resp.Items[0].RankChange)
	}

	// Price change: (120 - 100) / 100 * 100 = 20%
	if resp.Items[0].PriceChange == nil || *resp.Items[0].PriceChange != 20.0 {
		t.Errorf("Expected price change 20.0, got %v", resp.Items[0].PriceChange)
	}
}

// TestGetCurrentNewSymbol tests GetCurrent with new symbols.
func TestGetCurrentNewSymbol(t *testing.T) {
	store := NewStore("", 24*time.Hour)
	now := time.Now()

	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 100.0},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 105.0},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 2, Price: 50.0}, // New symbol
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})

	// Find ETHUSDT
	var ethItem *RankingItem
	for i := range resp.Items {
		if resp.Items[i].Symbol == "ETHUSDT" {
			ethItem = &resp.Items[i]
			break
		}
	}

	if ethItem == nil {
		t.Fatal("ETHUSDT not found in response")
	}

	if !ethItem.IsNew {
		t.Error("Expected ETHUSDT to be marked as new")
	}
	if ethItem.RankChange != nil {
		t.Error("Expected nil rank change for new symbol")
	}
}

// TestRankChangeProperty tests the rank change calculation property.
// Property 5: Rank Change Calculation
// Validates: Requirements 3.1, 3.2, 3.3
func TestRankChangeProperty(t *testing.T) {
	property := func(prevRank, currRank uint8) bool {
		// Ensure valid ranks (1-255)
		if prevRank == 0 {
			prevRank = 1
		}
		if currRank == 0 {
			currRank = 1
		}

		store := NewStore("", 24*time.Hour)
		now := time.Now()

		snap1 := &Snapshot{
			Timestamp: now.Add(-10 * time.Minute),
			Items: map[string]*SnapshotItem{
				"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: int(prevRank), Price: 100.0},
			},
		}
		snap2 := &Snapshot{
			Timestamp: now.Add(-5 * time.Minute),
			Items: map[string]*SnapshotItem{
				"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: int(currRank), Price: 100.0},
			},
		}

		store.Add(snap1)
		store.Add(snap2)

		resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})
		if len(resp.Items) != 1 {
			return false
		}

		expectedChange := int(prevRank) - int(currRank)
		if resp.Items[0].RankChange == nil {
			return false
		}
		return *resp.Items[0].RankChange == expectedChange
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Rank change property failed: %v", err)
	}
}

// TestPriceChangeProperty tests the price change calculation property.
// Property 6: Price Change Calculation
// Validates: Requirements 4.1, 4.2, 4.4
func TestPriceChangeProperty(t *testing.T) {
	property := func(prevPrice, currPrice uint16) bool {
		// Use uint16 to get reasonable price values
		prev := float64(prevPrice) + 0.01 // Ensure non-zero
		curr := float64(currPrice) + 0.01

		store := NewStore("", 24*time.Hour)
		now := time.Now()

		snap1 := &Snapshot{
			Timestamp: now.Add(-10 * time.Minute),
			Items: map[string]*SnapshotItem{
				"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: prev},
			},
		}
		snap2 := &Snapshot{
			Timestamp: now.Add(-5 * time.Minute),
			Items: map[string]*SnapshotItem{
				"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: curr},
			},
		}

		store.Add(snap1)
		store.Add(snap2)

		resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})
		if len(resp.Items) != 1 {
			return false
		}

		expectedChange := ((curr - prev) / prev) * 100
		if resp.Items[0].PriceChange == nil {
			return false
		}

		// Allow small floating point tolerance
		diff := *resp.Items[0].PriceChange - expectedChange
		if diff < 0 {
			diff = -diff
		}
		return diff < 0.0001
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Price change property failed: %v", err)
	}
}

// TestPriceChangeZeroPrevious tests price change when previous price is zero.
func TestPriceChangeZeroPrevious(t *testing.T) {
	store := NewStore("", 24*time.Hour)
	now := time.Now()

	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 0}, // Zero price
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 100.0},
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	resp := store.GetCurrent(CurrentOptions{Type: RankingTypeVolume})
	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}

	// Price change should be nil when previous price is zero
	if resp.Items[0].PriceChange != nil {
		t.Errorf("Expected nil price change for zero previous price, got %v", *resp.Items[0].PriceChange)
	}
}


// TestGetHistoryBasic tests basic GetHistory functionality.
func TestGetHistoryBasic(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	// Empty store
	resp := store.GetHistory("BTCUSDT")
	if resp.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", resp.Symbol)
	}
	if len(resp.Snapshots) != 0 {
		t.Errorf("Expected empty snapshots for empty store, got %d", len(resp.Snapshots))
	}

	// Add snapshots
	now := time.Now()
	snap1 := &Snapshot{
		Timestamp: now.Add(-20 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, TradesRank: 1, Price: 100.0, Volume: 1000, TradeCount: 500},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 2, TradesRank: 2, Price: 50.0, Volume: 800, TradeCount: 400},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 2, TradesRank: 1, Price: 105.0, Volume: 900, TradeCount: 550},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 1, TradesRank: 2, Price: 52.0, Volume: 1100, TradeCount: 420},
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	// Get BTCUSDT history
	resp = store.GetHistory("BTCUSDT")
	if len(resp.Snapshots) != 2 {
		t.Fatalf("Expected 2 snapshots, got %d", len(resp.Snapshots))
	}

	// Verify chronological order (oldest first)
	if !resp.Snapshots[0].Timestamp.Before(resp.Snapshots[1].Timestamp) {
		t.Error("Expected snapshots in chronological order")
	}

	// Verify first snapshot data
	if resp.Snapshots[0].VolumeRank != 1 {
		t.Errorf("Expected volume rank 1, got %d", resp.Snapshots[0].VolumeRank)
	}
	if resp.Snapshots[0].Price != 100.0 {
		t.Errorf("Expected price 100.0, got %f", resp.Snapshots[0].Price)
	}
}

// TestGetHistoryNonExistent tests GetHistory for non-existent symbol.
func TestGetHistoryNonExistent(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	snap := &Snapshot{
		Timestamp: time.Now(),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, Price: 100.0},
		},
	}
	store.Add(snap)

	resp := store.GetHistory("XYZUSDT")
	if resp.Symbol != "XYZUSDT" {
		t.Errorf("Expected symbol XYZUSDT, got %s", resp.Symbol)
	}
	if len(resp.Snapshots) != 0 {
		t.Errorf("Expected empty snapshots for non-existent symbol, got %d", len(resp.Snapshots))
	}
}

// TestHistoryChronologicalProperty tests the history chronological order property.
// Property 7: History Chronological Order
// Validates: Requirements 6.4
func TestHistoryChronologicalProperty(t *testing.T) {
	property := func(count uint8) bool {
		if count < 2 {
			return true
		}
		n := int(count%20) + 2 // 2-21 snapshots

		store := NewStore("", 24*time.Hour)
		baseTime := time.Now()

		// Add snapshots in chronological order
		for i := n - 1; i >= 0; i-- {
			snap := &Snapshot{
				Timestamp: baseTime.Add(-time.Duration(i*5) * time.Minute),
				Items: map[string]*SnapshotItem{
					"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: i + 1, Price: float64(100 + i)},
				},
			}
			store.Add(snap)
		}

		resp := store.GetHistory("BTCUSDT")

		// Verify chronological order
		for i := 1; i < len(resp.Snapshots); i++ {
			if !resp.Snapshots[i-1].Timestamp.Before(resp.Snapshots[i].Timestamp) {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("History chronological property failed: %v", err)
	}
}

// TestGetMoversBasic tests basic GetMovers functionality.
func TestGetMoversBasic(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	// Empty store
	resp := store.GetMovers(MoversOptions{Type: RankingTypeVolume, Direction: DirectionUp})
	if len(resp.Items) != 0 {
		t.Errorf("Expected empty items for empty store, got %d", len(resp.Items))
	}

	now := time.Now()
	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT":  {Symbol: "BTCUSDT", VolumeRank: 1, Price: 100.0},
			"ETHUSDT":  {Symbol: "ETHUSDT", VolumeRank: 2, Price: 50.0},
			"SOLUSDT":  {Symbol: "SOLUSDT", VolumeRank: 3, Price: 25.0},
			"DOGEUSDT": {Symbol: "DOGEUSDT", VolumeRank: 10, Price: 0.1},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT":  {Symbol: "BTCUSDT", VolumeRank: 3, Price: 100.0},  // Dropped 2
			"ETHUSDT":  {Symbol: "ETHUSDT", VolumeRank: 1, Price: 50.0},   // Up 1
			"SOLUSDT":  {Symbol: "SOLUSDT", VolumeRank: 2, Price: 25.0},   // Up 1
			"DOGEUSDT": {Symbol: "DOGEUSDT", VolumeRank: 4, Price: 0.1},   // Up 6
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	// Test up movers
	resp = store.GetMovers(MoversOptions{Type: RankingTypeVolume, Direction: DirectionUp})
	if len(resp.Items) != 3 {
		t.Fatalf("Expected 3 up movers, got %d", len(resp.Items))
	}

	// First should be DOGEUSDT (biggest change: +6)
	if resp.Items[0].Symbol != "DOGEUSDT" {
		t.Errorf("Expected DOGEUSDT as top mover, got %s", resp.Items[0].Symbol)
	}
	if resp.Items[0].RankChange == nil || *resp.Items[0].RankChange != 6 {
		t.Errorf("Expected rank change 6, got %v", resp.Items[0].RankChange)
	}

	// Test down movers
	resp = store.GetMovers(MoversOptions{Type: RankingTypeVolume, Direction: DirectionDown})
	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 down mover, got %d", len(resp.Items))
	}
	if resp.Items[0].Symbol != "BTCUSDT" {
		t.Errorf("Expected BTCUSDT as down mover, got %s", resp.Items[0].Symbol)
	}
	if resp.Items[0].RankChange == nil || *resp.Items[0].RankChange != -2 {
		t.Errorf("Expected rank change -2, got %v", resp.Items[0].RankChange)
	}
}

// TestGetMoversWithLimit tests GetMovers with limit.
func TestGetMoversWithLimit(t *testing.T) {
	store := NewStore("", 24*time.Hour)
	now := time.Now()

	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT":  {Symbol: "BTCUSDT", VolumeRank: 10},
			"ETHUSDT":  {Symbol: "ETHUSDT", VolumeRank: 20},
			"SOLUSDT":  {Symbol: "SOLUSDT", VolumeRank: 30},
			"DOGEUSDT": {Symbol: "DOGEUSDT", VolumeRank: 40},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT":  {Symbol: "BTCUSDT", VolumeRank: 5},  // Up 5
			"ETHUSDT":  {Symbol: "ETHUSDT", VolumeRank: 10}, // Up 10
			"SOLUSDT":  {Symbol: "SOLUSDT", VolumeRank: 15}, // Up 15
			"DOGEUSDT": {Symbol: "DOGEUSDT", VolumeRank: 20}, // Up 20
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	resp := store.GetMovers(MoversOptions{Type: RankingTypeVolume, Direction: DirectionUp, Limit: 2})
	if len(resp.Items) != 2 {
		t.Fatalf("Expected 2 items with limit, got %d", len(resp.Items))
	}

	// Should be DOGEUSDT (20) and SOLUSDT (15)
	if resp.Items[0].Symbol != "DOGEUSDT" {
		t.Errorf("Expected DOGEUSDT first, got %s", resp.Items[0].Symbol)
	}
	if resp.Items[1].Symbol != "SOLUSDT" {
		t.Errorf("Expected SOLUSDT second, got %s", resp.Items[1].Symbol)
	}
}

// TestMoversSortingProperty tests the movers sorting property.
// Property 8: Movers Sorting
// Validates: Requirements 7.5
func TestMoversSortingProperty(t *testing.T) {
	property := func(changes []int8) bool {
		if len(changes) < 2 {
			return true
		}

		store := NewStore("", 24*time.Hour)
		now := time.Now()

		// Build snapshots with specified rank changes
		items1 := make(map[string]*SnapshotItem)
		items2 := make(map[string]*SnapshotItem)

		for i, change := range changes {
			symbol := "SYM" + string(rune('A'+i)) + "USDT"
			prevRank := 50 // Start from middle
			currRank := prevRank - int(change)
			if currRank < 1 {
				currRank = 1
			}
			if currRank > 100 {
				currRank = 100
			}

			items1[symbol] = &SnapshotItem{Symbol: symbol, VolumeRank: prevRank}
			items2[symbol] = &SnapshotItem{Symbol: symbol, VolumeRank: currRank}
		}

		store.Add(&Snapshot{Timestamp: now.Add(-10 * time.Minute), Items: items1})
		store.Add(&Snapshot{Timestamp: now.Add(-5 * time.Minute), Items: items2})

		// Test up movers
		resp := store.GetMovers(MoversOptions{Type: RankingTypeVolume, Direction: DirectionUp})

		// Verify sorted by absolute change descending
		for i := 1; i < len(resp.Items); i++ {
			absI := *resp.Items[i-1].RankChange
			if absI < 0 {
				absI = -absI
			}
			absJ := *resp.Items[i].RankChange
			if absJ < 0 {
				absJ = -absJ
			}
			if absJ > absI {
				return false // Not sorted correctly
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Movers sorting property failed: %v", err)
	}
}
