# Design Document: K 线形态识别系统

## Overview

本设计文档描述了基于现有币安 WebSocket 价格流构建虚拟 K 线并进行形态识别的系统架构。系统将在内存中维护滚动 K 线数据，在每根 K 线收盘时触发形态检测，并将检测结果与现有枢轴点信号系统集成。

### 设计决策

1. **使用 talib-cdl-go 库**: 采用 `github.com/iwat/talib-cdl-go` 库进行 K 线形态识别，这是 TA-Lib CDL 模块的纯 Go 实现，支持多种经典形态。
2. **内存优先**: K 线数据仅存储在内存中，不持久化到磁盘，以保证性能。
3. **异步处理**: 形态检测在独立 goroutine 中执行，不阻塞主价格处理流程。
4. **事件驱动**: 仅在 K 线收盘时触发形态检测，而非每次价格更新。

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Binance WebSocket                            │
│                    (!markPrice@arr@1s)                              │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Monitor.onPrice()                           │
│                    (existing price handler)                         │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
┌───────────────────────────────┐   ┌───────────────────────────────┐
│      Pivot Check (existing)   │   │      KlineStore.Update()      │
│   checkPeriod() → emit()      │   │   Update OHLC, check close    │
└───────────────────────────────┘   └───────────────────────────────┘
                    │                           │
                    │                           │ (on kline close)
                    │                           ▼
                    │               ┌───────────────────────────────┐
                    │               │   PatternDetector.Detect()    │
                    │               │   (async goroutine)           │
                    │               └───────────────────────────────┘
                    │                           │
                    │                           │ (pattern found)
                    │                           ▼
                    │               ┌───────────────────────────────┐
                    │               │   PatternSignal.emit()        │
                    │               │   → History, SSE Broker       │
                    │               └───────────────────────────────┘
                    │                           │
                    └─────────────┬─────────────┘
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      SignalCombiner                                 │
│              Correlate pivot + pattern signals                      │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        SSE Broker                                   │
│                   Push to frontend                                  │
└─────────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Kline 数据结构

```go
// internal/kline/kline.go

package kline

import "time"

// Kline 表示单根 K 线数据
type Kline struct {
    Symbol    string    `json:"symbol"`
    Open      float64   `json:"open"`
    High      float64   `json:"high"`
    Low       float64   `json:"low"`
    Close     float64   `json:"close"`
    OpenTime  time.Time `json:"open_time"`
    CloseTime time.Time `json:"close_time"`
    IsClosed  bool      `json:"is_closed"`
}

// Body 返回 K 线实体大小（绝对值）
func (k *Kline) Body() float64 {
    return abs(k.Close - k.Open)
}

// UpperShadow 返回上影线长度
func (k *Kline) UpperShadow() float64 {
    if k.Close > k.Open {
        return k.High - k.Close
    }
    return k.High - k.Open
}

// LowerShadow 返回下影线长度
func (k *Kline) LowerShadow() float64 {
    if k.Close > k.Open {
        return k.Open - k.Low
    }
    return k.Close - k.Low
}

// IsBullish 判断是否为阳线
func (k *Kline) IsBullish() bool {
    return k.Close > k.Open
}

// IsBearish 判断是否为阴线
func (k *Kline) IsBearish() bool {
    return k.Close < k.Open
}

// Range 返回 K 线振幅
func (k *Kline) Range() float64 {
    return k.High - k.Low
}
```

### 2. KlineStore 存储模块

```go
// internal/kline/store.go

package kline

import (
    "sync"
    "time"
)

// Store 管理所有交易对的 K 线数据
type Store struct {
    mu       sync.RWMutex
    klines   map[string]*SymbolKlines  // symbol -> klines
    interval time.Duration             // K 线间隔（默认 5 分钟）
    maxCount int                       // 最大保留数量（默认 12）
    onClose  func(symbol string, klines []Kline)  // K 线收盘回调（传入深拷贝快照）
}

// SymbolKlines 单个交易对的 K 线数据
type SymbolKlines struct {
    Symbol   string
    Current  *Kline   // 当前正在形成的 K 线
    History  []Kline  // 已完成的历史 K 线
    // 重要约定：History 按时间顺序存储（最旧在前，最新在后）
    // 这与 talib-cdl-go 的输入要求一致
    LastSeen time.Time // 最后更新时间，用于清理长期无更新的 symbol
}

// NewStore 创建 K 线存储
func NewStore(interval time.Duration, maxCount int) *Store

// SetOnClose 设置 K 线收盘回调
// 重要：回调函数接收的 klines 是深拷贝快照，可安全在 goroutine 中使用
func (s *Store) SetOnClose(fn func(symbol string, klines []Kline))

// Update 更新价格数据
// 返回值: 是否触发了 K 线收盘
// 线程安全：内部使用互斥锁保护
func (s *Store) Update(symbol string, price float64, ts time.Time) bool

// GetKlines 获取指定交易对的 K 线数据（深拷贝）
// 返回值：按时间顺序排列（最旧在前，最新在后）
// 这是 talib-cdl-go 和自实现形态检测的标准输入格式
func (s *Store) GetKlines(symbol string) ([]Kline, bool)

// GetCurrentKline 获取当前正在形成的 K 线（深拷贝）
func (s *Store) GetCurrentKline(symbol string) (*Kline, bool)

// CleanupStale 清理长期无更新的 symbol（可选，定期调用）
// staleThreshold: 超过此时间无更新的 symbol 将被清理
func (s *Store) CleanupStale(staleThreshold time.Duration) int
```

### K 线序列方向约定（重要）

**全系统统一约定：K 线序列按时间顺序排列（最旧在前，最新在后）**

这与 talib-cdl-go 库的输入要求一致，也是大多数技术分析库的标准格式。

