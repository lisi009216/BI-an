[36mWHAT[0m: 更新README
[35mWHY[0m:  重写文档内容以反映当前功能
# Binance Pivot Monitor

[English](#english) | [中文](#中文)

---

## Screenshots / 截图预览

| Web Dashboard | Side Panel + TradingView |
|:---:|:---:|
| ![Browser](docs/screenshots/browser.png) | ![TradingView](docs/screenshots/tradingview.png) |

| Side Panel + Binance | iOS PWA |
|:---:|:---:|
| ![Binance](docs/screenshots/binance.png) | ![iOS](docs/screenshots/ios.jpg) |

---

## English

### Overview

Binance Pivot Monitor is a real‑time pivot signal and market monitoring suite for Binance USDT perpetual futures. It provides a live dashboard, K‑line preview, position calculator, candlestick pattern recognition, and a Chrome extension with side‑panel mode.

### Key Features

- **Real‑time signals** from Binance mark price + ticker streams
- **Camarilla pivots** (daily & weekly) with auto refresh
- **K‑line preview**: line/candle modes, zoom, pivot overlays, hover price line
- **Position calculator** with pivot level presets and risk controls
- **Pattern detection** (talib + custom) with confidence & direction
- **Rankings** for 24h volume and trade count
- **SSE streaming** to web dashboard & extension
- **Side panel / PWA‑friendly UI**

### Quick Start

```bash
# Run locally
cd /Users/lichen/CascadeProjects/windsurf-project

go run ./cmd/server
```

Open: `http://localhost:8080`

Build binary:

```bash
go build -o binance-pivot-monitor ./cmd/server
./binance-pivot-monitor -addr :8080 -data-dir ./data
```

### Configuration

#### CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | HTTP server address |
| `-data-dir` | `data` | Data directory path |
| `-cors-origins` | `*` | Allowed CORS origins |
| `-binance-rest` | `https://fapi.binance.com` | Binance REST API base URL |
| `-refresh-workers` | `16` | Pivot refresh workers |
| `-monitor-heartbeat` | `0` | Heartbeat log interval (0=disabled) |
| `-history-max` | `20000` | Max signal history in memory |
| `-history-file` | `signals/history.jsonl` | History file (relative to `-data-dir`) |
| `-ticker-batch-interval` | `500ms` | Ticker SSE batch interval |

#### Environment variables

| Env | Default | Description |
|-----|---------|-------------|
| `PATTERN_ENABLED` | `true` | Enable candlestick pattern detection |
| `KLINE_COUNT` | `12` | Number of klines kept per symbol |
| `KLINE_INTERVAL` | `15m` | Kline interval (`5m` or minutes like `5`) |
| `PATTERN_MIN_CONFIDENCE` | `60` | Minimum confidence threshold |
| `PATTERN_CRYPTO_MODE` | `true` | Relax gap constraints for crypto markets |
| `PATTERN_HISTORY_FILE` | `patterns/history.jsonl` | Pattern history file (relative to `-data-dir`) |
| `PATTERN_HISTORY_MAX` | `1000` | Max patterns kept in memory |
| `RANKING_ENABLED` | `true` | Enable volume/trade ranking monitor |

### Chrome Extension

1. Open `chrome://extensions/`
2. Enable **Developer mode**
3. Click **Load unpacked** and select `extension/`

### API (Quick List)

- `GET /api/history` – signal history
- `GET /api/sse` – SSE stream (signals, tickers, patterns)
- `GET /api/tickers` – current ticker map
- `GET /api/patterns` – pattern history
- `GET /api/klines` / `GET /api/klines/stats` – kline debug & stats
- `GET /api/runtime` – runtime stats
- `GET /api/pivot-status` – pivot refresh status
- `GET /healthz` – health check

### Data & Storage

Runtime data (pivots, signals, patterns, rankings) lives under `-data-dir`. Use a custom path for local runs to keep the repo clean.

### License

MIT

---

## 中文

### 概述

Binance Pivot Monitor 是一个面向币安 USDT 永续合约的实时枢轴信号与行情监控系统，提供 Web 看板、K 线预览、仓位计算、形态识别，以及带侧边栏模式的 Chrome 扩展。

### 核心功能

- **实时信号**：标记价格 + 行情流
- **Camarilla 枢轴点**：日线 / 周线自动刷新
- **K 线预览**：线/蜡烛切换、缩放、枢轴线、价格线
- **仓位计算**：点位选择 + 风险控制
- **形态识别**：talib + 自定义，含置信度与方向
- **排行面板**：24h 成交额 / 成交笔数
- **SSE 推送**：Web 与扩展同步
- **侧边栏 / PWA 友好 UI**

### 快速开始

```bash
cd /Users/lichen/CascadeProjects/windsurf-project

go run ./cmd/server
```

浏览器打开：`http://localhost:8080`

构建二进制：

```bash
go build -o binance-pivot-monitor ./cmd/server
./binance-pivot-monitor -addr :8080 -data-dir ./data
```

### 配置

#### 启动参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-addr` | `:8080` | HTTP 服务地址 |
| `-data-dir` | `data` | 数据目录 |
| `-cors-origins` | `*` | 允许的 CORS 来源 |
| `-binance-rest` | `https://fapi.binance.com` | 币安 REST API |
| `-refresh-workers` | `16` | 枢轴刷新并发 |
| `-monitor-heartbeat` | `0` | 心跳日志间隔（0=禁用） |
| `-history-max` | `20000` | 信号历史上限 |
| `-history-file` | `signals/history.jsonl` | 历史文件（相对 `-data-dir`） |
| `-ticker-batch-interval` | `500ms` | 行情推送批量间隔 |

#### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PATTERN_ENABLED` | `true` | 启用形态识别 |
| `KLINE_COUNT` | `12` | 每个交易对保存 K 线数量 |
| `KLINE_INTERVAL` | `15m` | K 线周期（如 `5m` 或纯数字 `5`） |
| `PATTERN_MIN_CONFIDENCE` | `60` | 置信度阈值 |
| `PATTERN_CRYPTO_MODE` | `true` | 加密市场模式 |
| `PATTERN_HISTORY_FILE` | `patterns/history.jsonl` | 形态历史文件（相对 `-data-dir`） |
| `PATTERN_HISTORY_MAX` | `1000` | 形态内存上限 |
| `RANKING_ENABLED` | `true` | 启用排行监控 |

### Chrome 扩展安装

1. 打开 `chrome://extensions/`
2. 开启「开发者模式」
3. 点击「加载已解压的扩展程序」，选择 `extension/`

### API 列表（简）

- `GET /api/history` – 信号历史
- `GET /api/sse` – SSE 推送
- `GET /api/tickers` – 行情数据
- `GET /api/patterns` – 形态历史
- `GET /api/klines` / `GET /api/klines/stats` – K 线调试
- `GET /api/runtime` – 运行时信息
- `GET /api/pivot-status` – 枢轴刷新状态
- `GET /healthz` – 健康检查

### 数据目录

运行时数据保存在 `-data-dir`，本地调试建议使用独立目录，避免污染仓库。

### 许可证

MIT
