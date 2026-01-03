# Implementation Plan: Ranking Monitor

## Overview

实现排名监控系统，包括后端采样器、存储、API 以及前端排名监控视图。使用 Go 语言实现后端，JavaScript 实现前端。

## Tasks

- [x] 1. 创建 ranking 包基础结构
  - 创建 `internal/ranking/` 目录
  - 定义数据类型：Snapshot, SnapshotItem, RankingItem, SymbolSnapshot
  - 定义响应类型：CurrentResponse, MoversResponse, HistoryResponse
    - HistoryResponse: {symbol: string, snapshots: []SymbolSnapshot}（无数据时 snapshots 为空数组）
  - 定义查询选项：CurrentOptions, MoversOptions
  - _Requirements: 1.3, 2.1-2.5_

- [x] 2. 实现 USDT 交易对过滤
  - [x] 2.1 实现 isUSDTPair 函数
    - 检查 symbol 是否以 "USDT" 结尾
    - _Requirements: 1.1, 1.2, 2.5_
  - [x] 2.2 编写 USDT 过滤属性测试
    - 测试 isUSDTPair 函数的正确性
    - 测试 Sampler.Sample() 返回的快照只包含 USDT 交易对
    - **Property 1: USDT Pair Filtering**
    - **Validates: Requirements 1.1, 1.2, 2.5**

- [x] 3. 实现排名计算逻辑
  - [x] 3.1 实现 calculateRanks 函数
    - 过滤 USDT 交易对
    - 按 volume 降序排序计算 volumeRanks
    - 按 trade_count 降序排序计算 tradesRanks
    - 使用 dense ranking 处理并列
    - _Requirements: 2.1, 2.2, 2.3, 2.4_
  - [x] 3.2 编写排名计算属性测试
    - **Property 2: Ranking Order Correctness**
    - **Property 3: Equal Value Rank Assignment (Dense Ranking)**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

- [x] 4. 实现 Ranking Store
  - [x] 4.1 实现 Store 结构和基础方法
    - NewStore 构造函数，接收 dataDir 参数（使用配置的 data 目录）
    - Add 添加快照方法
    - cleanup 清理过期快照方法
    - findSnapshotByTime 查找指定时间快照方法
      - 查找 timestamp <= targetTime 的最近快照
      - 如果没有符合条件的快照，返回最老的快照
    - _Requirements: 1.4, 1.5, 3.2, 3.3_
  - [x] 4.2 编写 24 小时保留窗口属性测试
    - **Property 4: 24-Hour Retention Window**
    - **Validates: Requirements 1.4, 1.5**

- [x] 5. 实现排名变化计算
  - [x] 5.1 实现 GetCurrent 方法
    - 获取最新快照
    - 根据 compare 参数查找比较快照
    - 计算 rank_change 和 price_change
    - 处理新交易对标记
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.1, 4.2, 4.3, 4.4_
  - [x] 5.2 编写排名变化计算属性测试
    - **Property 5: Rank Change Calculation**
    - **Property 6: Price Change Calculation**
    - **Validates: Requirements 3.1-3.6, 4.1-4.4**

- [x] 6. 实现历史查询和异动查询
  - [x] 6.1 实现 GetHistory 方法
    - 遍历所有快照提取指定 symbol 的数据
    - 按时间升序返回
    - 返回 HistoryResponse 结构（无数据时 snapshots 为空数组）
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  - [x] 6.2 实现 GetMovers 方法
    - 根据 compare 参数使用 findSnapshotByTime 查找比较快照
    - 计算所有交易对的排名变化
    - 按 direction 过滤（up: rank_change > 0, down: rank_change < 0）
    - 按绝对变化值降序排序
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_
  - [x] 6.3 编写历史和异动查询属性测试
    - **Property 7: History Chronological Order**
    - **Property 8: Movers Sorting**
    - **Validates: Requirements 6.4, 7.5**