```go
// 正确的序列方向示例
// klines[0] = 最旧的 K 线（例如 10:00）
// klines[1] = 较新的 K 线（例如 10:05）
// klines[n-1] = 最新的 K 线（例如 10:55）

// 在 Store.Update 中，收盘时的处理逻辑：
func (s *Store) Update(symbol string, price float64, ts time.Time) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    sk := s.getOrCreate(symbol)
    sk.LastSeen = ts
    
    // ... 更新当前 K 线的 OHLC ...
    
    // 检查是否需要收盘
    if shouldClose(sk.Current, ts, s.interval) {
        // 1. 将当前 K 线追加到历史末尾（保持时间顺序）
        sk.History = append(sk.History, *sk.Current)
        
        // 2. 维护滚动窗口大小
        if len(sk.History) > s.maxCount {
            sk.History = sk.History[len(sk.History)-s.maxCount:]
        }
        
        // 3. 生成深拷贝快照（在锁内完成）
        snapshot := make([]Kline, len(sk.History))
        copy(snapshot, sk.History)
        
        // 4. 创建新 K 线
        sk.Current = &Kline{
            Symbol:   symbol,
            Open:     price,
            High:     price,
            Low:      price,
            Close:    price,
            OpenTime: getKlineOpenTime(ts, s.interval),
        }
        
        // 5. 解锁后异步调用回调（避免长时间持锁）
        if s.onClose != nil {
            go s.onClose(symbol, snapshot)
        }
        
        return true
    }
    
    return false
}
```

### 3. PatternDetector 形态检测模块

```go
// internal/pattern/detector.go

package pattern

import (
    talibcdl "github.com/iwat/talib-cdl-go"
    "example.com/binance-pivot-monitor/internal/kline"
)

// PatternType 形态类型
type PatternType string

const (
    // === talib-cdl-go 库支持的形态 ===
    
    // 反转形态
    PatternDoji              PatternType = "doji"              // 十字星
    PatternDojiStar          PatternType = "doji_star"         // 十字星线
    PatternEveningStar       PatternType = "evening_star"      // 暮星
    PatternPiercing          PatternType = "piercing"          // 刺透形态
    PatternAbandonedBaby     PatternType = "abandoned_baby"    // 弃婴形态
    PatternMatchingLow       PatternType = "matching_low"      // 相同低价
    
    // 持续形态
    PatternThreeWhite        PatternType = "three_white"       // 三白兵
    PatternThreeBlack        PatternType = "three_black"       // 三只乌鸦
    PatternThreeInside       PatternType = "three_inside"      // 三内部
    PatternThreeOutside      PatternType = "three_outside"     // 三外部
    PatternThreeLineStrike   PatternType = "three_line_strike" // 三线打击
    PatternThreeStarsInSouth PatternType = "three_stars_south" // 南方三星
    
    // 其他形态
    PatternAdvanceBlock      PatternType = "advance_block"     // 前进受阻
    PatternBeltHold          PatternType = "belt_hold"         // 捉腰带线
    PatternBreakAway         PatternType = "break_away"        // 脱离形态
    PatternClosingMarubozu   PatternType = "closing_marubozu"  // 收盘光头光脚
    PatternTwoCrows          PatternType = "two_crows"         // 两只乌鸦
    PatternStickSandwich     PatternType = "stick_sandwich"    // 条形三明治
    PatternConcealBabySwall  PatternType = "conceal_baby"      // 藏婴吞没
    
    // === 需要自实现的高效形态（talib-cdl-go 未实现）===
    
    PatternHammer            PatternType = "hammer"            // 锤子线
    PatternInvertedHammer    PatternType = "inverted_hammer"   // 倒锤子线
    PatternHangingMan        PatternType = "hanging_man"       // 上吊线
    PatternShootingStar      PatternType = "shooting_star"     // 流星线
    PatternEngulfing         PatternType = "engulfing"         // 吞没形态
    PatternMorningStar       PatternType = "morning_star"      // 晨星
    PatternMorningDojiStar   PatternType = "morning_doji_star" // 晨十字星
    PatternEveningDojiStar   PatternType = "evening_doji_star" // 暮十字星
    PatternDarkCloudCover    PatternType = "dark_cloud_cover"  // 乌云盖顶
    PatternHarami            PatternType = "harami"            // 孕线
    PatternHaramiCross       PatternType = "harami_cross"      // 十字孕线
    PatternKicking           PatternType = "kicking"           // 反冲形态
    PatternDragonflyDoji     PatternType = "dragonfly_doji"    // 蜻蜓十字
    PatternGravestoneDoji    PatternType = "gravestone_doji"   // 墓碑十字
)

// Direction 形态方向
type Direction string

const (
    DirectionBullish Direction = "bullish"  // 看涨
    DirectionBearish Direction = "bearish"  // 看跌
    DirectionNeutral Direction = "neutral"  // 中性
)

// DetectedPattern 检测到的形态
type DetectedPattern struct {
    Type       PatternType
    Direction  Direction
    Confidence int  // 0-100，基于 talib-cdl-go 返回值计算
}

// Detector 形态检测器
type Detector struct {
    minConfidence    int  // 最小置信度阈值
    highEfficiencyOnly bool // 是否只检测高效形态
}

// NewDetector 创建检测器
func NewDetector(minConfidence int, highEfficiencyOnly bool) *Detector

// Detect 检测 K 线形态
// klines: 按时间顺序排列的 K 线（最旧在前）
// 返回检测到的所有形态
func (d *Detector) Detect(klines []kline.Kline) []DetectedPattern

// toSeries 将 K 线数据转换为 talib-cdl-go 的 Series 格式
func toSeries(klines []kline.Kline) talibcdl.Series {
    n := len(klines)
    series := talibcdl.Series{
        Open:  make([]float64, n),
        High:  make([]float64, n),
        Low:   make([]float64, n),
        Close: make([]float64, n),
    }
    for i, k := range klines {
        series.Open[i] = k.Open
        series.High[i] = k.High
        series.Low[i] = k.Low
        series.Close[i] = k.Close
    }
    return series
}
```

### 自实现形态检测算法

以下形态 talib-cdl-go 库未实现，需要自行实现：

