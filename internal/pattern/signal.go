package pattern

import (
	"fmt"
	"time"
)

// Signal represents a detected pattern signal.
type Signal struct {
	ID             string      `json:"id"`
	Symbol         string      `json:"symbol"`
	Pattern        PatternType `json:"pattern"`
	PatternCN      string      `json:"pattern_cn"`      // Chinese name
	Direction      Direction   `json:"direction"`
	Confidence     int         `json:"confidence"`      // 0-100
	UpPercent      int         `json:"up_percent"`      // Historical up probability
	DownPercent    int         `json:"down_percent"`    // Historical down probability
	EfficiencyRank string      `json:"efficiency_rank"` // Efficiency rank
	Source         string      `json:"source"`          // Detection source: "talib" or "custom"
	StatsSource    string      `json:"stats_source"`    // Statistics data source
	IsEstimated    bool        `json:"is_estimated"`    // Whether stats are estimated
	KlineTime      time.Time   `json:"kline_time"`      // Kline close time
	DetectedAt     time.Time   `json:"detected_at"`
}

// NewSignal creates a new pattern signal with statistics populated.
func NewSignal(symbol string, pattern PatternType, direction Direction, confidence int, klineTime time.Time) Signal {
	stats := PatternStatsMap[pattern]
	return Signal{
		ID:             generateID(symbol, pattern, klineTime),
		Symbol:         symbol,
		Pattern:        pattern,
		PatternCN:      PatternNames[pattern],
		Direction:      direction,
		Confidence:     confidence,
		UpPercent:      stats.UpPercent,
		DownPercent:    stats.DownPercent,
		EfficiencyRank: stats.EfficiencyRank,
		Source:         stats.Source,
		StatsSource:    stats.StatsSource,
		IsEstimated:    stats.IsEstimated,
		KlineTime:      klineTime,
		DetectedAt:     time.Now().UTC(),
	}
}

// generateID generates a unique signal ID using symbol + pattern + klineTime.
// Format: {klineTime_unix_nano}-{symbol}-{pattern}
func generateID(symbol string, pattern PatternType, klineTime time.Time) string {
	return fmt.Sprintf("%d-%s-%s", klineTime.UnixNano(), symbol, pattern)
}

// DetectedPattern represents a pattern detected by the detector.
type DetectedPattern struct {
	Type       PatternType
	Direction  Direction
	Confidence int // 0-100, based on talib-cdl-go return value
}

// IsValid returns true if the signal has all required fields.
func (s *Signal) IsValid() bool {
	if s.ID == "" {
		return false
	}
	if s.Symbol == "" {
		return false
	}
	if s.Pattern == "" {
		return false
	}
	if s.Direction != DirectionBullish && s.Direction != DirectionBearish && s.Direction != DirectionNeutral {
		return false
	}
	if s.Confidence < 0 || s.Confidence > 100 {
		return false
	}
	if s.KlineTime.IsZero() {
		return false
	}
	if s.DetectedAt.IsZero() {
		return false
	}
	return true
}
