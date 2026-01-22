package kline

import (
	"log"
	"sync"
	"time"
)

// SymbolKlines holds kline data for a single trading pair.
type SymbolKlines struct {
	Symbol   string
	Current  *Kline  // Current forming kline
	History  []Kline // Completed historical klines (oldest first, newest last)
	LastSeen time.Time
}

// Store manages kline data for all trading pairs.
type Store struct {
	mu       sync.RWMutex
	klines   map[string]*SymbolKlines
	interval time.Duration
	maxCount int
	onClose  func(symbol string, klines []Kline)
	stopCh   chan struct{}
}

// DefaultKlineCount is the default number of klines to maintain per symbol.
// Set to 20 to support pattern detection (max requirement is 14 klines).
const DefaultKlineCount = 20

// NewStore creates a new kline store.
// interval: kline interval (e.g., 5 * time.Minute)
// maxCount: maximum number of historical klines to keep per symbol
func NewStore(interval time.Duration, maxCount int) *Store {
	// 参数校验：防止负数或零导致 panic
	if maxCount <= 0 {
		log.Printf("WARN: invalid KLINE_COUNT=%d, using default %d", maxCount, DefaultKlineCount)
		maxCount = DefaultKlineCount
	}
	return &Store{
		klines:   make(map[string]*SymbolKlines),
		interval: interval,
		maxCount: maxCount,
		stopCh:   make(chan struct{}),
	}
}

// SetOnClose sets the callback function called when a kline closes.
// The callback receives a deep copy snapshot of klines, safe for async use.
func (s *Store) SetOnClose(fn func(symbol string, klines []Kline)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onClose = fn
}

// StartCloseTimer starts a timer that triggers kline close at interval boundaries.
// This ensures all symbols close their klines at the same time (e.g., 00, 15, 30, 45 for 15m interval).
func (s *Store) StartCloseTimer() {
	go func() {
		for {
			// Calculate time until next interval boundary
			now := time.Now().UTC()
			currentOpen := getKlineOpenTime(now, s.interval)
			nextClose := getKlineCloseTime(currentOpen, s.interval)
			// Add a small delay (2 seconds) to ensure we're past the boundary
			sleepDuration := nextClose.Sub(now) + 2*time.Second
			if sleepDuration < 0 {
				sleepDuration = s.interval
			}

			log.Printf("kline: timer scheduled - now=%s currentOpen=%s nextClose=%s sleepFor=%v",
				now.Format("15:04:05"), currentOpen.Format("15:04:05"), nextClose.Format("15:04:05"), sleepDuration.Round(time.Second))

			select {
			case <-s.stopCh:
				return
			case <-time.After(sleepDuration):
				s.closeAllKlines()
			}
		}
	}()
}

// StopCloseTimer stops the close timer.
func (s *Store) StopCloseTimer() {
	close(s.stopCh)
}

