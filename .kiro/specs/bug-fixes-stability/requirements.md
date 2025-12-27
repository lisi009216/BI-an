# Requirements Document

## Introduction

本功能修复系统中发现的多个稳定性和正确性问题，包括周枢轴点刷新逻辑错误、参数校验缺失、历史文件无限增长以及数据解析兼容性问题。这些修复将提高系统的健壮性和长期运行稳定性。

## Glossary

- **Pivot_Refresher**: 枢轴点刷新模块，负责判断和更新日/周枢轴点数据
- **Kline_Store**: K线存储模块，负责维护虚拟K线数据
- **Pattern_History**: 形态历史模块，负责持久化形态识别结果
- **TickerEvent**: Binance WebSocket 返回的行情事件数据结构
- **Stale_Check**: 过期检查逻辑，判断枢轴点数据是否需要刷新

## Requirements

### Requirement 1: 周枢轴点刷新逻辑修复

**User Story:** As a system administrator, I want the weekly pivot refresh logic to work correctly on all days of the week, so that pivot data is always up-to-date.

#### Acceptance Criteria

1. WHEN checking if weekly pivot is stale on Sunday, THE Pivot_Refresher SHALL correctly identify that the current week's Monday has passed
2. WHEN calculating the "current week's Monday" for staleness check, THE Pivot_Refresher SHALL use the Monday of the current ISO week, not the next Monday
3. WHEN the weekly pivot refresh fails on Monday, THE Pivot_Refresher SHALL continue to identify the pivot as stale on subsequent days (Tuesday through Sunday) until successfully refreshed
4. THE Pivot_Refresher SHALL include unit tests covering the Sunday edge case scenario
5. WHEN running on any day of the week, THE Pivot_Refresher SHALL produce consistent staleness results for the same pivot data

### Requirement 2: 参数校验与防护

**User Story:** As a developer, I want the system to validate configuration parameters, so that invalid values do not cause runtime panics.

#### Acceptance Criteria

1. WHEN PATTERN_HISTORY_MAX is set to a negative value, THE Pattern_History module SHALL use the default value (1000) instead of panicking
2. WHEN KLINE_COUNT is set to a negative or zero value, THE Kline_Store SHALL use the default value (12) instead of panicking
3. WHEN any numeric configuration parameter is invalid, THE system SHALL log a warning with the invalid value and the fallback value being used
4. THE system SHALL validate all numeric configuration parameters at startup before any data structures are initialized
5. WHEN PATTERN_MIN_CONFIDENCE is outside the range 0-100, THE system SHALL clamp it to the valid range and log a warning

### Requirement 3: 形态历史文件截断

**User Story:** As a system administrator, I want the pattern history file to have a size limit, so that disk space is not exhausted during long-term operation.

#### Acceptance Criteria

1. THE Pattern_History module SHALL implement file truncation similar to the existing signal history behavior
2. WHEN the pattern history file exceeds MAX_FILE_LINES (default: 10000), THE Pattern_History module SHALL truncate the file to retain only the most recent entries
3. THE truncation operation SHALL preserve the most recent N entries where N equals PATTERN_HISTORY_MAX
4. WHEN truncation occurs, THE system SHALL log the number of entries removed
5. THE truncation check SHALL occur periodically (e.g., every 100 new entries) rather than on every write for performance
6. IF truncation fails due to I/O error, THEN THE system SHALL log the error and continue operation without crashing

### Requirement 4: TickerEvent 数据解析兼容性

**User Story:** As a developer, I want the TickerEvent parser to handle both string and numeric values from Binance, so that data is correctly parsed regardless of format.

#### Acceptance Criteria

1. WHEN Binance returns the event time (E) as a string, THE TickerEvent parser SHALL correctly parse it as an integer
2. WHEN Binance returns the trade count (n) as a string, THE TickerEvent parser SHALL correctly parse it as an integer
3. WHEN parsing fails for a numeric field, THE TickerEvent parser SHALL log a warning with the field name and raw value
4. THE TickerEvent parser SHALL handle both JSON number and JSON string representations for all numeric fields
5. WHEN a numeric field cannot be parsed, THE system SHALL use a sensible default (0) and continue processing rather than failing silently

### Requirement 5: 测试覆盖

**User Story:** As a developer, I want comprehensive tests for the fixed issues, so that regressions can be detected early.

#### Acceptance Criteria

1. THE test suite SHALL include tests for weekly pivot staleness check on all days of the week (Monday through Sunday)
2. THE test suite SHALL include tests for parameter validation with edge cases (negative, zero, very large values)
3. THE test suite SHALL include tests for pattern history truncation behavior
4. THE test suite SHALL include tests for TickerEvent parsing with both string and numeric field values
5. WHEN running `go test ./...`, all new tests SHALL pass

