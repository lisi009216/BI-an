# Design Document: Bug Fixes & Stability

## Overview

本设计文档描述了系统稳定性修复的技术方案，涵盖四个主要问题：
1. 周枢轴点刷新逻辑在周日计算错误
2. 配置参数缺乏校验导致 panic
3. 形态历史文件无限增长
4. TickerEvent 解析不兼容字符串格式

## Architecture

修复涉及以下模块，均为现有代码的局部修改：

```
internal/
├── pivot/
│   └── refresher.go      # 修复 needsRefresh 周一计算逻辑
├── pattern/
│   └── history.go        # 添加文件截断功能
├── kline/
│   └── store.go          # 添加参数校验
└── binance/
    └── ws_ticker.go      # 修复 parseInt 字符串处理
```

## Components and Interfaces

### 1. Pivot Refresher 修复

**问题分析**：
当前 `needsRefresh` 函数中计算"本周一"的逻辑有误：

```go
// 当前错误代码 (refresher.go:236)
delta := (int(time.Monday) - int(now.Weekday()) + 7) % 7
thisMonday8am02 := today.AddDate(0, 0, -((7 - delta) % 7))
```

当 `now.Weekday()` 是 Sunday (0) 时：
- `delta = (1 - 0 + 7) % 7 = 1`
- 这会指向下周一，而不是本周一

**修复方案**：

```go
// 修复后的逻辑
func getThisWeekMonday(now time.Time, loc *time.Location) time.Time {
    today := time.Date(now.Year(), now.Month(), now.Day(), 8, 2, 0, 0, loc)
    
    // 计算距离本周一的天数差
    // Sunday = 0, Monday = 1, ..., Saturday = 6
    weekday := int(now.Weekday())
    if weekday == 0 {
        weekday = 7 // 将周日视为 7，这样计算更直观
    }
    daysFromMonday := weekday - 1 // Monday = 0, Tuesday = 1, ..., Sunday = 6
    
    return today.AddDate(0, 0, -daysFromMonday)
}
```

### 2. 参数校验

**修复位置**：
- `internal/kline/store.go`: `NewStore` 函数
- `internal/pattern/history.go`: `NewHistory` 函数

**校验逻辑**：

```go
// kline/store.go
func NewStore(interval time.Duration, maxCount int) *Store {
    if maxCount <= 0 {
        log.Printf("WARN: invalid KLINE_COUNT=%d, using default 12", maxCount)
        maxCount = 12
    }
    // ...
}

// pattern/history.go
func NewHistory(filePath string, maxSize int) (*History, error) {
    if maxSize <= 0 {
        log.Printf("WARN: invalid PATTERN_HISTORY_MAX=%d, using default 1000", maxSize)
        maxSize = 1000
    }
    // ...
}
```

### 3. 形态历史文件截断

**参考实现**：`internal/signal/history.go` 的 `compactLocked` 方法

**新增字段**：

```go
type History struct {
    // ... 现有字段
    fileLines int  // 跟踪文件行数
}
```

**截断逻辑**：

```go
func (h *History) Add(sig Signal) error {
    // ... 现有逻辑
    
    // 持久化后检查是否需要截断
    if h.persistMode && h.file != nil {
        h.fileLines++
        
        // 每 100 条检查一次，文件行数超过 maxSize*2 时截断
        if h.fileLines%100 == 0 && h.fileLines > h.maxSize*2 {
            if err := h.compact(); err != nil {
                log.Printf("WARN: pattern history compact failed: %v", err)
                // 继续运行，不中断
            }
        }
    }
    
    return nil
}

func (h *History) compact() error {
    // 1. 创建临时文件
    // 2. 写入最新的 maxSize 条记录
    // 3. 原子替换原文件
    // 4. 重新打开文件句柄
}
```

### 4. TickerEvent 解析修复

**问题分析**：
当前 `parseInt` 函数不处理字符串类型：

```go
// 当前代码
parseInt := func(key string) int64 {
    switch val := v.(type) {
    case json.Number:
        i, _ := val.Int64()
        return i
    case float64:
        return int64(val)
    }
    return 0  // 字符串会走到这里，返回 0
}
```

**修复方案**：

```go
parseInt := func(key string) int64 {
    v, ok := raw[key]
    if !ok {
        return 0
    }
    switch val := v.(type) {
    case json.Number:
        i, _ := val.Int64()
        return i
    case string:
        // 新增：处理字符串格式
        i, _ := json.Number(val).Int64()
        return i
    case float64:
        return int64(val)
    }
    return 0
}
```

## Data Models

无新增数据模型，仅修改现有实现。

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: 周一计算一致性

*For any* date in a week (Monday through Sunday), the calculated "this week's Monday" should always be the same date and should be on or before the current date.

**Validates: Requirements 1.2, 1.5**

### Property 2: 过期检测持续性

*For any* weekly pivot data that was last updated before this week's Monday, the `needsRefresh` function should return `true` on all days of the current week (Monday through Sunday).

**Validates: Requirements 1.1, 1.3**

### Property 3: 参数校验防护

*For any* negative or zero value passed to `NewStore` or `NewHistory`, the system should not panic and should use a positive default value.

**Validates: Requirements 2.1, 2.2**

### Property 4: 文件截断保留最新记录

*For any* sequence of pattern signals written to history, after truncation, the file should contain exactly the most recent `maxSize` entries in chronological order.

**Validates: Requirements 3.2, 3.3**

### Property 5: JSON 数值解析等价性

*For any* numeric value, parsing it from JSON number format and JSON string format should produce the same result.

**Validates: Requirements 4.1, 4.2, 4.4**

## Error Handling

| 场景 | 处理方式 |
|------|----------|
| 参数校验失败 | 记录警告日志，使用默认值继续 |
| 文件截断 I/O 错误 | 记录错误日志，继续运行 |
| JSON 解析失败 | 返回默认值 0，继续处理 |

## Testing Strategy

### 单元测试

1. **周一计算测试** (`internal/pivot/refresher_test.go`)
   - 测试一周七天的周一计算
   - 重点测试周日边界情况

2. **参数校验测试** (`internal/kline/store_test.go`, `internal/pattern/history_test.go`)
   - 测试负数、零、正常值

3. **文件截断测试** (`internal/pattern/history_test.go`)
   - 测试截断触发条件
   - 测试截断后数据完整性

4. **JSON 解析测试** (`internal/binance/ws_ticker_test.go`)
   - 测试数字格式和字符串格式

### 属性测试

使用 `testing/quick` 包进行属性测试：

1. **Property 1**: 生成随机日期，验证周一计算一致性
2. **Property 2**: 生成随机 pivot 数据和日期，验证过期检测
3. **Property 5**: 生成随机数值，验证两种 JSON 格式解析等价

**测试配置**：
- 每个属性测试至少运行 100 次迭代
- 使用 `testing/quick.Config{MaxCount: 100}`