// closeAllKlines forces all current klines to close.
// Called by the timer at interval boundaries.
func (s *Store) closeAllKlines() {
	s.mu.Lock()

	now := time.Now().UTC()
	// Calculate the previous interval's close time (which is the current interval's open time)
	currentOpen := getKlineOpenTime(now, s.interval)
	previousClose := currentOpen // The previous kline should close at currentOpen

	log.Printf("kline: closeAllKlines triggered at %s, tracking %d symbols, previousClose=%s",
		now.Format("15:04:05"), len(s.klines), previousClose.Format("15:04:05"))

	var toClose []struct {
		symbol   string
		snapshot []Kline
	}

	skippedNoKline := 0
	skippedNotReady := 0

	for symbol, sk := range s.klines {
		if sk.Current == nil {
			skippedNoKline++
			continue
		}

		// Check if current kline belongs to the previous interval
		// A kline should be closed if its OpenTime is before the current interval's open time
		if !sk.Current.OpenTime.Before(currentOpen) {
			skippedNotReady++
			continue
		}

		// Close current kline
		sk.Current.IsClosed = true
		sk.Current.CloseTime = currentOpen // Close time is the start of the new interval

		// Append to history
		sk.History = append(sk.History, *sk.Current)

		// Maintain rolling window size
		if len(sk.History) > s.maxCount {
			sk.History = sk.History[len(sk.History)-s.maxCount:]
		}

		// Create deep copy snapshot for callback
		snapshot := make([]Kline, len(sk.History))
		copy(snapshot, sk.History)

		toClose = append(toClose, struct {
			symbol   string
			snapshot []Kline
		}{symbol, snapshot})

		// Create new kline
		sk.Current = &Kline{
			Symbol:   symbol,
			Open:     sk.Current.Close, // Use last close as new open
			High:     sk.Current.Close,
			Low:      sk.Current.Close,
			Close:    sk.Current.Close,
			OpenTime: currentOpen,
		}
	}

	onClose := s.onClose
	s.mu.Unlock()

	// Log statistics
	if skippedNoKline > 0 || skippedNotReady > 0 {
		log.Printf("kline: skipped %d (no current), %d (not ready)", skippedNoKline, skippedNotReady)
	}

	// Call callbacks outside lock
	if onClose != nil {
		for _, item := range toClose {
			go onClose(item.symbol, item.snapshot)
		}
	}

	if len(toClose) > 0 {
		log.Printf("kline: closed %d symbols at interval boundary", len(toClose))
	}
}

// getKlineOpenTime calculates the kline open time aligned to interval boundary.
// For 5-minute intervals: 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55
func getKlineOpenTime(ts time.Time, interval time.Duration) time.Time {
	intervalMinutes := int(interval.Minutes())
	if intervalMinutes == 0 {
		intervalMinutes = 1
	}
	minute := ts.Minute()
	alignedMinute := (minute / intervalMinutes) * intervalMinutes
	return time.Date(
		ts.Year(), ts.Month(), ts.Day(),
		ts.Hour(), alignedMinute, 0, 0,
		ts.Location(),
	)
}

// getKlineCloseTime calculates the kline close time.
func getKlineCloseTime(openTime time.Time, interval time.Duration) time.Time {
	return openTime.Add(interval)
}

// getOrCreate returns the SymbolKlines for a symbol, creating if needed.
func (s *Store) getOrCreate(symbol string) *SymbolKlines {
	sk, ok := s.klines[symbol]
	if !ok {
		sk = &SymbolKlines{
			Symbol:  symbol,
			History: make([]Kline, 0, s.maxCount),
		}
		s.klines[symbol] = sk
	}
	return sk
}

// shouldClose checks if the current kline should be closed based on timestamp.
func shouldClose(current *Kline, ts time.Time, interval time.Duration) bool {
	if current == nil {
		return false
	}
	closeTime := getKlineCloseTime(current.OpenTime, interval)
	return !ts.Before(closeTime)
}

// Update updates the kline data with a new price.
// Returns true if a kline was closed.
func (s *Store) Update(symbol string, price float64, ts time.Time) bool {
	if price <= 0 {
		return false
	}

	s.mu.Lock()

	sk := s.getOrCreate(symbol)
	sk.LastSeen = ts

	// Check if we need to close the current kline
	if shouldClose(sk.Current, ts, s.interval) {
		// Close current kline
		sk.Current.IsClosed = true
		sk.Current.CloseTime = getKlineCloseTime(sk.Current.OpenTime, s.interval)

		// Append to history (oldest first, newest last)
		sk.History = append(sk.History, *sk.Current)

		// Maintain rolling window size
		if len(sk.History) > s.maxCount {
			sk.History = sk.History[len(sk.History)-s.maxCount:]
		}

		// Create deep copy snapshot for callback
		snapshot := make([]Kline, len(sk.History))
		copy(snapshot, sk.History)

		// Create new kline
		openTime := getKlineOpenTime(ts, s.interval)
		sk.Current = &Kline{
			Symbol:   symbol,
			Open:     price,
			High:     price,
			Low:      price,
			Close:    price,
			OpenTime: openTime,
		}

		// Get callback reference while holding lock
		onClose := s.onClose

		s.mu.Unlock()

		// Call callback outside lock to avoid deadlock
		if onClose != nil {
			go onClose(symbol, snapshot)
		}

		return true
	}

	// Initialize current kline if needed
	if sk.Current == nil {
		openTime := getKlineOpenTime(ts, s.interval)
		sk.Current = &Kline{
			Symbol:   symbol,
			Open:     price,
			High:     price,
			Low:      price,
			Close:    price,
			OpenTime: openTime,
		}
		s.mu.Unlock()
		return false
	}

	// Update current kline OHLC
	if price > sk.Current.High {
		sk.Current.High = price
	}
	if price < sk.Current.Low {
		sk.Current.Low = price
	}
	sk.Current.Close = price

	s.mu.Unlock()
	return false
}

