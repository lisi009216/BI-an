package ranking

import (
	"sync"
	"time"
)

const (
	// DefaultMaxAge is the default retention period for snapshots (24 hours).
	DefaultMaxAge = 24 * time.Hour
)

// Store stores and manages ranking snapshots.
type Store struct {
	mu        sync.RWMutex
	snapshots []*Snapshot // Ordered by timestamp, newest at the end
	maxAge    time.Duration
	dataDir   string
}

// NewStore creates a new ranking store.
// dataDir: directory for persistence (e.g., "/path/to/data")
// maxAge: maximum age for snapshots (0 uses DefaultMaxAge)
func NewStore(dataDir string, maxAge time.Duration) *Store {
	if maxAge <= 0 {
		maxAge = DefaultMaxAge
	}
	return &Store{
		snapshots: make([]*Snapshot, 0),
		maxAge:    maxAge,
		dataDir:   dataDir,
	}
}

// Add adds a snapshot to the store and triggers cleanup.
func (s *Store) Add(snapshot *Snapshot) {
	if snapshot == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Set timestamp if not set
	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now()
	}

	s.snapshots = append(s.snapshots, snapshot)
	s.cleanupLocked()
}

// cleanup removes snapshots older than maxAge.
// Must be called with lock held.
func (s *Store) cleanupLocked() {
	if len(s.snapshots) == 0 {
		return
	}

	cutoff := time.Now().Add(-s.maxAge)
	firstValid := 0

	for i, snap := range s.snapshots {
		if snap.Timestamp.After(cutoff) || snap.Timestamp.Equal(cutoff) {
			firstValid = i
			break
		}
		firstValid = i + 1
	}

	if firstValid > 0 {
		s.snapshots = s.snapshots[firstValid:]
	}
}

// Cleanup removes expired snapshots (public method).
func (s *Store) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
}

// Latest returns the most recent snapshot, or nil if none.
func (s *Store) Latest() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) == 0 {
		return nil
	}
	return s.snapshots[len(s.snapshots)-1]
}

// Previous returns the second most recent snapshot, or nil if none.
func (s *Store) Previous() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) < 2 {
		return nil
	}
	return s.snapshots[len(s.snapshots)-2]
}

// FindSnapshotByTime finds the snapshot with timestamp closest to but not exceeding targetTime.
// If no snapshot exists with timestamp <= targetTime, returns the oldest snapshot.
// Returns nil if no snapshots exist.
func (s *Store) FindSnapshotByTime(targetTime time.Time) *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) == 0 {
		return nil
	}

	// Find the latest snapshot with timestamp <= targetTime
	var result *Snapshot
	for i := len(s.snapshots) - 1; i >= 0; i-- {
		snap := s.snapshots[i]
		if snap.Timestamp.Before(targetTime) || snap.Timestamp.Equal(targetTime) {
			result = snap
			break
		}
	}

	// If no snapshot found with timestamp <= targetTime, return the oldest
	if result == nil {
		result = s.snapshots[0]
	}

	return result
}

// Count returns the number of snapshots in the store.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.snapshots)
}

// All returns all snapshots (for testing/debugging).
func (s *Store) All() []*Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Snapshot, len(s.snapshots))
	copy(result, s.snapshots)
	return result
}

// DataDir returns the data directory path.
func (s *Store) DataDir() string {
	return s.dataDir
}

// GetCurrent returns the current ranking with rank changes calculated.
func (s *Store) GetCurrent(opts CurrentOptions) *CurrentResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) == 0 {
		return &CurrentResponse{Items: []RankingItem{}}
	}

	// Get the latest snapshot
	current := s.snapshots[len(s.snapshots)-1]

	// Find comparison snapshot
	var compare *Snapshot
	if opts.Compare > 0 {
		// Find snapshot at targetTime = now - compare duration
		targetTime := current.Timestamp.Add(-opts.Compare)
		compare = s.findSnapshotByTimeLocked(targetTime)
	} else {
		// Use previous snapshot
		if len(s.snapshots) >= 2 {
			compare = s.snapshots[len(s.snapshots)-2]
		}
	}

	// Build response items
	items := s.buildRankingItems(current, compare, opts.Type)

	// Sort by rank
	sortRankingItemsByRank(items)

	// Apply limit
	if opts.Limit > 0 && len(items) > opts.Limit {
		items = items[:opts.Limit]
	}

	resp := &CurrentResponse{
		Timestamp: current.Timestamp,
		Items:     items,
	}
	if compare != nil {
		resp.CompareTo = compare.Timestamp
	}

	return resp
}

// findSnapshotByTimeLocked finds snapshot by time (must hold read lock).
func (s *Store) findSnapshotByTimeLocked(targetTime time.Time) *Snapshot {
	if len(s.snapshots) == 0 {
		return nil
	}

	// Find the latest snapshot with timestamp <= targetTime
	var result *Snapshot
	for i := len(s.snapshots) - 1; i >= 0; i-- {
		snap := s.snapshots[i]
		if snap.Timestamp.Before(targetTime) || snap.Timestamp.Equal(targetTime) {
			result = snap
			break
		}
	}

	// If no snapshot found with timestamp <= targetTime, return the oldest
	if result == nil {
		result = s.snapshots[0]
	}

	return result
}