```go
// internal/pattern/custom.go

package pattern

import "example.com/binance-pivot-monitor/internal/kline"

// detectHammer 检测锤子线
// 条件: 下影线 >= 实体2倍，上影线很小，出现在下跌后
// 改进：使用 3 根 K 线判断趋势，而非仅看前一根
func detectHammer(klines []kline.Kline) (bool, Direction, int) {
    if len(klines) < 3 { // 需要至少 3 根判断趋势
        return false, "", 0
    }
    k := klines[len(klines)-1]
    
    body := k.Body()
    if body == 0 || k.Range() == 0 {
        return false, "", 0
    }
    
    lowerShadow := k.LowerShadow()
    upperShadow := k.UpperShadow()
    
    // 下影线至少是实体的 2 倍
    if lowerShadow < body*2 {
        return false, "", 0
    }
    // 上影线很小（< 实体的 30%）
    if upperShadow > body*0.3 {
        return false, "", 0
    }
    
    // 改进的趋势判断：检查最近 3 根 K 线是否呈下跌趋势
    // 条件：收盘价递减 或 至少 2 根阴线
    if !isDowntrend(klines[len(klines)-3:]) {
        return false, "", 0
    }
    
    confidence := 70
    if lowerShadow >= body*3 {
        confidence = 85
    }
    return true, DirectionBullish, confidence
}

// isDowntrend 判断是否为下跌趋势
// 条件：收盘价递减 或 至少 2/3 为阴线
func isDowntrend(klines []kline.Kline) bool {
    if len(klines) < 2 {
        return false
    }
    
    // 方法1：收盘价递减
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
    
    // 方法2：至少 2/3 为阴线
    bearishCount := 0
    for _, k := range klines {
        if k.IsBearish() {
            bearishCount++
        }
    }
    return bearishCount >= (len(klines)*2)/3
}

// isUptrend 判断是否为上涨趋势
func isUptrend(klines []kline.Kline) bool {
    if len(klines) < 2 {
        return false
    }
    
    // 方法1：收盘价递增
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
    
    // 方法2：至少 2/3 为阳线
    bullishCount := 0
    for _, k := range klines {
        if k.IsBullish() {
            bullishCount++
        }
    }
    return bullishCount >= (len(klines)*2)/3
}

// detectEngulfing 检测吞没形态
func detectEngulfing(klines []kline.Kline) (bool, Direction, int) {
    if len(klines) < 2 {
        return false, "", 0
    }
    curr := klines[len(klines)-1]
    prev := klines[len(klines)-2]
    
    // 看涨吞没: 前阴后阳，当前实体包含前一根
    if prev.IsBearish() && curr.IsBullish() {
        if curr.Open <= prev.Close && curr.Close >= prev.Open {
            confidence := 75
            if curr.Body() > prev.Body()*1.5 {
                confidence = 90
            }
            return true, DirectionBullish, confidence
        }
    }
    
    // 看跌吞没: 前阳后阴，当前实体包含前一根
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

// detectMorningStar 检测晨星
func detectMorningStar(klines []kline.Kline) (bool, Direction, int) {
    if len(klines) < 3 {
        return false, "", 0
    }
    first := klines[len(klines)-3]
    second := klines[len(klines)-2]
    third := klines[len(klines)-1]
    
    // 第一根大阴线
    if !first.IsBearish() || first.Body() < first.Range()*0.6 {
        return false, "", 0
    }
    // 第二根小实体
    if second.Body() > first.Body()*0.3 {
        return false, "", 0
    }
    // 第三根大阳线
    if !third.IsBullish() || third.Body() < third.Range()*0.6 {
        return false, "", 0
    }
    // 第三根收盘进入第一根实体
    midFirst := (first.Open + first.Close) / 2
    if third.Close < midFirst {
        return false, "", 0
    }
    
    return true, DirectionBullish, 80
}

// detectDarkCloudCover 检测乌云盖顶
// 注意：加密市场 24/7 连续交易，几乎无跳空
// 因此放宽条件：当前开盘 >= 前收盘（而非严格 > 前最高）
func detectDarkCloudCover(klines []kline.Kline) (bool, Direction, int) {
    if len(klines) < 2 {
        return false, "", 0
    }
    prev := klines[len(klines)-2]
    curr := klines[len(klines)-1]
    
    // 前一根大阳线
    if !prev.IsBullish() || prev.Body() < prev.Range()*0.6 {
        return false, "", 0
    }
    // 当前阴线
    if !curr.IsBearish() {
        return false, "", 0
    }
    // 放宽条件：当前开盘 >= 前收盘（加密市场适配）
    // 传统定义要求 curr.Open > prev.High（跳空高开）
    if curr.Open < prev.Close {
        return false, "", 0
    }
    // 收盘深入前实体50%以上
    midPrev := (prev.Open + prev.Close) / 2
    if curr.Close > midPrev {
        return false, "", 0
    }
    
    // 如果有跳空（curr.Open > prev.High），置信度更高
    confidence := 70
    if curr.Open > prev.High {
        confidence = 85
    }
    
    return true, DirectionBearish, confidence
}

// detectShootingStar 检测流星线
// 改进：使用 3 根 K 线判断上涨趋势
func detectShootingStar(klines []kline.Kline) (bool, Direction, int) {
    if len(klines) < 3 { // 需要至少 3 根判断趋势
        return false, "", 0
    }
    k := klines[len(klines)-1]
    
    body := k.Body()
    if body == 0 || k.Range() == 0 {
        return false, "", 0
    }
    
    upperShadow := k.UpperShadow()
    lowerShadow := k.LowerShadow()
    
    // 上影线至少是实体的 2 倍
    if upperShadow < body*2 {
        return false, "", 0
    }
    // 下影线很小
    if lowerShadow > body*0.3 {
        return false, "", 0
    }
    // 改进的趋势判断：检查最近 3 根 K 线是否呈上涨趋势
    if !isUptrend(klines[len(klines)-3:]) {
        return false, "", 0
    }
    
    confidence := 70
    if upperShadow >= body*3 {
        confidence = 85
    }
    return true, DirectionBearish, confidence
}
```

### 4. PatternSignal 信号结构