// GetKlines returns a deep copy of historical klines for a symbol.
// Returns klines in time order (oldest first, newest last).
func (s *Store) GetKlines(symbol string) ([]Kline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sk, ok := s.klines[symbol]
	if !ok || len(sk.History) == 0 {
		return nil, false
	}

	// Deep copy
	result := make([]Kline, len(sk.History))
	copy(result, sk.History)
	return result, true
}

// GetCurrentKline returns a deep copy of the current forming kline.
func (s *Store) GetCurrentKline(symbol string) (*Kline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sk, ok := s.klines[symbol]
	if !ok || sk.Current == nil {
		return nil, false
	}

	clone := sk.Current.Clone()
	return &clone, true
}

// GetAllKlines returns historical klines plus current kline (deep copy).
func (s *Store) GetAllKlines(symbol string) ([]Kline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sk, ok := s.klines[symbol]
	if !ok {
		return nil, false
	}

	result := make([]Kline, 0, len(sk.History)+1)
	result = append(result, sk.History...)
	if sk.Current != nil {
		result = append(result, sk.Current.Clone())
	}

	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// CleanupStale removes symbols that haven't been updated for staleThreshold.
// Returns the number of symbols removed.
func (s *Store) CleanupStale(staleThreshold time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	removed := 0

	for symbol, sk := range s.klines {
		if now.Sub(sk.LastSeen) > staleThreshold {
			delete(s.klines, symbol)
			removed++
		}
	}

	return removed
}

// SymbolCount returns the number of symbols being tracked.
func (s *Store) SymbolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.klines)
}

// KlineCount returns the number of historical klines for a symbol.
func (s *Store) KlineCount(symbol string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sk, ok := s.klines[symbol]
	if !ok {
		return 0
	}
	return len(sk.History)
}

// StoreStats contains statistics about the kline store.
type StoreStats struct {
	Enabled      bool              `json:"enabled"`
	SymbolCount  int               `json:"symbol_count"`
	Interval     string            `json:"interval"`
	MaxCount     int               `json:"max_count"`
	Symbols      []SymbolStats     `json:"symbols,omitempty"`
}

// SymbolStats contains statistics for a single symbol.
type SymbolStats struct {
	Symbol       string    `json:"symbol"`
	KlineCount   int       `json:"kline_count"`
	HasCurrent   bool      `json:"has_current"`
	LastSeen     time.Time `json:"last_seen"`
	CurrentOpen  float64   `json:"current_open,omitempty"`
	CurrentClose float64   `json:"current_close,omitempty"`
}

// Stats returns statistics about the kline store.
func (s *Store) Stats() StoreStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := StoreStats{
		Enabled:     true,
		SymbolCount: len(s.klines),
		Interval:    s.interval.String(),
		MaxCount:    s.maxCount,
		Symbols:     make([]SymbolStats, 0, len(s.klines)),
	}

	for symbol, sk := range s.klines {
		ss := SymbolStats{
			Symbol:     symbol,
			KlineCount: len(sk.History),
			HasCurrent: sk.Current != nil,
			LastSeen:   sk.LastSeen,
		}
		if sk.Current != nil {
			ss.CurrentOpen = sk.Current.Open
			ss.CurrentClose = sk.Current.Close
		}
		stats.Symbols = append(stats.Symbols, ss)
	}

	return stats
}
