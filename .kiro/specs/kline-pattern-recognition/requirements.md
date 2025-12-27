# Requirements Document

## Introduction

本功能基于现有的币安 WebSocket 标记价格流（!markPrice@arr@1s），在内存中构建虚拟 5 分钟 K 线数据，并使用技术分析（TA）模块进行 K 线形态识别。识别结果将与现有的枢轴点信号系统结合，为用户提供更丰富的交易信号参考。

## Glossary

- **Kline_Store**: 虚拟 K 线存储模块，负责维护每个交易对的滚动 K 线数据
- **Kline**: 单根 K 线数据结构，包含 Open、High、Low、Close 四个价格
- **Pattern_Detector**: K 线形态识别模块，使用 TA 库检测特定形态
- **Pattern_Signal**: 形态信号数据结构，包含识别到的形态信息
- **Signal_Combiner**: 信号组合模块，将形态信号与枢轴点信号关联
- **Monitor**: 现有的价格监控模块
- **SSE_Broker**: 现有的 Server-Sent Events 广播器

## Requirements

### Requirement 1: 虚拟 K 线数据结构

**User Story:** As a developer, I want to maintain virtual 5-minute kline data in memory, so that I can perform technical analysis without additional API calls.

#### Acceptance Criteria

1. THE Kline_Store SHALL maintain a rolling window of X klines per symbol, where X defaults to 12 (representing 1 hour of data)
2. WHEN a new price arrives, THE Kline_Store SHALL update the current kline's High if the price is higher than the existing High
3. WHEN a new price arrives, THE Kline_Store SHALL update the current kline's Low if the price is lower than the existing Low
4. WHEN a new price arrives, THE Kline_Store SHALL update the current kline's Close to the latest price
5. WHEN the 5-minute boundary is crossed (0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55), THE Kline_Store SHALL close the current kline and start a new one
6. WHEN a new kline starts, THE Kline_Store SHALL set Open, High, Low, Close all to the first received price
7. WHEN the kline count exceeds X, THE Kline_Store SHALL remove the oldest kline to maintain the rolling window
8. THE Kline_Store SHALL store kline data using a thread-safe data structure to support concurrent access

### Requirement 2: K 线形态识别

**User Story:** As a trader, I want the system to detect candlestick patterns, so that I can receive pattern-based trading signals.

#### Acceptance Criteria

1. WHEN a kline closes, THE Pattern_Detector SHALL analyze the recent klines for pattern recognition
2. THE Pattern_Detector SHALL detect the following reversal patterns:
   - Hammer (锤子线)
   - Inverted Hammer (倒锤子线)
   - Bullish Engulfing (看涨吞没)
   - Bearish Engulfing (看跌吞没)
   - Morning Star (晨星)
   - Evening Star (暮星)
   - Doji (十字星)
3. THE Pattern_Detector SHALL detect the following continuation patterns:
   - Three White Soldiers (三白兵)
   - Three Black Crows (三只乌鸦)
4. WHEN a pattern is detected, THE Pattern_Detector SHALL emit a Pattern_Signal with pattern name, direction (bullish/bearish), and confidence level
5. THE Pattern_Detector SHALL use a Go TA library (such as github.com/markcheno/go-talib or similar) for pattern detection
6. IF no suitable Go TA library is available, THEN THE Pattern_Detector SHALL implement pattern detection logic based on standard candlestick pattern definitions

### Requirement 3: 形态信号数据结构

**User Story:** As a developer, I want a well-defined pattern signal structure, so that I can integrate it with the existing signal system.

#### Acceptance Criteria

1. THE Pattern_Signal SHALL contain the following fields:
   - ID (unique identifier)
   - Symbol (trading pair)
   - Pattern (pattern name)
   - Direction (bullish/bearish)
   - Confidence (0-100)
   - KlineTime (the closing time of the kline that triggered the pattern)
   - DetectedAt (timestamp when pattern was detected)
2. WHEN serializing Pattern_Signal to JSON, THE system SHALL use snake_case field names for consistency with existing Signal structure
3. THE Pattern_Signal SHALL be stored in a separate history from pivot signals

