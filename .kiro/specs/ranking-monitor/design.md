# Design Document: Ranking Monitor

## Overview

排名监控系统通过定时采样机制追踪 USDT 交易对的成交量和成交笔数排名变化。系统每 5 分钟采集一次快照，保留 24 小时滚动数据，支持排名变化计算和价格联动分析。

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP API Layer                            │
│  /api/ranking/current  /api/ranking/history  /api/ranking/movers│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Ranking Store                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  Snapshots  │  │ Symbol Index│  │  Query & Compare Logic  │  │
│  │  (Ring Buf) │  │  (by symbol)│  │                         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ 每5分钟采样
┌─────────────────────────────────────────────────────────────────┐
│                     Ranking Sampler                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ USDT Filter │  │ Rank Calc   │  │  Snapshot Builder       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ 读取 ticker 数据
┌─────────────────────────────────────────────────────────────────┐
│                     Ticker Store (existing)                      │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Ranking Sampler

负责定时采样和排名计算。

```go
package ranking

// Sampler 排名采样器
type Sampler struct {
    tickerStore   *ticker.Store
    rankingStore  *Store
    interval      time.Duration // 默认 5 分钟
}

// NewSampler 创建采样器
func NewSampler(tickerStore *ticker.Store, rankingStore *Store) *Sampler

// Run 启动采样循环
func (s *Sampler) Run(ctx context.Context)

// Sample 执行一次采样，计算排名并返回快照
func (s *Sampler) Sample() *Snapshot

// calculateRanks 计算排名（使用 dense ranking）
// 返回 volumeRanks 和 tradesRanks 两个 map[string]int
func (s *Sampler) calculateRanks(tickers map[string]*ticker.Ticker) (volumeRanks, tradesRanks map[string]int)

// isUSDTPair 检查是否为 USDT 交易对
func isUSDTPair(symbol string) bool {
    return strings.HasSuffix(symbol, "USDT")
}
```

### 2. Ranking Store

负责存储和查询排名快照。

```go
package ranking

// Store 排名存储
type Store struct {
    mu          sync.RWMutex
    snapshots   []*Snapshot          // 按时间顺序存储，新的在后
    maxAge      time.Duration        // 最大保留时间，默认 24h
    dataDir     string               // 持久化目录
}

// NewStore 创建存储
func NewStore(dataDir string, maxAge time.Duration) *Store

// Add 添加快照，自动触发清理
func (s *Store) Add(snapshot *Snapshot)

// GetCurrent 获取当前排名（带变化计算）
func (s *Store) GetCurrent(opts CurrentOptions) *CurrentResponse

// GetHistory 获取指定交易对的历史（遍历所有快照提取该 symbol 的数据）
func (s *Store) GetHistory(symbol string) []SymbolSnapshot

// GetMovers 获取排名变化最大的交易对
func (s *Store) GetMovers(opts MoversOptions) *MoversResponse

// findSnapshotByTime 查找指定时间点之前最近的快照
// 返回 timestamp <= targetTime 的最近快照，如果没有则返回最老的快照
func (s *Store) findSnapshotByTime(targetTime time.Time) *Snapshot

// cleanup 清理过期快照（timestamp < now - maxAge）
func (s *Store) cleanup()

// persist 持久化到磁盘
func (s *Store) persist() error

// load 从磁盘加载
func (s *Store) load() error
```

### 3. HTTP API Handlers

```go
package httpapi

// handleRankingCurrent GET /api/ranking/current
// Query params:
//   - type: volume|trades (default: volume)
//   - compare: 5m|15m|30m|1h|6h|24h (default: previous snapshot)
//   - limit: int (default: 100)
func (s *Server) handleRankingCurrent(w http.ResponseWriter, r *http.Request)

// handleRankingHistory GET /api/ranking/history/{symbol}
func (s *Server) handleRankingHistory(w http.ResponseWriter, r *http.Request)

// handleRankingMovers GET /api/ranking/movers
// Query params:
//   - type: volume|trades (default: volume)
//   - direction: up|down (required)
//   - compare: 5m|15m|30m|1h|6h|24h (default: previous snapshot)
//   - limit: int (default: 20)
func (s *Server) handleRankingMovers(w http.ResponseWriter, r *http.Request)
```

## Data Models

### Snapshot（快照）

```go
// Snapshot 单次采样快照
type Snapshot struct {
    Timestamp time.Time              `json:"timestamp"`
    Items     map[string]*SnapshotItem `json:"items"` // symbol -> item
}

// SnapshotItem 单个交易对的快照数据
type SnapshotItem struct {
    Symbol      string  `json:"symbol"`
    VolumeRank  int     `json:"volume_rank"`
    TradesRank  int     `json:"trades_rank"`
    Price       float64 `json:"price"`
    Volume      float64 `json:"volume"`       // 成交额
    TradeCount  int64   `json:"trade_count"`  // 成交笔数
}
```

