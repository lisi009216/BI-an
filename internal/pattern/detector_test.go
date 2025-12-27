package pattern

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"example.com/binance-pivot-monitor/internal/kline"
)

func makeKline(open, high, low, close float64) kline.Kline {
	return kline.Kline{
		Symbol:   "TEST",
		Open:     open,
		High:     high,
		Low:      low,
		Close:    close,
		OpenTime: time.Now(),
	}
}

func TestDetector_Detect_Engulfing(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Bullish engulfing
	klines := []kline.Kline{
		makeKline(100, 100, 95, 96),  // Bearish
		makeKline(95, 105, 94, 104),  // Bullish engulfing
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternEngulfing && p.Direction == DirectionBullish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected bullish engulfing pattern")
	}

	// Bearish engulfing
	klines = []kline.Kline{
		makeKline(95, 105, 95, 104),  // Bullish
		makeKline(105, 106, 93, 94),  // Bearish engulfing
	}

	patterns = detector.Detect(klines)
	found = false
	for _, p := range patterns {
		if p.Type == PatternEngulfing && p.Direction == DirectionBearish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected bearish engulfing pattern")
	}
}

func TestDetector_Detect_Hammer(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Hammer after downtrend - need 4 klines minimum
	// Hammer conditions: lower shadow >= 2x body, upper shadow < 0.3x body
	klines := []kline.Kline{
		makeKline(115, 115, 110, 111), // Bearish
		makeKline(111, 111, 106, 107), // Bearish
		makeKline(107, 107, 102, 103), // Bearish
		makeKline(103, 103, 97, 98),   // Bearish (trend)
		// Hammer: body = |99-98| = 1, lower shadow = 98-88 = 10 (>= 2*1), upper shadow = 100-99 = 1 (< 0.3*1? No)
		// Need: body small, lower shadow >= 2x body, upper shadow < 0.3x body
		// Let's make: open=98, high=99, low=88, close=99 -> body=1, lower=10, upper=0
		makeKline(98, 99, 88, 99), // Hammer: body=1, lower=10, upper=0
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternHammer && p.Direction == DirectionBullish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected hammer pattern")
	}
}

func TestDetector_Detect_ShootingStar(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Shooting star after uptrend - need 4 klines minimum
	// Shooting star conditions: upper shadow >= 2x body, lower shadow < 0.3x body
	klines := []kline.Kline{
		makeKline(85, 90, 85, 89),   // Bullish
		makeKline(89, 95, 89, 94),   // Bullish
		makeKline(94, 100, 94, 99),  // Bullish
		makeKline(99, 105, 99, 104), // Bullish (trend)
		// Shooting star: body=1, upper=15, lower=0
		// open=105, high=120, low=105, close=104 -> body=1, upper=16, lower=0
		makeKline(105, 120, 104, 104), // Shooting star: body=1, upper=16, lower=0
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternShootingStar && p.Direction == DirectionBearish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected shooting star pattern")
	}
}

func TestDetector_Detect_MorningStar(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Morning star
	klines := []kline.Kline{
		makeKline(110, 110, 95, 96),  // Large bearish
		makeKline(96, 98, 94, 97),    // Small body (star)
		makeKline(97, 115, 96, 112),  // Large bullish closing above mid of first
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternMorningStar && p.Direction == DirectionBullish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected morning star pattern")
	}
}

func TestDetector_Detect_DarkCloudCover(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0, CryptoMode: true})

	// Dark cloud cover (crypto mode - relaxed gap)
	klines := []kline.Kline{
		makeKline(90, 110, 90, 108),  // Large bullish
		makeKline(108, 112, 95, 96),  // Bearish opening at/above prev close, closing below mid
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternDarkCloudCover && p.Direction == DirectionBearish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected dark cloud cover pattern")
	}
}

func TestDetector_Detect_Harami(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Bullish harami
	klines := []kline.Kline{
		makeKline(110, 110, 90, 92),  // Large bearish
		makeKline(95, 100, 94, 98),   // Small bullish inside prev body
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternHarami && p.Direction == DirectionBullish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected bullish harami pattern")
	}
}

func TestDetector_Detect_DragonflyDoji(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Dragonfly doji - need at least 2 klines for detector
	klines := []kline.Kline{
		makeKline(95, 100, 90, 98),    // Some previous kline
		makeKline(100, 100.2, 90, 100), // Dragonfly doji: open=close at top, long lower shadow
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternDragonflyDoji && p.Direction == DirectionBullish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected dragonfly doji pattern")
	}
}

func TestDetector_Detect_GravestoneDoji(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0})

	// Gravestone doji - need at least 2 klines for detector
	klines := []kline.Kline{
		makeKline(95, 100, 90, 98),    // Some previous kline
		makeKline(100, 112, 99.8, 100), // Gravestone doji: open=close at bottom, long upper shadow
	}

	patterns := detector.Detect(klines)
	found := false
	for _, p := range patterns {
		if p.Type == PatternGravestoneDoji && p.Direction == DirectionBearish {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected gravestone doji pattern")
	}
}

