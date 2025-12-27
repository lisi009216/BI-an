package pattern

import (
	"example.com/binance-pivot-monitor/internal/kline"
)

// detectCustomPatterns detects patterns not available in talib-cdl-go.
func (d *Detector) detectCustomPatterns(klines []kline.Kline) []DetectedPattern {
	var patterns []DetectedPattern

	// Hammer
	if found, dir, conf := detectHammer(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternHammer, Direction: dir, Confidence: conf})
	}

	// Inverted Hammer
	if found, dir, conf := detectInvertedHammer(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternInvertedHammer, Direction: dir, Confidence: conf})
	}

	// Hanging Man
	if found, dir, conf := detectHangingMan(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternHangingMan, Direction: dir, Confidence: conf})
	}

	// Shooting Star
	if found, dir, conf := detectShootingStar(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternShootingStar, Direction: dir, Confidence: conf})
	}

	// Engulfing
	if found, dir, conf := detectEngulfing(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternEngulfing, Direction: dir, Confidence: conf})
	}

	// Morning Star
	if found, dir, conf := detectMorningStar(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternMorningStar, Direction: dir, Confidence: conf})
	}

	// Morning Doji Star
	if found, dir, conf := detectMorningDojiStar(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternMorningDojiStar, Direction: dir, Confidence: conf})
	}

	// Evening Doji Star
	if found, dir, conf := detectEveningDojiStar(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternEveningDojiStar, Direction: dir, Confidence: conf})
	}

	// Dark Cloud Cover
	if found, dir, conf := detectDarkCloudCover(klines, d.config.CryptoMode); found {
		patterns = append(patterns, DetectedPattern{Type: PatternDarkCloudCover, Direction: dir, Confidence: conf})
	}

	// Harami
	if found, dir, conf := detectHarami(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternHarami, Direction: dir, Confidence: conf})
	}

	// Harami Cross
	if found, dir, conf := detectHaramiCross(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternHaramiCross, Direction: dir, Confidence: conf})
	}

	// Dragonfly Doji
	if found, dir, conf := detectDragonflyDoji(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternDragonflyDoji, Direction: dir, Confidence: conf})
	}

	// Gravestone Doji
	if found, dir, conf := detectGravestoneDoji(klines); found {
		patterns = append(patterns, DetectedPattern{Type: PatternGravestoneDoji, Direction: dir, Confidence: conf})
	}

	return patterns
}

// isDowntrend checks if the klines show a downtrend.
// Condition: closing prices decreasing OR at least 2/3 bearish.
func isDowntrend(klines []kline.Kline) bool {
	if len(klines) < 2 {
		return false
	}

	// Method 1: Closing prices decreasing
	decreasing := true
	for i := 1; i < len(klines); i++ {
		if klines[i].Close >= klines[i-1].Close {
			decreasing = false
			break
		}
	}
	if decreasing {
		return true
	}

	// Method 2: At least 2/3 bearish
	bearishCount := 0
	for _, k := range klines {
		if k.IsBearish() {
			bearishCount++
		}
	}
	return bearishCount >= (len(klines)*2)/3
}

// isUptrend checks if the klines show an uptrend.
func isUptrend(klines []kline.Kline) bool {
	if len(klines) < 2 {
		return false
	}

	// Method 1: Closing prices increasing
	increasing := true
	for i := 1; i < len(klines); i++ {
		if klines[i].Close <= klines[i-1].Close {
			increasing = false
			break
		}
	}
	if increasing {
		return true
	}

	// Method 2: At least 2/3 bullish
	bullishCount := 0
	for _, k := range klines {
		if k.IsBullish() {
			bullishCount++
		}
	}
	return bullishCount >= (len(klines)*2)/3
}

// isDoji checks if a kline is a doji (very small body).
// Excludes zero-range klines to avoid false positives in low-liquidity data.
func isDoji(k *kline.Kline) bool {
	if k.Range() == 0 {
		return false // 零波动不算 doji，避免极端数据误报
	}
	return k.Body()/k.Range() < 0.1
}