```go
// internal/pattern/signal.go

package pattern

import "time"

// Signal 形态信号
type Signal struct {
    ID             string      `json:"id"`
    Symbol         string      `json:"symbol"`
    Pattern        PatternType `json:"pattern"`
    PatternCN      string      `json:"pattern_cn"`      // 中文名称
    Direction      Direction   `json:"direction"`
    Confidence     int         `json:"confidence"`      // 0-100
    UpPercent      int         `json:"up_percent"`      // 历史上涨概率
    DownPercent    int         `json:"down_percent"`    // 历史下跌概率
    EfficiencyRank string      `json:"efficiency_rank"` // 效率排名
    Source         string      `json:"source"`          // 检测来源: "talib" 或 "custom"
    StatsSource    string      `json:"stats_source"`    // 统计数据来源
    IsEstimated    bool        `json:"is_estimated"`    // 统计数据是否为估算
    KlineTime      time.Time   `json:"kline_time"`      // K 线收盘时间
    DetectedAt     time.Time   `json:"detected_at"`
}

// PatternNames 形态中文名称映射
var PatternNames = map[PatternType]string{
    // talib-cdl-go 库支持的形态
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
    
    // 自实现的高效形态
    PatternHammer:            "锤子线",
    PatternInvertedHammer:    "倒锤子线",
    PatternHangingMan:        "上吊线",
    PatternShootingStar:      "流星线",
    PatternEngulfing:         "吞没形态",
    PatternMorningStar:       "晨星",
    PatternMorningDojiStar:   "晨十字星",
    PatternEveningDojiStar:   "暮十字星",
    PatternDarkCloudCover:    "乌云盖顶",
    PatternHarami:            "孕线",
    PatternHaramiCross:       "十字孕线",
    PatternKicking:           "反冲形态",
    PatternDragonflyDoji:     "蜻蜓十字",
    PatternGravestoneDoji:    "墓碑十字",
}

// NewSignal 创建信号并填充统计数据
func NewSignal(symbol string, pattern PatternType, direction Direction, confidence int, klineTime time.Time) Signal {
    stats := PatternStatsMap[pattern]
    return Signal{
        ID:             generateID(symbol, pattern, klineTime), // 使用 symbol+pattern+time 避免冲突
        Symbol:         symbol,
        Pattern:        pattern,
        PatternCN:      PatternNames[pattern],
        Direction:      direction,
        Confidence:     confidence,
        UpPercent:      stats.UpPercent,
        DownPercent:    stats.DownPercent,
        EfficiencyRank: stats.EfficiencyRank,
        KlineTime:      klineTime,
        DetectedAt:     time.Now().UTC(),
    }
}

// generateID 生成唯一信号 ID
// 使用 symbol + pattern + klineTime 组合，避免同一根 K 线多个形态 ID 冲突
func generateID(symbol string, pattern PatternType, klineTime time.Time) string {
    // 格式: {klineTime_unix_nano}-{symbol}-{pattern}
    return fmt.Sprintf("%d-%s-%s", klineTime.UnixNano(), symbol, pattern)
}
```

### 加密市场适配说明

加密货币市场 24/7 连续交易，与传统股票市场有显著差异：

**1. 几乎无跳空（Gap）**
- 传统形态如 AbandonedBaby、Kicking、DarkCloudCover 依赖跳空
- 在加密市场中，这些形态触发频率极低

**2. 适配策略**
```go
// CryptoMode 配置（可通过环境变量 PATTERN_CRYPTO_MODE=true 开启）
type DetectorConfig struct {
    CryptoMode    bool    // 加密市场模式
    GapThreshold  float64 // 跳空阈值（默认 0.001 = 0.1%）
}

// 在加密模式下：
// - 放宽跳空条件：curr.Open >= prev.Close * (1 + GapThreshold)
// - 降低依赖跳空形态的置信度
// - 默认禁用纯跳空形态（AbandonedBaby、Kicking）
```

**3. 形态分类**

| 形态 | 跳空依赖 | 加密市场建议 |
|------|----------|--------------|
| AbandonedBaby | 强依赖 | 默认禁用 |
| Kicking | 强依赖 | 默认禁用 |
| DarkCloudCover | 中等依赖 | 放宽条件 |
| MorningStar/EveningStar | 弱依赖 | 正常使用 |
| Engulfing | 无依赖 | 正常使用 |
| Hammer/ShootingStar | 无依赖 | 正常使用 |

### 5. SignalCombiner 信号组合模块

```go
// internal/signal/combiner.go

package signal

import (
    "sync"
    "time"
    
    "example.com/binance-pivot-monitor/internal/pattern"
)

// CorrelationStrength 相关性强度
type CorrelationStrength string

const (
    CorrelationStrong   CorrelationStrength = "strong"   // 方向一致
    CorrelationModerate CorrelationStrength = "moderate" // 中等
    CorrelationWeak     CorrelationStrength = "weak"     // 方向冲突
)

// CombinedSignal 组合信号
type CombinedSignal struct {
    PivotSignal   *Signal          `json:"pivot_signal"`
    PatternSignal *pattern.Signal  `json:"pattern_signal"`
    Correlation   CorrelationStrength `json:"correlation"`
    CombinedAt    time.Time        `json:"combined_at"`
}

// Combiner 信号组合器
type Combiner struct {
    mu             sync.RWMutex
    recentPivots   map[string][]Signal         // symbol -> recent pivot signals
    recentPatterns map[string][]pattern.Signal // symbol -> recent pattern signals
    window         time.Duration               // 关联时间窗口（默认 15 分钟）
    onCombined     func(CombinedSignal)
}

// NewCombiner 创建组合器
func NewCombiner(window time.Duration) *Combiner

// SetOnCombined 设置组合信号回调
func (c *Combiner) SetOnCombined(fn func(CombinedSignal))

// AddPivotSignal 添加枢轴点信号
func (c *Combiner) AddPivotSignal(sig Signal)

// AddPatternSignal 添加形态信号
func (c *Combiner) AddPatternSignal(sig pattern.Signal)

// checkCorrelation 检查信号相关性
func (c *Combiner) checkCorrelation(pivot Signal, pat pattern.Signal) CorrelationStrength
```

### 6. PatternHistory 历史记录

```go
// internal/pattern/history.go

package pattern

import (
    "bufio"
    "encoding/json"
    "os"
    "sync"
)

// History 形态信号历史
// 存储策略：内存为主，落盘可选
// - 默认：仅内存存储，重启后清空
// - 可选：配置 PATTERN_HISTORY_FILE 环境变量开启落盘
type History struct {
    mu          sync.RWMutex
    signals     []Signal
    maxSize     int
    filePath    string // 为空则不落盘
    persistMode bool   // 是否开启持久化
}

// NewHistory 创建历史记录
// filePath: 为空字符串则仅内存模式，非空则开启落盘
func NewHistory(filePath string, maxSize int) (*History, error)

// Add 添加信号
// 如果开启持久化，会同步写入文件
func (h *History) Add(sig Signal) error

// Recent 获取最近的信号
func (h *History) Recent(limit int) []Signal

// Query 查询信号
func (h *History) Query(opts QueryOptions) []Signal

// IsPersistent 返回是否开启了持久化
func (h *History) IsPersistent() bool

// QueryOptions 查询选项
type QueryOptions struct {
    Symbol    string
    Pattern   PatternType
    Direction Direction
    Limit     int
    Since     time.Time
}
```

