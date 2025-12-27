# Implementation Plan: K 线形态识别系统

## Overview

本实现计划将 K 线形态识别系统分解为可增量执行的任务，每个任务构建在前一个任务之上。使用 Go 语言实现，采用 `github.com/iwat/talib-cdl-go` 库进行形态检测，并自实现部分高效形态。

## Tasks

- [x] 1. 项目结构和依赖设置
  - 创建 `internal/kline/` 目录结构
  - 创建 `internal/pattern/` 目录结构
  - 添加 `github.com/iwat/talib-cdl-go` 依赖到 go.mod
  - 添加 `github.com/leanovate/gopter` 测试依赖
  - _Requirements: 1.8, 2.5_

- [x] 2. 实现 Kline 数据结构
  - [x] 2.1 创建 `internal/kline/kline.go`
    - 定义 Kline 结构体（Symbol, Open, High, Low, Close, OpenTime, CloseTime, IsClosed）
    - 实现辅助方法：Body(), UpperShadow(), LowerShadow(), IsBullish(), IsBearish(), Range()
    - _Requirements: 1.1_
  - [x] 2.2 编写 Kline 结构体单元测试
    - 测试 Body, UpperShadow, LowerShadow 计算
    - 测试 IsBullish, IsBearish 判断
    - _Requirements: 1.1_

- [x] 3. 实现 KlineStore 存储模块
  - [x] 3.1 创建 `internal/kline/store.go`
    - 实现 Store 结构体和 SymbolKlines
    - 实现 NewStore() 构造函数
    - 实现 SetOnClose() 回调设置
    - _Requirements: 1.1, 1.8_
  - [x] 3.2 实现 K 线时间边界计算
    - 实现 getKlineOpenTime() 和 getKlineCloseTime()
    - 确保 5 分钟边界对齐（0, 5, 10, 15...）
    - _Requirements: 1.5_
  - [x] 3.3 编写属性测试：K 线时间边界对齐
    - **Property 1: K 线时间边界对齐**
    - **Validates: Requirements 1.5**
  - [x] 3.4 实现 Update() 方法
    - 更新当前 K 线的 OHLC
    - 检测收盘并触发回调（传深拷贝快照）
    - 维护滚动窗口大小
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_
  - [x] 3.5 编写属性测试：OHLC 不变量
    - **Property 2: K 线 OHLC 不变量**
    - **Validates: Requirements 1.2, 1.3, 1.4, 1.6**
  - [x] 3.6 编写属性测试：滚动窗口大小限制
    - **Property 3: 滚动窗口大小限制**
    - **Validates: Requirements 1.1, 1.7**
  - [x] 3.7 实现 GetKlines() 和 GetCurrentKline()
    - 返回深拷贝，按时间顺序（最旧在前）
    - _Requirements: 1.8_
  - [x] 3.8 实现 CleanupStale() 方法
    - 清理长期无更新的 symbol
    - _Requirements: 7.4_

- [x] 4. Checkpoint - 确保 KlineStore 测试通过
  - 运行所有 kline 包测试
  - 确保属性测试通过

- [x] 5. 实现 PatternStats 统计数据
  - [x] 5.1 创建 `internal/pattern/stats.go`
    - 定义 PatternStats 结构体
    - 定义 PatternStatsMap 映射（包含所有形态的统计数据）
    - 实现 IsHighEfficiency() 和 GetHighEfficiencyPatterns()
    - _Requirements: 2.2, 2.3_
  - [x] 5.2 创建 `internal/pattern/types.go`
    - 定义 PatternType 常量
    - 定义 Direction 常量
    - 定义 PatternNames 中文名称映射
    - _Requirements: 3.1_

