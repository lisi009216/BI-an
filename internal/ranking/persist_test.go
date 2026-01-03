package ranking

import (
	"os"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"
)

func TestPersistAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ranking-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create store with data
	store := NewStore(tmpDir, 24*time.Hour)
	now := time.Now()

	snap1 := &Snapshot{
		Timestamp: now.Add(-10 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, TradesRank: 1, Price: 100.0, Volume: 1000, TradeCount: 500},
		},
	}
	snap2 := &Snapshot{
		Timestamp: now.Add(-5 * time.Minute),
		Items: map[string]*SnapshotItem{
			"BTCUSDT": {Symbol: "BTCUSDT", VolumeRank: 1, TradesRank: 1, Price: 105.0, Volume: 1100, TradeCount: 550},
			"ETHUSDT": {Symbol: "ETHUSDT", VolumeRank: 2, TradesRank: 2, Price: 50.0, Volume: 800, TradeCount: 400},
		},
	}

	store.Add(snap1)
	store.Add(snap2)

	// Persist
	if err := store.Persist(); err != nil {
		t.Fatalf("Persist failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, "ranking", "snapshots.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Snapshots file not created")
	}

	// Create new store and load
	store2 := NewStore(tmpDir, 24*time.Hour)
	if err := store2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data
	if store2.Count() != 2 {
		t.Errorf("Expected 2 snapshots after load, got %d", store2.Count())
	}

	latest := store2.Latest()
	if latest == nil {
		t.Fatal("Latest snapshot is nil after load")
	}

	if len(latest.Items) != 2 {
		t.Errorf("Expected 2 items in latest snapshot, got %d", len(latest.Items))
	}

	btc, ok := latest.Items["BTCUSDT"]
	if !ok {
		t.Fatal("BTCUSDT not found in loaded snapshot")
	}
	if btc.Price != 105.0 {
		t.Errorf("Expected price 105.0, got %f", btc.Price)
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ranking-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(tmpDir, 24*time.Hour)

	// Load should not error on non-existent file
	if err := store.Load(); err != nil {
		t.Errorf("Load should not error on non-existent file: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Expected 0 snapshots, got %d", store.Count())
	}
}

func TestPersistNoDataDir(t *testing.T) {
	store := NewStore("", 24*time.Hour)

	snap := &Snapshot{
		Timestamp: time.Now(),
		Items:     map[string]*SnapshotItem{},
	}
	store.Add(snap)

	// Persist should be no-op with empty dataDir
	if err := store.Persist(); err != nil {
		t.Errorf("Persist should not error with empty dataDir: %v", err)
	}
}

// TestPersistenceRoundTripProperty tests the persistence round trip property.
// Property 9: Persistence Round Trip
// Validates: Requirements 10.1, 10.2
func TestPersistenceRoundTripProperty(t *testing.T) {
	property := func(count uint8) bool {
		if count == 0 {
			return true
		}
		n := int(count%10) + 1 // 1-10 snapshots

		tmpDir, err := os.MkdirTemp("", "ranking-test-*")
		if err != nil {
			return false
		}
		defer os.RemoveAll(tmpDir)

		store := NewStore(tmpDir, 24*time.Hour)
		baseTime := time.Now()

		// Add snapshots
		for i := n - 1; i >= 0; i-- {
			snap := &Snapshot{
				Timestamp: baseTime.Add(-time.Duration(i*5) * time.Minute),
				Items: map[string]*SnapshotItem{
					"BTCUSDT": {
						Symbol:     "BTCUSDT",
						VolumeRank: i + 1,
						TradesRank: i + 1,
						Price:      float64(100 + i),
						Volume:     float64(1000 + i*100),
						TradeCount: int64(500 + i*50),
					},
				},
			}
			store.Add(snap)
		}

		originalCount := store.Count()

		// Persist
		if err := store.Persist(); err != nil {
			return false
		}

		// Load into new store
		store2 := NewStore(tmpDir, 24*time.Hour)
		if err := store2.Load(); err != nil {
			return false
		}

		// Verify count matches
		if store2.Count() != originalCount {
			return false
		}

		// Verify latest snapshot data
		orig := store.Latest()
		loaded := store2.Latest()

		if orig == nil || loaded == nil {
			return false
		}

		if len(orig.Items) != len(loaded.Items) {
			return false
		}

		for symbol, origItem := range orig.Items {
			loadedItem, ok := loaded.Items[symbol]
			if !ok {
				return false
			}
			if origItem.VolumeRank != loadedItem.VolumeRank ||
				origItem.TradesRank != loadedItem.TradesRank ||
				origItem.Price != loadedItem.Price {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Persistence round trip property failed: %v", err)
	}
}
