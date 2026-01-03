// Package ranking provides ranking monitoring for trading pairs.
package ranking

import "time"

// Snapshot 单次采样快照
type Snapshot struct {
	Timestamp time.Time               `json:"timestamp"`
	Items     map[string]*SnapshotItem `json:"items"` // symbol -> item
}

// SnapshotItem 单个交易对的快照数据
type SnapshotItem struct {
	Symbol     string  `json:"symbol"`
	VolumeRank int     `json:"volume_rank"`
	TradesRank int     `json:"trades_rank"`
	Price      float64 `json:"price"`
	Volume     float64 `json:"volume"`      // 成交额
	TradeCount int64   `json:"trade_count"` // 成交笔数
}

// RankingItem 排名查询响应项
type RankingItem struct {
	Symbol       string   `json:"symbol"`
	Rank         int      `json:"rank"`
	RankChange   *int     `json:"rank_change,omitempty"`   // 排名变化，正数表示上升
	Price        float64  `json:"price"`
	PriceChange  *float64 `json:"price_change,omitempty"`  // 价格变化百分比
	Volume       float64  `json:"volume"`
	VolumeChange *float64 `json:"volume_change,omitempty"` // 成交额变化百分比
	TradeCount   int64    `json:"trade_count"`
	TradeChange  *float64 `json:"trade_change,omitempty"` // 成交笔数变化百分比
	IsNew        bool     `json:"is_new,omitempty"`        // 是否新上榜
}

// SymbolSnapshot 单个交易对的历史快照
type SymbolSnapshot struct {
	Timestamp  time.Time `json:"timestamp"`
	VolumeRank int       `json:"volume_rank"`
	TradesRank int       `json:"trades_rank"`
	Price      float64   `json:"price"`
	Volume     float64   `json:"volume"`
	TradeCount int64     `json:"trade_count"`
}

// CurrentOptions 当前排名查询选项
type CurrentOptions struct {
	Type    string        // "volume" or "trades"
	Compare time.Duration // 比较时间窗口，0 表示与上一快照比较
	Limit   int
}

// CurrentResponse 当前排名响应
type CurrentResponse struct {
	Timestamp time.Time     `json:"timestamp,omitempty"`
	CompareTo time.Time     `json:"compare_to,omitempty"`
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
	Timestamp time.Time     `json:"timestamp,omitempty"`
	CompareTo time.Time     `json:"compare_to,omitempty"`
	Direction string        `json:"direction"`
	Items     []RankingItem `json:"items"`
}

// HistoryResponse 历史响应
type HistoryResponse struct {
	Symbol    string           `json:"symbol"`
	Snapshots []SymbolSnapshot `json:"snapshots"`
}

// RankingType 排名类型常量
const (
	RankingTypeVolume = "volume"
	RankingTypeTrades = "trades"
)

// Direction 方向常量
const (
	DirectionUp   = "up"
	DirectionDown = "down"
)