- [x] 6. 实现 PatternSignal 信号结构
  - [x] 6.1 创建 `internal/pattern/signal.go`
    - 定义 Signal 结构体（包含 Source, StatsSource, IsEstimated 字段）
    - 实现 NewSignal() 构造函数
    - 实现 generateID() 函数（使用 symbol+pattern+klineTime）
    - _Requirements: 3.1, 3.2_
  - [x] 6.2 编写属性测试：信号完整性
    - **Property 5: 形态信号完整性**
    - **Validates: Requirements 3.1, 3.2**

- [x] 7. 实现 talib-cdl-go 形态检测
  - [x] 7.1 创建 `internal/pattern/detector.go`
    - 定义 Detector 结构体和 DetectorConfig
    - 实现 NewDetector() 构造函数
    - 实现 toSeries() 转换函数
    - _Requirements: 2.1, 2.5_
  - [x] 7.2 实现 talib-cdl-go 形态检测
    - 调用 Doji, EveningStar, ThreeWhiteSoldiers, ThreeBlackCrows 等
    - 处理返回值（正值看涨，负值看跌）
    - _Requirements: 2.2, 2.3, 2.4_
  - [x] 7.3 编写属性测试：形态检测确定性
    - **Property 4: 形态检测确定性**
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

- [x] 8. 实现自定义形态检测
  - [x] 8.1 创建 `internal/pattern/custom.go`
    - 实现 isDowntrend() 和 isUptrend() 趋势判断
    - _Requirements: 2.2_
  - [x] 8.2 实现 detectHammer() 锤子线检测
    - 使用 3 根 K 线判断趋势
    - _Requirements: 2.2_
  - [x] 8.3 实现 detectShootingStar() 流星线检测
    - _Requirements: 2.2_
  - [x] 8.4 实现 detectEngulfing() 吞没形态检测
    - _Requirements: 2.2_
  - [x] 8.5 实现 detectMorningStar() 晨星检测
    - _Requirements: 2.2_
  - [x] 8.6 实现 detectDarkCloudCover() 乌云盖顶检测
    - 放宽跳空条件适配加密市场
    - _Requirements: 2.2_
  - [x] 8.7 实现其他自定义形态
    - InvertedHammer, HangingMan, Harami, HaramiCross
    - DragonflyDoji, GravestoneDoji
    - _Requirements: 2.2, 2.3_
  - [x] 8.8 编写自定义形态单元测试
    - 测试各形态的正例和反例
    - _Requirements: 2.2, 2.3_

- [x] 9. Checkpoint - 确保形态检测测试通过
  - 运行所有 pattern 包测试
  - 确保属性测试通过

- [x] 10. 实现 PatternHistory 历史记录
  - [x] 10.1 创建 `internal/pattern/history.go`
    - 实现 History 结构体（内存为主，落盘可选）
    - 实现 NewHistory() 构造函数
    - 实现 Add(), Recent(), Query() 方法
    - _Requirements: 6.1, 6.3, 6.5_
  - [x] 10.2 编写属性测试：历史记录持久化往返
    - **Property 10: 历史记录持久化往返**
    - **Validates: Requirements 6.1, 6.3, 6.5**

- [x] 11. 实现 SignalCombiner 信号组合
  - [x] 11.1 创建 `internal/signal/combiner.go`
    - 定义 Combiner 结构体和 CombinedSignal
    - 实现 NewCombiner() 构造函数
    - _Requirements: 4.1, 4.2_
  - [x] 11.2 实现信号相关性检测
    - 实现 AddPivotSignal() 和 AddPatternSignal()
    - 实现 checkCorrelation() 方法
    - _Requirements: 4.3, 4.4, 4.5, 4.6_
  - [x] 11.3 编写属性测试：信号相关性
    - **Property 6: 信号相关性时间窗口**
    - **Property 7: 组合信号完整性**
    - **Property 8: 方向匹配相关性强度**
    - **Property 9: 方向冲突相关性强度**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6**

- [x] 12. Checkpoint - 确保信号组合测试通过
  - 运行所有 signal 包测试
  - 确保属性测试通过

