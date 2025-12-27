// Package pattern provides candlestick pattern detection and signal generation.
package pattern

// PatternType represents a candlestick pattern type.
type PatternType string

const (
	// === talib-cdl-go library supported patterns ===

	// Reversal patterns
	PatternDoji              PatternType = "doji"              // 十字星
	PatternDojiStar          PatternType = "doji_star"         // 十字星线
	PatternEveningStar       PatternType = "evening_star"      // 暮星
	PatternPiercing          PatternType = "piercing"          // 刺透形态
	PatternAbandonedBaby     PatternType = "abandoned_baby"    // 弃婴形态
	PatternMatchingLow       PatternType = "matching_low"      // 相同低价
	PatternThreeWhite        PatternType = "three_white"       // 三白兵
	PatternThreeBlack        PatternType = "three_black"       // 三只乌鸦
	PatternThreeInside       PatternType = "three_inside"      // 三内部
	PatternThreeOutside      PatternType = "three_outside"     // 三外部
	PatternThreeLineStrike   PatternType = "three_line_strike" // 三线打击
	PatternThreeStarsInSouth PatternType = "three_stars_south" // 南方三星
	PatternAdvanceBlock      PatternType = "advance_block"     // 前进受阻
	PatternBeltHold          PatternType = "belt_hold"         // 捉腰带线
	PatternBreakAway         PatternType = "break_away"        // 脱离形态
	PatternClosingMarubozu   PatternType = "closing_marubozu"  // 收盘光头光脚
	PatternTwoCrows          PatternType = "two_crows"         // 两只乌鸦
	PatternStickSandwich     PatternType = "stick_sandwich"    // 条形三明治
	PatternConcealBabySwall  PatternType = "conceal_baby"      // 藏婴吞没

	// === Custom implemented patterns (not in talib-cdl-go) ===

	PatternHammer          PatternType = "hammer"            // 锤子线
	PatternInvertedHammer  PatternType = "inverted_hammer"   // 倒锤子线
	PatternHangingMan      PatternType = "hanging_man"       // 上吊线
	PatternShootingStar    PatternType = "shooting_star"     // 流星线
	PatternEngulfing       PatternType = "engulfing"         // 吞没形态
	PatternMorningStar     PatternType = "morning_star"      // 晨星
	PatternMorningDojiStar PatternType = "morning_doji_star" // 晨十字星
	PatternEveningDojiStar PatternType = "evening_doji_star" // 暮十字星
	PatternDarkCloudCover  PatternType = "dark_cloud_cover"  // 乌云盖顶
	PatternHarami          PatternType = "harami"            // 孕线
	PatternHaramiCross     PatternType = "harami_cross"      // 十字孕线
	PatternKicking         PatternType = "kicking"           // 反冲形态
	PatternDragonflyDoji   PatternType = "dragonfly_doji"    // 蜻蜓十字
	PatternGravestoneDoji  PatternType = "gravestone_doji"   // 墓碑十字
)

// Direction represents the pattern direction.
type Direction string

const (
	DirectionBullish Direction = "bullish" // 看涨
	DirectionBearish Direction = "bearish" // 看跌
	DirectionNeutral Direction = "neutral" // 中性
)

// PatternNames maps pattern types to Chinese names.
var PatternNames = map[PatternType]string{
	// talib-cdl-go library supported patterns
	PatternDoji:              "十字星",
	PatternDojiStar:          "十字星线",
	PatternEveningStar:       "暮星",
	PatternPiercing:          "刺透形态",
	PatternAbandonedBaby:     "弃婴形态",
	PatternMatchingLow:       "相同低价",
	PatternThreeWhite:        "三白兵",
	PatternThreeBlack:        "三只乌鸦",
	PatternThreeInside:       "三内部",
	PatternThreeOutside:      "三外部",
	PatternThreeLineStrike:   "三线打击",
	PatternThreeStarsInSouth: "南方三星",
	PatternAdvanceBlock:      "前进受阻",
	PatternBeltHold:          "捉腰带线",
	PatternBreakAway:         "脱离形态",
	PatternClosingMarubozu:   "收盘光头光脚",
	PatternTwoCrows:          "两只乌鸦",
	PatternStickSandwich:     "条形三明治",
	PatternConcealBabySwall:  "藏婴吞没",

	// Custom implemented patterns
	PatternHammer:          "锤子线",
	PatternInvertedHammer:  "倒锤子线",
	PatternHangingMan:      "上吊线",
	PatternShootingStar:    "流星线",
	PatternEngulfing:       "吞没形态",
	PatternMorningStar:     "晨星",
	PatternMorningDojiStar: "晨十字星",
	PatternEveningDojiStar: "暮十字星",
	PatternDarkCloudCover:  "乌云盖顶",
	PatternHarami:          "孕线",
	PatternHaramiCross:     "十字孕线",
	PatternKicking:         "反冲形态",
	PatternDragonflyDoji:   "蜻蜓十字",
	PatternGravestoneDoji:  "墓碑十字",
}