### RankingItem（API 响应项）

```go
// RankingItem 排名查询响应项
type RankingItem struct {
    Symbol       string   `json:"symbol"`
    Rank         int      `json:"rank"`
    RankChange   *int     `json:"rank_change,omitempty"`   // 排名变化，正数表示上升
    Price        float64  `json:"price"`
    PriceChange  *float64 `json:"price_change,omitempty"`  // 价格变化百分比
    Volume       float64  `json:"volume"`
    TradeCount   int64    `json:"trade_count"`
    IsNew        bool     `json:"is_new,omitempty"`        // 是否新上榜
}
```

### SymbolSnapshot（交易对历史项）

```go
// SymbolSnapshot 单个交易对的历史快照
type SymbolSnapshot struct {
    Timestamp   time.Time `json:"timestamp"`
    VolumeRank  int       `json:"volume_rank"`
    TradesRank  int       `json:"trades_rank"`
    Price       float64   `json:"price"`
    Volume      float64   `json:"volume"`
    TradeCount  int64     `json:"trade_count"`
}
```

### Query Options

```go
// CurrentOptions 当前排名查询选项
type CurrentOptions struct {
    Type      string        // "volume" or "trades"
    Compare   time.Duration // 比较时间窗口，0 表示与上一快照比较
    Limit     int
}

// CurrentResponse 当前排名响应
type CurrentResponse struct {
    Timestamp time.Time     `json:"timestamp"`
    CompareTo time.Time     `json:"compare_to"`
    Items     []RankingItem `json:"items"`
}

// MoversOptions 异动查询选项
type MoversOptions struct {
    Type      string        // "volume" or "trades"
    Direction string        // "up" or "down" (required)
    Compare   time.Duration
    Limit     int
}

// MoversResponse 异动响应
type MoversResponse struct {
    Timestamp time.Time     `json:"timestamp"`
    CompareTo time.Time     `json:"compare_to"`
    Direction string        `json:"direction"`
    Items     []RankingItem `json:"items"`
}

// HistoryResponse 历史响应
type HistoryResponse struct {
    Symbol    string           `json:"symbol"`
    Snapshots []SymbolSnapshot `json:"snapshots"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: USDT Pair Filtering

*For any* ticker data containing mixed trading pairs (USDT and non-USDT), when the Ranking_Sampler creates a snapshot, all items in the snapshot should have symbols ending with "USDT", and no non-USDT pairs should be included.

**Validates: Requirements 1.1, 1.2, 2.5**

### Property 2: Ranking Order Correctness

*For any* set of USDT trading pairs with distinct volume/trade values, the assigned ranks should be in descending order of the respective metric (volume or trades), with rank 1 assigned to the highest value.

**Validates: Requirements 2.1, 2.2, 2.3**

### Property 3: Equal Value Rank Assignment (Dense Ranking)

*For any* two or more symbols with equal volume (or trade count) values, they should be assigned the same rank, and the next distinct value should receive rank = previous_rank + 1 (dense ranking, not competition ranking).

**Validates: Requirements 2.4**

### Property 4: 24-Hour Retention Window

*For any* sequence of snapshots added to the store, after cleanup, only snapshots with timestamps within the last 24 hours should remain.

**Validates: Requirements 1.4, 1.5**

### Property 5: Rank Change Calculation

*For any* two consecutive snapshots where a symbol exists in both, the rank change should equal (previous_rank - current_rank), positive indicating improvement (lower rank number = better).

**Validates: Requirements 3.1, 3.2, 3.3**

### Property 6: Price Change Calculation

*For any* two snapshots where a symbol exists in both with non-zero previous price, the price change percentage should equal ((current_price - previous_price) / previous_price) * 100. When previous_price is zero, price change should be null.

**Validates: Requirements 4.1, 4.2, 4.4**

### Property 7: History Chronological Order

*For any* symbol's history query result, the snapshots should be ordered by timestamp in ascending order (oldest first).

**Validates: Requirements 6.4**

### Property 8: Movers Sorting

*For any* movers query result, the items should be sorted by absolute rank change in descending order.

**Validates: Requirements 7.5**

### Property 9: Persistence Round Trip

*For any* valid store state, persisting to disk and then loading should produce an equivalent state with all snapshots preserved.

**Validates: Requirements 10.1, 10.2**

## Error Handling

1. **Ticker Store Unavailable**: If ticker store returns no data, sampler should skip the sampling cycle and log a warning.

2. **Persistence Failure**: If disk write fails, log error and continue with in-memory data. Do not crash the service.

3. **Invalid Query Parameters**: Return 400 Bad Request with descriptive error message for invalid compare duration or type values.