- [x] 7. 实现 Ranking Sampler
  - [x] 7.1 实现 Sampler 结构
    - NewSampler 构造函数
    - Sample 执行单次采样方法
    - Run 启动采样循环方法（5分钟间隔）
    - _Requirements: 1.1, 1.3, 1.6_
  - [x] 7.2 集成 Sampler 到 main.go
    - 创建 Sampler 实例
    - 启动采样协程
    - _Requirements: 1.6_

- [x] 8. Checkpoint - 确保后端核心逻辑测试通过
  - 运行 `go test ./internal/ranking/...`
  - 确保所有属性测试通过
  - 如有问题请询问用户

- [x] 9. 实现数据持久化
  - [x] 9.1 实现 persist 和 load 方法
    - 使用 JSON 格式存储到配置的 dataDir/ranking/ 子目录
    - dataDir 通过 NewStore 参数传入，与服务器 -data-dir 标志一致
    - 定期持久化（每次添加快照后）
    - 启动时加载历史数据
    - _Requirements: 10.1, 10.2, 10.3, 10.4_
  - [x] 9.2 编写持久化往返属性测试
    - **Property 9: Persistence Round Trip**
    - **Validates: Requirements 10.1, 10.2**

- [x] 10. 实现 HTTP API
  - [x] 10.1 实现 handleRankingCurrent
    - GET /api/ranking/current
    - 支持 type, compare, limit 参数
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_
  - [x] 10.2 实现 handleRankingHistory
    - GET /api/ranking/history/{symbol}
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
  - [x] 10.3 实现 handleRankingMovers
    - GET /api/ranking/movers
    - 支持 type, direction, compare, limit 参数
    - direction 为必填参数
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_
  - [x] 10.4 注册路由到 Server
    - 在 httpapi/server.go 中添加路由
    - 添加 RankingStore 到 Server 结构

- [x] 11. Checkpoint - 确保 API 测试通过
  - 运行 `go test ./internal/httpapi/...`
  - 手动测试 API 端点
  - 如有问题请询问用户

- [x] 12. 实现前端排名监控视图
  - [x] 12.1 添加导航和视图切换
    - 在 index.html 添加 "Ranking" 导航按钮
    - 添加 rankingMonitorScroll 容器
    - _Requirements: 8.1_
  - [x] 12.2 实现排名列表渲染
    - 创建 renderRankingMonitorItem 函数
    - 显示排名、变化箭头、价格变化
    - 绿色上升箭头、红色下降箭头
    - _Requirements: 8.2, 8.3, 8.4_
  - [x] 12.3 实现筛选控件
    - 添加 type 切换（Volume/Trades）
    - 添加 compare 时间窗口选择
    - _Requirements: 8.5, 8.6_
  - [x] 12.4 实现数据加载和刷新
    - 调用 /api/ranking/current API
    - 定时刷新（可选）

- [x] 13. 实现交易对详情弹窗
  - [x] 13.1 创建详情弹窗 HTML 结构
    - 添加 rankingDetailModal 容器
    - 添加图表容器
    - _Requirements: 9.1, 9.2_
  - [x] 13.2 实现历史数据加载
    - 调用 /api/ranking/history/{symbol} API
    - 处理数据格式
  - [x] 13.3 实现图表渲染
    - 使用简单的 SVG 或 Canvas 绘制趋势图
    - 显示 volume rank、trades rank、price 三条线
    - 同一时间轴展示
    - _Requirements: 9.3, 9.4, 9.5, 9.6_

- [x] 14. Final Checkpoint - 完整功能测试
  - 运行 `go test ./...`
  - 启动服务器测试前端功能
  - 验证采样、API、前端展示完整流程
  - 如有问题请询问用户

## Notes

- 所有任务均为必需，包括属性测试
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- 后端使用 Go 语言，前端使用原生 JavaScript
- 采样间隔 5 分钟，保留 24 小时数据
