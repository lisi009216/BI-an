package ranking

import (
	"context"
	"log"
	"time"

	"example.com/binance-pivot-monitor/internal/ticker"
)

const (
	// DefaultSampleInterval is the default sampling interval (5 minutes).
	DefaultSampleInterval = 5 * time.Minute
)

// Sampler samples ticker data and builds ranking snapshots.
type Sampler struct {
	tickerStore  *ticker.Store
	rankingStore *Store
	interval     time.Duration
}

// NewSampler creates a new ranking sampler.
func NewSampler(tickerStore *ticker.Store, rankingStore *Store) *Sampler {
	return &Sampler{
		tickerStore:  tickerStore,
		rankingStore: rankingStore,
		interval:     DefaultSampleInterval,
	}
}

// SetInterval sets the sampling interval.
func (s *Sampler) SetInterval(interval time.Duration) {
	if interval > 0 {
		s.interval = interval
	}
}

// Run starts the sampling loop.
func (s *Sampler) Run(ctx context.Context) {
	// Do an initial sample; if no data yet, wait for ticker data and try again.
	if s.Sample() == nil {
		s.waitForTickerData(ctx, 2*time.Second)
		s.Sample()
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Sample()
		}
	}
}

// waitForTickerData waits until ticker data is available or context is canceled.
func (s *Sampler) waitForTickerData(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(s.tickerStore.GetAll()) > 0 {
				return
			}
		}
	}
}

// Sample executes a single sampling cycle.
func (s *Sampler) Sample() *Snapshot {
	tickers := s.tickerStore.GetAll()
	if len(tickers) == 0 {
		log.Printf("ranking sampler: no ticker data available, skipping")
		return nil
	}

	snapshot := BuildSnapshot(tickers)
	if snapshot == nil || len(snapshot.Items) == 0 {
		log.Printf("ranking sampler: no USDT pairs found, skipping")
		return nil
	}

	s.rankingStore.Add(snapshot)
	log.Printf("ranking sampler: snapshot added with %d USDT pairs", len(snapshot.Items))

	return snapshot
}