// detectHammer detects hammer pattern.
// Conditions: long lower shadow (>= 2x body), small upper shadow, appears after downtrend.
func detectHammer(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 4 { // Need at least 4 klines (3 for trend + 1 current)
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	body := k.Body()
	if body == 0 || k.Range() == 0 {
		return false, "", 0
	}

	lowerShadow := k.LowerShadow()
	upperShadow := k.UpperShadow()

	// Lower shadow at least 2x body
	if lowerShadow < body*2 {
		return false, "", 0
	}
	// Upper shadow small (< 30% of body)
	if upperShadow > body*0.3 {
		return false, "", 0
	}

	// Check downtrend using last 3 klines (excluding current)
	if !isDowntrend(klines[len(klines)-4 : len(klines)-1]) {
		return false, "", 0
	}

	confidence := 70
	if lowerShadow >= body*3 {
		confidence = 85
	}
	return true, DirectionBullish, confidence
}

// detectInvertedHammer detects inverted hammer pattern.
func detectInvertedHammer(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 4 { // Need at least 4 klines (3 for trend + 1 current)
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	body := k.Body()
	if body == 0 || k.Range() == 0 {
		return false, "", 0
	}

	upperShadow := k.UpperShadow()
	lowerShadow := k.LowerShadow()

	// Upper shadow at least 2x body
	if upperShadow < body*2 {
		return false, "", 0
	}
	// Lower shadow small
	if lowerShadow > body*0.3 {
		return false, "", 0
	}

	// Check downtrend
	if !isDowntrend(klines[len(klines)-4 : len(klines)-1]) {
		return false, "", 0
	}

	confidence := 65
	if upperShadow >= body*3 {
		confidence = 80
	}
	return true, DirectionBullish, confidence
}

// detectHangingMan detects hanging man pattern (hammer at top).
func detectHangingMan(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 4 { // Need at least 4 klines (3 for trend + 1 current)
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	body := k.Body()
	if body == 0 || k.Range() == 0 {
		return false, "", 0
	}

	lowerShadow := k.LowerShadow()
	upperShadow := k.UpperShadow()

	// Lower shadow at least 2x body
	if lowerShadow < body*2 {
		return false, "", 0
	}
	// Upper shadow small
	if upperShadow > body*0.3 {
		return false, "", 0
	}

	// Check uptrend (opposite of hammer)
	if !isUptrend(klines[len(klines)-4 : len(klines)-1]) {
		return false, "", 0
	}

	confidence := 70
	if lowerShadow >= body*3 {
		confidence = 85
	}
	return true, DirectionBearish, confidence
}

// detectShootingStar detects shooting star pattern.
func detectShootingStar(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 4 { // Need at least 4 klines (3 for trend + 1 current)
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	body := k.Body()
	if body == 0 || k.Range() == 0 {
		return false, "", 0
	}

	upperShadow := k.UpperShadow()
	lowerShadow := k.LowerShadow()

	// Upper shadow at least 2x body
	if upperShadow < body*2 {
		return false, "", 0
	}
	// Lower shadow small
	if lowerShadow > body*0.3 {
		return false, "", 0
	}

	// Check uptrend
	if !isUptrend(klines[len(klines)-4 : len(klines)-1]) {
		return false, "", 0
	}

	confidence := 70
	if upperShadow >= body*3 {
		confidence = 85
	}
	return true, DirectionBearish, confidence
}

// detectEngulfing detects engulfing pattern.
func detectEngulfing(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 2 {
		return false, "", 0
	}
	curr := &klines[len(klines)-1]
	prev := &klines[len(klines)-2]

	// Bullish engulfing: prev bearish, curr bullish, curr body contains prev body
	if prev.IsBearish() && curr.IsBullish() {
		if curr.Open <= prev.Close && curr.Close >= prev.Open {
			confidence := 75
			if curr.Body() > prev.Body()*1.5 {
				confidence = 90
			}
			return true, DirectionBullish, confidence
		}
	}

	// Bearish engulfing: prev bullish, curr bearish, curr body contains prev body
	if prev.IsBullish() && curr.IsBearish() {
		if curr.Open >= prev.Close && curr.Close <= prev.Open {
			confidence := 75
			if curr.Body() > prev.Body()*1.5 {
				confidence = 90
			}
			return true, DirectionBearish, confidence
		}
	}

	return false, "", 0
}

// detectMorningStar detects morning star pattern.
func detectMorningStar(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 3 {
		return false, "", 0
	}
	first := &klines[len(klines)-3]
	second := &klines[len(klines)-2]
	third := &klines[len(klines)-1]

	// First: large bearish candle
	if !first.IsBearish() || first.Body() < first.Range()*0.6 {
		return false, "", 0
	}
	// Second: small body (star)
	if second.Body() > first.Body()*0.3 {
		return false, "", 0
	}
	// Third: large bullish candle
	if !third.IsBullish() || third.Body() < third.Range()*0.6 {
		return false, "", 0
	}
	// Third closes into first body
	midFirst := (first.Open + first.Close) / 2
	if third.Close < midFirst {
		return false, "", 0
	}

	return true, DirectionBullish, 80
}

// detectMorningDojiStar detects morning doji star pattern.
func detectMorningDojiStar(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 3 {
		return false, "", 0
	}
	first := &klines[len(klines)-3]
	second := &klines[len(klines)-2]
	third := &klines[len(klines)-1]

	// First: large bearish candle
	if !first.IsBearish() || first.Body() < first.Range()*0.6 {
		return false, "", 0
	}
	// Second: doji
	if !isDoji(second) {
		return false, "", 0
	}
	// Third: large bullish candle
	if !third.IsBullish() || third.Body() < third.Range()*0.6 {
		return false, "", 0
	}
	// Third closes into first body
	midFirst := (first.Open + first.Close) / 2
	if third.Close < midFirst {
		return false, "", 0
	}

	return true, DirectionBullish, 78
}

// detectEveningDojiStar detects evening doji star pattern.
func detectEveningDojiStar(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 3 {
		return false, "", 0
	}
	first := &klines[len(klines)-3]
	second := &klines[len(klines)-2]
	third := &klines[len(klines)-1]

	// First: large bullish candle
	if !first.IsBullish() || first.Body() < first.Range()*0.6 {
		return false, "", 0
	}
	// Second: doji
	if !isDoji(second) {
		return false, "", 0
	}
	// Third: large bearish candle
	if !third.IsBearish() || third.Body() < third.Range()*0.6 {
		return false, "", 0
	}
	// Third closes into first body
	midFirst := (first.Open + first.Close) / 2
	if third.Close > midFirst {
		return false, "", 0
	}

	return true, DirectionBearish, 78
}

// detectDarkCloudCover detects dark cloud cover pattern.
// In crypto mode, gap condition is relaxed.
func detectDarkCloudCover(klines []kline.Kline, cryptoMode bool) (bool, Direction, int) {
	if len(klines) < 2 {
		return false, "", 0
	}
	prev := &klines[len(klines)-2]
	curr := &klines[len(klines)-1]

	// Previous: large bullish candle
	if !prev.IsBullish() || prev.Body() < prev.Range()*0.6 {
		return false, "", 0
	}
	// Current: bearish candle
	if !curr.IsBearish() {
		return false, "", 0
	}

	// Gap condition
	if cryptoMode {
		// Relaxed: current open >= prev close
		if curr.Open < prev.Close {
			return false, "", 0
		}
	} else {
		// Traditional: current open > prev high (gap up)
		if curr.Open <= prev.High {
			return false, "", 0
		}
	}

	// Close penetrates into prev body by at least 50%
	midPrev := (prev.Open + prev.Close) / 2
	if curr.Close > midPrev {
		return false, "", 0
	}

	// Higher confidence if there's a gap
	confidence := 70
	if curr.Open > prev.High {
		confidence = 85
	}

	return true, DirectionBearish, confidence
}

// detectHarami detects harami pattern.
func detectHarami(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 2 {
		return false, "", 0
	}
	prev := &klines[len(klines)-2]
	curr := &klines[len(klines)-1]

	// Previous must have significant body (exclude zero-range)
	if prev.Range() == 0 || prev.Body() < prev.Range()*0.5 {
		return false, "", 0
	}

	// Current body must be inside previous body
	prevBodyHigh := max(prev.Open, prev.Close)
	prevBodyLow := min(prev.Open, prev.Close)
	currBodyHigh := max(curr.Open, curr.Close)
	currBodyLow := min(curr.Open, curr.Close)

	if currBodyHigh > prevBodyHigh || currBodyLow < prevBodyLow {
		return false, "", 0
	}

	// Current body should be smaller
	if curr.Body() > prev.Body()*0.5 {
		return false, "", 0
	}

	// Direction based on previous candle
	if prev.IsBearish() {
		return true, DirectionBullish, 65
	}
	return true, DirectionBearish, 65
}

// detectHaramiCross detects harami cross pattern (harami with doji).
func detectHaramiCross(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 2 {
		return false, "", 0
	}
	prev := &klines[len(klines)-2]
	curr := &klines[len(klines)-1]

	// Previous must have significant body (exclude zero-range)
	if prev.Range() == 0 || prev.Body() < prev.Range()*0.5 {
		return false, "", 0
	}

	// Current must be a doji
	if !isDoji(curr) {
		return false, "", 0
	}

	// Current doji's BODY (not high/low) must be inside previous body
	// This allows shadows to extend beyond, which is standard for harami cross
	prevBodyHigh := max(prev.Open, prev.Close)
	prevBodyLow := min(prev.Open, prev.Close)
	currBodyHigh := max(curr.Open, curr.Close)
	currBodyLow := min(curr.Open, curr.Close)

	if currBodyHigh > prevBodyHigh || currBodyLow < prevBodyLow {
		return false, "", 0
	}

	// Direction based on previous candle
	if prev.IsBearish() {
		return true, DirectionBullish, 70
	}
	return true, DirectionBearish, 70
}

// detectDragonflyDoji detects dragonfly doji pattern.
func detectDragonflyDoji(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 1 {
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	if k.Range() == 0 {
		return false, "", 0
	}

	// Must be a doji
	if !isDoji(k) {
		return false, "", 0
	}

	// Long lower shadow, almost no upper shadow
	lowerShadow := k.LowerShadow()
	upperShadow := k.UpperShadow()

	if lowerShadow < k.Range()*0.6 {
		return false, "", 0
	}
	if upperShadow > k.Range()*0.1 {
		return false, "", 0
	}

	return true, DirectionBullish, 65
}

// detectGravestoneDoji detects gravestone doji pattern.
func detectGravestoneDoji(klines []kline.Kline) (bool, Direction, int) {
	if len(klines) < 1 {
		return false, "", 0
	}
	k := &klines[len(klines)-1]

	if k.Range() == 0 {
		return false, "", 0
	}

	// Must be a doji
	if !isDoji(k) {
		return false, "", 0
	}

	// Long upper shadow, almost no lower shadow
	upperShadow := k.UpperShadow()
	lowerShadow := k.LowerShadow()

	if upperShadow < k.Range()*0.6 {
		return false, "", 0
	}
	if lowerShadow > k.Range()*0.1 {
		return false, "", 0
	}

	return true, DirectionBearish, 65
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