### 存储策略说明

**K 线数据**：仅内存存储，不持久化
- 原因：K 线数据可从价格流实时重建，无需持久化
- 重启后：需要等待足够的 K 线数据积累后才能进行形态检测

**形态信号**：内存为主，落盘可选
- 默认模式：仅内存存储，重启后历史清空
- 持久化模式：配置 `PATTERN_HISTORY_FILE=data/patterns/history.jsonl` 开启
- 持久化时：信号同步追加到 JSONL 文件，启动时加载最近 N 条

## Data Models

### K 线时间边界计算

```go
// getKlineOpenTime 计算 K 线开盘时间
// 对于 5 分钟 K 线，时间边界为 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55
func getKlineOpenTime(ts time.Time, interval time.Duration) time.Time {
    intervalMinutes := int(interval.Minutes())
    minute := ts.Minute()
    alignedMinute := (minute / intervalMinutes) * intervalMinutes
    return time.Date(
        ts.Year(), ts.Month(), ts.Day(),
        ts.Hour(), alignedMinute, 0, 0,
        ts.Location(),
    )
}

// getKlineCloseTime 计算 K 线收盘时间
func getKlineCloseTime(openTime time.Time, interval time.Duration) time.Time {
    return openTime.Add(interval)
}
```

### 形态检测使用 talib-cdl-go

talib-cdl-go 库提供了以下形态检测函数，每个函数返回 `[]int`，其中：
- 正值表示看涨信号
- 负值表示看跌信号
- 0 表示未检测到形态
- 绝对值表示信号强度（通常为 100）

```go
// 使用示例
series := talibcdl.Series{
    Open:  []float64{...},
    High:  []float64{...},
    Low:   []float64{...},
    Close: []float64{...},
}

// 检测十字星
results := talibcdl.Doji(series)
if results[len(results)-1] != 0 {
    // 检测到十字星
}

// 检测暮星（需要 penetration 参数）
results := talibcdl.EveningStar(series, 0.3)

// 检测三白兵
results := talibcdl.ThreeWhiteSoldiers(series)

// 检测三只乌鸦
results := talibcdl.ThreeBlackCrows(series)
```

### 支持的形态列表

#### talib-cdl-go 库支持的形态

| 函数名 | 形态 | 方向 | Up% | Down% | 效率 | 说明 |
|--------|------|------|-----|-------|------|------|
| Doji | 十字星 | 中性 | 43% | 57% | J+ | 开盘价≈收盘价 |
| DojiStar | 十字星线 | 看跌 | 36% | 64% | E- | 上涨后出现十字星 |
| EveningStar | 暮星 | 看跌 | 28% | 72% | A | 三根 K 线反转形态 |
| Piercing | 刺透形态 | 看涨 | 64% | 39% | B+ | 阴线后阳线穿透 |
| AbandonedBaby | 弃婴形态 | 反转 | 70% | 30% | A- | 跳空十字星 |
| ThreeWhiteSoldiers | 三白兵 | 看涨 | 82% | 18% | D+ | 连续三根阳线 |
| ThreeBlackCrows | 三只乌鸦 | 看跌 | 22% | 78% | A+ | 连续三根阴线 |
| ThreeInside | 三内部 | 反转 | 40% | 60% | F | 孕线确认形态 |
| ThreeOutside | 三外部 | 反转 | 31% | 69% | D- | 吞没确认形态 |
| ThreeLineStrike | 三线打击 | 反转 | 35% | 65% | A+ | 三线后反向大线 |
| ThreeStarsInSouth | 南方三星 | 看涨 | 86% | 14% | J- | 三根递减阴线 |
| AdvanceBlock | 前进受阻 | 看跌 | 64% | 36% | F | 上涨动能减弱 |
| BeltHold | 捉腰带线 | 反转 | 71% | 29% | G+ | 光头/光脚线 |
| BreakAway | 脱离形态 | 反转 | 63% | 37% | B+ | 跳空后回补 |
| ClosingMarubozu | 收盘光头光脚 | 持续 | 52% | 48% | E+ | 收盘价=最高/最低 |
| TwoCrows | 两只乌鸦 | 看跌 | 46% | 54% | G+ | 上涨后两根阴线 |
| MatchingLow | 相同低价 | 看涨 | 39% | 61% | A- | 两根阴线低点相同 |
| StickSandwich | 条形三明治 | 看涨 | 38% | 62% | B | 阴-阳-阴形态 |

#### 自实现的高效形态

