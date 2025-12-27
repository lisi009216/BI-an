# Implementation Plan: Bug Fixes & Stability

## Overview

本计划修复系统中的四个稳定性问题，按优先级排序实现。

## Tasks

- [x] 1. 修复周枢轴点刷新逻辑
  - [x] 1.1 修复 `needsRefresh` 函数中的周一计算逻辑
    - 修改 `internal/pivot/refresher.go`
    - 将周日 (Weekday=0) 视为 7，确保计算出的周一在当前日期之前
    - _Requirements: 1.1, 1.2_
  - [x] 1.2 添加周一计算的单元测试
    - 在 `internal/pivot/refresher_test.go` 中添加测试
    - 覆盖周一到周日所有情况，重点测试周日边界
    - _Requirements: 1.4_
  - [x] 1.3 添加属性测试验证周一计算一致性
    - **Property 1: 周一计算一致性**
    - **Validates: Requirements 1.2, 1.5**

- [x] 2. 添加参数校验防护
  - [x] 2.1 修复 `Kline Store` 参数校验
    - 修改 `internal/kline/store.go` 的 `NewStore` 函数
    - 当 `maxCount <= 0` 时使用默认值 12 并记录警告
    - _Requirements: 2.2_
  - [x] 2.2 修复 `Pattern History` 参数校验
    - 修改 `internal/pattern/history.go` 的 `NewHistory` 函数
    - 当 `maxSize <= 0` 时使用默认值 1000 并记录警告
    - _Requirements: 2.1_
  - [x] 2.3 添加参数校验的单元测试
    - 测试负数、零、正常值的处理
    - 验证不会 panic 且使用正确的默认值
    - **Property 3: 参数校验防护**
    - **Validates: Requirements 2.1, 2.2**

- [x] 3. 实现形态历史文件截断
  - [x] 3.1 添加文件行数跟踪
    - 在 `History` 结构体中添加 `fileLines` 字段
    - 在 `load()` 时统计行数
    - _Requirements: 3.2_
  - [x] 3.2 实现 `compact` 方法
    - 参考 `internal/signal/history.go` 的实现
    - 创建临时文件，写入最新记录，原子替换
    - _Requirements: 3.2, 3.3_
  - [x] 3.3 在 `Add` 方法中集成截断检查
    - 每 100 条检查一次
    - 当 `fileLines > maxSize*2` 时触发截断
    - 截断失败时记录日志但继续运行
    - _Requirements: 3.5, 3.6_
  - [x] 3.4 添加文件截断的单元测试
    - 测试截断触发条件
    - 验证截断后保留最新的 maxSize 条记录
    - **Property 4: 文件截断保留最新记录**
    - **Validates: Requirements 3.2, 3.3**

- [x] 4. 修复 TickerEvent 解析兼容性
  - [x] 4.1 修复 `parseInt` 函数
    - 修改 `internal/binance/ws_ticker.go`
    - 添加 `case string:` 分支处理字符串格式
    - _Requirements: 4.1, 4.2_
  - [x] 4.2 添加 JSON 解析的单元测试
    - 测试数字格式和字符串格式的解析
    - 验证两种格式产生相同结果
    - **Property 5: JSON 数值解析等价性**
    - **Validates: Requirements 4.1, 4.2, 4.4**

- [x] 5. Checkpoint - 运行完整测试套件
  - 运行 `go test ./...` 确保所有测试通过
  - 如有问题请告知

## Notes

- 每个任务引用了对应的需求编号
- 属性测试使用 `testing/quick` 包
- 修复按优先级排序：High → Medium → Low