### Requirement 4: 信号组合与关联

**User Story:** As a trader, I want to see pattern signals combined with pivot signals, so that I can make more informed trading decisions.

#### Acceptance Criteria

1. WHEN a pivot signal is emitted, THE Signal_Combiner SHALL check for recent pattern signals on the same symbol within the last 15 minutes
2. WHEN a pattern signal is emitted, THE Signal_Combiner SHALL check for recent pivot signals on the same symbol within the last 15 minutes
3. WHEN both signals exist for the same symbol, THE Signal_Combiner SHALL create a combined signal indicating the correlation
4. THE combined signal SHALL include:
   - The original pivot signal information
   - The related pattern signal information
   - A correlation strength indicator (strong/moderate/weak)
5. WHEN the pivot signal direction matches the pattern direction (e.g., up + bullish, down + bearish), THE correlation strength SHALL be marked as "strong"
6. WHEN the pivot signal direction conflicts with the pattern direction, THE correlation strength SHALL be marked as "weak"

### Requirement 5: 前端 UI 展示

**User Story:** As a user, I want to see pattern signals and their correlation with pivot signals in the UI, so that I can quickly assess trading opportunities.

#### Acceptance Criteria

1. THE UI SHALL display pattern signals in a new "Patterns" tab alongside the existing "Signals" tab
2. WHEN displaying a pattern signal, THE UI SHALL show:
   - Symbol name
   - Pattern name (with Chinese translation)
   - Direction indicator (bullish/bearish with color coding)
   - Confidence level as a percentage
   - Time since detection
3. WHEN a pivot signal has a correlated pattern signal, THE UI SHALL display a pattern badge on the signal item
4. THE pattern badge SHALL be color-coded based on correlation strength:
   - Green for strong correlation (direction match)
   - Yellow for moderate correlation
   - Red for weak correlation (direction conflict)
5. WHEN clicking on a pattern badge, THE UI SHALL show a tooltip with pattern details
6. THE UI SHALL support filtering signals by pattern type
7. THE UI SHALL push pattern signals via SSE in real-time, similar to pivot signals

### Requirement 6: 数据存储

**User Story:** As a system administrator, I want pattern data to be persisted, so that I can review historical patterns and system performance.

#### Acceptance Criteria

1. THE system SHALL store pattern signals in a JSONL file similar to the existing signal history
2. THE pattern history file SHALL be located at data/patterns/history.jsonl
3. WHEN the system starts, THE system SHALL load the last N pattern signals from the history file, where N defaults to 1000
4. THE system SHALL NOT persist the raw kline data to disk (memory-only for performance)
5. WHEN a pattern signal is emitted, THE system SHALL append it to the history file immediately
6. THE system SHALL provide an API endpoint to query pattern history with filtering options

### Requirement 7: 性能优化

**User Story:** As a developer, I want the system to handle high-frequency price updates efficiently, so that pattern detection does not impact the existing monitoring functionality.

#### Acceptance Criteria

1. THE Kline_Store SHALL process price updates in O(1) time complexity
2. WHEN processing price updates, THE system SHALL use goroutines to avoid blocking the main price processing loop
3. THE Pattern_Detector SHALL only run when a kline closes, not on every price update
4. THE system SHALL limit pattern detection to symbols that have pivot data loaded (to avoid unnecessary computation)
5. IF pattern detection takes longer than 100ms for a single symbol, THEN THE system SHALL log a warning

### Requirement 8: 配置选项

**User Story:** As a system administrator, I want to configure the kline and pattern detection parameters, so that I can tune the system for different use cases.

#### Acceptance Criteria

1. THE system SHALL support the following configuration options via environment variables:
   - KLINE_COUNT: Number of klines to maintain (default: 12)
   - KLINE_INTERVAL: Kline interval in minutes (default: 5)
   - PATTERN_ENABLED: Enable/disable pattern detection (default: true)
   - PATTERN_MIN_CONFIDENCE: Minimum confidence threshold for emitting signals (default: 60)
2. WHEN PATTERN_ENABLED is false, THE system SHALL skip all pattern detection logic
3. THE configuration SHALL be logged at startup for debugging purposes