以下形态的统计数据来源于多个研究：
- [fivehundred.co](https://fivehundred.co/) - 40年127百万根K线的统计分析
- [patternswizard.com](https://patternswizard.com/) - 4120个市场的形态回测
- [feedroll.com](http://feedroll.com/) - TA-Lib 官方引用的统计来源

| 形态 | 方向 | 成功率 | 效率 | 数据来源 | 说明 |
|------|------|--------|------|----------|------|
| Hammer | 锤子线 | 看涨 | 60% | B+ | fivehundred.co | 下影线长，上影线短 |
| InvertedHammer | 倒锤子线 | 看涨 | 55% | C+ | fivehundred.co | 上影线长，下影线短 |
| HangingMan | 上吊线 | 看跌 | 59% | B | fivehundred.co | 锤子线出现在顶部 |
| ShootingStar | 流星线 | 看跌 | 62% | A- | fivehundred.co | 倒锤子出现在顶部 |
| Engulfing | 吞没形态 | 反转 | 67% | A | patternswizard.com | 实体完全包含前一根 |
| MorningStar | 晨星 | 看涨 | 70% | A | stockgro.club | 三根 K 线反转形态 |
| MorningDojiStar | 晨十字星 | 看涨 | 68% | A- | 估算 | 晨星中间为十字星 |
| EveningDojiStar | 暮十字星 | 看跌 | 68% | A- | 估算 | 暮星中间为十字星 |
| DarkCloudCover | 乌云盖顶 | 看跌 | 70% | A | fivehundred.co | 阳线后阴线深入 |
| Harami | 孕线 | 反转 | 53% | C | fivehundred.co | 小实体在大实体内 |
| HaramiCross | 十字孕线 | 反转 | 55% | B- | 估算 | 孕线中间为十字星 |
| Kicking | 反冲形态 | 反转 | 69% | A+ | feedroll.com | 跳空光头光脚线 |
| DragonflyDoji | 蜻蜓十字 | 看涨 | 57% | C+ | fivehundred.co | 长下影线十字星 |
| GravestoneDoji | 墓碑十字 | 看跌 | 57% | C+ | fivehundred.co | 长上影线十字星 |

**注意**: 标记为"估算"的数据是基于相似形态推算的，实际效果可能有差异。建议用户在实盘中验证。

### 形态筛选策略

基于历史统计数据，我们将形态分为三个优先级：

**高优先级（Efficiency Rank A/B）- 推荐启用**:
- EveningStar (暮星) - 72% 下跌概率，效率 A
- ThreeBlackCrows (三只乌鸦) - 78% 下跌概率，效率 A+
- ThreeLineStrike (三线打击) - 65% 下跌概率，效率 A+
- AbandonedBaby (弃婴形态) - 70% 上涨概率，效率 A-
- Piercing (刺透形态) - 64% 上涨概率，效率 B+
- BreakAway (脱离形态) - 63% 上涨概率，效率 B+
- MatchingLow (相同低价) - 效率 A-
- StickSandwich (条形三明治) - 效率 B
- **Engulfing (吞没形态)** - 63% 准确率，效率 A
- **MorningStar (晨星)** - 78% 上涨概率，效率 A
- **DarkCloudCover (乌云盖顶)** - 70% 下跌概率，效率 A
- **ShootingStar (流星线)** - 62% 下跌概率，效率 A-
- **Kicking (反冲形态)** - 69% 准确率，效率 A+
- **Hammer (锤子线)** - 60% 上涨概率，效率 B+
- **HangingMan (上吊线)** - 59% 下跌概率，效率 B

**中优先级（Efficiency Rank C/D/E）**:
- ThreeWhiteSoldiers (三白兵) - 82% 上涨概率，效率 D+
- ThreeOutside (三外部) - 69% 下跌概率，效率 D-
- DojiStar (十字星线) - 64% 下跌概率，效率 E-
- ClosingMarubozu (收盘光头光脚) - 效率 E+
- InvertedHammer (倒锤子线) - 效率 C+
- Harami (孕线) - 效率 C
- DragonflyDoji (蜻蜓十字) - 效率 C+
- GravestoneDoji (墓碑十字) - 效率 C+

**低优先级（Efficiency Rank F/G/J）**:
- Doji, ThreeInside, AdvanceBlock, BeltHold, TwoCrows, ThreeStarsInSouth

### PatternStats 形态统计数据

```go
// internal/pattern/stats.go

package pattern

// PatternStats 形态统计数据
type PatternStats struct {
    UpPercent      int    // 上涨概率
    DownPercent    int    // 下跌概率
    EfficiencyRank string // 效率排名 A+ ~ J-
    CommonRank     string // 常见度排名
    Source         string // 检测来源: "talib" 或 "custom"
    StatsSource    string // 统计数据来源
    IsEstimated    bool   // 是否为估算数据
}

// PatternStatsMap 形态统计数据映射
// 数据来源: feedroll.com (talib-cdl-go), fivehundred.co, patternswizard.com
var PatternStatsMap = map[PatternType]PatternStats{
    // talib-cdl-go 库支持的形态 (数据来源: feedroll.com)
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
    
    // 自实现的形态 (数据来源: fivehundred.co, patternswizard.com)
    PatternHammer:            {60, 40, "B+", "C", "custom", "fivehundred.co", false},
    PatternInvertedHammer:    {55, 45, "C+", "D", "custom", "fivehundred.co", false},
    PatternHangingMan:        {41, 59, "B", "C", "custom", "fivehundred.co", false},
    PatternShootingStar:      {38, 62, "A-", "C", "custom", "fivehundred.co", false},
    PatternEngulfing:         {67, 33, "A", "B", "custom", "patternswizard.com", false},
    PatternMorningStar:       {70, 30, "A", "G", "custom", "stockgro.club", false},
    PatternMorningDojiStar:   {68, 32, "A-", "H", "custom", "estimated", true},
    PatternEveningDojiStar:   {32, 68, "A-", "H", "custom", "estimated", true},
    PatternDarkCloudCover:    {30, 70, "A", "E", "custom", "fivehundred.co", false},
    PatternHarami:            {53, 47, "C", "B", "custom", "fivehundred.co", false},
    PatternHaramiCross:       {55, 45, "B-", "D", "custom", "estimated", true},
    PatternKicking:           {69, 31, "A+", "J", "custom", "feedroll.com", false},
    PatternDragonflyDoji:     {57, 43, "C+", "E", "custom", "fivehundred.co", false},
    PatternGravestoneDoji:    {43, 57, "C+", "E", "custom", "fivehundred.co", false},
}

// IsHighEfficiency 判断是否为高效形态（A/B 级）
func IsHighEfficiency(pt PatternType) bool {
    stats, ok := PatternStatsMap[pt]
    if !ok {
        return false
    }
    return stats.EfficiencyRank[0] == 'A' || stats.EfficiencyRank[0] == 'B'
}

// GetHighEfficiencyPatterns 获取所有高效形态列表
func GetHighEfficiencyPatterns() []PatternType {
    var result []PatternType
    for pt := range PatternStatsMap {
        if IsHighEfficiency(pt) {
            result = append(result, pt)
        }
    }
    return result
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: K 线时间边界对齐

*For any* timestamp and interval, the calculated kline open time should always be aligned to the interval boundary. For 5-minute intervals, the minute component should be one of: 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, or 55. The second and nanosecond components should always be zero.

**Validates: Requirements 1.5**

### Property 2: K 线 OHLC 不变量

*For any* kline after receiving a sequence of price updates, the following invariants must hold:
- Open equals the first price received for this kline period
- High >= all prices received during this kline period
- Low <= all prices received during this kline period
- Close equals the most recent price received
- High >= Low (always)
- High >= Open and High >= Close
- Low <= Open and Low <= Close

**Validates: Requirements 1.2, 1.3, 1.4, 1.6**

### Property 3: 滚动窗口大小限制

*For any* symbol's kline history after any number of price updates, the count of stored historical klines should never exceed the configured maximum count (default 12).

**Validates: Requirements 1.1, 1.7**

### Property 4: 形态检测确定性

*For any* set of klines, calling Detect() multiple times with the same input should return identical patterns in the same order. Pattern detection is a pure function with no side effects.

**Validates: Requirements 2.1, 2.2, 2.3, 2.4**

### Property 5: 形态信号完整性

*For any* detected pattern, the emitted Pattern_Signal must contain all required fields: ID (non-empty), Symbol (non-empty), Pattern (valid PatternType), Direction (bullish/bearish/neutral), Confidence (0-100), KlineTime (valid timestamp), and DetectedAt (valid timestamp).

**Validates: Requirements 3.1, 3.2**

### Property 6: 信号相关性时间窗口

*For any* pivot signal and pattern signal on the same symbol, they should be correlated if and only if the time difference between them is within the configured window (default 15 minutes). Signals outside the window should not be correlated.

**Validates: Requirements 4.1, 4.2**

### Property 7: 组合信号完整性

*For any* correlated pivot and pattern signal pair, the created CombinedSignal must contain both the original pivot signal, the original pattern signal, and a valid correlation strength.

**Validates: Requirements 4.3, 4.4**

### Property 8: 方向匹配相关性强度

*For any* pivot signal with direction "up" paired with a pattern signal with direction "bullish", the correlation strength should be "strong". Similarly, "down" paired with "bearish" should yield "strong" correlation.

**Validates: Requirements 4.5**

### Property 9: 方向冲突相关性强度

*For any* pivot signal with direction "up" paired with a pattern signal with direction "bearish", the correlation strength should be "weak". Similarly, "down" paired with "bullish" should yield "weak" correlation.

**Validates: Requirements 4.6**

### Property 10: 历史记录持久化往返

*For any* pattern signal, serializing to JSON, writing to file, reading back, and deserializing should produce a signal equivalent to the original (all fields match).

**Validates: Requirements 6.1, 6.3, 6.5**

### Property 11: 形态检测范围限制

*For any* symbol without loaded pivot data, the Pattern_Detector should not perform pattern detection for that symbol, even if kline data exists.

**Validates: Requirements 7.4**

## Error Handling

### 1. 价格更新错误

- **无效价格**: 忽略 price <= 0 的更新
- **时间戳异常**: 如果时间戳早于当前 K 线开盘时间，记录警告但仍处理

### 2. 形态检测错误

- **数据不足**: 如果 K 线数量不足以检测某形态，跳过该形态
- **检测超时**: 如果单个 symbol 检测超过 100ms，记录警告

### 3. 存储错误

- **文件写入失败**: 记录错误，信号仍通过 SSE 推送
- **文件读取失败**: 启动时记录警告，使用空历史

### 4. 并发安全

- 所有共享数据结构使用 `sync.RWMutex` 保护
- K 线更新和形态检测在不同 goroutine 中执行

## Testing Strategy

### 单元测试

1. **K 线计算测试**
   - 测试时间边界对齐
   - 测试 OHLC 更新逻辑
   - 测试滚动窗口维护

2. **形态检测测试**
   - 每种形态的正例和反例
   - 边界条件（数据不足、极端价格）

3. **信号组合测试**
   - 相关性计算
   - 时间窗口过滤

### 属性测试

使用 `github.com/leanovate/gopter` 进行属性测试：

1. **Property 1**: 生成随机时间戳，验证对齐结果（分钟为 0/5/10/.../55，秒和纳秒为 0）
2. **Property 2**: 生成随机价格序列，验证 OHLC 不变量
3. **Property 3**: 生成大量更新，验证窗口大小不超过配置值
4. **Property 4**: 生成随机 K 线数据，多次调用 Detect() 验证结果一致
5. **Property 5**: 生成随机形态信号，验证所有必需字段存在且有效
6. **Property 6**: 生成不同时间差的信号对，验证时间窗口过滤
7. **Property 7**: 生成相关信号对，验证组合信号包含所有必需信息
8. **Property 8-9**: 生成不同方向组合的信号对，验证相关性强度
9. **Property 10**: 生成随机信号，序列化后反序列化验证等价性
10. **Property 11**: 生成有/无 pivot 数据的 symbol，验证检测范围

### 集成测试

1. 模拟 WebSocket 价格流
2. 验证 K 线收盘触发形态检测
3. 验证 SSE 推送

## API Endpoints

### 新增 API

```go
// GET /api/patterns?limit=100&symbol=BTCUSDT&pattern=hammer
// 获取形态信号历史

// GET /api/klines?symbol=BTCUSDT
// 获取指定交易对的当前 K 线数据（调试用）

// SSE event: "pattern"
// 实时推送形态信号
```

### 修改现有 API

```go
// GET /api/history 响应增加 related_pattern 字段
// 如果枢轴点信号有关联的形态信号，包含形态信息
```

## Frontend Changes

### 新增 Patterns 标签页

由于现有 UI 比较紧凑，我们采用以下策略展示形态信号和统计数据：

#### 1. 信号列表项设计（紧凑模式）

```javascript
// 形态信号列表项 - 紧凑设计
function renderPatternItem(signal) {
    // 效率等级颜色
    const efficiencyColor = {
        'A': '#22c55e',  // 绿色 - 高效
        'B': '#84cc16',  // 黄绿 - 较高效
        'C': '#eab308',  // 黄色 - 中等
        'D': '#f97316',  // 橙色 - 较低
        'E': '#ef4444',  // 红色 - 低效
    }[signal.efficiency_rank[0]] || '#6b7280';
    
    // 方向颜色
    const dirColor = signal.direction === 'bullish' ? '#22c55e' : 
                     signal.direction === 'bearish' ? '#ef4444' : '#6b7280';
    
    // 成功率（取方向对应的概率）
    const successRate = signal.direction === 'bullish' ? signal.up_percent : 
                        signal.direction === 'bearish' ? signal.down_percent : 
                        Math.max(signal.up_percent, signal.down_percent);
    
    return `
        <div class="item pattern-item" data-symbol="${signal.symbol}">
            <div class="top">
                <div class="sym">${signal.symbol}</div>
                <div class="tags">
                    <span class="tag pattern">${signal.pattern_cn}</span>
                    <span class="tag" style="background:${dirColor}">${signal.direction === 'bullish' ? '↑' : '↓'}</span>
                    <span class="tag rate" style="background:${efficiencyColor}">${successRate}%</span>
                </div>
            </div>
            <div class="sub">
                <span class="efficiency" title="效率等级">${signal.efficiency_rank}</span>
                <span class="muted time-rel">${fmtRelTime(signal.detected_at)}</span>
            </div>
        </div>
    `;
}
```

#### 2. 详情弹窗（点击查看完整统计）

```javascript
// 点击信号项显示详情弹窗
function showPatternDetail(signal) {
    const modal = document.createElement('div');
    modal.className = 'pattern-modal';
    modal.innerHTML = `
        <div class="pattern-modal-content">
            <div class="modal-header">
                <span class="symbol">${signal.symbol}</span>
                <span class="pattern-name">${signal.pattern_cn}</span>
                <button class="close-btn">&times;</button>
            </div>
            <div class="modal-body">
                <div class="stat-row">
                    <span class="label">形态方向</span>
                    <span class="value ${signal.direction}">${signal.direction === 'bullish' ? '看涨 ↑' : '看跌 ↓'}</span>
                </div>
                <div class="stat-row">
                    <span class="label">历史上涨概率</span>
                    <span class="value">${signal.up_percent}%</span>
                    <div class="bar"><div class="fill up" style="width:${signal.up_percent}%"></div></div>
                </div>
                <div class="stat-row">
                    <span class="label">历史下跌概率</span>
                    <span class="value">${signal.down_percent}%</span>
                    <div class="bar"><div class="fill down" style="width:${signal.down_percent}%"></div></div>
                </div>
                <div class="stat-row">
                    <span class="label">预测效率</span>
                    <span class="value efficiency-${signal.efficiency_rank[0]}">${signal.efficiency_rank}</span>
                </div>
                <div class="stat-row">
                    <span class="label">检测时间</span>
                    <span class="value">${formatTime(signal.detected_at)}</span>
                </div>
                <div class="stat-row">
                    <span class="label">K线收盘时间</span>
                    <span class="value">${formatTime(signal.kline_time)}</span>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-trade" data-symbol="${signal.symbol}">去交易</button>
                <button class="btn-filter" data-symbol="${signal.symbol}">筛选此币</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
}
```

#### 3. 枢轴点信号上的形态徽章

```javascript
// 在枢轴点信号项上显示关联形态徽章
function renderPatternBadge(relatedPattern) {
    if (!relatedPattern) return '';
    
    // 相关性颜色
    const corrColor = {
        'strong': '#22c55e',   // 绿色 - 方向一致
        'moderate': '#eab308', // 黄色 - 中等
        'weak': '#ef4444'      // 红色 - 方向冲突
    }[relatedPattern.correlation];
    
    // 简短显示：形态名 + 成功率
    const rate = relatedPattern.direction === 'bullish' ? 
                 relatedPattern.up_percent : relatedPattern.down_percent;
    
    return `
        <span class="pattern-badge" 
              style="border-color:${corrColor}"
              title="${relatedPattern.pattern_cn} (${relatedPattern.direction}) - ${rate}% 成功率"
              data-pattern-id="${relatedPattern.id}">
            ${relatedPattern.pattern_cn.slice(0,2)} ${rate}%
        </span>
    `;
}
```

#### 4. 筛选器增强

```javascript
// 新增形态筛选下拉框
<select id="patternFilter">
    <option value="">所有形态</option>
    <optgroup label="高效形态 (A/B级)">
        <option value="engulfing">吞没形态 (67%)</option>
        <option value="morning_star">晨星 (70%)</option>
        <option value="evening_star">暮星 (72%)</option>
        <option value="three_black">三只乌鸦 (78%)</option>
        <option value="dark_cloud_cover">乌云盖顶 (70%)</option>
        <option value="hammer">锤子线 (60%)</option>
    </optgroup>
    <optgroup label="中效形态 (C/D级)">
        <option value="three_white">三白兵 (82%)</option>
        <option value="harami">孕线 (53%)</option>
        <option value="doji">十字星 (57%)</option>
    </optgroup>
</select>

// 效率等级筛选
<select id="efficiencyFilter">
    <option value="">所有效率</option>
    <option value="A">A级 (最高效)</option>
    <option value="B">B级 (高效)</option>
    <option value="C">C级 (中等)</option>
</select>
```

#### 5. CSS 样式

```css
/* 形态信号项 */
.pattern-item .tags .rate {
    font-weight: bold;
    min-width: 40px;
    text-align: center;
}

.pattern-item .efficiency {
    font-size: 11px;
    color: #888;
    margin-right: 8px;
}

/* 形态徽章 */
.pattern-badge {
    display: inline-block;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid;
    background: rgba(255,255,255,0.1);
    cursor: pointer;
    margin-left: 4px;
}

.pattern-badge:hover {
    background: rgba(255,255,255,0.2);
}

/* 详情弹窗 */
.pattern-modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.pattern-modal-content {
    background: #1a1a2e;
    border-radius: 12px;
    width: 320px;
    max-width: 90vw;
}

.stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid #333;
}

.stat-row .bar {
    width: 60px;
    height: 6px;
    background: #333;
    border-radius: 3px;
    margin-left: 8px;
}

.stat-row .bar .fill {
    height: 100%;
    border-radius: 3px;
}

.stat-row .bar .fill.up { background: #22c55e; }
.stat-row .bar .fill.down { background: #ef4444; }

/* 效率等级颜色 */
.efficiency-A { color: #22c55e; }
.efficiency-B { color: #84cc16; }
.efficiency-C { color: #eab308; }
.efficiency-D { color: #f97316; }
.efficiency-E { color: #ef4444; }
```

### UI 交互流程

1. **列表视图（紧凑）**: 显示币种、形态名、方向箭头、成功率百分比
2. **悬停提示**: 显示效率等级和简要说明
3. **点击详情**: 弹窗显示完整统计数据（上涨/下跌概率、效率等级、时间等）
4. **快捷操作**: 弹窗底部提供"去交易"和"筛选此币"按钮

### 响应式设计

```css
/* 移动端适配 */
@media (max-width: 480px) {
    .pattern-item .tags .rate {
        font-size: 11px;
        min-width: 36px;
    }
    
    .pattern-badge {
        font-size: 9px;
        padding: 1px 4px;
    }
    
    .pattern-modal-content {
        width: 95vw;
        margin: 10px;
    }
}
```