func TestDetector_MinConfidenceFilter(t *testing.T) {
	// Create detector with high min confidence
	detector := NewDetector(DetectorConfig{MinConfidence: 95})

	klines := []kline.Kline{
		makeKline(100, 100, 95, 96),
		makeKline(95, 105, 94, 104),
	}

	patterns := detector.Detect(klines)
	// Most patterns have confidence < 95, so should be filtered
	for _, p := range patterns {
		if p.Confidence < 95 {
			t.Errorf("Pattern %v with confidence %d should have been filtered", p.Type, p.Confidence)
		}
	}
}

func TestDetector_HighEfficiencyOnlyFilter(t *testing.T) {
	detector := NewDetector(DetectorConfig{MinConfidence: 0, HighEfficiencyOnly: true})

	klines := []kline.Kline{
		makeKline(100, 100, 95, 96),
		makeKline(95, 105, 94, 104),
	}

	patterns := detector.Detect(klines)
	for _, p := range patterns {
		if !IsHighEfficiency(p.Type) {
			t.Errorf("Pattern %v is not high efficiency but was not filtered", p.Type)
		}
	}
}

// Property test: Detection determinism
func TestProperty_DetectionDeterminism(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Detect returns same results for same input", prop.ForAll(
		func(prices []float64) bool {
			if len(prices) < 8 { // Need at least 2 klines with 4 prices each
				return true
			}

			// Build klines from prices
			var klines []kline.Kline
			for i := 0; i+3 < len(prices); i += 4 {
				open := prices[i]
				high := prices[i+1]
				low := prices[i+2]
				close := prices[i+3]

				// Ensure valid OHLC
				if open <= 0 || high <= 0 || low <= 0 || close <= 0 {
					continue
				}
				if high < open || high < close || low > open || low > close {
					// Fix invalid OHLC
					high = max(max(open, close), high)
					low = min(min(open, close), low)
				}

				klines = append(klines, kline.Kline{
					Symbol:   "TEST",
					Open:     open,
					High:     high,
					Low:      low,
					Close:    close,
					OpenTime: time.Now(),
				})
			}

			if len(klines) < 2 {
				return true
			}

			detector := NewDetector(DetectorConfig{MinConfidence: 0})

			// Call Detect multiple times
			result1 := detector.Detect(klines)
			result2 := detector.Detect(klines)
			result3 := detector.Detect(klines)

			// Results should be identical
			if len(result1) != len(result2) || len(result1) != len(result3) {
				return false
			}

			for i := range result1 {
				if result1[i].Type != result2[i].Type || result1[i].Type != result3[i].Type {
					return false
				}
				if result1[i].Direction != result2[i].Direction || result1[i].Direction != result3[i].Direction {
					return false
				}
				if result1[i].Confidence != result2[i].Confidence || result1[i].Confidence != result3[i].Confidence {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.Float64Range(0.01, 1000)),
	))

	properties.TestingRun(t)
}

func TestIsDowntrend(t *testing.T) {
	tests := []struct {
		name     string
		klines   []kline.Kline
		expected bool
	}{
		{
			name: "decreasing closes",
			klines: []kline.Kline{
				makeKline(100, 100, 95, 98),
				makeKline(98, 98, 93, 95),
				makeKline(95, 95, 90, 92),
			},
			expected: true,
		},
		{
			name: "mostly bearish",
			klines: []kline.Kline{
				makeKline(100, 100, 95, 96), // Bearish
				makeKline(96, 98, 94, 97),   // Bullish
				makeKline(97, 97, 92, 93),   // Bearish
			},
			expected: true,
		},
		{
			name: "uptrend",
			klines: []kline.Kline{
				makeKline(90, 95, 90, 94),
				makeKline(94, 100, 94, 99),
				makeKline(99, 105, 99, 104),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDowntrend(tt.klines); got != tt.expected {
				t.Errorf("isDowntrend() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsUptrend(t *testing.T) {
	tests := []struct {
		name     string
		klines   []kline.Kline
		expected bool
	}{
		{
			name: "increasing closes",
			klines: []kline.Kline{
				makeKline(90, 95, 90, 94),
				makeKline(94, 100, 94, 99),
				makeKline(99, 105, 99, 104),
			},
			expected: true,
		},
		{
			name: "mostly bullish",
			klines: []kline.Kline{
				makeKline(90, 95, 90, 94),   // Bullish
				makeKline(94, 95, 92, 93),   // Bearish
				makeKline(93, 100, 93, 99),  // Bullish
			},
			expected: true,
		},
		{
			name: "downtrend",
			klines: []kline.Kline{
				makeKline(100, 100, 95, 96),
				makeKline(96, 96, 90, 91),
				makeKline(91, 91, 85, 86),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUptrend(tt.klines); got != tt.expected {
				t.Errorf("isUptrend() = %v, want %v", got, tt.expected)
			}
		})
	}
}
