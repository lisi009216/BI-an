package ranking

import (
	"sort"

	"example.com/binance-pivot-monitor/internal/ticker"
)

// tickerItem holds ticker data for sorting
type tickerItem struct {
	Symbol     string
	Volume     float64
	TradeCount int64
	Price      float64
}

// CalculateRanks calculates volume and trades ranks for USDT pairs.
// Uses dense ranking: equal values get the same rank, next distinct value gets rank+1.
// Returns two maps: volumeRanks and tradesRanks (symbol -> rank).
func CalculateRanks(tickers map[string]*ticker.Ticker) (volumeRanks, tradesRanks map[string]int) {
	// Filter USDT pairs
	var items []tickerItem
	for symbol, t := range tickers {
		if !IsUSDTPair(symbol) {
			continue
		}
		items = append(items, tickerItem{
			Symbol:     symbol,
			Volume:     t.QuoteVolume,
			TradeCount: t.TradeCount,
			Price:      t.LastPrice,
		})
	}

	volumeRanks = calculateDenseRanks(items, func(a, b tickerItem) bool {
		return a.Volume > b.Volume // Descending order
	}, func(a, b tickerItem) bool {
		return a.Volume == b.Volume
	})

	tradesRanks = calculateDenseRanks(items, func(a, b tickerItem) bool {
		return a.TradeCount > b.TradeCount // Descending order
	}, func(a, b tickerItem) bool {
		return a.TradeCount == b.TradeCount
	})

	return volumeRanks, tradesRanks
}

// calculateDenseRanks calculates dense ranks for items.
// less: comparison function for sorting (descending order)
// equal: function to check if two items have equal values
func calculateDenseRanks(items []tickerItem, less func(a, b tickerItem) bool, equal func(a, b tickerItem) bool) map[string]int {
	if len(items) == 0 {
		return make(map[string]int)
	}

	// Make a copy to avoid modifying the original
	sorted := make([]tickerItem, len(items))
	copy(sorted, items)

	// Sort in descending order
	sort.Slice(sorted, func(i, j int) bool {
		return less(sorted[i], sorted[j])
	})

	ranks := make(map[string]int, len(sorted))
	currentRank := 1

	for i, item := range sorted {
		if i > 0 && !equal(sorted[i-1], item) {
			// Different value from previous, increment rank
			currentRank++
		}
		ranks[item.Symbol] = currentRank
	}

	return ranks
}

// BuildSnapshot creates a snapshot from ticker data.
// It automatically calculates volume and trades ranks.
func BuildSnapshot(tickers map[string]*ticker.Ticker) *Snapshot {
	volumeRanks, tradesRanks := CalculateRanks(tickers)

	items := make(map[string]*SnapshotItem)

	for symbol, t := range tickers {
		if !IsUSDTPair(symbol) {
			continue
		}

		volRank, hasVolRank := volumeRanks[symbol]
		tradeRank, hasTradeRank := tradesRanks[symbol]

		if !hasVolRank || !hasTradeRank {
			continue
		}

		items[symbol] = &SnapshotItem{
			Symbol:     symbol,
			VolumeRank: volRank,
			TradesRank: tradeRank,
			Price:      t.LastPrice,
			Volume:     t.QuoteVolume,
			TradeCount: t.TradeCount,
		}
	}

	return &Snapshot{
		Items: items,
	}
}
