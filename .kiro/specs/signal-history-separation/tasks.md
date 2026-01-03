# Implementation Plan: Signal History Separation

## Overview

重构 `internal/signal/history.go`，在内部实现按周期分离存储，同时保持所有公开接口不变。

## Tasks

- [x] 1. 添加 periodBucket 内部结构
  - 在 history.go 中添加 periodBucket 结构体
  - 包含 signals、symbolsUpper、max、filePath、fileLines 等字段
  - 添加 normalizePeriod 函数处理周期标准化
  - _Requirements: 1.1, 1.3_

- [x] 2. 重构 History 结构
  - [x] 2.1 修改 History 结构体字段
    - 添加 buckets map[string]*periodBucket
    - 添加 periodMax map[string]int 配置
    - 添加 defaultMax、baseDir、baseName 字段
    - 保留 totalMax 用于向后兼容
    - _Requirements: 1.3, 2.1, 2.2_

  - [x] 2.2 修改 NewHistory 函数
    - 保持签名 NewHistory(max int) *History 不变
    - 内部计算各周期的默认容量分配
    - 初始化 buckets map
    - _Requirements: 2.1, 2.2, 4.1_

- [x] 3. 实现 Add 方法的分桶逻辑
  - [x] 3.1 重构 Add 方法
    - 根据 signal.Period 确定目标桶
    - 调用 normalizePeriod 标准化周期
    - 添加到对应桶，仅驱逐同桶信号
    - _Requirements: 1.1, 1.2, 4.2_

  - [x] 3.2 编写属性测试：Period-specific storage
    - **Property 1: Period-specific storage**
    - **Validates: Requirements 1.1**

  - [x] 3.3 编写属性测试：Cross-period isolation
    - **Property 2: Cross-period isolation on eviction**
    - **Validates: Requirements 1.2**

- [x] 4. 实现 Query 方法的合并逻辑
  - [x] 4.1 重构 Query 方法
    - 有 period 过滤时只查对应桶
    - 无 period 过滤时合并所有桶
    - 合并后按 triggered_at 降序排序
    - _Requirements: 1.4, 4.4, 4.5, 4.6_

  - [x] 4.2 编写属性测试：Merge and sort
    - **Property 3: Merge and chronological sort**
    - **Validates: Requirements 1.4, 4.5, 4.6**

  - [x] 4.3 编写属性测试：Period filter
    - **Property 4: Period filter queries correct bucket**
    - **Validates: Requirements 4.4**

- [x] 5. 实现 Count 和 SymbolCount 方法
  - 遍历所有桶累加计数
  - 保持方法签名不变
  - _Requirements: 4.3_

- [x] 6. Checkpoint - 确保内存模式测试通过
  - 运行 go test ./internal/signal/... 确保现有测试通过
  - 如有问题请询问用户

- [x] 7. 实现分离的持久化
  - [x] 7.1 重构 EnablePersistence 方法
    - 解析 baseName 和 baseDir
    - 检测是否存在旧的统一文件需要迁移
    - 为每个周期创建独立文件路径
    - _Requirements: 3.1, 5.4, 5.8_

  - [x] 7.2 实现迁移逻辑
    - 读取旧的统一文件
    - 按周期分类信号
    - 写入各周期文件
    - 重命名旧文件为 .migrated
    - _Requirements: 5.4, 5.5, 5.6, 5.7_

  - [x] 7.3 实现各桶独立的 append 和 compact
    - 每个桶独立追加到自己的文件
    - 每个桶独立触发 compact
    - _Requirements: 3.2, 3.3_

  - [x] 7.4 编写属性测试：Persistence round-trip
    - **Property 5: Persistence round-trip**
    - **Validates: Requirements 3.2, 5.1, 5.3**

- [x] 8. 错误处理
  - 单个周期文件损坏时记录警告并继续
  - 迁移失败时回退到统一存储模式
  - _Requirements: 3.4, 5.6_

- [x] 9. Final Checkpoint - 确保所有测试通过
  - 运行 go test ./internal/signal/... -v
  - 确保现有测试和新属性测试都通过
  - 如有问题请询问用户

## Notes

- 所有任务均为必需，包括属性测试
- 所有公开方法签名保持不变，调用方无需修改
- 使用 gopter 库进行属性测试
