# Requirements Document

## Introduction

排名监控系统用于追踪和分析交易对的成交量（Vol）和成交笔数（Trades）排名变化，通过定时采样记录历史数据，帮助用户发现异动交易对并分析量价关系。

## Glossary

- **Ranking_Sampler**: 排名采样器，负责定时采集所有交易对的排名数据
- **Ranking_Store**: 排名存储，负责保存和管理历史排名快照数据
- **Ranking_API**: 排名查询接口，提供 HTTP API 供前端查询排名数据
- **Snapshot**: 快照，某一时刻所有交易对的排名和价格数据
- **Rank_Change**: 排名变化，当前排名与历史排名的差值
- **Price_Change**: 价格变化，当前价格与历史价格的百分比变化
- **Movers**: 异动交易对，排名变化显著的交易对
- **USDT_Pair**: USDT 交易对，以 USDT 为报价货币的交易对（如 BTCUSDT、ETHUSDT）

## Requirements

### Requirement 1: 定时采样

**User Story:** As a user, I want the system to periodically sample ranking data, so that I can track ranking changes over time.

#### Acceptance Criteria

1. THE Ranking_Sampler SHALL sample USDT trading pairs only every 5 minutes
2. THE Ranking_Sampler SHALL exclude non-USDT trading pairs (such as BTC, ETH, BNB, FDUSD quoted pairs)
3. WHEN sampling, THE Ranking_Sampler SHALL record volume rank, trades rank, price, volume value, and trade count for each USDT_Pair
4. THE Ranking_Store SHALL retain snapshots for a rolling 24-hour window (approximately 288 snapshots)
5. WHEN a snapshot exceeds the 24-hour retention period, THE Ranking_Store SHALL automatically remove it
6. WHEN the server starts, THE Ranking_Sampler SHALL begin sampling immediately after ticker data is available

### Requirement 2: 排名计算

**User Story:** As a user, I want to see accurate rankings for volume and trades, so that I can identify the most active trading pairs.

#### Acceptance Criteria

1. WHEN calculating volume rank, THE Ranking_Sampler SHALL sort all USDT_Pairs by quote volume in descending order
2. WHEN calculating trades rank, THE Ranking_Sampler SHALL sort all USDT_Pairs by trade count in descending order
3. THE Ranking_Sampler SHALL assign rank positions starting from 1 for the highest value
4. WHEN two symbols have equal values, THE Ranking_Sampler SHALL assign the same rank to both using dense ranking (next rank is current + 1, not current + count of ties)
5. THE Ranking_Sampler SHALL identify USDT_Pairs by checking if the symbol ends with "USDT"

### Requirement 3: 排名变化计算

**User Story:** As a user, I want to see how rankings have changed over time, so that I can identify trending or declining trading pairs.

#### Acceptance Criteria

1. WHEN querying current rankings, THE Ranking_API SHALL calculate rank change compared to the previous snapshot
2. WHEN a compare parameter is provided, THE Ranking_API SHALL find the snapshot with timestamp closest to but not exceeding (current_time - compare_duration)
3. IF no snapshot exists within the compare window, THE Ranking_API SHALL use the oldest available snapshot
4. THE Ranking_API SHALL express rank change as a signed integer (positive for improvement, negative for decline)
5. WHEN a symbol has no historical data for comparison, THE Ranking_API SHALL return null for rank change
6. WHEN a symbol is new (not in previous snapshot), THE Ranking_API SHALL indicate it as a new entry

### Requirement 4: 价格联动分析

**User Story:** As a user, I want to see price changes alongside ranking changes, so that I can understand the relationship between volume activity and price movement.

#### Acceptance Criteria

1. WHEN returning ranking data, THE Ranking_API SHALL include price change percentage compared to the comparison point
2. THE Ranking_API SHALL calculate price change as ((current_price - previous_price) / previous_price) * 100
3. WHEN a symbol has no historical price data, THE Ranking_API SHALL return null for price change
4. WHEN previous_price is zero or null, THE Ranking_API SHALL return null for price change

