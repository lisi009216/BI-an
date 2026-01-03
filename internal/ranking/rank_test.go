package ranking

import (
	"math/rand"
	"testing"
	"testing/quick"

	"example.com/binance-pivot-monitor/internal/ticker"
)

// TestCalculateRanksBasic tests basic ranking calculation.
func TestCalculateRanksBasic(t *testing.T) {
	tickers := map[string]*ticker.Ticker{
		"BTCUSDT": {Symbol: "BTCUSDT", QuoteVolume: 1000, TradeCount: 100, LastPrice: 50000},
		"ETHUSDT": {Symbol: "ETHUSDT", QuoteVolume: 500, TradeCount: 200, LastPrice: 3000},
		"SOLUSDT": {Symbol: "SOLUSDT", QuoteVolume: 200, TradeCount: 50, LastPrice: 100},
		"ETHBTC":  {Symbol: "ETHBTC", QuoteVolume: 800, TradeCount: 150, LastPrice: 0.06}, // Non-USDT
	}

	volumeRanks, tradesRanks := CalculateRanks(tickers)

	// Check volume ranks (BTC > ETH > SOL)
	if volumeRanks["BTCUSDT"] != 1 {
		t.Errorf("BTCUSDT volume rank = %d, want 1", volumeRanks["BTCUSDT"])
	}
	if volumeRanks["ETHUSDT"] != 2 {
		t.Errorf("ETHUSDT volume rank = %d, want 2", volumeRanks["ETHUSDT"])
	}
	if volumeRanks["SOLUSDT"] != 3 {
		t.Errorf("SOLUSDT volume rank = %d, want 3", volumeRanks["SOLUSDT"])
	}

	// Check trades ranks (ETH > BTC > SOL)
	if tradesRanks["ETHUSDT"] != 1 {
		t.Errorf("ETHUSDT trades rank = %d, want 1", tradesRanks["ETHUSDT"])
	}
	if tradesRanks["BTCUSDT"] != 2 {
		t.Errorf("BTCUSDT trades rank = %d, want 2", tradesRanks["BTCUSDT"])
	}
	if tradesRanks["SOLUSDT"] != 3 {
		t.Errorf("SOLUSDT trades rank = %d, want 3", tradesRanks["SOLUSDT"])
	}

	// Non-USDT pair should not be in ranks
	if _, ok := volumeRanks["ETHBTC"]; ok {
		t.Error("ETHBTC should not be in volume ranks")
	}
}

// TestCalculateRanksDenseRanking tests dense ranking with equal values.
// Property 3: Equal Value Rank Assignment (Dense Ranking)
// Validates: Requirements 2.4
func TestCalculateRanksDenseRanking(t *testing.T) {
	tickers := map[string]*ticker.Ticker{
		"BTCUSDT":  {Symbol: "BTCUSDT", QuoteVolume: 1000, TradeCount: 100, LastPrice: 50000},
		"ETHUSDT":  {Symbol: "ETHUSDT", QuoteVolume: 1000, TradeCount: 100, LastPrice: 3000}, // Same as BTC
		"SOLUSDT":  {Symbol: "SOLUSDT", QuoteVolume: 500, TradeCount: 50, LastPrice: 100},
		"ADAUSDT":  {Symbol: "ADAUSDT", QuoteVolume: 500, TradeCount: 50, LastPrice: 0.5}, // Same as SOL
		"DOGEUSDT": {Symbol: "DOGEUSDT", QuoteVolume: 200, TradeCount: 25, LastPrice: 0.1},
	}

	volumeRanks, tradesRanks := CalculateRanks(tickers)

	// BTC and ETH should have same rank (1)
	if volumeRanks["BTCUSDT"] != volumeRanks["ETHUSDT"] {
		t.Errorf("BTCUSDT and ETHUSDT should have same volume rank, got %d and %d",
			volumeRanks["BTCUSDT"], volumeRanks["ETHUSDT"])
	}
	if volumeRanks["BTCUSDT"] != 1 {
		t.Errorf("BTCUSDT volume rank = %d, want 1", volumeRanks["BTCUSDT"])
	}

	// SOL and ADA should have same rank (2, not 3 - dense ranking)
	if volumeRanks["SOLUSDT"] != volumeRanks["ADAUSDT"] {
		t.Errorf("SOLUSDT and ADAUSDT should have same volume rank, got %d and %d",
			volumeRanks["SOLUSDT"], volumeRanks["ADAUSDT"])
	}
	if volumeRanks["SOLUSDT"] != 2 {
		t.Errorf("SOLUSDT volume rank = %d, want 2 (dense ranking)", volumeRanks["SOLUSDT"])
	}

	// DOGE should have rank 3 (not 5 - dense ranking)
	if volumeRanks["DOGEUSDT"] != 3 {
		t.Errorf("DOGEUSDT volume rank = %d, want 3 (dense ranking)", volumeRanks["DOGEUSDT"])
	}

	// Same checks for trades ranks
	if tradesRanks["BTCUSDT"] != tradesRanks["ETHUSDT"] {
		t.Errorf("BTCUSDT and ETHUSDT should have same trades rank")
	}
	if tradesRanks["SOLUSDT"] != tradesRanks["ADAUSDT"] {
		t.Errorf("SOLUSDT and ADAUSDT should have same trades rank")
	}
}

