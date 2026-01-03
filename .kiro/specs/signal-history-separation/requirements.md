# Requirements Document

## Introduction

本功能旨在解决信号历史记录中日级(D)信号和周级(W)信号混合存储导致的周级信号丢失问题。由于日级信号频率远高于周级信号（约为N倍），在统一的内存限制下，周级信号会被高频的日级信号挤出历史记录。

**设计原则：**
- 最小化代码改动：保持现有 History 结构的公开接口不变
- 数据一致性：查询结果的排序和格式与之前完全一致
- 平滑迁移：自动处理现有数据，无需手动干预

## Glossary

- **Signal_History**: 信号历史记录管理器，负责存储和查询交易信号
- **Period**: 信号的时间周期，如 "1d"（日级）或 "1w"（周级）
- **Signal**: 包含交易对、周期、价格、方向等信息的交易信号记录
- **Capacity**: 每个周期独立的最大信号存储数量

## Requirements

### Requirement 1: 按周期分离信号存储

**User Story:** 作为系统运维人员，我希望日级和周级信号分开存储，以便周级信号不会被高频的日级信号挤出历史记录。

#### Acceptance Criteria

1. WHEN a signal is added to history, THE Signal_History SHALL store it in a period-specific storage bucket
2. WHEN the storage bucket for a period reaches its capacity, THE Signal_History SHALL only evict signals from that same period
3. THE Signal_History SHALL maintain separate capacity limits for each period type
4. WHEN querying signals without period filter, THE Signal_History SHALL merge results from all period buckets in chronological order

### Requirement 2: 独立的周期容量配置

**User Story:** 作为系统运维人员，我希望能够为不同周期配置不同的容量限制，以便根据信号频率合理分配存储资源。

#### Acceptance Criteria

1. THE Signal_History SHALL support configuring capacity per period via a map of period to max count
2. WHEN a period is not explicitly configured, THE Signal_History SHALL use a default capacity value
3. THE Signal_History SHALL allow runtime inspection of current capacity settings

### Requirement 3: 分离的持久化文件

**User Story:** 作为系统运维人员，我希望不同周期的信号存储在不同的文件中，以便独立管理和备份。

#### Acceptance Criteria

1. WHEN persistence is enabled, THE Signal_History SHALL create separate files for each period (e.g., history_1d.jsonl, history_1w.jsonl)
2. WHEN loading from persistence, THE Signal_History SHALL load each period file independently
3. WHEN compacting history, THE Signal_History SHALL compact each period file independently
4. IF a period file is corrupted or missing, THEN THE Signal_History SHALL log a warning and continue with other periods

### Requirement 4: 向后兼容的查询接口

**User Story:** 作为前端开发者，我希望查询接口保持向后兼容，以便现有的前端代码无需修改。

#### Acceptance Criteria

1. THE Signal_History SHALL maintain the existing Query method signature unchanged
2. THE Signal_History SHALL maintain the existing Add method signature unchanged
3. THE Signal_History SHALL maintain the existing Count and SymbolCount method signatures unchanged
4. WHEN querying with a period filter, THE Signal_History SHALL only search the corresponding period bucket
5. WHEN querying without a period filter, THE Signal_History SHALL search all period buckets and merge results
6. THE Signal_History SHALL return results sorted by triggered_at in descending order (newest first), consistent with current behavior

### Requirement 5: 数据一致性保证

**User Story:** 作为系统运维人员，我希望新旧系统的查询结果在逻辑上保持一致，以便验证迁移正确性。

#### Acceptance Criteria

1. FOR ALL signals in the old unified history, THE Signal_History SHALL preserve them in the corresponding period bucket after migration
2. WHEN querying all signals without filters, THE Signal_History SHALL return results in the same chronological order as before
3. THE Signal_History SHALL preserve all signal fields (ID, Symbol, Period, Level, Price, Direction, TriggeredAt, Source) without modification

### Requirement 5: 迁移现有数据

**User Story:** 作为系统运维人员，我希望现有的混合历史文件能够自动迁移到新的分离存储结构。

#### Acceptance Criteria

1. WHEN the system starts with an existing unified history file (history.jsonl), THE Signal_History SHALL automatically migrate signals to period-specific files
2. WHEN migration completes successfully, THE Signal_History SHALL rename the old file with a .migrated suffix as backup
3. IF migration fails, THEN THE Signal_History SHALL log an error and fall back to unified storage mode
4. THE Signal_History SHALL only attempt migration once per startup
5. WHEN no unified history file exists, THE Signal_History SHALL directly use period-separated storage
