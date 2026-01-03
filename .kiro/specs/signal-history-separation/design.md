# Design Document: Signal History Separation

## Overview

本设计通过在现有 `History` 结构内部引入按周期(Period)分离的存储桶，解决日级信号挤占周级信号的问题。设计的核心原则是：

1. **内部重构，外部不变**：所有公开方法签名保持不变
2. **最小化改动**：仅修改 `internal/signal/history.go`，不影响调用方
3. **平滑迁移**：自动检测并迁移现有数据

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      History (公开接口不变)                   │
│  NewHistory(max int) *History                               │
│  EnablePersistence(filePath string) error                   │
│  Add(s Signal)                                              │
│  Query(...) []Signal                                        │
│  Count() int                                                │
│  SymbolCount() int                                          │
├─────────────────────────────────────────────────────────────┤
│                      内部实现 (重构)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ bucket["1d"]│  │ bucket["1w"]│  │ bucket[""]  │         │
│  │ max: 8000   │  │ max: 2000   │  │ max: 1000   │         │
│  │ file: _1d   │  │ file: _1w   │  │ file: _other│         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### periodBucket (新增内部结构)

每个周期的独立存储桶：

```go
type periodBucket struct {
    mu           sync.RWMutex
    max          int
    signals      []Signal
    symbolsUpper []string
    
    fileMu    sync.Mutex
    filePath  string
    fileLines int
}
```

### History (重构)

```go
type History struct {
    // 容量配置
    totalMax    int                    // 总容量（向后兼容）
    periodMax   map[string]int         // 每个周期的容量
    defaultMax  int                    // 未配置周期的默认容量
    
    // 分离存储
    buckets     map[string]*periodBucket  // period -> bucket
    bucketsMu   sync.RWMutex
    
    // 持久化
    baseDir     string                 // 基础目录
    baseName    string                 // 基础文件名
}
```

### 容量分配策略

默认容量分配（基于 totalMax=20000）：
- `1d` (日级): 80% = 16000
- `1w` (周级): 15% = 3000  
- 其他: 5% = 1000

这个比例确保周级信号有足够的独立空间。

### 文件命名约定

基于原始文件路径 `signals/history.jsonl`：
- 日级: `signals/history_1d.jsonl`
- 周级: `signals/history_1w.jsonl`
- 其他: `signals/history_other.jsonl`

## Data Models

### Signal (不变)

```go
type Signal struct {
    ID          string    `json:"id"`
    Symbol      string    `json:"symbol"`
    Period      string    `json:"period"`      // "1d", "1w", etc.
    Level       string    `json:"level"`
    Price       float64   `json:"price"`
    Direction   string    `json:"direction"`
    TriggeredAt time.Time `json:"triggered_at"`
    Source      string    `json:"source"`
}
```

### 周期分类

```go
func normalizePeriod(period string) string {
    switch strings.ToLower(period) {
    case "1d", "d", "daily":
        return "1d"
    case "1w", "w", "weekly":
        return "1w"
    default:
        return "other"
    }
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Period-specific storage

*For any* signal with a given period, when added to history, it SHALL be stored in the bucket corresponding to that period and not in any other bucket.

**Validates: Requirements 1.1**

### Property 2: Cross-period isolation on eviction

*For any* history state where one period bucket is at capacity, adding a new signal to that bucket SHALL only evict signals from that same bucket, leaving signals in other period buckets unchanged.

**Validates: Requirements 1.2**

### Property 3: Merge and chronological sort

*For any* set of signals across multiple periods, querying without a period filter SHALL return all signals merged and sorted by triggered_at in descending order (newest first).

**Validates: Requirements 1.4, 4.5, 4.6, 5.2**

### Property 4: Period filter queries correct bucket

*For any* query with a period filter, the results SHALL only contain signals from that specific period bucket.

**Validates: Requirements 4.4**

### Property 5: Persistence round-trip

*For any* set of signals added to history with persistence enabled, reloading the history SHALL restore all signals to their correct period buckets with all fields preserved.

**Validates: Requirements 3.2, 5.1, 5.3**

## Error Handling

### Migration Errors

- If the unified history file cannot be read, log error and start with empty history
- If individual period files are corrupted, log warning and skip that file
- If migration fails mid-way, keep the original file intact

### File I/O Errors

- Append failures: log error, signal still stored in memory
- Compact failures: log error, keep original file

### Invalid Period Values

- Empty or unknown periods are normalized to "other" bucket
- No panic on unexpected period values

## Testing Strategy

### Unit Tests

- Test bucket creation and capacity limits
- Test period normalization logic
- Test file naming convention
- Test migration detection logic

### Property-Based Tests

使用 Go 的 `gopter` 库进行属性测试，每个属性至少运行 100 次迭代。

**Property Test 1: Period-specific storage**
- Generate random signals with random periods
- Add to history
- Verify each signal is in the correct bucket

**Property Test 2: Cross-period isolation**
- Fill one bucket to capacity
- Add more signals to that bucket
- Verify other buckets are unchanged

**Property Test 3: Merge and sort**
- Generate signals with random timestamps across periods
- Query without filter
- Verify results are sorted by triggered_at descending

**Property Test 4: Period filter**
- Add signals to multiple buckets
- Query with period filter
- Verify only matching signals returned

**Property Test 5: Persistence round-trip**
- Generate random signals
- Add to history with persistence
- Reload history
- Verify all signals preserved with correct fields

### Integration Tests

- Test migration from existing unified history file
- Test concurrent Add and Query operations
- Test compaction triggers correctly