- [x] 13. 集成到 Monitor
  - [x] 13.1 修改 `internal/monitor/monitor.go`
    - 添加 KlineStore 字段
    - 在 onPrice() 中调用 KlineStore.Update()
    - _Requirements: 7.2, 7.3_
  - [x] 13.2 实现形态检测触发
    - 设置 onClose 回调
    - 在回调中异步调用 PatternDetector
    - _Requirements: 2.1, 7.2_
  - [x] 13.3 实现形态信号发射
    - 检测到形态后创建 Signal
    - 通过 SSE Broker 推送
    - 记录到 PatternHistory
    - _Requirements: 2.4, 5.7_
  - [x] 13.4 编写属性测试：检测范围限制
    - **Property 11: 形态检测范围限制**
    - **Validates: Requirements 7.4**

- [x] 14. 实现 HTTP API
  - [x] 14.1 修改 `internal/httpapi/server.go`
    - 添加 GET /api/patterns 端点
    - 添加 GET /api/klines 端点（调试用）
    - _Requirements: 6.6_
  - [x] 14.2 实现 SSE pattern 事件
    - 在 SSE 连接中推送 pattern 事件
    - _Requirements: 5.7_
  - [x] 14.3 修改 GET /api/history 响应
    - 添加 related_pattern 字段
    - _Requirements: 4.3, 4.4_

- [x] 15. Checkpoint - 确保后端集成测试通过
  - 运行完整后端测试
  - 手动测试 API 端点

- [x] 16. 实现前端 Patterns 标签页
  - [x] 16.1 修改 `internal/httpapi/static/index.html`
    - 添加 Patterns 标签
    - 添加形态筛选下拉框
    - 添加效率等级筛选
    - _Requirements: 5.1, 5.6_
  - [x] 16.2 修改 `internal/httpapi/static/app.js` - 数据层
    - 添加 masterPatterns 数组
    - 添加 SSE pattern 事件监听
    - _Requirements: 5.7_
  - [x] 16.3 实现 renderPatternItem() 函数
    - 紧凑模式：币种、形态名、方向箭头、成功率
    - 效率等级颜色编码
    - _Requirements: 5.2, 5.3_
  - [x] 16.4 实现 showPatternDetail() 弹窗
    - 显示完整统计数据
    - 上涨/下跌概率条形图
    - _Requirements: 5.2, 5.5_
  - [x] 16.5 实现 renderPatternBadge() 函数
    - 在枢轴点信号上显示形态徽章
    - 相关性颜色编码
    - _Requirements: 5.3, 5.4_
  - [x] 16.6 添加 CSS 样式
    - 形态信号项样式
    - 详情弹窗样式
    - 响应式适配
    - _Requirements: 5.1, 5.2_

- [x] 17. 实现配置选项
  - [x] 17.1 修改 `cmd/server/main.go`
    - 读取 KLINE_COUNT 环境变量
    - 读取 KLINE_INTERVAL 环境变量
    - 读取 PATTERN_ENABLED 环境变量
    - 读取 PATTERN_MIN_CONFIDENCE 环境变量
    - 读取 PATTERN_HISTORY_FILE 环境变量
    - 读取 PATTERN_CRYPTO_MODE 环境变量
    - _Requirements: 8.1, 8.2, 8.3_
  - [x] 17.2 启动时日志输出配置
    - _Requirements: 8.3_

- [x] 18. Final Checkpoint - 完整系统测试
  - [x] 运行所有测试 - 全部通过
  - [x] 修复 TestProperty_TimeWindowCorrelation 测试边界条件问题
  - 手动测试完整流程
  - 验证 SSE 推送
  - 验证前端展示

## Notes

- 所有测试任务都必须执行（完整测试覆盖）
- 每个 Checkpoint 确保增量验证
- 属性测试使用 `github.com/leanovate/gopter` 库
- K 线序列统一为"最旧在前，最新在后"
- 形态检测在独立 goroutine 中异步执行