// buildRankingItems builds ranking items from current and compare snapshots.
func (s *Store) buildRankingItems(current, compare *Snapshot, rankType string) []RankingItem {
	items := make([]RankingItem, 0, len(current.Items))

	for symbol, item := range current.Items {
		ri := RankingItem{
			Symbol:     symbol,
			Price:      item.Price,
			Volume:     item.Volume,
			TradeCount: item.TradeCount,
		}

		// Set rank based on type
		if rankType == RankingTypeTrades {
			ri.Rank = item.TradesRank
		} else {
			ri.Rank = item.VolumeRank
		}

		// Calculate changes if we have a comparison snapshot
		if compare != nil {
			if prevItem, exists := compare.Items[symbol]; exists {
				// Calculate rank change (positive = improved = lower rank number)
				var prevRank int
				if rankType == RankingTypeTrades {
					prevRank = prevItem.TradesRank
				} else {
					prevRank = prevItem.VolumeRank
				}
				rankChange := prevRank - ri.Rank
				ri.RankChange = &rankChange

				// Calculate price change percentage
				if prevItem.Price > 0 {
					priceChange := ((item.Price - prevItem.Price) / prevItem.Price) * 100
					ri.PriceChange = &priceChange
				}

				// Calculate volume change percentage
				if prevItem.Volume > 0 {
					volumeChange := ((item.Volume - prevItem.Volume) / prevItem.Volume) * 100
					ri.VolumeChange = &volumeChange
				}

				// Calculate trade count change percentage
				if prevItem.TradeCount > 0 {
					tradeChange := (float64(item.TradeCount-prevItem.TradeCount) / float64(prevItem.TradeCount)) * 100
					ri.TradeChange = &tradeChange
				}
			} else {
				// Symbol is new (not in comparison snapshot)
				ri.IsNew = true
			}
		}

		items = append(items, ri)
	}

	return items
}

// sortRankingItemsByRank sorts items by rank in ascending order.
func sortRankingItemsByRank(items []RankingItem) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Rank < items[i].Rank {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// GetHistory returns the history of a specific symbol.
// Returns HistoryResponse with empty snapshots array if symbol not found.
func (s *Store) GetHistory(symbol string) *HistoryResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &HistoryResponse{
		Symbol:    symbol,
		Snapshots: []SymbolSnapshot{},
	}

	// Iterate through all snapshots in chronological order (oldest first)
	for _, snap := range s.snapshots {
		if item, exists := snap.Items[symbol]; exists {
			resp.Snapshots = append(resp.Snapshots, SymbolSnapshot{
				Timestamp:  snap.Timestamp,
				VolumeRank: item.VolumeRank,
				TradesRank: item.TradesRank,
				Price:      item.Price,
				Volume:     item.Volume,
				TradeCount: item.TradeCount,
			})
		}
	}

	return resp
}

// GetMovers returns symbols with the largest rank changes.
func (s *Store) GetMovers(opts MoversOptions) *MoversResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := &MoversResponse{
		Direction: opts.Direction,
		Items:     []RankingItem{},
	}

	if len(s.snapshots) == 0 {
		return resp
	}

	// Get the latest snapshot
	current := s.snapshots[len(s.snapshots)-1]
	resp.Timestamp = current.Timestamp

	// Find comparison snapshot
	var compare *Snapshot
	if opts.Compare > 0 {
		targetTime := current.Timestamp.Add(-opts.Compare)
		compare = s.findSnapshotByTimeLocked(targetTime)
	} else {
		if len(s.snapshots) >= 2 {
			compare = s.snapshots[len(s.snapshots)-2]
		}
	}

	if compare == nil {
		return resp
	}
	resp.CompareTo = compare.Timestamp

	// Build ranking items with changes
	items := s.buildRankingItems(current, compare, opts.Type)

	// Filter by direction and collect movers
	var movers []RankingItem
	for _, item := range items {
		if item.RankChange == nil {
			continue // Skip new symbols
		}
		change := *item.RankChange
		if opts.Direction == DirectionUp && change > 0 {
			movers = append(movers, item)
		} else if opts.Direction == DirectionDown && change < 0 {
			movers = append(movers, item)
		}
	}

	// Sort by absolute rank change in descending order
	sortRankingItemsByAbsChange(movers)

	// Apply limit
	if opts.Limit > 0 && len(movers) > opts.Limit {
		movers = movers[:opts.Limit]
	}

	resp.Items = movers
	return resp
}

// sortRankingItemsByAbsChange sorts items by absolute rank change in descending order.
func sortRankingItemsByAbsChange(items []RankingItem) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			absI := *items[i].RankChange
			if absI < 0 {
				absI = -absI
			}
			absJ := *items[j].RankChange
			if absJ < 0 {
				absJ = -absJ
			}
			if absJ > absI {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
