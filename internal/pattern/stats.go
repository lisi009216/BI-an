package pattern

// PatternStats holds statistical data for a pattern.
type PatternStats struct {
	UpPercent      int    // Historical up probability
	DownPercent    int    // Historical down probability
	EfficiencyRank string // Efficiency rank A+ ~ J-
	CommonRank     string // Commonality rank
	Source         string // Detection source: "talib" or "custom"
	StatsSource    string // Statistics data source
	IsEstimated    bool   // Whether stats are estimated
}

// PatternStatsMap maps pattern types to their statistics.
// Data sources: feedroll.com (talib-cdl-go), fivehundred.co, patternswizard.com
var PatternStatsMap = map[PatternType]PatternStats{
	// talib-cdl-go library supported patterns (source: feedroll.com)
	PatternDoji:              {43, 57, "J+", "E-", "talib", "feedroll.com", false},
	PatternDojiStar:          {36, 64, "E-", "F+", "talib", "feedroll.com", false},
	PatternEveningStar:       {28, 72, "A", "H+", "talib", "feedroll.com", false},
	PatternPiercing:          {64, 39, "B+", "D-", "talib", "feedroll.com", false},
	PatternAbandonedBaby:     {70, 30, "A-", "J+", "talib", "feedroll.com", false},
	PatternThreeWhite:        {82, 18, "D+", "G", "talib", "feedroll.com", false},
	PatternThreeBlack:        {22, 78, "A+", "F-", "talib", "feedroll.com", false},
	PatternThreeInside:       {40, 60, "F", "D+", "talib", "feedroll.com", false},
	PatternThreeOutside:      {31, 69, "D-", "C+", "talib", "feedroll.com", false},
	PatternThreeLineStrike:   {35, 65, "A+", "J", "talib", "feedroll.com", false},
	PatternThreeStarsInSouth: {86, 14, "J-", "J-", "talib", "feedroll.com", false},
	PatternAdvanceBlock:      {64, 36, "F", "G", "talib", "feedroll.com", false},
	PatternBeltHold:          {71, 29, "G+", "C+", "talib", "feedroll.com", false},
	PatternBreakAway:         {63, 37, "B+", "J-", "talib", "feedroll.com", false},
	PatternClosingMarubozu:   {52, 48, "E+", "B-", "talib", "feedroll.com", false},
	PatternTwoCrows:          {46, 54, "G+", "G", "talib", "feedroll.com", false},
	PatternMatchingLow:       {39, 61, "A-", "F-", "talib", "feedroll.com", false},
	PatternStickSandwich:     {38, 62, "B", "F-", "talib", "feedroll.com", false},
	PatternConcealBabySwall:  {25, 75, "J-", "J-", "talib", "feedroll.com", false},

	// Custom implemented patterns (sources: fivehundred.co, patternswizard.com)
	PatternHammer:          {60, 40, "B+", "C", "custom", "fivehundred.co", false},
	PatternInvertedHammer:  {55, 45, "C+", "D", "custom", "fivehundred.co", false},
	PatternHangingMan:      {41, 59, "B", "C", "custom", "fivehundred.co", false},
	PatternShootingStar:    {38, 62, "A-", "C", "custom", "fivehundred.co", false},
	PatternEngulfing:       {67, 33, "A", "B", "custom", "patternswizard.com", false},
	PatternMorningStar:     {70, 30, "A", "G", "custom", "stockgro.club", false},
	PatternMorningDojiStar: {68, 32, "A-", "H", "custom", "estimated", true},
	PatternEveningDojiStar: {32, 68, "A-", "H", "custom", "estimated", true},
	PatternDarkCloudCover:  {30, 70, "A", "E", "custom", "fivehundred.co", false},
	PatternHarami:          {53, 47, "C", "B", "custom", "fivehundred.co", false},
	PatternHaramiCross:     {55, 45, "B-", "D", "custom", "estimated", true},
	PatternKicking:         {69, 31, "A+", "J", "custom", "feedroll.com", false},
	PatternDragonflyDoji:   {57, 43, "C+", "E", "custom", "fivehundred.co", false},
	PatternGravestoneDoji:  {43, 57, "C+", "E", "custom", "fivehundred.co", false},
}

// IsHighEfficiency returns true if the pattern has efficiency rank A or B.
func IsHighEfficiency(pt PatternType) bool {
	stats, ok := PatternStatsMap[pt]
	if !ok {
		return false
	}
	if len(stats.EfficiencyRank) == 0 {
		return false
	}
	return stats.EfficiencyRank[0] == 'A' || stats.EfficiencyRank[0] == 'B'
}

// GetHighEfficiencyPatterns returns all patterns with efficiency rank A or B.
func GetHighEfficiencyPatterns() []PatternType {
	var result []PatternType
	for pt := range PatternStatsMap {
		if IsHighEfficiency(pt) {
			result = append(result, pt)
		}
	}
	return result
}

// GetStats returns the statistics for a pattern type.
func GetStats(pt PatternType) (PatternStats, bool) {
	stats, ok := PatternStatsMap[pt]
	return stats, ok
}