### Requirement 5: 当前排名查询 API

**User Story:** As a developer, I want to query current rankings with change indicators, so that I can display ranking data in the frontend.

#### Acceptance Criteria

1. THE Ranking_API SHALL provide a GET /api/ranking/current endpoint
2. WHEN the compare parameter is omitted, THE Ranking_API SHALL compare with the previous snapshot
3. WHEN the compare parameter is provided (e.g., 30m, 1h), THE Ranking_API SHALL compare with the snapshot closest to that time ago
4. THE Ranking_API SHALL return a response object containing timestamp, compare_to timestamp, and items array
5. THE items array SHALL be sorted by the specified type (volume or trades)
6. WHEN the type parameter is omitted, THE Ranking_API SHALL default to volume ranking
7. THE Ranking_API SHALL support a limit parameter to restrict the number of results

### Requirement 6: 交易对历史查询 API

**User Story:** As a user, I want to view the ranking history of a specific trading pair, so that I can analyze its activity trend.

#### Acceptance Criteria

1. THE Ranking_API SHALL provide a GET /api/ranking/history/{symbol} endpoint
2. THE Ranking_API SHALL return all snapshots for the specified symbol within the 24-hour retention window
3. WHEN the symbol has no data, THE Ranking_API SHALL return an empty array
4. THE Ranking_API SHALL return snapshots in chronological order (oldest first)

### Requirement 7: 异动交易对查询 API

**User Story:** As a user, I want to find trading pairs with significant ranking changes, so that I can quickly identify unusual activity.

#### Acceptance Criteria

1. THE Ranking_API SHALL provide a GET /api/ranking/movers endpoint
2. THE Ranking_API SHALL support type parameter (volume or trades) to specify which ranking to analyze
3. THE Ranking_API SHALL require direction parameter (up or down) to filter by rank improvement or decline
4. THE Ranking_API SHALL support limit parameter to restrict the number of results (default 20)
5. THE Ranking_API SHALL sort results by absolute rank change in descending order
6. THE Ranking_API SHALL return a response object containing timestamp, compare_to timestamp, direction, and items array

### Requirement 8: 前端排名监控视图

**User Story:** As a user, I want a dedicated view to monitor rankings, so that I can easily track market activity.

#### Acceptance Criteria

1. THE Dashboard SHALL provide a "Ranking Monitor" view accessible from the navigation
2. WHEN displaying ranking items, THE Dashboard SHALL show symbol, current rank, rank change indicator, price change percentage, and absolute values
3. THE Dashboard SHALL display rank improvement with an up arrow (↑) in green
4. THE Dashboard SHALL display rank decline with a down arrow (↓) in red
5. THE Dashboard SHALL support switching between volume ranking and trades ranking
6. THE Dashboard SHALL support selecting comparison time window (5m, 15m, 30m, 1h, 6h, 24h)

### Requirement 9: 交易对详情弹窗

**User Story:** As a user, I want to view detailed ranking history for a trading pair, so that I can analyze its activity pattern.

#### Acceptance Criteria

1. WHEN a user clicks on a ranking item, THE Dashboard SHALL display a detail modal
2. THE Detail_Modal SHALL show the symbol name and current values
3. THE Detail_Modal SHALL display a chart showing volume rank history over 24 hours
4. THE Detail_Modal SHALL display a chart showing trades rank history over 24 hours
5. THE Detail_Modal SHALL display a chart showing price change history over 24 hours
6. THE Detail_Modal SHALL display rank changes and price changes on the same time axis to show correlation visually

### Requirement 10: 数据持久化

**User Story:** As a system administrator, I want ranking data to persist across server restarts, so that historical data is not lost.

#### Acceptance Criteria

1. THE Ranking_Store SHALL persist snapshots to disk periodically
2. WHEN the server starts, THE Ranking_Store SHALL load existing snapshots from disk
3. THE Ranking_Store SHALL store data in the configured data directory
4. IF persistence fails, THE Ranking_Store SHALL log the error and continue operating with in-memory data