// TestCalculateRanksEmpty tests with empty input.
func TestCalculateRanksEmpty(t *testing.T) {
	volumeRanks, tradesRanks := CalculateRanks(nil)
	if len(volumeRanks) != 0 {
		t.Errorf("Expected empty volume ranks, got %d", len(volumeRanks))
	}
	if len(tradesRanks) != 0 {
		t.Errorf("Expected empty trades ranks, got %d", len(tradesRanks))
	}

	volumeRanks, tradesRanks = CalculateRanks(map[string]*ticker.Ticker{})
	if len(volumeRanks) != 0 {
		t.Errorf("Expected empty volume ranks, got %d", len(volumeRanks))
	}
	if len(tradesRanks) != 0 {
		t.Errorf("Expected empty trades ranks, got %d", len(tradesRanks))
	}
}

// TestRankingOrderProperty tests that ranks are correctly ordered.
// Property 2: Ranking Order Correctness
// Validates: Requirements 2.1, 2.2, 2.3
func TestRankingOrderProperty(t *testing.T) {
	// Property: For any set of tickers, higher volume should have lower (better) rank
	property := func(volumes []uint16) bool {
		if len(volumes) < 2 {
			return true
		}

		// Create tickers with random volumes
		tickers := make(map[string]*ticker.Ticker)
		symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "ADAUSDT", "DOGEUSDT",
			"XRPUSDT", "DOTUSDT", "LINKUSDT", "AVAXUSDT", "MATICUSDT"}

		for i, vol := range volumes {
			if i >= len(symbols) {
				break
			}
			tickers[symbols[i]] = &ticker.Ticker{
				Symbol:      symbols[i],
				QuoteVolume: float64(vol),
				TradeCount:  int64(vol),
				LastPrice:   float64(vol),
			}
		}

		volumeRanks, _ := CalculateRanks(tickers)

		// Verify: for any two symbols, higher volume should have lower or equal rank
		for sym1, t1 := range tickers {
			for sym2, t2 := range tickers {
				if sym1 == sym2 {
					continue
				}
				rank1, ok1 := volumeRanks[sym1]
				rank2, ok2 := volumeRanks[sym2]
				if !ok1 || !ok2 {
					continue
				}

				if t1.QuoteVolume > t2.QuoteVolume && rank1 > rank2 {
					return false // Higher volume should have lower rank
				}
				if t1.QuoteVolume == t2.QuoteVolume && rank1 != rank2 {
					return false // Equal volume should have equal rank
				}
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Ranking order property failed: %v", err)
	}
}

// TestBuildSnapshot tests snapshot building.
func TestBuildSnapshot(t *testing.T) {
	tickers := map[string]*ticker.Ticker{
		"BTCUSDT": {Symbol: "BTCUSDT", QuoteVolume: 1000, TradeCount: 100, LastPrice: 50000},
		"ETHUSDT": {Symbol: "ETHUSDT", QuoteVolume: 500, TradeCount: 200, LastPrice: 3000},
		"ETHBTC":  {Symbol: "ETHBTC", QuoteVolume: 800, TradeCount: 150, LastPrice: 0.06}, // Non-USDT
	}

	snapshot := BuildSnapshot(tickers)

	// Should only have USDT pairs
	if len(snapshot.Items) != 2 {
		t.Errorf("Expected 2 items in snapshot, got %d", len(snapshot.Items))
	}

	// Check BTC item
	btc, ok := snapshot.Items["BTCUSDT"]
	if !ok {
		t.Fatal("BTCUSDT not found in snapshot")
	}
	if btc.VolumeRank != 1 {
		t.Errorf("BTCUSDT volume rank = %d, want 1", btc.VolumeRank)
	}
	if btc.Price != 50000 {
		t.Errorf("BTCUSDT price = %f, want 50000", btc.Price)
	}

	// Non-USDT should not be in snapshot
	if _, ok := snapshot.Items["ETHBTC"]; ok {
		t.Error("ETHBTC should not be in snapshot")
	}
}

// TestSnapshotOnlyContainsUSDTPairs tests that snapshots only contain USDT pairs.
// Property 1: USDT Pair Filtering (at snapshot level)
// Validates: Requirements 1.1, 1.2, 2.5
func TestSnapshotOnlyContainsUSDTPairs(t *testing.T) {
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))

		// Generate random tickers with mixed pairs
		allSymbols := []string{
			"BTCUSDT", "ETHUSDT", "SOLUSDT", // USDT pairs
			"ETHBTC", "SOLBTC", "ADABTC", // BTC pairs
			"SOLETH", "ADAETH", // ETH pairs
			"BTCFDUSD", "ETHFDUSD", // FDUSD pairs
		}

		tickers := make(map[string]*ticker.Ticker)
		for _, sym := range allSymbols {
			if rng.Float32() > 0.3 { // 70% chance to include
				tickers[sym] = &ticker.Ticker{
					Symbol:      sym,
					QuoteVolume: rng.Float64() * 1000000,
					TradeCount:  rng.Int63n(100000),
					LastPrice:   rng.Float64() * 100000,
				}
			}
		}

		snapshot := BuildSnapshot(tickers)

		// Verify all items in snapshot are USDT pairs
		for symbol := range snapshot.Items {
			if !IsUSDTPair(symbol) {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Snapshot USDT filtering property failed: %v", err)
	}
}