4. **Symbol Not Found**: Return empty array for history queries on non-existent symbols.

5. **No Snapshots Available**: Return empty array for current/movers queries when no snapshots exist.

## Testing Strategy

### Unit Tests

- Test `isUSDTPair()` function with various symbol formats
- Test rank calculation with edge cases (empty data, single item, ties)
- Test time window comparison logic
- Test cleanup logic with various timestamp scenarios

### Property-Based Tests

Use Go's `testing/quick` package or a property-based testing library:

1. **USDT Filtering Property**: Generate random ticker data, verify only USDT pairs in snapshot
2. **Ranking Order Property**: Generate random volumes/trades, verify correct ordering
3. **Rank Change Property**: Generate snapshot pairs, verify change calculation
4. **Price Change Property**: Generate price pairs, verify percentage calculation
5. **Retention Property**: Generate snapshots over time, verify cleanup
6. **Persistence Property**: Generate store state, verify round-trip

### Integration Tests

- Test full sampling cycle with mock ticker store
- Test API endpoints with various query parameters
- Test persistence across simulated restarts

## API Response Examples

### GET /api/ranking/current?type=volume&compare=30m&limit=10

```json
{
  "timestamp": "2026-01-03T10:00:00Z",
  "compare_to": "2026-01-03T09:30:00Z",
  "items": [
    {
      "symbol": "BTCUSDT",
      "rank": 1,
      "rank_change": 0,
      "price": 98500.50,
      "price_change": 1.25,
      "volume": 1250000000,
      "trade_count": 850000
    },
    {
      "symbol": "ETHUSDT",
      "rank": 2,
      "rank_change": 1,
      "price": 3450.25,
      "price_change": 2.15,
      "volume": 680000000,
      "trade_count": 520000
    },
    {
      "symbol": "XRPUSDT",
      "rank": 3,
      "rank_change": -1,
      "price": 2.35,
      "price_change": -0.85,
      "volume": 450000000,
      "trade_count": 380000
    }
  ]
}
```

### GET /api/ranking/history/BTCUSDT

```json
{
  "symbol": "BTCUSDT",
  "snapshots": [
    {
      "timestamp": "2026-01-02T10:00:00Z",
      "volume_rank": 1,
      "trades_rank": 1,
      "price": 97000.00,
      "volume": 1100000000,
      "trade_count": 780000
    },
    {
      "timestamp": "2026-01-02T10:05:00Z",
      "volume_rank": 1,
      "trades_rank": 1,
      "price": 97250.50,
      "volume": 1150000000,
      "trade_count": 800000
    }
  ]
}
```

### GET /api/ranking/movers?type=volume&direction=up&limit=5

```json
{
  "timestamp": "2026-01-03T10:00:00Z",
  "compare_to": "2026-01-03T09:55:00Z",
  "direction": "up",
  "items": [
    {
      "symbol": "PEPEUSDT",
      "rank": 15,
      "rank_change": 25,
      "price": 0.00001850,
      "price_change": 8.5,
      "volume": 85000000,
      "trade_count": 125000
    },
    {
      "symbol": "DOGEUSDT",
      "rank": 8,
      "rank_change": 12,
      "price": 0.125,
      "price_change": 5.2,
      "volume": 180000000,
      "trade_count": 95000
    }
  ]
}
```

## Frontend Design

### Ranking Monitor View

```
┌─────────────────────────────────────────────────────────────┐
│  [Volume ▼] [Trades]    Compare: [5m ▼]    [Refresh 🔄]    │
├─────────────────────────────────────────────────────────────┤
│  #1  BTCUSDT     ─      $1.25B   850K trades   +1.25%      │
│  #2  ETHUSDT     ↑1     $680M    520K trades   +2.15%      │
│  #3  XRPUSDT     ↓1     $450M    380K trades   -0.85%      │
│  #4  SOLUSDT     ↑3     $320M    280K trades   +4.50%      │
│  #5  BNBUSDT     ↓2     $280M    195K trades   -1.20%      │
│  ...                                                        │
└─────────────────────────────────────────────────────────────┘
```

### Detail Modal

```
┌─────────────────────────────────────────────────────────────┐
│  BTCUSDT Ranking History                              [×]   │
├─────────────────────────────────────────────────────────────┤
│  Current: Vol #1 | Trades #1 | $98,500.50 (+1.25%)         │
├─────────────────────────────────────────────────────────────┤
│  Volume Rank (24h)                                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1 ─────────────────────────────────────────────────│   │
│  │ 2                                                   │   │
│  │ 3                                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Trades Rank (24h)                                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1 ─────────────────────────────────────────────────│   │
│  │ 2                                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Price Change (24h)                                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │     ╱╲    ╱╲                                        │   │
│  │ ───╱  ╲──╱  ╲───────────────────────────────────── │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```
