package pattern

import (
	talibcdl "github.com/iwat/talib-cdl-go"

	"example.com/binance-pivot-monitor/internal/kline"
)

// DetectorConfig holds configuration for the pattern detector.
type DetectorConfig struct {
	MinConfidence      int  // Minimum confidence threshold (0-100)
	HighEfficiencyOnly bool // Only detect high efficiency patterns (A/B rank)
	CryptoMode         bool // Crypto market mode (relaxed gap conditions)
	GapThreshold       float64 // Gap threshold for crypto mode (default 0.001 = 0.1%)
}

// DefaultDetectorConfig returns the default detector configuration.
func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		MinConfidence:      60,
		HighEfficiencyOnly: false,
		CryptoMode:         true,
		GapThreshold:       0.001,
	}
}

// Detector detects candlestick patterns in kline data.
type Detector struct {
	config DetectorConfig
}

// NewDetector creates a new pattern detector.
func NewDetector(config DetectorConfig) *Detector {
	return &Detector{config: config}
}

// toSeries converts klines to talib-cdl-go SimpleSeries format.
// klines must be in time order (oldest first, newest last).
func toSeries(klines []kline.Kline) talibcdl.SimpleSeries {
	n := len(klines)
	series := talibcdl.SimpleSeries{
		Opens:  make([]float64, n),
		Highs:  make([]float64, n),
		Lows:   make([]float64, n),
		Closes: make([]float64, n),
	}
	for i, k := range klines {
		series.Opens[i] = k.Open
		series.Highs[i] = k.High
		series.Lows[i] = k.Low
		series.Closes[i] = k.Close
	}
	return series
}

// Detect detects patterns in the given klines.
// klines must be in time order (oldest first, newest last).
// Returns all detected patterns.
func (d *Detector) Detect(klines []kline.Kline) []DetectedPattern {
	if len(klines) < 2 {
		return nil
	}

	var patterns []DetectedPattern

	// Detect talib-cdl-go patterns
	patterns = append(patterns, d.detectTalibPatterns(klines)...)

	// Detect custom patterns
	patterns = append(patterns, d.detectCustomPatterns(klines)...)

	// Filter by minimum confidence
	var filtered []DetectedPattern
	for _, p := range patterns {
		if p.Confidence >= d.config.MinConfidence {
			if d.config.HighEfficiencyOnly && !IsHighEfficiency(p.Type) {
				continue
			}
			filtered = append(filtered, p)
		}
	}

	return filtered
}

// detectTalibPatterns detects patterns using talib-cdl-go library.
func (d *Detector) detectTalibPatterns(klines []kline.Kline) []DetectedPattern {
	if len(klines) < 3 {
		return nil
	}

	series := toSeries(klines)
	var patterns []DetectedPattern
	lastIdx := len(klines) - 1

	// Doji
	if results := talibcdl.Doji(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternDoji,
			Direction:  DirectionNeutral,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// DojiStar
	if results := talibcdl.DojiStar(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBearish
		if results[lastIdx] > 0 {
			dir = DirectionBullish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternDojiStar,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// EveningStar
	if results := talibcdl.EveningStar(series, 0.3); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternEveningStar,
			Direction:  DirectionBearish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// Piercing
	if results := talibcdl.Piercing(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternPiercing,
			Direction:  DirectionBullish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// AbandonedBaby (skip in crypto mode due to gap dependency)
	if !d.config.CryptoMode {
		if results := talibcdl.AbandonedBaby(series, 0.3); len(results) > lastIdx && results[lastIdx] != 0 {
			dir := DirectionBullish
			if results[lastIdx] < 0 {
				dir = DirectionBearish
			}
			patterns = append(patterns, DetectedPattern{
				Type:       PatternAbandonedBaby,
				Direction:  dir,
				Confidence: absInt(results[lastIdx]),
			})
		}
	}

	// ThreeWhiteSoldiers
	if results := talibcdl.ThreeWhiteSoldiers(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeWhite,
			Direction:  DirectionBullish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ThreeBlackCrows
	if results := talibcdl.ThreeBlackCrows(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeBlack,
			Direction:  DirectionBearish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ThreeInside
	if results := talibcdl.ThreeInside(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeInside,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ThreeOutside
	if results := talibcdl.ThreeOutside(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeOutside,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ThreeLineStrike
	if results := talibcdl.ThreeLineStrike(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeLineStrike,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ThreeStarsInSouth
	if results := talibcdl.ThreeStarsInSouth(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternThreeStarsInSouth,
			Direction:  DirectionBullish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// AdvanceBlock
	if results := talibcdl.AdvanceBlock(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternAdvanceBlock,
			Direction:  DirectionBearish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// BeltHold
	if results := talibcdl.BeltHold(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternBeltHold,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// BreakAway
	if results := talibcdl.BreakAway(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternBreakAway,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ClosingMarubozu
	if results := talibcdl.ClosingMarubozu(series); len(results) > lastIdx && results[lastIdx] != 0 {
		dir := DirectionBullish
		if results[lastIdx] < 0 {
			dir = DirectionBearish
		}
		patterns = append(patterns, DetectedPattern{
			Type:       PatternClosingMarubozu,
			Direction:  dir,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// TwoCrows
	if results := talibcdl.TwoCrows(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternTwoCrows,
			Direction:  DirectionBearish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// MatchingLow
	if results := talibcdl.MatchingLow(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternMatchingLow,
			Direction:  DirectionBullish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// StickSandwich
	if results := talibcdl.StickSandwich(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternStickSandwich,
			Direction:  DirectionBullish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	// ConcealBabySwall
	if results := talibcdl.ConcealBabySwall(series); len(results) > lastIdx && results[lastIdx] != 0 {
		patterns = append(patterns, DetectedPattern{
			Type:       PatternConcealBabySwall,
			Direction:  DirectionBearish,
			Confidence: absInt(results[lastIdx]),
		})
	}

	return patterns
}

// absInt returns the absolute value of an integer.
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
