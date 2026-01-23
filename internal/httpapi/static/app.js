// Pivot Monitor Frontend v2.0 - Virtual List + Local Filtering
(function () {
    'use strict';

    const $ = id => document.getElementById(id);

    const LANG_KEY = "pivot_lang";
    const I18N = {
        zh: {
            app_title: "枢轴监控",
            placeholder_symbol: "币种",
            title_click_refresh: "点击刷新",
            title_sound_alert: "声音提醒",
            lang_zh: "中文",
            lang_en: "英文",
            status_connected: "已连接",
            status_reconnecting: "连接中",
            status_disconnected: "未连接",
            status_unknown: "未知",
            btn_refresh: "刷新",
            label_daily: "日",
            label_weekly: "周",
            pivot_period_day: "日线",
            pivot_period_week: "周线",
            option_all: "全部",
            option_up: "上涨",
            option_down: "下跌",
            placeholder_diff: "差值%",
            placeholder_vol: "成交额",
            label_levels: "点位",
            label_sound: "提示音",
            tab_signals: "信号",
            tab_patterns: "形态",
            tab_volume: "成交额",
            tab_trades: "成交笔数",
            hint_signals: "信号: {filtered}/{total}",
            hint_patterns: "形态: {filtered}/{total}",
            hint_loading: "加载中...",
            hint_load_failed: "加载失败: {error}",
            pattern_all: "全部形态",
            pattern_efficiency_high: "高效率 (A/B)",
            pattern_efficiency_medium: "中效率 (C/D)",
            pattern_rank_all: "全部等级",
            pattern_rank_a: "A (最高)",
            pattern_rank_b: "B (较高)",
            pattern_rank_c: "C (中等)",
            pattern_rank_d: "D (较低)",
            pattern_direction_all: "全部方向",
            pattern_direction_bullish: "看涨 ↑",
            pattern_direction_bearish: "看跌 ↓",
            pattern_count: "共{count}个形态",
            ranking_compare: "对比",
            ranking_sort: "排序",
            ranking_sort_order: "顺序",
            ranking_prev: "上一快照",
            ranking_5m: "5分钟",
            ranking_15m: "15分钟",
            ranking_30m: "30分钟",
            ranking_1h: "1小时",
            ranking_6h: "6小时",
            ranking_24h: "24小时",
            ranking_sort_default: "默认",
            ranking_sort_rank: "排名",
            ranking_sort_growth: "增长%",
            ranking_order_asc: "正序",
            ranking_order_desc: "倒序",
            modal_direction: "方向",
            modal_up_prob: "上涨概率",
            modal_down_prob: "下跌概率",
            modal_efficiency: "效率等级",
            modal_confidence: "置信度",
            modal_source: "来源",
            modal_detected: "发现时间",
            btn_trade: "交易",
            btn_filter: "筛选",
            btn_kline: "K线",
            btn_filter_short: "筛选",
            btn_filter_collapse: "收起筛选",
            btn_filter_expand: "展开筛选",
            action_jump_trade: "🚀 跳转交易",
            action_pivot_levels: "📊 枢轴点位",
            action_position_calc: "⚖️ 仓位计算",
            action_copy_symbol: "📋 复制币种",
            action_filter_symbol: "🔍 筛选币种",
            action_show_signals: "📊 查看信号",
            action_kline_preview: "K线预览",
            position_title: "仓位计算",
            position_direction: "方向",
            position_long: "做多",
            position_short: "做空",
            position_level: "参考点位",
            position_custom_price: "自定义价",
            position_offset: "偏移%",
            position_stop_pct: "止损比例",
            position_stop_price: "止损价格",
            position_max_loss: "最大亏损",
            position_entry: "开仓价",
            position_stop: "止损价",
            position_risk: "风险距离",
            position_size: "仓位大小",
            position_notional: "数量",
            position_diff: "当前价 {current} · 点位 {level} · 差 {diff}",
            position_error_no_price: "暂无当前价格",
            position_error_no_level: "请选择有效点位或自定义价",
            position_error_no_stop: "请填写止损比例或止损价格",
            position_error_invalid_stop: "止损价方向不正确",
            position_error_max_loss: "最大亏损需大于 0",
            toast_copied_value: "已复制: {value}",
            panel_signals: "信号",
            panel_volume: "成交额",
            panel_trades: "成交笔数",
            pivot_levels_title: "枢轴点位",
            ranking_history_title: "排名历史",
            chart_volume_rank: "成交额排名（24h）",
            chart_trades_rank: "成交笔数排名（24h）",
            chart_price_change: "价格变化（24h）",
            footer_sse: "SSE",
            footer_goroutines: "协程",
            footer_heap: "堆内存",
            footer_symbols: "币种",
            footer_signals: "信号",
            footer_uptime: "运行时长",
            no_signals: "暂无信号",
            no_patterns: "暂无形态",
            no_ranking: "暂无排名数据",
            time_just_now: "刚刚",
            time_seconds_ago: "{s}秒前",
            time_minutes_ago: "{m}分钟前",
            time_hours_minutes_ago: "{h}小时{m}分钟前",
            time_days_hours_ago: "{d}天{h}小时前",
            time_now: "刚刚",
            time_hm: "{h}小时{m}分钟",
            time_minutes: "{m}分钟",
            time_in_minutes: "{m}分钟后",
            period_day_short: "日",
            period_week_short: "周",
            label_volume_rank: "成交额排名",
            label_trades_rank: "成交笔数排名",
            label_price: "价格",
            label_volume: "成交额",
            label_trade_count: "成交笔数",
            label_rank: "等级",
            label_source_unknown: "未知",
            label_custom: "自定义",
            label_neutral: "中性",
            label_bullish: "看涨",
            label_bearish: "看跌",
            label_new: "新",
            label_trades_unit: "笔成交",
            label_current_price: "当前价格",
            label_symbols: "{count} 个币种",
            status_stale: "过期",
            status_ok: "正常",
            toast_copied: "已复制: {symbol}",
            toast_copy_failed: "复制失败",
            toast_filtered: "已筛选: {symbol}",
            toast_filter_cleared: "已清除筛选",
            toast_showing_signals: "已显示 {symbol} 的信号",
            toast_no_pivot: "无枢轴数据: {symbol}",
            toast_no_data: "暂无数据",
            text_loading: "加载中...",
            text_no_history: "暂无历史数据",
            text_not_enough_data: "数据不足",
            error_http: "请求失败: {status}",
            hint_ranking: "{type}: {count} 项（对比 {time}）",
            hint_ranking_simple: "{type}: {count} 项",
            type_volume: "成交额",
            type_trades: "成交笔数",
            label_signal: "信号",
            new_signals_tip: "有 {count} 条新信号",
            new_signals_action: "点击查看",
            new_signals_badge: "新 {count}",
            kline_title: "K线预览",
            kline_interval: "周期",
            kline_loading: "加载K线中...",
            kline_unavailable: "当前网络无法访问 Binance 永续合约",
            kline_network_fail: "K线加载失败，请检查网络",
            kline_no_data: "暂无K线数据",
            kline_price: "价格",
            kline_time: "时间",
            kline_click_load: "点击时间轴加载该时间段"
        },
        en: {
            app_title: "Pivot Monitor",
            placeholder_symbol: "Symbol",
            title_click_refresh: "Click to refresh",
            title_sound_alert: "Sound Alert",
            lang_zh: "Chinese",
            lang_en: "English",
            status_connected: "connected",
            status_reconnecting: "reconnecting",
            status_disconnected: "disconnected",
            status_unknown: "unknown",
            btn_refresh: "Refresh",
            label_daily: "Daily",
            label_weekly: "Weekly",
            pivot_period_day: "Daily",
            pivot_period_week: "Weekly",
            option_all: "All",
            option_up: "up",
            option_down: "down",
            placeholder_diff: "Diff%",
            placeholder_vol: "Vol",
            label_levels: "Levels",
            label_sound: "Sound",
            tab_signals: "Signals",
            tab_patterns: "Patterns",
            tab_volume: "Volume",
            tab_trades: "Trades",
            hint_signals: "Signals: {filtered}/{total}",
            hint_patterns: "Patterns: {filtered}/{total}",
            hint_loading: "Loading...",
            hint_load_failed: "Load failed: {error}",
            pattern_all: "All Patterns",
            pattern_efficiency_high: "High Efficiency (A/B)",
            pattern_efficiency_medium: "Medium Efficiency (C/D)",
            pattern_rank_all: "All Ranks",
            pattern_rank_a: "A (Highest)",
            pattern_rank_b: "B (High)",
            pattern_rank_c: "C (Medium)",
            pattern_rank_d: "D (Low)",
            pattern_direction_all: "All Directions",
            pattern_direction_bullish: "Bullish ↑",
            pattern_direction_bearish: "Bearish ↓",
            pattern_count: "{count} patterns",
            ranking_compare: "Compare",
            ranking_sort: "Sort",
            ranking_sort_order: "Order",
            ranking_prev: "Previous",
            ranking_5m: "5 min",
            ranking_15m: "15 min",
            ranking_30m: "30 min",
            ranking_1h: "1 hour",
            ranking_6h: "6 hours",
            ranking_24h: "24 hours",
            ranking_sort_default: "Default",
            ranking_sort_rank: "Rank",
            ranking_sort_growth: "Growth %",
            ranking_order_asc: "Asc",
            ranking_order_desc: "Desc",
            modal_direction: "Direction",
            modal_up_prob: "Up Probability",
            modal_down_prob: "Down Probability",
            modal_efficiency: "Efficiency Rank",
            modal_confidence: "Confidence",
            modal_source: "Source",
            modal_detected: "Detected At",
            btn_trade: "Trade",
            btn_filter: "Filter",
            btn_kline: "Kline",
            btn_filter_short: "Filters",
            btn_filter_collapse: "Hide Filters",
            btn_filter_expand: "Show Filters",
            action_jump_trade: "🚀 Jump to Trade",
            action_pivot_levels: "📊 Pivot Levels",
            action_position_calc: "⚖️ Position Calc",
            action_copy_symbol: "📋 Copy Symbol",
            action_filter_symbol: "🔍 Filter Symbol",
            action_show_signals: "📊 Show Signals",
            action_kline_preview: "Kline Preview",
            position_title: "Position Calc",
            position_direction: "Direction",
            position_long: "Long",
            position_short: "Short",
            position_level: "Reference",
            position_custom_price: "Custom price",
            position_offset: "Offset %",
            position_stop_pct: "Stop %",
            position_stop_price: "Stop Price",
            position_max_loss: "Max Loss",
            position_entry: "Entry",
            position_stop: "Stop",
            position_risk: "Risk Distance",
            position_size: "Position Size",
            position_notional: "Quantity",
            position_diff: "Current {current} · Level {level} · Diff {diff}",
            position_error_no_price: "No current price",
            position_error_no_level: "Select a valid level or custom price",
            position_error_no_stop: "Enter stop % or stop price",
            position_error_invalid_stop: "Stop price is invalid for the direction",
            position_error_max_loss: "Max loss must be > 0",
            toast_copied_value: "Copied: {value}",
            panel_signals: "Signals",
            panel_volume: "Volume",
            panel_trades: "Trades",
            pivot_levels_title: "Pivot Levels",
            ranking_history_title: "Ranking History",
            chart_volume_rank: "Volume Rank (24h)",
            chart_trades_rank: "Trades Rank (24h)",
            chart_price_change: "Price Change (24h)",
            footer_sse: "SSE",
            footer_goroutines: "Goroutines",
            footer_heap: "Heap",
            footer_symbols: "Symbols",
            footer_signals: "Signals",
            footer_uptime: "Uptime",
            no_signals: "No signals",
            no_patterns: "No patterns",
            no_ranking: "No ranking data",
            time_just_now: "just now",
            time_seconds_ago: "{s}s ago",
            time_minutes_ago: "{m}m ago",
            time_hours_minutes_ago: "{h}h {m}m ago",
            time_days_hours_ago: "{d}d {h}h ago",
            time_now: "now",
            time_hm: "{h}h {m}m",
            time_minutes: "{m}m",
            time_in_minutes: "in {m}m",
            period_day_short: "D",
            period_week_short: "W",
            label_volume_rank: "Volume Rank",
            label_trades_rank: "Trades Rank",
            label_price: "Price",
            label_volume: "Volume",
            label_trade_count: "Trade Count",
            label_rank: "Rank",
            label_source_unknown: "unknown",
            label_custom: "Custom",
            label_neutral: "Neutral",
            label_bullish: "Bullish",
            label_bearish: "Bearish",
            label_new: "NEW",
            label_trades_unit: "trades",
            label_current_price: "Current Price",
            label_symbols: "{count} symbols",
            status_stale: "STALE",
            status_ok: "OK",
            toast_copied: "Copied: {symbol}",
            toast_copy_failed: "Copy failed",
            toast_filtered: "Filtered: {symbol}",
            toast_filter_cleared: "Filter cleared",
            toast_showing_signals: "Showing signals for {symbol}",
            toast_no_pivot: "No pivot data for {symbol}",
            toast_no_data: "No data",
            text_loading: "Loading...",
            text_no_history: "No history data",
            text_not_enough_data: "Not enough data",
            error_http: "Request failed: {status}",
            hint_ranking: "{type}: {count} items (vs {time})",
            hint_ranking_simple: "{type}: {count} items",
            type_volume: "Volume",
            type_trades: "Trades",
            label_signal: "Signal",
            new_signals_tip: "{count} new signals",
            new_signals_action: "View now",
            new_signals_badge: "New {count}",
            kline_title: "Kline Preview",
            kline_interval: "Interval",
            kline_loading: "Loading klines...",
            kline_unavailable: "Binance Futures is not reachable",
            kline_network_fail: "Failed to load klines",
            kline_no_data: "No kline data",
            kline_price: "Price",
            kline_time: "Time",
            kline_click_load: "Click the time axis to load around that point"
        }
    };

    let currentLang = (localStorage.getItem(LANG_KEY) ||
        (navigator.language && navigator.language.toLowerCase().startsWith("zh") ? "zh" : "en"));

    function t(key, vars) {
        const table = I18N[currentLang] || I18N.zh;
        let text = table[key] || I18N.zh[key] || key;
        if (vars) {
            for (const [name, value] of Object.entries(vars)) {
                text = text.replace(new RegExp(`\\{${name}\\}`, "g"), String(value));
            }
        }
        return text;
    }

    function applyI18n() {
        document.documentElement.lang = currentLang === "zh" ? "zh-CN" : "en";
        document.querySelectorAll("[data-i18n]").forEach(el => {
            el.textContent = t(el.dataset.i18n);
        });
        document.querySelectorAll("[data-i18n-placeholder]").forEach(el => {
            el.setAttribute("placeholder", t(el.dataset.i18nPlaceholder));
        });
        document.querySelectorAll("[data-i18n-title]").forEach(el => {
            el.setAttribute("title", t(el.dataset.i18nTitle));
        });
        document.querySelectorAll("[data-i18n-label]").forEach(el => {
            el.setAttribute("label", t(el.dataset.i18nLabel));
        });

        const title = t("app_title");
        document.title = title;
        const appleTitle = document.querySelector('meta[name="apple-mobile-web-app-title"]');
        if (appleTitle) {
            appleTitle.setAttribute("content", title);
        }

        const langSelect = $("langSelect");
        if (langSelect) {
            langSelect.value = currentLang;
        }

        if (signalCluster) signalCluster.options.no_data_text = t("no_signals");
        if (patternCluster) patternCluster.options.no_data_text = t("no_patterns");
        if (rankingCluster) rankingCluster.options.no_data_text = t("no_ranking");

        if ($("status")) {
            setStatus(currentStatus);
        }

        updateNewSignalBanner();
        updateFilterToggleLabel();
        if (typeof calculatePosition === "function") {
            calculatePosition();
        }
        if (typeof updateKlineIntervalLabel === "function") {
            updateKlineIntervalLabel();
        }
        if (typeof updateKlineMeta === "function") {
            const point = (typeof klineHoverIndex === "number" && currentKlineData[klineHoverIndex])
                ? currentKlineData[klineHoverIndex]
                : null;
            updateKlineMeta(point);
        }
    }

    function setLanguage(lang) {
        if (!I18N[lang]) return;
        currentLang = lang;
        localStorage.setItem(LANG_KEY, lang);
        applyI18n();
        updateHint();
        updateView();
    }

    // ==================== 格式化函数 ====================
    const fmtRelTime = v => {
        try {
            const d = new Date(v);
            if (isNaN(d)) return String(v);
            const now = Date.now(), diff = Math.floor((now - d.getTime()) / 1000);
            if (diff < 0) return t("time_just_now");
            if (diff < 60) return t("time_seconds_ago", { s: diff });
            if (diff < 3600) return t("time_minutes_ago", { m: Math.floor(diff / 60) });
            if (diff < 86400) {
                return t("time_hours_minutes_ago", {
                    h: Math.floor(diff / 3600),
                    m: Math.floor((diff % 3600) / 60)
                });
            }
            return t("time_days_hours_ago", {
                d: Math.floor(diff / 86400),
                h: Math.floor((diff % 86400) / 3600)
            });
        } catch (_) { return String(v); }
    };

    const fmtPrice = v => {
        if (typeof v === "number") {
            const a = Math.abs(v);
            return a >= 1000 ? v.toFixed(2) : a >= 1 ? v.toFixed(4) : v.toPrecision(6);
        }
        return String(v);
    };

    const fmtPct = v => {
        if (typeof v !== "number") return "";
        const sign = v >= 0 ? "+" : "";
        return sign + v.toFixed(2) + "%";
    };

    const fmtTradeCount = n => {
        if (typeof n !== "number") return "";
        if (n >= 1e6) return (n / 1e6).toFixed(1) + "M";
        if (n >= 1e3) return (n / 1e3).toFixed(1) + "K";
        return String(n);
    };

    const fmtVolume = v => {
        if (typeof v !== "number") return "";
        if (v >= 1e9) return "$" + (v / 1e9).toFixed(2) + "B";
        if (v >= 1e6) return "$" + (v / 1e6).toFixed(1) + "M";
        if (v >= 1e3) return "$" + (v / 1e3).toFixed(0) + "K";
        return "$" + v.toFixed(0);
    };

    const QUOTE_SUFFIXES = ["USDT", "USDC", "BUSD", "USD", "BTC", "ETH"];
    const getBaseSymbol = symbol => {
        if (!symbol) return "";
        for (const q of QUOTE_SUFFIXES) {
            if (symbol.endsWith(q) && symbol.length > q.length) {
                return symbol.slice(0, -q.length);
            }
        }
        return symbol;
    };

    const fmtDur = s => {
        if (s < 0) return t("time_now");
        const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60);
        return h > 0 ? t("time_hm", { h, m }) : t("time_minutes", { m });
    };

    // 效率等级颜色
    const getEfficiencyColor = rank => {
        if (!rank) return '#6b7280';
        const first = rank[0];
        return {
            'A': '#22c55e',  // 绿色 - 高效
            'B': '#84cc16',  // 黄绿 - 较高效
            'C': '#eab308',  // 黄色 - 中等
            'D': '#f97316',  // 橙色 - 较低
            'E': '#ef4444',  // 红色 - 低效
            'F': '#ef4444',
            'G': '#6b7280',
            'J': '#6b7280'
        }[first] || '#6b7280';
    };

    // 方向颜色
    const getDirectionColor = dir => {
        return dir === 'bullish' ? '#22c55e' : dir === 'bearish' ? '#ef4444' : '#6b7280';
    };

    // 格式化时间
    const fmtTime = v => {
        try {
            const d = new Date(v);
            if (isNaN(d)) return String(v);
            const locale = currentLang === "zh" ? "zh-CN" : "en-US";
            return d.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        } catch (_) { return String(v); }
    };

    // ==================== 状态管理 ====================
    let masterSignals = [];      // 主数据（从后端加载）
    let filteredSignals = [];    // 过滤后数据（前端计算）
    let masterPatterns = [];     // 形态信号数据
    let filteredPatterns = [];   // 过滤后形态数据
    let tickerData = new Map();  // Ticker 数据
    let symbolRanking = { volume: new Map(), trades: new Map() };

    // Ranking monitor data (from backend API)
    let rankingData = { volume: [], trades: [] };
    let rankingCompare = ''; // Compare duration: '', '5m', '15m', '30m', '1h', '6h', '24h'
    let rankingSort = 'default'; // 'default', 'rank' or 'growth'
    let rankingSortOrder = 'asc'; // 'asc' or 'desc'

    // 枢轴点缓存 (5分钟过期)
    const pivotCache = new Map(); // symbol -> { data, timestamp }
    const PIVOT_CACHE_TTL = 5 * 60 * 1000; // 5 minutes

    let selectedLevels = new Set();
    let soundLevels = new Set(["R4", "R5", "S4", "S5"]);
    let currentView = 'signals';
    let filterCollapsed = false;
    let menuSymbol = null;
    let menuFromRanking = false;
    let currentPatternSignal = null; // 当前显示详情的形态信号
    let currentPivotPreviewSymbol = null; // 当前预览的交易对
    let currentPivotPreviewPeriod = '1d'; // 当前预览的周期
    let currentPositionSymbol = null; // 当前仓位计算的交易对
    let positionLevels = []; // 当前可选点位
    let positionStopMode = 'pct'; // 'pct' or 'price'
    let positionLastPrice = 0;
    let pendingSignalCount = 0; // 滚动期间积压的新信号数
    const SIGNAL_SCROLL_THRESHOLD = 8;
    const SIGNAL_SCROLL_IDLE_MS = 220;
    const signalScrollState = { isScrolling: false, awayFromTop: false, idleTimer: null };
    const wideSignalScrollState = { isScrolling: false, awayFromTop: false, idleTimer: null };

    // Clusterize 实例
    let signalCluster = null;
    let patternCluster = null;
    let rankingCluster = null;

    // ==================== 宽屏多面板模式 ====================
    const WIDESCREEN_BREAKPOINT = 1200; // 宽屏断点
    let isWidescreen = false; // 当前是否为宽屏模式

    // 宽屏模式下的 Clusterize 实例
    let wideSignalCluster = null;
    let wideVolumeCluster = null;
    let wideTradesCluster = null;

    // 宽屏面板独立的过滤状态
    let wideVolumeCompare = '';
    let wideVolumeSort = 'default';
    let wideVolumeSortOrder = 'asc';
    let wideTradesCompare = '';
    let wideTradesSort = 'default';
    let wideTradesSortOrder = 'asc';

    // localStorage keys
    const STORAGE_KEYS = {
        soundLevels: "pivot_sound_levels",
        soundEnabled: "pivot_sound_enabled",
        limit: "pivot_limit",
        minDiff: "pivot_min_diff",
        minVolume: "pivot_min_volume",
        volumeUnit: "pivot_volume_unit",
        filterLevels: "pivot_filter_levels",
        filterPeriod: "pivot_filter_period",
        filterDirection: "pivot_filter_direction",
        filterCollapsed: "pivot_filter_collapsed"
    };

    // ==================== 工具函数 ====================
    let currentStatus = "disconnected";
    const setStatus = s => {
        const e = $("status");
        currentStatus = s || "disconnected";
        const label = s === "connected" ? t("status_connected") :
            s === "reconnecting" ? t("status_reconnecting") :
                s === "disconnected" ? t("status_disconnected") :
                    t("status_unknown");
        e.textContent = label;
        e.classList.remove("connected", "reconnecting", "disconnected");
        if (s) e.classList.add(s);
    };

    const showToast = (msg, duration = 2000) => {
        const t = $("toast");
        t.textContent = msg;
        t.classList.add("show");
        setTimeout(() => t.classList.remove("show"), duration);
    };

    const debounce = (fn, ms) => { let t; return (...args) => { clearTimeout(t); t = setTimeout(() => fn(...args), ms); }; };

    function updateFilterToggleLabel() {
        const btn = $("filterToggle");
        if (!btn) return;
        btn.setAttribute("title", t(filterCollapsed ? "btn_filter_expand" : "btn_filter_collapse"));
        btn.setAttribute("aria-expanded", filterCollapsed ? "false" : "true");
        btn.classList.toggle("collapsed", filterCollapsed);
    }

    function setFilterCollapsed(collapsed, save = true) {
        filterCollapsed = !!collapsed;
        document.body.classList.toggle("filter-collapsed", filterCollapsed);
        updateFilterToggleLabel();
        if (save) {
            try {
                localStorage.setItem(STORAGE_KEYS.filterCollapsed, filterCollapsed ? "1" : "0");
            } catch (_) { }
        }
        calcScrollHeight();
    }

    function updateNewSignalBanner() {
        const count = pendingSignalCount;
        const banner = $("signalNewBanner");
        const bannerWide = $("signalNewBannerWide");
        const text = t("new_signals_tip", { count });
        const actionText = t("new_signals_action");
        const badge = $("pendingBadge");
        const badgeText = t("new_signals_badge", { count });

        if (banner) {
            const textEl = $("signalNewText");
            const btn = $("signalNewApply");
            if (textEl) textEl.textContent = text;
            if (btn) btn.textContent = actionText;
            banner.style.display = (!isWidescreen && currentView === 'signals' && count > 0) ? 'flex' : 'none';
        }

        if (bannerWide) {
            const textElWide = $("signalNewTextWide");
            const btnWide = $("signalNewApplyWide");
            if (textElWide) textElWide.textContent = text;
            if (btnWide) btnWide.textContent = actionText;
            bannerWide.style.display = (isWidescreen && count > 0) ? 'flex' : 'none';
        }

        if (badge) {
            badge.textContent = badgeText;
            badge.style.display = count > 0 ? 'inline-flex' : 'none';
        }
    }

    function clearPendingSignals() {
        if (pendingSignalCount === 0) return;
        pendingSignalCount = 0;
        updateNewSignalBanner();
    }

    function applyPendingSignals() {
        if (pendingSignalCount === 0) return;
        pendingSignalCount = 0;
        const scrollEl = isWidescreen ? $("wideSignalScroll") : $("signalScroll");
        if (scrollEl) scrollEl.scrollTop = 0;

        applyFilters();
        if (isWidescreen) {
            updateWideSignalPanel();
        } else if (currentView === 'signals') {
            updateSignalList();
        }
        updateNewSignalBanner();
    }

    function handleSignalScroll(el, state) {
        state.awayFromTop = el.scrollTop > SIGNAL_SCROLL_THRESHOLD;
        state.isScrolling = true;
        if (state.idleTimer) clearTimeout(state.idleTimer);
        state.idleTimer = setTimeout(() => {
            state.isScrolling = false;
            state.awayFromTop = el.scrollTop > SIGNAL_SCROLL_THRESHOLD;
            const canAutoApply = pendingSignalCount > 0 && !state.awayFromTop &&
                (isWidescreen || currentView === 'signals');
            if (canAutoApply) {
                applyPendingSignals();
                return;
            }
            updateNewSignalBanner();
        }, SIGNAL_SCROLL_IDLE_MS);
        updateNewSignalBanner();
    }

    function isSignalScrollBusy(isWide) {
        const scrollEl = isWide ? $("wideSignalScroll") : $("signalScroll");
        const state = isWide ? wideSignalScrollState : signalScrollState;
        if (!scrollEl) return false;
        state.awayFromTop = scrollEl.scrollTop > SIGNAL_SCROLL_THRESHOLD;
        return state.isScrolling || state.awayFromTop;
    }

    function shouldDeferSignalUpdate() {
        if (isWidescreen) {
            return isSignalScrollBusy(true);
        }
        if (currentView !== 'signals') return false;
        return isSignalScrollBusy(false);
    }

    function setupSignalScrollTracking() {
        const signalScroll = $("signalScroll");
        if (signalScroll) {
            signalScroll.addEventListener('scroll', () => {
                handleSignalScroll(signalScroll, signalScrollState);
            }, { passive: true });
        }
        const wideSignalScroll = $("wideSignalScroll");
        if (wideSignalScroll) {
            wideSignalScroll.addEventListener('scroll', () => {
                handleSignalScroll(wideSignalScroll, wideSignalScrollState);
            }, { passive: true });
        }
    }

    function setupSignalNewBanner() {
        const applyBtn = $("signalNewApply");
        if (applyBtn) applyBtn.onclick = applyPendingSignals;
        const applyBtnWide = $("signalNewApplyWide");
        if (applyBtnWide) applyBtnWide.onclick = applyPendingSignals;
    }

    // ==================== 设置持久化 ====================
    function loadSettings() {
        try {
            const saved = localStorage.getItem(STORAGE_KEYS.soundLevels);
            if (saved) soundLevels = new Set(JSON.parse(saved));

            const enabled = localStorage.getItem(STORAGE_KEYS.soundEnabled);
            if (enabled !== null) $("soundEnabled").checked = enabled === "true";

            const limit = localStorage.getItem(STORAGE_KEYS.limit);
            if (limit) $("limit").value = limit;

            const minDiff = localStorage.getItem(STORAGE_KEYS.minDiff);
            if (minDiff) $("minDiff").value = minDiff;

            const minVolume = localStorage.getItem(STORAGE_KEYS.minVolume);
            if (minVolume) $("minVolume").value = minVolume;

            const volumeUnit = localStorage.getItem(STORAGE_KEYS.volumeUnit);
            if (volumeUnit) $("volumeUnit").value = volumeUnit;

            // 恢复过滤器 Levels
            const filterLevels = localStorage.getItem(STORAGE_KEYS.filterLevels);
            if (filterLevels) {
                selectedLevels = new Set(JSON.parse(filterLevels));
            }

            // 恢复 Period 和 Direction
            const filterPeriod = localStorage.getItem(STORAGE_KEYS.filterPeriod);
            if (filterPeriod) $("period").value = filterPeriod;

            const filterDirection = localStorage.getItem(STORAGE_KEYS.filterDirection);
            if (filterDirection) $("direction").value = filterDirection;

            const collapsed = localStorage.getItem(STORAGE_KEYS.filterCollapsed);
            if (collapsed !== null) {
                setFilterCollapsed(collapsed === "1" || collapsed === "true", false);
            }
        } catch (_) { }
    }

    function saveSettings() {
        try {
            localStorage.setItem(STORAGE_KEYS.soundLevels, JSON.stringify(Array.from(soundLevels)));
            localStorage.setItem(STORAGE_KEYS.soundEnabled, $("soundEnabled").checked);
            localStorage.setItem(STORAGE_KEYS.limit, $("limit").value);
            const minDiff = $("minDiff").value;
            if (minDiff) localStorage.setItem(STORAGE_KEYS.minDiff, minDiff);
            else localStorage.removeItem(STORAGE_KEYS.minDiff);

            const minVolume = $("minVolume").value;
            if (minVolume) localStorage.setItem(STORAGE_KEYS.minVolume, minVolume);
            else localStorage.removeItem(STORAGE_KEYS.minVolume);

            localStorage.setItem(STORAGE_KEYS.volumeUnit, $("volumeUnit").value);

            // 保存过滤器 Levels
            localStorage.setItem(STORAGE_KEYS.filterLevels, JSON.stringify(Array.from(selectedLevels)));

            // 保存 Period 和 Direction
            localStorage.setItem(STORAGE_KEYS.filterPeriod, $("period").value);
            localStorage.setItem(STORAGE_KEYS.filterDirection, $("direction").value);
        } catch (_) { }
    }

    // ==================== 成交额过滤 ====================
    const VOLUME_UNITS = { K: 1e3, M: 1e6, B: 1e9 };

    function parseVolumeInput() {
        const value = parseFloat($("minVolume").value) || 0;
        const unit = $("volumeUnit").value || 'M';
        return value * (VOLUME_UNITS[unit] || 1e6);
    }

    // ==================== 过滤逻辑 ====================
    function getFilters() {
        return {
            symbol: $("symbol").value.trim(),
            period: $("period").value,
            levels: Array.from(selectedLevels),
            direction: $("direction").value,
            minDiff: parseFloat($("minDiff").value) || 0,
            minVolume: parseVolumeInput()
        };
    }

    function matchSignal(signal, filters) {
        if (!signal) return false;

        // Symbol 过滤：$开头精确匹配，否则模糊匹配
        if (filters.symbol) {
            const query = filters.symbol.toUpperCase();
            const sym = String(signal.symbol || "").toUpperCase();
            if (query.startsWith("$")) {
                // 精确匹配
                if (sym !== query.slice(1)) return false;
            } else {
                // 模糊匹配
                if (!sym.includes(query)) return false;
            }
        }

        // Period 过滤
        if (filters.period && signal.period !== filters.period) return false;

        // Level 过滤
        if (filters.levels.length && !filters.levels.includes(signal.level)) return false;

        // Direction 过滤
        if (filters.direction && signal.direction !== filters.direction) return false;

        // Diff% 过滤
        if (filters.minDiff > 0) {
            const ticker = tickerData.get(signal.symbol);
            if (ticker && signal.price > 0) {
                const diffPct = Math.abs((ticker.last_price - signal.price) / signal.price * 100);
                if (diffPct < filters.minDiff) return false;
            }
        }

        // 成交额过滤
        if (filters.minVolume > 0) {
            const ticker = tickerData.get(signal.symbol);
            if (!ticker || (ticker.quote_volume || 0) < filters.minVolume) {
                return false;
            }
        }

        return true;
    }

    function applyFilters() {
        const filters = getFilters();
        filteredSignals = masterSignals.filter(s => matchSignal(s, filters));
        updateHint();

        // 宽屏模式下同步更新信号面板
        if (isWidescreen && wideSignalCluster) {
            updateWideSignalPanel();
        }
    }

    function updateHint() {
        if (currentView === 'patterns') {
            $("hint").textContent = t("hint_patterns", {
                filtered: filteredPatterns.length,
                total: masterPatterns.length
            });
        } else if (isWidescreen) {
            // 宽屏模式下显示信号计数
            $("hint").textContent = t("hint_signals", {
                filtered: filteredSignals.length,
                total: masterSignals.length
            });
        } else {
            $("hint").textContent = t("hint_signals", {
                filtered: filteredSignals.length,
                total: masterSignals.length
            });
        }
    }

    // ==================== Pattern 过滤逻辑 ====================
    function getPatternFilters() {
        return {
            symbol: $("symbol").value.trim(),
            patternType: $("patternType").value,
            efficiencyRank: $("efficiencyRank").value,
            direction: $("patternDirection").value
        };
    }

    function matchPattern(pattern, filters) {
        if (!pattern) return false;

        // Symbol 过滤
        if (filters.symbol) {
            const query = filters.symbol.toUpperCase();
            const sym = String(pattern.symbol || "").toUpperCase();
            if (query.startsWith("$")) {
                if (sym !== query.slice(1)) return false;
            } else {
                if (!sym.includes(query)) return false;
            }
        }

        // Pattern type 过滤
        if (filters.patternType && pattern.pattern !== filters.patternType) return false;

        // Efficiency rank 过滤
        if (filters.efficiencyRank) {
            const rank = pattern.efficiency_rank || '';
            if (!rank.startsWith(filters.efficiencyRank)) return false;
        }

        // Direction 过滤
        if (filters.direction && pattern.direction !== filters.direction) return false;

        return true;
    }

    function applyPatternFilters() {
        const filters = getPatternFilters();
        filteredPatterns = masterPatterns.filter(p => matchPattern(p, filters));
        updateHint();
    }

    // ==================== 排行计算 ====================
    function computeRanking() {
        // 基于 masterSignals 中的交易对计算排行
        const signalSymbols = new Set(masterSignals.map(s => s.symbol));
        const items = [];

        for (const symbol of signalSymbols) {
            const ticker = tickerData.get(symbol);
            if (ticker) {
                items.push({
                    symbol,
                    volume: ticker.quote_volume || 0,
                    trades: ticker.trade_count || 0
                });
            }
        }

        // 成交额排名
        const byVolume = [...items].sort((a, b) => b.volume - a.volume);
        symbolRanking.volume.clear();
        byVolume.forEach((it, i) => symbolRanking.volume.set(it.symbol, i + 1));

        // 交易数排名
        const byTrades = [...items].sort((a, b) => b.trades - a.trades);
        symbolRanking.trades.clear();
        byTrades.forEach((it, i) => symbolRanking.trades.set(it.symbol, i + 1));

        return { byVolume, byTrades };
    }

    // ==================== 渲染函数 ====================
    function directionLabel(direction) {
        const dir = String(direction || "").toLowerCase();
        if (dir === "bullish" || dir === "up") return t("label_bullish");
        if (dir === "bearish" || dir === "down") return t("label_bearish");
        if (dir === "neutral") return t("label_neutral");
        return direction || "-";
    }

    function rankingTypeLabel(type) {
        return type === "trades" ? t("type_trades") : t("type_volume");
    }

    function renderSignalItem(signal, index) {
        const ticker = tickerData.get(signal.symbol);
        const volRank = symbolRanking.volume.get(signal.symbol);
        const tradeRank = symbolRanking.trades.get(signal.symbol);

        // 排行徽章
        let rankHtml = '';
        if (volRank || tradeRank) {
            const vr = volRank ? `<span class="rank-badge vol" title="${t("label_volume_rank")}">#${volRank}V</span>` : '';
            const tr = tradeRank ? `<span class="rank-badge trd" title="${t("label_trades_rank")}">#${tradeRank}T</span>` : '';
            rankHtml = `<div class="ranks">${vr}${tr}</div>`;
        }

        // 关联形态徽章
        let patternBadgeHtml = '';
        if (signal.related_pattern) {
            patternBadgeHtml = renderPatternBadge(signal.related_pattern);
        }

        // 价格差异
        let diffHtml = '';
        let tickerHtml = '';
        if (ticker) {
            const diff = ticker.last_price - signal.price;
            const diffPct = signal.price > 0 ? ((diff / signal.price) * 100) : 0;
            const diffSign = diff >= 0 ? '+' : '';
            const diffClass = diff >= 0 ? 'up' : 'down';
            diffHtml = `<span class="price-diff ${diffClass}">${diffSign}${diffPct.toFixed(2)}%</span>`;

            const pctClass = ticker.price_percent >= 0 ? 'up' : 'down';
            tickerHtml = `
                <div class="price-info">
                    <span class="price-now">${fmtPrice(ticker.last_price)}</span>
                    <span class="price-pct ${pctClass}">${fmtPct(ticker.price_percent)}</span>
                    <span class="volume">${fmtVolume(ticker.quote_volume)}</span>
                    <span class="trades">${fmtTradeCount(ticker.trade_count)} ${t("label_trades_unit")}</span>
                </div>
            `;
        }

        return `
            <div class="item" data-index="${index}" data-symbol="${signal.symbol}" data-price="${signal.price}" data-time="${signal.triggered_at}">
                <div class="top">
                    <div class="sym"><span class="symbol-link" data-symbol="${signal.symbol}">${signal.symbol}</span> ${rankHtml}</div>
                    <div class="tags">
                        <span class="tag">${signal.period}</span>
                        <span class="tag">${signal.level}</span>
                        <span class="tag ${signal.direction}">${directionLabel(signal.direction)}</span>
                        ${patternBadgeHtml}
                    </div>
                </div>
                <div class="sub">
                    <div>${t("label_signal")}: ${fmtPrice(signal.price)} ${diffHtml}</div>
                    <div class="muted time-rel">${fmtRelTime(signal.triggered_at)}</div>
                </div>
                ${tickerHtml}
                <div class="pivot-levels-row" data-symbol="${signal.symbol}"></div>
            </div>
        `;
    }

    function renderRankingItem(item, index, type) {
        const rankClass = index < 3 ? 'ranking-rank top3' : 'ranking-rank';
        const value = type === 'volume'
            ? fmtVolume(item.volume)
            : fmtTradeCount(item.trades) + ' ' + t("label_trades_unit");

        return `
            <div class="ranking-item" data-symbol="${item.symbol}">
                <span class="${rankClass}">#${index + 1}</span>
                <span class="ranking-symbol"><span class="symbol-link" data-symbol="${item.symbol}">${item.symbol}</span></span>
                <span class="ranking-value">${value}</span>
            </div>
        `;
    }

    // ==================== Pattern 渲染函数 ====================
    function renderPatternItem(pattern, index) {
        const effColor = getEfficiencyColor(pattern.efficiency_rank);
        const dirColor = getDirectionColor(pattern.direction);
        const dirArrow = pattern.direction === 'bullish' ? '↑' : pattern.direction === 'bearish' ? '↓' : '•';
        const sourceLabel = pattern.source === 'talib' ? 'TA-Lib' :
            pattern.source === 'custom' ? t("label_custom") : t("label_source_unknown");

        // 显示 confidence（需求 5.2）
        const confidence = pattern.confidence || 0;

        return `
            <div class="item pattern-item" data-index="${index}" data-symbol="${pattern.symbol}" data-pattern-id="${pattern.id}">
                <div class="top">
                    <div class="sym"><span class="symbol-link" data-symbol="${pattern.symbol}">${pattern.symbol}</span></div>
                    <div class="tags">
                        <span class="tag">${pattern.pattern_cn || pattern.pattern}</span>
                        <span class="tag" style="background:${dirColor};color:#fff">${dirArrow}</span>
                        <span class="tag rate" style="background:${effColor};color:#fff">${confidence}%</span>
                    </div>
                </div>
                <div class="sub">
                    <div>
                        <span class="efficiency">${t("label_rank")}: ${pattern.efficiency_rank || '-'}</span>
                        <span class="source-tag">${sourceLabel}</span>
                    </div>
                    <div class="muted time-rel" data-time="${pattern.detected_at}">${fmtRelTime(pattern.detected_at)}</div>
                </div>
            </div>
        `;
    }

    // 在枢轴点信号上显示关联形态徽章
    function renderPatternBadge(relatedPattern) {
        if (!relatedPattern) return '';

        const isBullish = relatedPattern.direction === 'bullish';
        const isBearish = relatedPattern.direction === 'bearish';

        // 上涨绿色，下跌红色
        const colorClass = isBullish ? 'bullish' : isBearish ? 'bearish' : 'neutral';
        const arrow = isBullish ? '↑' : isBearish ? '↓' : '•';

        // 显示对应方向的概率
        const rate = isBullish ? relatedPattern.up_percent :
            isBearish ? relatedPattern.down_percent :
                Math.max(relatedPattern.up_percent || 0, relatedPattern.down_percent || 0);

        // 形态名称（中文优先）
        const name = relatedPattern.pattern_cn || relatedPattern.pattern || '';

        // 时间差和数量信息
        const timeDiff = relatedPattern.time_diff || '';
        const count = relatedPattern.count || 1;
        const countInfo = count > 1 ? `(${count})` : '';

        // 构建 title 提示
        const dirText = isBullish ? t("label_bullish") : isBearish ? t("label_bearish") : t("label_neutral");
        const countText = count > 1 ? t("pattern_count", { count }) : "";
        const title = `${name} - ${dirText} ${rate}%${timeDiff ? ' - ' + timeDiff : ''}${countText ? ' - ' + countText : ''}`;

        return `
            <span class="pattern-badge ${colorClass}" title="${title}">
                ${arrow}${rate}% ${name}${countInfo} ${timeDiff}
            </span>
        `;
    }

    // ==================== Clusterize 管理 ====================
    function initClusterize() {
        // 信号列表
        signalCluster = new Clusterize({
            rows: [],
            scrollId: 'signalScroll',
            contentId: 'signalList',
            rows_in_block: 20,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_signals"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindSignalItemEvents();
                }
            }
        });

        // 形态列表
        patternCluster = new Clusterize({
            rows: [],
            scrollId: 'patternScroll',
            contentId: 'patternList',
            rows_in_block: 20,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_patterns"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindPatternItemEvents();
                }
            }
        });

        // 排行榜
        rankingCluster = new Clusterize({
            rows: [],
            scrollId: 'rankingScroll',
            contentId: 'rankingList',
            rows_in_block: 15,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_ranking"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindRankingItemEvents();
                }
            }
        });
    }

    // 初始化宽屏 Clusterize 实例
    function initWidescreenClusters() {
        if (wideSignalCluster) return; // 已初始化

        wideSignalCluster = new Clusterize({
            rows: [],
            scrollId: 'wideSignalScroll',
            contentId: 'wideSignalList',
            rows_in_block: 15,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_signals"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindWideSignalEvents();
                }
            }
        });

        wideVolumeCluster = new Clusterize({
            rows: [],
            scrollId: 'wideVolumeScroll',
            contentId: 'wideVolumeList',
            rows_in_block: 15,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_ranking"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindWideVolumeEvents();
                }
            }
        });

        wideTradesCluster = new Clusterize({
            rows: [],
            scrollId: 'wideTradesScroll',
            contentId: 'wideTradesList',
            rows_in_block: 15,
            blocks_in_cluster: 4,
            tag: null,
            no_data_text: t("no_ranking"),
            no_data_class: 'clusterize-no-data',
            callbacks: {
                clusterChanged: function () {
                    bindWideTradesEvents();
                }
            }
        });
    }

    // 销毁宽屏 Clusterize 实例
    function destroyWidescreenClusters() {
        if (wideSignalCluster) {
            wideSignalCluster.destroy();
            wideSignalCluster = null;
        }
        if (wideVolumeCluster) {
            wideVolumeCluster.destroy();
            wideVolumeCluster = null;
        }
        if (wideTradesCluster) {
            wideTradesCluster.destroy();
            wideTradesCluster = null;
        }
    }

    // ==================== 宽屏面板数据更新 ====================
    // 更新所有宽屏面板
    function updateWidescreenPanels() {
        if (!isWidescreen) return;

        // 更新信号面板
        updateWideSignalPanel();

        // 更新成交额面板
        updateWideVolumePanel();

        // 更新成交笔数面板
        updateWideTradesPanel();
    }

    // 更新信号面板
    function updateWideSignalPanel() {
        if (!wideSignalCluster) return;

        computeRanking();
        // 注意：不要调用 applyFilters()，直接使用已过滤的数据，避免无限递归
        const rows = filteredSignals.map((s, i) => renderSignalItem(s, i));
        wideSignalCluster.update(rows);

        // 更新计数
        const countEl = $("signalPanelCount");
        if (countEl) {
            countEl.textContent = `${filteredSignals.length}/${masterSignals.length}`;
        }

        clearPendingSignals();
    }

    // 更新成交额面板
    async function updateWideVolumePanel() {
        if (!wideVolumeCluster) return;

        try {
            await loadRanking('volume', wideVolumeCompare);
            const items = rankingData.volume || [];
            const filtered = filterRankingItems(items);
            const sorted = sortRankingItems(filtered, 'volume', wideVolumeSort, wideVolumeSortOrder);
            const rows = sorted.map((item, i) => renderRankingItemFromAPI(item, i, 'volume'));
            wideVolumeCluster.update(rows);

            // 更新计数
            const countEl = $("volumePanelCount");
            if (countEl) {
                countEl.textContent = String(filtered.length);
            }
        } catch (e) {
            console.error('Failed to update volume panel:', e);
        }
    }

    // 更新成交笔数面板
    async function updateWideTradesPanel() {
        if (!wideTradesCluster) return;

        try {
            await loadRanking('trades', wideTradesCompare);
            const items = rankingData.trades || [];
            const filtered = filterRankingItems(items);
            const sorted = sortRankingItems(filtered, 'trades', wideTradesSort, wideTradesSortOrder);
            const rows = sorted.map((item, i) => renderRankingItemFromAPI(item, i, 'trades'));
            wideTradesCluster.update(rows);

            // 更新计数
            const countEl = $("tradesPanelCount");
            if (countEl) {
                countEl.textContent = String(filtered.length);
            }
        } catch (e) {
            console.error('Failed to update trades panel:', e);
        }
    }

    function updateSignalList() {
        computeRanking();
        const rows = filteredSignals.map((s, i) => renderSignalItem(s, i));
        signalCluster.update(rows);
        clearPendingSignals();
    }

    function updatePatternList() {
        const rows = filteredPatterns.map((p, i) => renderPatternItem(p, i));
        patternCluster.update(rows);
    }

    function filterRankingItems(items) {
        const query = $("symbol").value.trim().toUpperCase();
        if (!query) {
            return items;
        }
        return items.filter(item => {
            const sym = String(item.symbol || "").toUpperCase();
            if (query.startsWith("$")) {
                return sym === query.slice(1);
            }
            return sym.includes(query);
        });
    }

    function sortRankingItems(items, type, customSort, customSortOrder) {
        if (items.length === 0) {
            return items;
        }
        const sort = customSort !== undefined ? customSort : rankingSort;
        const sortOrder = customSortOrder !== undefined ? customSortOrder : rankingSortOrder;
        if (sort === 'default') {
            return items;
        }
        const sorted = [...items];
        const order = sortOrder === 'desc' ? -1 : 1;
        const missingValue = sortOrder === 'desc' ? -Infinity : Infinity;
        if (sort === 'growth') {
            const field = type === 'volume' ? 'volume_change' : 'trade_change';
            sorted.sort((a, b) => {
                const av = typeof a[field] === "number" ? a[field] : missingValue;
                const bv = typeof b[field] === "number" ? b[field] : missingValue;
                return order * (av - bv);
            });
            return sorted;
        }

        if (sorted[0].rank !== undefined) {
            sorted.sort((a, b) => {
                const av = typeof a.rank === "number" ? a.rank : missingValue;
                const bv = typeof b.rank === "number" ? b.rank : missingValue;
                return order * (av - bv);
            });
            return sorted;
        }

        if (type === 'volume') {
            sorted.sort((a, b) => order * ((a.volume || 0) - (b.volume || 0)));
        } else {
            sorted.sort((a, b) => order * ((a.trades || 0) - (b.trades || 0)));
        }
        return sorted;
    }

    function updateRankingList() {
        const type = currentView; // 'volume' or 'trades'

        // Use backend API data if available
        const items = rankingData[type] || [];
        if (items.length > 0) {
            const filtered = filterRankingItems(items);
            const sorted = sortRankingItems(filtered, type);
            const rows = sorted.map((item, i) => renderRankingItemFromAPI(item, i, type));
            rankingCluster.update(rows);
        } else {
            // Fallback to local computation
            const { byVolume, byTrades } = computeRanking();
            const localItems = type === 'volume' ? byVolume : byTrades;
            const filtered = filterRankingItems(localItems);
            const sorted = sortRankingItems(filtered, type);
            const rows = sorted.map((item, i) => renderRankingItem(item, i, type));
            rankingCluster.update(rows);
        }
    }

    // Render ranking item from backend API response
    function renderRankingItemFromAPI(item, index, type) {
        const rankClass = item.rank <= 3 ? 'ranking-rank top3' : 'ranking-rank';
        const value = type === 'volume'
            ? fmtVolume(item.volume)
            : fmtTradeCount(item.trade_count) + ' ' + t("label_trades_unit");

        // Rank change indicator
        let changeHtml = '';
        if (item.rank_change !== null && item.rank_change !== undefined) {
            const change = item.rank_change;
            if (change > 0) {
                changeHtml = `<span class="rank-change up">↑${change}</span>`;
            } else if (change < 0) {
                changeHtml = `<span class="rank-change down">↓${Math.abs(change)}</span>`;
            } else {
                changeHtml = `<span class="rank-change">-</span>`;
            }
        } else if (item.is_new) {
            changeHtml = `<span class="rank-change new">${t("label_new")}</span>`;
        }

        // Price change indicator
        let priceChangeHtml = '';
        if (item.price_change !== null && item.price_change !== undefined) {
            const pctClass = item.price_change >= 0 ? 'up' : 'down';
            priceChangeHtml = `<span class="price-pct ${pctClass}">${fmtPct(item.price_change)}</span>`;
        }

        return `
            <div class="ranking-item" data-symbol="${item.symbol}">
                <span class="${rankClass}">#${item.rank}</span>
                <span class="ranking-symbol"><span class="symbol-link" data-symbol="${item.symbol}">${item.symbol}</span></span>
                ${changeHtml}
                <span class="ranking-value">${value}</span>
                ${priceChangeHtml}
            </div>
        `;
    }

    // ==================== 宽屏布局检测与切换 ====================
    // 检测并更新布局模式
    function checkWidescreenMode() {
        const width = window.innerWidth;
        // 防御性检查
        if (!width || width <= 0) {
            isWidescreen = false;
            return;
        }
        const newIsWidescreen = width > WIDESCREEN_BREAKPOINT;
        if (newIsWidescreen !== isWidescreen) {
            isWidescreen = newIsWidescreen;
            updateLayoutMode();
        }
    }

    // 切换布局模式
    function updateLayoutMode() {
        const multiPanel = $("multiPanelContainer");
        const scrollArea = document.querySelector(".scroll-area");
        const tabs = document.querySelector(".tabs");

        if (isWidescreen) {
            // 切换到宽屏模式
            if (multiPanel) multiPanel.style.display = 'flex';
            if (scrollArea) scrollArea.style.display = 'none';
            if (tabs) tabs.style.display = 'none';

            // 初始化宽屏 Clusterize
            initWidescreenClusters();

            // 加载所有面板数据
            updateWidescreenPanels();
        } else {
            // 切换到窄屏模式
            if (multiPanel) multiPanel.style.display = 'none';
            if (scrollArea) scrollArea.style.display = '';
            if (tabs) tabs.style.display = '';

            // 恢复原有视图
            updateView();
        }

        updateNewSignalBanner();
    }

    function updateView() {
        const signalScroll = $("signalScroll");
        const patternScroll = $("patternScroll");
        const rankingScroll = $("rankingScroll");
        const patternFilters = $("patternFilters");
        const rankingFilters = $("rankingFilters");
        const showSignalsBtn = document.querySelector('[data-action="signals"]');

        // 隐藏所有滚动区域
        signalScroll.style.display = 'none';
        patternScroll.style.display = 'none';
        rankingScroll.style.display = 'none';
        patternFilters.style.display = 'none';
        if (rankingFilters) rankingFilters.style.display = 'none';

        if (currentView === 'signals') {
            signalScroll.style.display = '';
            showSignalsBtn.style.display = 'none';
            applyFilters();
            updateSignalList();
        } else if (currentView === 'patterns') {
            patternScroll.style.display = '';
            patternFilters.style.display = 'flex';
            showSignalsBtn.style.display = 'none';
            applyPatternFilters();
            updatePatternList();
        } else {
            // Volume or Trades ranking view
            rankingScroll.style.display = '';
            if (rankingFilters) rankingFilters.style.display = 'flex';
            showSignalsBtn.style.display = '';
            // Load ranking data from backend API
            loadRanking(currentView).then(() => {
                updateRankingList();
            });
        }

        updateNewSignalBanner();
    }

    // ==================== Intersection Observer for Pivot Levels ====================
    const visibleSymbols = new Set();
    let pivotObserver = null;
    let pendingPivotSymbols = new Set(); // 待获取枢轴点的交易对
    let pivotFetchTimer = null;
    const PIVOT_FETCH_DEBOUNCE = 300; // 防抖延迟

    function initPivotObserver() {
        if (pivotObserver) return;

        pivotObserver = new IntersectionObserver((entries) => {
            let needsFetch = false;
            entries.forEach(entry => {
                const item = entry.target;
                const symbol = item.dataset.symbol;
                if (!symbol) return;

                if (entry.isIntersecting) {
                    visibleSymbols.add(symbol);
                    // 检查是否需要获取枢轴点数据
                    if (!pivotCache.has(symbol)) {
                        pendingPivotSymbols.add(symbol);
                        needsFetch = true;
                    } else {
                        // 已有缓存，直接更新
                        updateItemPivotLevelsSync(item, symbol);
                    }
                } else {
                    visibleSymbols.delete(symbol);
                    pendingPivotSymbols.delete(symbol);
                }
            });

            // 防抖批量获取
            if (needsFetch) {
                schedulePivotFetch();
            }
        }, { threshold: 0.1, rootMargin: '50px' });
    }

    function schedulePivotFetch() {
        if (pivotFetchTimer) clearTimeout(pivotFetchTimer);
        pivotFetchTimer = setTimeout(async () => {
            pivotFetchTimer = null;
            if (pendingPivotSymbols.size === 0) return;

            // 批量获取（目前 API 是单个获取，但我们可以并行请求）
            const symbols = Array.from(pendingPivotSymbols);
            pendingPivotSymbols.clear();

            // 并行获取所有待处理的交易对
            await Promise.all(symbols.map(async (symbol) => {
                if (!visibleSymbols.has(symbol)) return; // 已滚出可视区域
                await getPivotData(symbol);
            }));

            // 更新所有可视项的枢轴点位
            updateAllVisiblePivotLevels();
        }, PIVOT_FETCH_DEBOUNCE);
    }

    function updateAllVisiblePivotLevels() {
        document.querySelectorAll("#signalList .item[data-symbol]").forEach(item => {
            const symbol = item.dataset.symbol;
            if (visibleSymbols.has(symbol)) {
                updateItemPivotLevelsSync(item, symbol);
            }
        });
    }

    // 同步版本，使用已缓存的数据
    function updateItemPivotLevelsSync(item, symbol) {
        const pivotRow = item.querySelector('.pivot-levels-row');
        if (!pivotRow) return;

        const ticker = tickerData.get(symbol);
        if (!ticker || !ticker.last_price) {
            pivotRow.innerHTML = '';
            return;
        }

        const cached = pivotCache.get(symbol);
        if (!cached || !cached.data) {
            pivotRow.innerHTML = '';
            return;
        }

        const pivotData = cached.data;
        const currentPrice = ticker.last_price;

        // 日级点位
        const dailyLevels = findNearestLevels(pivotData, currentPrice, '1d');
        const dailyAbove = formatPivotLevel(dailyLevels.above, currentPrice, '1d');
        const dailyBelow = formatPivotLevel(dailyLevels.below, currentPrice, '1d');

        // 周级点位
        const weeklyLevels = findNearestLevels(pivotData, currentPrice, '1w');
        const weeklyAbove = formatPivotLevel(weeklyLevels.above, currentPrice, '1w');
        const weeklyBelow = formatPivotLevel(weeklyLevels.below, currentPrice, '1w');

        let html = '';
        if (dailyAbove || dailyBelow) {
            html += `<span class="pivot-pair">${dailyAbove} ${dailyBelow}</span>`;
        }
        if (weeklyAbove || weeklyBelow) {
            html += `<span class="pivot-pair">${weeklyAbove} ${weeklyBelow}</span>`;
        }

        pivotRow.innerHTML = html;
    }

    async function updateItemPivotLevels(item, symbol) {
        const pivotRow = item.querySelector('.pivot-levels-row');
        if (!pivotRow) return;

        const ticker = tickerData.get(symbol);
        if (!ticker || !ticker.last_price) {
            pivotRow.innerHTML = '';
            return;
        }

        const pivotData = await getPivotData(symbol);
        if (!pivotData) {
            pivotRow.innerHTML = '';
            return;
        }

        const currentPrice = ticker.last_price;

        // 日级点位
        const dailyLevels = findNearestLevels(pivotData, currentPrice, '1d');
        const dailyAbove = formatPivotLevel(dailyLevels.above, currentPrice, '1d');
        const dailyBelow = formatPivotLevel(dailyLevels.below, currentPrice, '1d');

        // 周级点位
        const weeklyLevels = findNearestLevels(pivotData, currentPrice, '1w');
        const weeklyAbove = formatPivotLevel(weeklyLevels.above, currentPrice, '1w');
        const weeklyBelow = formatPivotLevel(weeklyLevels.below, currentPrice, '1w');

        let html = '';
        if (dailyAbove || dailyBelow) {
            html += `<span class="pivot-pair">${dailyAbove} ${dailyBelow}</span>`;
        }
        if (weeklyAbove || weeklyBelow) {
            html += `<span class="pivot-pair">${weeklyAbove} ${weeklyBelow}</span>`;
        }

        pivotRow.innerHTML = html;
    }

    // ==================== 事件绑定 ====================
    function bindSymbolPreview(containerSelector) {
        document.querySelectorAll(`${containerSelector} .symbol-link`).forEach(link => {
            link.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                const symbol = link.dataset.symbol;
                if (symbol) showKlineModal(symbol);
            };
        });
    }

    function bindSignalItemEvents() {
        initPivotObserver();

        document.querySelectorAll("#signalList .item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                menuFromRanking = false;
                showActionMenu(e, item.dataset.symbol);
            };

            // 观察可视状态以更新枢轴点位
            if (pivotObserver) {
                pivotObserver.observe(item);
            }
        });

        bindSymbolPreview("#signalList");
    }

    function bindRankingItemEvents() {
        document.querySelectorAll("#rankingList .ranking-item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                const symbol = item.dataset.symbol;
                // Show ranking detail modal
                showRankingDetail(symbol);
            };
        });

        bindSymbolPreview("#rankingList");
    }

    function bindPatternItemEvents() {
        document.querySelectorAll("#patternList .pattern-item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                const patternId = item.dataset.patternId;
                const pattern = masterPatterns.find(p => p.id === patternId);
                if (pattern) {
                    showPatternDetail(pattern);
                }
            };
        });

        bindSymbolPreview("#patternList");
    }

    // ==================== 宽屏面板事件绑定 ====================
    function bindWideSignalEvents() {
        initPivotObserver();

        document.querySelectorAll("#wideSignalList .item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                menuFromRanking = false;
                showActionMenu(e, item.dataset.symbol);
            };

            // 观察可视状态以更新枢轴点位
            if (pivotObserver) {
                pivotObserver.observe(item);
            }
        });

        bindSymbolPreview("#wideSignalList");
    }

    function bindWideVolumeEvents() {
        document.querySelectorAll("#wideVolumeList .ranking-item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                const symbol = item.dataset.symbol;
                showRankingDetail(symbol);
            };
        });

        bindSymbolPreview("#wideVolumeList");
    }

    function bindWideTradesEvents() {
        document.querySelectorAll("#wideTradesList .ranking-item").forEach(item => {
            item.onclick = (e) => {
                e.preventDefault();
                e.stopPropagation();
                const symbol = item.dataset.symbol;
                showRankingDetail(symbol);
            };
        });

        bindSymbolPreview("#wideTradesList");
    }

    // ==================== Pattern 详情弹窗 ====================
    function showPatternDetail(pattern) {
        currentPatternSignal = pattern;

        $("modalSymbol").textContent = pattern.symbol;
        $("modalPatternName").textContent = pattern.pattern_cn || pattern.pattern;

        const dirText = pattern.direction === 'bullish' ? t("pattern_direction_bullish") :
            pattern.direction === 'bearish' ? t("pattern_direction_bearish") : t("label_neutral");
        $("modalDirection").textContent = dirText;
        $("modalDirection").className = 'value ' + pattern.direction;

        $("modalUpPct").textContent = (pattern.up_percent || 0) + '%';
        $("modalUpBar").style.width = (pattern.up_percent || 0) + '%';

        $("modalDownPct").textContent = (pattern.down_percent || 0) + '%';
        $("modalDownBar").style.width = (pattern.down_percent || 0) + '%';

        const effRank = pattern.efficiency_rank || '-';
        $("modalEfficiency").textContent = effRank;
        $("modalEfficiency").className = 'value efficiency-' + (effRank[0] || '');

        $("modalConfidence").textContent = (pattern.confidence || 0) + '%';

        const sourceText = pattern.source === 'talib' ? 'TA-Lib' :
            pattern.source === 'custom' ? t("label_custom") : (pattern.source || t("label_source_unknown"));
        $("modalSource").textContent = sourceText + (pattern.stats_source ? ' (' + pattern.stats_source + ')' : '');

        $("modalDetectedAt").textContent = fmtTime(pattern.detected_at);

        $("patternModal").style.display = 'flex';
    }

    function hidePatternModal() {
        $("patternModal").style.display = 'none';
        currentPatternSignal = null;
    }

    // 暴露到全局以便 HTML onclick 调用
    window.hidePatternModal = hidePatternModal;

    // ==================== Ranking Detail Modal ====================
    let currentRankingSymbol = null;

    async function showRankingDetail(symbol) {
        currentRankingSymbol = symbol;
        const modal = $("rankingModal");
        if (!modal) return;

        $("rankingModalSymbol").textContent = symbol;

        // Show loading state
        const currentDiv = $("rankingModalCurrent");
        currentDiv.innerHTML = `<div style="text-align:center;color:var(--text-tertiary)">${t("text_loading")}</div>`;

        modal.style.display = 'flex';

        // Load history data
        const historyData = await loadRankingHistory(symbol);
        if (!historyData || !historyData.snapshots || historyData.snapshots.length === 0) {
            currentDiv.innerHTML = `<div style="text-align:center;color:var(--text-tertiary)">${t("text_no_history")}</div>`;
            return;
        }

        // Render current stats
        const latest = historyData.snapshots[historyData.snapshots.length - 1];
        const ticker = tickerData.get(symbol);
        currentDiv.innerHTML = `
            <div class="current-row">
                <span class="current-label">${t("label_volume_rank")}</span>
                <span class="current-value">#${latest.volume_rank}</span>
            </div>
            <div class="current-row">
                <span class="current-label">${t("label_trades_rank")}</span>
                <span class="current-value">#${latest.trades_rank}</span>
            </div>
            <div class="current-row">
                <span class="current-label">${t("label_price")}</span>
                <span class="current-value">${fmtPrice(ticker ? ticker.last_price : latest.price)}</span>
            </div>
            <div class="current-row">
                <span class="current-label">${t("label_volume")}</span>
                <span class="current-value">${fmtVolume(latest.volume)}</span>
            </div>
        `;

        // Render charts
        renderRankingChart('rankingChartVolume', historyData.snapshots, 'volume_rank', true);
        renderRankingChart('rankingChartTrades', historyData.snapshots, 'trades_rank', true);

        const priceChangeSnapshots = historyData.snapshots.map((snap, i, arr) => {
            const prev = arr[i - 1];
            let priceChange = 0;
            if (prev && prev.price > 0) {
                priceChange = ((snap.price - prev.price) / prev.price) * 100;
            }
            return { ...snap, price_change: priceChange };
        });
        renderRankingChart('rankingChartPrice', priceChangeSnapshots, 'price_change', false);
    }

    function renderRankingChart(canvasId, snapshots, field, invertY) {
        const canvas = $(canvasId);
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;

        // Clear canvas
        ctx.fillStyle = '#29313D';
        ctx.fillRect(0, 0, width, height);

        if (snapshots.length < 2) {
            ctx.fillStyle = '#707A8A';
            ctx.font = '11px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(t("text_not_enough_data"), width / 2, height / 2);
            return;
        }

        // Extract values
        const values = snapshots.map(s => s[field]);
        const minVal = Math.min(...values);
        const maxVal = Math.max(...values);
        const range = maxVal - minVal || 1;

        // Draw line
        ctx.strokeStyle = field === 'price' ? '#F0B90B' : '#0ECB81';
        ctx.lineWidth = 1.5;
        ctx.beginPath();

        const padding = 10;
        const chartWidth = width - padding * 2;
        const chartHeight = height - padding * 2;

        snapshots.forEach((snap, i) => {
            const x = padding + (i / (snapshots.length - 1)) * chartWidth;
            let y;
            if (invertY) {
                // For ranks, lower is better, so invert Y axis
                y = padding + ((snap[field] - minVal) / range) * chartHeight;
            } else {
                y = padding + chartHeight - ((snap[field] - minVal) / range) * chartHeight;
            }

            if (i === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });

        ctx.stroke();

        // Draw min/max labels
        ctx.fillStyle = '#707A8A';
        ctx.font = '9px sans-serif';
        ctx.textAlign = 'left';

        if (field === 'price') {
            ctx.fillText(fmtPrice(maxVal), padding, padding + 8);
            ctx.fillText(fmtPrice(minVal), padding, height - padding);
        } else {
            ctx.fillText('#' + (invertY ? minVal : maxVal), padding, padding + 8);
            ctx.fillText('#' + (invertY ? maxVal : minVal), padding, height - padding);
        }
    }

    function hideRankingModal() {
        $("rankingModal").style.display = 'none';
        currentRankingSymbol = null;
    }

    window.hideRankingModal = hideRankingModal;

    function setupRankingModal() {
        const modal = $("rankingModal");
        if (!modal) return;

        // Click background to close
        modal.onclick = (e) => {
            if (e.target === modal) {
                hideRankingModal();
            }
        };

        // Trade button
        const tradeBtn = $("rankingModalTradeBtn");
        if (tradeBtn) {
            tradeBtn.onclick = () => {
                if (currentRankingSymbol) {
                    jumpToTrade(currentRankingSymbol);
                    hideRankingModal();
                }
            };
        }

        const klineBtn = $("rankingModalKlineBtn");
        if (klineBtn) {
            klineBtn.onclick = () => {
                if (currentRankingSymbol) {
                    showKlineModal(currentRankingSymbol);
                    hideRankingModal();
                }
            };
        }

        // Filter button
        const filterBtn = $("rankingModalFilterBtn");
        if (filterBtn) {
            filterBtn.onclick = () => {
                if (currentRankingSymbol) {
                    $("symbol").value = "$" + currentRankingSymbol;
                    currentView = 'signals';
                    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
                    document.querySelector('.tab[data-view="signals"]').classList.add('active');
                    applyFilters();
                    updateView();
                    hideRankingModal();
                    showToast(t("toast_filtered", { symbol: currentRankingSymbol }));
                }
            };
        }
    }

    // ==================== 枢轴点预览 Modal ====================
    async function showPivotPreview(symbol) {
        currentPivotPreviewSymbol = symbol;
        const pivotData = await getPivotData(symbol);
        if (!pivotData) {
            showToast(t("toast_no_pivot", { symbol }));
            return;
        }

        const modal = $("pivotModal");
        if (!modal) return;

        $("pivotModalSymbol").textContent = symbol;
        renderPivotPreview(pivotData, currentPivotPreviewPeriod);
        modal.style.display = 'flex';
    }

    function renderPivotPreview(pivotData, period) {
        const container = $("pivotLevelsList");
        if (!container) return;

        const levels = period === '1w' ? pivotData.weekly : pivotData.daily;
        if (!levels) {
            container.innerHTML = `<div class="no-data">${t("toast_no_data")}</div>`;
            return;
        }

        const ticker = tickerData.get(currentPivotPreviewSymbol);
        const currentPrice = ticker ? ticker.last_price : 0;

        // 构建所有点位数组
        const allLevels = [
            { name: 'R5', price: levels.r5, type: 'resistance' },
            { name: 'R4', price: levels.r4, type: 'resistance' },
            { name: 'R3', price: levels.r3, type: 'resistance' },
            { name: 'R2', price: levels.r2, type: 'resistance' },
            { name: 'R1', price: levels.r1, type: 'resistance' },
            { name: 'PP', price: levels.pp, type: 'pivot' },
            { name: 'S1', price: levels.s1, type: 'support' },
            { name: 'S2', price: levels.s2, type: 'support' },
            { name: 'S3', price: levels.s3, type: 'support' },
            { name: 'S4', price: levels.s4, type: 'support' },
            { name: 'S5', price: levels.s5, type: 'support' },
        ].filter(l => l.price > 0);

        // 按价格降序排列
        allLevels.sort((a, b) => b.price - a.price);

        // 找到当前价格应该插入的位置
        let currentPriceInserted = false;
        let html = '';

        for (let i = 0; i < allLevels.length; i++) {
            const level = allLevels[i];
            const pct = currentPrice > 0 ? ((level.price - currentPrice) / currentPrice * 100) : 0;
            const pctStr = (pct >= 0 ? '+' : '') + pct.toFixed(2) + '%';

            // 在适当位置插入当前价格
            if (!currentPriceInserted && currentPrice > 0 && level.price < currentPrice) {
                html += `<div class="pivot-level current-price">
                    <span class="level-name">${t("label_current_price")}</span>
                    <span class="level-price">${fmtPrice(currentPrice)}</span>
                    <span class="level-pct">--</span>
                </div>`;
                currentPriceInserted = true;
            }

            const typeClass = level.type === 'resistance' ? 'resistance' :
                level.type === 'support' ? 'support' : 'pivot-point';
            const pctClass = pct >= 0 ? 'pct-up' : 'pct-down';

            html += `<div class="pivot-level ${typeClass}">
                <span class="level-name">${level.name}</span>
                <span class="level-price">${fmtPrice(level.price)}</span>
                <span class="level-pct ${pctClass}">${pctStr}</span>
            </div>`;
        }

        // 如果当前价格低于所有点位
        if (!currentPriceInserted && currentPrice > 0) {
            html += `<div class="pivot-level current-price">
                <span class="level-name">${t("label_current_price")}</span>
                <span class="level-price">${fmtPrice(currentPrice)}</span>
                <span class="level-pct">--</span>
            </div>`;
        }

        container.innerHTML = html;

        // 更新周期按钮状态
        document.querySelectorAll('.pivot-period-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.period === period);
        });
    }

    function hidePivotModal() {
        const modal = $("pivotModal");
        if (modal) modal.style.display = 'none';
        currentPivotPreviewSymbol = null;
    }

    function switchPivotPeriod(period) {
        currentPivotPreviewPeriod = period;
        if (currentPivotPreviewSymbol) {
            getPivotData(currentPivotPreviewSymbol).then(data => {
                if (data) renderPivotPreview(data, period);
            });
        }
    }

    // 暴露到全局
    window.hidePivotModal = hidePivotModal;
    window.switchPivotPeriod = switchPivotPeriod;

    // ==================== 仓位计算 ====================
    const POSITION_CUSTOM_VALUE = "__custom__";
    const POSITION_CURRENT_VALUE = "__current__";

    function buildPositionLevels(pivotData) {
        const levels = [];
        if (!pivotData) return levels;

        const addLevels = (prefix, data) => {
            if (!data) return;
            [
                ["R5", data.r5],
                ["R4", data.r4],
                ["R3", data.r3],
                ["R2", data.r2],
                ["R1", data.r1],
                ["PP", data.pp],
                ["S1", data.s1],
                ["S2", data.s2],
                ["S3", data.s3],
                ["S4", data.s4],
                ["S5", data.s5]
            ].forEach(([name, price]) => {
                if (price && price > 0) {
                    levels.push({
                        key: `${prefix}-${name}`,
                        label: `${prefix}${name}`,
                        price
                    });
                }
            });
        };

        addLevels("D", pivotData.daily);
        addLevels("W", pivotData.weekly);
        return levels;
    }

    function formatInputPrice(v) {
        if (!isFinite(v)) return "";
        const a = Math.abs(v);
        if (a >= 1000) return v.toFixed(2);
        if (a >= 1) return v.toFixed(4);
        return v.toPrecision(6);
    }

    function formatAmount(v) {
        if (!isFinite(v)) return "-";
        const a = Math.abs(v);
        if (a >= 1000) return v.toFixed(2);
        if (a >= 1) return v.toFixed(4);
        return v.toPrecision(6);
    }

    function inferPrecisionFromPrice(price) {
        if (!isFinite(price)) return 4;
        const s = String(price);
        if (s.includes("e-")) {
            const [base, exp] = s.split("e-");
            const baseDecimals = (base.split(".")[1] || "").length;
            const expNum = parseInt(exp, 10) || 0;
            return expNum + baseDecimals;
        }
        const parts = s.split(".");
        return parts[1] ? parts[1].length : 0;
    }

    function getPricePrecision(symbol, fallbackPrice) {
        const ticker = symbol ? tickerData.get(symbol) : null;
        const refPrice = ticker && ticker.last_price ? ticker.last_price : fallbackPrice;
        return inferPrecisionFromPrice(refPrice);
    }

    function formatPriceWithPrecision(price, precision) {
        if (!isFinite(price)) return "-";
        if (precision <= 0) return Math.round(price).toString();
        return price.toFixed(precision);
    }

    function snapPrice(price, precision, mode = "round") {
        if (!isFinite(price)) return price;
        if (precision <= 0) return Math[mode === "floor" ? "floor" : mode === "ceil" ? "ceil" : "round"](price);
        const step = Math.pow(10, -precision);
        const scaled = price / step;
        const fn = mode === "floor" ? Math.floor : mode === "ceil" ? Math.ceil : Math.round;
        return fn(scaled) * step;
    }

    function formatDiffPct(levelPrice, currentPrice) {
        if (!levelPrice || !currentPrice) return "";
        const diffPct = ((levelPrice - currentPrice) / currentPrice) * 100;
        const sign = diffPct >= 0 ? "+" : "";
        return sign + diffPct.toFixed(2) + "%";
    }

    function formatLevelOption(label, price, currentPrice, precision) {
        const priceText = formatPriceWithPrecision(price, precision);
        const diffText = formatDiffPct(price, currentPrice);
        return diffText ? `${label} ${priceText} (${diffText})` : `${label} ${priceText}`;
    }

    function getSelectedPositionPrice() {
        const select = $("positionLevelSelect");
        if (!select) return 0;
        if (select.value === POSITION_CURRENT_VALUE) {
            const ticker = tickerData.get(currentPositionSymbol);
            return ticker ? ticker.last_price : 0;
        }
        if (select.value === POSITION_CUSTOM_VALUE) {
            return parseFloat($("positionLevelCustom").value) || 0;
        }
        const level = positionLevels.find(l => l.key === select.value);
        return level ? level.price : 0;
    }

    function refreshPositionOptions(currentPrice) {
        const select = $("positionLevelSelect");
        if (!select || !currentPrice) return;
        const precision = getPricePrecision(currentPositionSymbol, currentPrice);
        select.querySelectorAll("option").forEach(opt => {
            const value = opt.value;
            if (value === POSITION_CUSTOM_VALUE) return;
            if (value === POSITION_CURRENT_VALUE) {
                opt.dataset.price = String(currentPrice);
                opt.dataset.label = t("label_current_price");
            }
            const price = parseFloat(opt.dataset.price) || 0;
            const label = opt.dataset.label || opt.textContent || "";
            const nextText = formatLevelOption(label, price, currentPrice, precision);
            if (opt.textContent !== nextText) {
                opt.textContent = nextText;
            }
        });
    }

    function updatePositionDiff(currentPrice, refPrice) {
        const diffEl = $("positionDiff");
        if (!diffEl) return;
        if (!currentPrice || !refPrice) {
            diffEl.textContent = "-";
            return;
        }
        const diffPct = ((refPrice - currentPrice) / currentPrice) * 100;
        const diffText = (diffPct >= 0 ? "+" : "") + diffPct.toFixed(2) + "%";
        const precision = getPricePrecision(currentPositionSymbol, currentPrice);
        diffEl.textContent = t("position_diff", {
            current: formatPriceWithPrecision(currentPrice, precision),
            level: formatPriceWithPrecision(refPrice, precision),
            diff: diffText
        });
    }

    function calculatePosition() {
        const modal = $("positionModal");
        if (!modal || modal.style.display === 'none' || !currentPositionSymbol) return;

        const ticker = tickerData.get(currentPositionSymbol);
        const currentPrice = ticker ? ticker.last_price : 0;
        const refPrice = getSelectedPositionPrice();
        const precision = getPricePrecision(currentPositionSymbol, currentPrice);
        const roundedCurrent = currentPrice ? snapPrice(currentPrice, precision, "round") : 0;
        if (roundedCurrent && roundedCurrent !== positionLastPrice) {
            positionLastPrice = roundedCurrent;
            refreshPositionOptions(roundedCurrent);
        }
        updatePositionDiff(currentPrice, refPrice);

        const entryEl = $("positionEntry");
        const stopEl = $("positionStop");
        const riskEl = $("positionRisk");
        const sizeEl = $("positionSize");
        const notionalEl = $("positionNotional");
        const errorEl = $("positionError");

        const setError = (msg) => {
            if (!errorEl) return;
            if (!msg) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
                return;
            }
            errorEl.textContent = msg;
            errorEl.style.display = 'block';
        };

        if (!currentPrice) {
            setError(t("position_error_no_price"));
            if (entryEl) {
                entryEl.textContent = "-";
                entryEl.dataset.value = "";
            }
            if (stopEl) {
                stopEl.textContent = "-";
                stopEl.dataset.value = "";
            }
            if (riskEl) riskEl.textContent = "-";
            if (sizeEl) sizeEl.textContent = "-";
            if (notionalEl) notionalEl.textContent = "-";
            return;
        }

        if (!refPrice) {
            setError(t("position_error_no_level"));
            if (entryEl) {
                entryEl.textContent = "-";
                entryEl.dataset.value = "";
            }
            if (stopEl) {
                stopEl.textContent = "-";
                stopEl.dataset.value = "";
            }
            if (riskEl) riskEl.textContent = "-";
            if (sizeEl) sizeEl.textContent = "-";
            if (notionalEl) notionalEl.textContent = "-";
            return;
        }

        const sideBtn = document.querySelector("#positionSide button.active");
        const side = sideBtn ? sideBtn.dataset.side : 'long';
        const offset = parseFloat($("positionOffset").value) || 0;
        const entry = side === 'long'
            ? refPrice * (1 - offset / 100)
            : refPrice * (1 + offset / 100);
        const entryPrice = snapPrice(entry, precision, "round");

        if (entryEl) {
            const entryText = formatPriceWithPrecision(entryPrice, precision);
            entryEl.textContent = entryText;
            entryEl.dataset.value = entryText;
        }

        const stopPctInput = parseFloat($("positionStopPct").value);
        const stopPriceInput = parseFloat($("positionStopPrice").value);
        let stopPrice = NaN;
        let stopPct = NaN;
        const stopMode = side === 'long' ? "floor" : "ceil";

        if (positionStopMode === 'price' && isFinite(stopPriceInput)) {
            stopPrice = snapPrice(stopPriceInput, precision, stopMode);
            if (entryPrice > 0) {
                stopPct = side === 'long'
                    ? ((entryPrice - stopPrice) / entryPrice) * 100
                    : ((stopPrice - entryPrice) / entryPrice) * 100;
                if (isFinite(stopPct)) $("positionStopPct").value = stopPct.toFixed(2);
            }
        } else if (positionStopMode === 'pct' && isFinite(stopPctInput)) {
            stopPct = stopPctInput;
            const rawStop = side === 'long'
                ? entryPrice * (1 - stopPct / 100)
                : entryPrice * (1 + stopPct / 100);
            stopPrice = snapPrice(rawStop, precision, stopMode);
            if (isFinite(stopPrice)) $("positionStopPrice").value = formatPriceWithPrecision(stopPrice, precision);
        } else if (isFinite(stopPriceInput)) {
            stopPrice = snapPrice(stopPriceInput, precision, stopMode);
        } else if (isFinite(stopPctInput)) {
            stopPct = stopPctInput;
            const rawStop = side === 'long'
                ? entryPrice * (1 - stopPct / 100)
                : entryPrice * (1 + stopPct / 100);
            stopPrice = snapPrice(rawStop, precision, stopMode);
        }

        if (!isFinite(stopPrice)) {
            setError(t("position_error_no_stop"));
            if (stopEl) {
                stopEl.textContent = "-";
                stopEl.dataset.value = "";
            }
            if (riskEl) riskEl.textContent = "-";
            if (sizeEl) sizeEl.textContent = "-";
            if (notionalEl) notionalEl.textContent = "-";
            return;
        }

        if ((side === 'long' && stopPrice >= entryPrice) || (side === 'short' && stopPrice <= entryPrice)) {
            setError(t("position_error_invalid_stop"));
            if (stopEl) {
                const stopText = formatPriceWithPrecision(stopPrice, precision);
                stopEl.textContent = stopText;
                stopEl.dataset.value = stopText;
            }
            if (riskEl) riskEl.textContent = "-";
            if (sizeEl) sizeEl.textContent = "-";
            if (notionalEl) notionalEl.textContent = "-";
            return;
        }

        const riskPerUnit = Math.abs(entryPrice - stopPrice);
        const riskPct = entryPrice > 0 ? (riskPerUnit / entryPrice) * 100 : 0;
        if (stopEl) {
            const stopText = formatPriceWithPrecision(stopPrice, precision);
            stopEl.textContent = stopText;
            stopEl.dataset.value = stopText;
        }
        if (riskEl) riskEl.textContent = `${formatPriceWithPrecision(riskPerUnit, precision)} (${riskPct.toFixed(2)}%)`;

        const maxLoss = parseFloat($("positionMaxLoss").value) || 0;
        if (maxLoss <= 0) {
            setError(t("position_error_max_loss"));
            if (sizeEl) sizeEl.textContent = "-";
            if (notionalEl) notionalEl.textContent = "-";
            return;
        }

        const size = riskPerUnit > 0 ? maxLoss / riskPerUnit : 0;
        const notional = size * entryPrice;
        if (sizeEl) sizeEl.textContent = formatAmount(notional) + " USDT";
        if (notionalEl) {
            const base = getBaseSymbol(currentPositionSymbol);
            notionalEl.textContent = formatAmount(size) + (base ? " " + base : "");
        }
        setError("");
    }

    async function showPositionModal(symbol) {
        currentPositionSymbol = symbol;
        const modal = $("positionModal");
        if (!modal) return;

        $("positionSymbol").textContent = symbol;
        positionStopMode = 'pct';

        let pivotData = null;
        try {
            pivotData = await getPivotData(symbol);
        } catch (_) {
            pivotData = null;
        }

        positionLevels = buildPositionLevels(pivotData);
        const select = $("positionLevelSelect");
        if (select) {
            const ticker = tickerData.get(symbol);
            const currentPrice = ticker ? ticker.last_price : 0;
            const precision = getPricePrecision(symbol, currentPrice);
            positionLastPrice = currentPrice ? snapPrice(currentPrice, precision, "round") : 0;

            if (currentPrice) {
                positionLevels.push({
                    key: POSITION_CURRENT_VALUE,
                    label: t("label_current_price"),
                    price: currentPrice
                });
            }

            positionLevels.sort((a, b) => b.price - a.price);

            const options = positionLevels.map(level => {
                const label = level.key === POSITION_CURRENT_VALUE ? t("label_current_price") : level.label;
                const text = formatLevelOption(label, level.price, currentPrice, precision);
                return `<option value="${level.key}" data-price="${level.price}" data-label="${label}">${text}</option>`;
            }).join("");
            select.innerHTML = options + `<option value="${POSITION_CUSTOM_VALUE}">${t("label_custom")}</option>`;

            select.value = currentPrice ? POSITION_CURRENT_VALUE : (positionLevels[0] ? positionLevels[0].key : POSITION_CUSTOM_VALUE);
        }

        const customInput = $("positionLevelCustom");
        if (customInput) {
            customInput.style.display = (select && select.value === POSITION_CUSTOM_VALUE) ? '' : 'none';
            if (customInput.style.display !== 'none' && !customInput.value) {
                const ticker = tickerData.get(symbol);
                if (ticker && ticker.last_price) {
                    customInput.value = formatInputPrice(ticker.last_price);
                }
            }
        }

        modal.style.display = 'flex';
        calculatePosition();
    }

    function hidePositionModal() {
        const modal = $("positionModal");
        if (modal) modal.style.display = 'none';
        currentPositionSymbol = null;
    }

    function setupPositionModal() {
        const modal = $("positionModal");
        if (!modal) return;

        modal.onclick = (e) => {
            if (e.target === modal) hidePositionModal();
        };

        const bindCopy = (el) => {
            if (!el) return;
            el.onclick = (e) => {
                e.stopPropagation();
                const val = el.dataset.value || "";
                if (!val) return;
                copyToClipboard(val)
                    .then(() => showToast(t("toast_copied_value", { value: val })))
                    .catch(() => showToast(t("toast_copy_failed")));
            };
        };
        bindCopy($("positionEntry"));
        bindCopy($("positionStop"));

        const sideBtns = document.querySelectorAll("#positionSide button");
        sideBtns.forEach(btn => {
            btn.onclick = () => {
                sideBtns.forEach(b => b.classList.remove("active"));
                btn.classList.add("active");
                calculatePosition();
            };
        });

        const select = $("positionLevelSelect");
        const customInput = $("positionLevelCustom");
        if (select) {
            select.onchange = () => {
                if (customInput) {
                    customInput.style.display = select.value === POSITION_CUSTOM_VALUE ? '' : 'none';
                }
                calculatePosition();
            };
        }
        if (customInput) {
            customInput.oninput = () => calculatePosition();
        }

        $("positionOffset").oninput = () => calculatePosition();
        $("positionStopPct").oninput = () => {
            positionStopMode = 'pct';
            calculatePosition();
        };
        $("positionStopPrice").oninput = () => {
            positionStopMode = 'price';
            calculatePosition();
        };
        $("positionMaxLoss").oninput = () => calculatePosition();
    }

    function updatePositionModalIfOpen() {
        const modal = $("positionModal");
        if (!modal || modal.style.display === 'none' || !currentPositionSymbol) return;
        calculatePosition();
    }

    window.hidePositionModal = hidePositionModal;

    // ==================== Kline 预览 ====================
    const BINANCE_PING_URL = "https://fapi.binance.com/fapi/v1/ping";
    const BINANCE_KLINE_URL = "https://fapi.binance.com/fapi/v1/klines";
    const KLINE_INTERVALS = ["1m", "5m", "15m", "1h", "4h", "1d"];
    const KLINE_INTERVAL_MS = {
        "1m": 60 * 1000,
        "5m": 5 * 60 * 1000,
        "15m": 15 * 60 * 1000,
        "1h": 60 * 60 * 1000,
        "4h": 4 * 60 * 60 * 1000,
        "1d": 24 * 60 * 60 * 1000
    };
    const KLINE_LIMIT = 160;
    const KLINE_CACHE_TTL = 45 * 1000;
    let binanceAvailable = false;
    let binanceChecked = false;
    let binanceCheckPromise = null;
    let currentKlineSymbol = null;
    let currentKlineInterval = "15m";
    let currentKlineData = [];
    let currentKlineCenterTime = null;
    let currentKlinePivotData = null;
    let klineHoverIndex = null;
    let klineFetchController = null;
    const klineCache = new Map();

    async function checkBinanceAvailability() {
        if (binanceCheckPromise) return binanceCheckPromise;
        binanceCheckPromise = (async () => {
            const controller = new AbortController();
            const timer = setTimeout(() => controller.abort(), 2500);
            try {
                const r = await fetch(BINANCE_PING_URL, { signal: controller.signal, cache: "no-store" });
                binanceAvailable = r.ok;
            } catch (_) {
                binanceAvailable = false;
            } finally {
                clearTimeout(timer);
                binanceChecked = true;
            }
            return binanceAvailable;
        })();
        return binanceCheckPromise;
    }

    function setKlineLoading(show, msg) {
        const loading = $("klineLoading");
        if (!loading) return;
        if (msg) loading.textContent = msg;
        loading.classList.toggle("show", show);
    }

    function showKlineError(msg) {
        const meta = $("klineMeta");
        if (meta) meta.textContent = msg;
        setKlineLoading(true, msg);
        setTimeout(() => setKlineLoading(false), 1200);
    }

    function updateKlineIntervalLabel() {
        const label = $("klineIntervalText");
        if (label) label.textContent = currentKlineInterval;
        document.querySelectorAll("#klineIntervals button").forEach(btn => {
            btn.classList.toggle("active", btn.dataset.interval === currentKlineInterval);
        });
    }

    function fmtKlineTime(ts) {
        try {
            const d = new Date(ts);
            if (isNaN(d)) return String(ts);
            const locale = currentLang === "zh" ? "zh-CN" : "en-US";
            if (currentKlineInterval.includes("d")) {
                return d.toLocaleDateString(locale, { month: '2-digit', day: '2-digit' });
            }
            return d.toLocaleString(locale, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
        } catch (_) {
            return String(ts);
        }
    }

    function updateKlineMeta(point) {
        const meta = $("klineMeta");
        if (!meta) return;
        if (!point) {
            meta.textContent = t("kline_no_data");
            return;
        }
        const priceText = `${t("kline_price")}: ${fmtPrice(point.close)}`;
        const timeText = `${t("kline_time")}: ${fmtKlineTime(point.openTime)}`;
        meta.textContent = `${priceText}  ${timeText}`;
    }

    function getKlinePivotLines() {
        if (!currentKlinePivotData) return [];
        const ticker = tickerData.get(currentKlineSymbol);
        const lastPoint = currentKlineData.length ? currentKlineData[currentKlineData.length - 1] : null;
        const currentPrice = ticker ? ticker.last_price : (lastPoint ? lastPoint.close : 0);
        if (!currentPrice) return [];

        const daily = findNearestLevels(currentKlinePivotData, currentPrice, '1d');
        const weekly = findNearestLevels(currentKlinePivotData, currentPrice, '1w');
        const lines = [];

        if (daily.above) lines.push({ price: daily.above.price, label: `D:${daily.above.name}`, type: 'resistance' });
        if (daily.below) lines.push({ price: daily.below.price, label: `D:${daily.below.name}`, type: 'support' });
        if (weekly.above) lines.push({ price: weekly.above.price, label: `W:${weekly.above.name}`, type: 'resistance' });
        if (weekly.below) lines.push({ price: weekly.below.price, label: `W:${weekly.below.name}`, type: 'support' });

        return lines;
    }

    function buildKlineRange(centerTime, interval) {
        if (!centerTime) return {};
        const step = KLINE_INTERVAL_MS[interval] || 60 * 1000;
        const half = Math.floor(KLINE_LIMIT / 2);
        const startTime = Math.max(0, centerTime - step * half);
        const endTime = startTime + step * KLINE_LIMIT;
        return { startTime, endTime };
    }

    async function fetchKlines(symbol, interval, centerTime, signal) {
        const range = buildKlineRange(centerTime, interval);
        const params = new URLSearchParams({
            symbol,
            interval,
            limit: String(KLINE_LIMIT)
        });
        if (range.startTime) {
            params.set("startTime", String(range.startTime));
            params.set("endTime", String(range.endTime));
        }
        const r = await fetch(`${BINANCE_KLINE_URL}?${params.toString()}`, { cache: "no-store", signal });
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        const raw = await r.json();
        if (!Array.isArray(raw)) return [];
        return raw.map(row => ({
            openTime: row[0],
            open: parseFloat(row[1]),
            high: parseFloat(row[2]),
            low: parseFloat(row[3]),
            close: parseFloat(row[4]),
            volume: parseFloat(row[5]),
            closeTime: row[6]
        }));
    }

    async function loadKlineData(symbol, interval, centerTime) {
        if (!symbol) return;
        if (!binanceAvailable) {
            showKlineError(t("kline_unavailable"));
            return;
        }

        const key = `${symbol}_${interval}_${centerTime || 'latest'}`;
        const cached = klineCache.get(key);
        const now = Date.now();

        if (cached && (now - cached.ts) < KLINE_CACHE_TTL) {
            currentKlineData = cached.data;
            currentKlineCenterTime = centerTime || null;
            klineHoverIndex = currentKlineData.length ? currentKlineData.length - 1 : null;
            drawKlineChart();
            updateKlineMeta(klineHoverIndex !== null ? currentKlineData[klineHoverIndex] : null);
            return;
        }

        if (klineFetchController) klineFetchController.abort();
        klineFetchController = new AbortController();
        setKlineLoading(true, t("kline_loading"));

        try {
            const data = await fetchKlines(symbol, interval, centerTime, klineFetchController.signal);
            currentKlineData = data;
            currentKlineCenterTime = centerTime || null;
            klineHoverIndex = data.length ? data.length - 1 : null;
            klineCache.set(key, { data, ts: now });
            drawKlineChart();
            updateKlineMeta(klineHoverIndex !== null ? data[klineHoverIndex] : null);
            setKlineLoading(false);
        } catch (e) {
            if (e.name === "AbortError") return;
            showKlineError(t("kline_network_fail"));
        }
    }

    function drawKlineChart() {
        const canvas = $("klineCanvas");
        if (!canvas) return;
        const rect = canvas.getBoundingClientRect();
        const dpr = window.devicePixelRatio || 1;
        canvas.width = rect.width * dpr;
        canvas.height = rect.height * dpr;

        const ctx = canvas.getContext('2d');
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
        ctx.clearRect(0, 0, rect.width, rect.height);

        const data = currentKlineData || [];
        if (data.length < 2) {
            ctx.fillStyle = '#707A8A';
            ctx.font = '12px "Exo 2", sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(t("kline_no_data"), rect.width / 2, rect.height / 2);
            return;
        }

        const values = data.map(d => d.close);
        let minVal = Math.min(...values);
        let maxVal = Math.max(...values);
        const paddingRange = (maxVal - minVal) * 0.08 || 1;
        minVal -= paddingRange;
        maxVal += paddingRange;

        const padding = { left: 44, right: 12, top: 12, bottom: 24 };
        const width = rect.width - padding.left - padding.right;
        const height = rect.height - padding.top - padding.bottom;

        // Grid
        ctx.strokeStyle = 'rgba(255,255,255,0.06)';
        ctx.lineWidth = 1;
        for (let i = 0; i <= 3; i++) {
            const y = padding.top + (height / 3) * i;
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(rect.width - padding.right, y);
            ctx.stroke();
        }

        // Line + area
        const gradient = ctx.createLinearGradient(0, padding.top, 0, rect.height - padding.bottom);
        gradient.addColorStop(0, 'rgba(240,185,11,0.35)');
        gradient.addColorStop(1, 'rgba(240,185,11,0)');

        ctx.beginPath();
        data.forEach((point, i) => {
            const x = padding.left + (i / (data.length - 1)) * width;
            const y = padding.top + height - ((point.close - minVal) / (maxVal - minVal)) * height;
            if (i === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
        });
        ctx.strokeStyle = '#F0B90B';
        ctx.lineWidth = 1.6;
        ctx.stroke();

        ctx.lineTo(padding.left + width, padding.top + height);
        ctx.lineTo(padding.left, padding.top + height);
        ctx.closePath();
        ctx.fillStyle = gradient;
        ctx.fill();

        // Axis labels
        ctx.fillStyle = '#707A8A';
        ctx.font = '10px "Exo 2", sans-serif';
        ctx.textAlign = 'left';
        ctx.fillText(fmtPrice(maxVal), 6, padding.top + 8);
        ctx.fillText(fmtPrice(minVal), 6, rect.height - padding.bottom);

        // Pivot lines (do not affect scale)
        const pivotLines = getKlinePivotLines();
        pivotLines.forEach((line) => {
            const y = padding.top + height - ((line.price - minVal) / (maxVal - minVal)) * height;
            if (y < padding.top - 4 || y > padding.top + height + 4) return;

            ctx.setLineDash(line.label.startsWith('W:') ? [6, 4] : [3, 3]);
            ctx.strokeStyle = line.type === 'support' ? 'rgba(14,203,129,0.45)' : 'rgba(246,70,93,0.45)';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(rect.width - padding.right, y);
            ctx.stroke();
            ctx.setLineDash([]);

            const label = `${line.label} ${fmtPrice(line.price)}`;
            ctx.fillStyle = line.type === 'support' ? 'rgba(14,203,129,0.9)' : 'rgba(246,70,93,0.9)';
            ctx.font = '10px "Exo 2", sans-serif';
            ctx.textAlign = 'right';
            ctx.fillText(label, rect.width - padding.right, Math.max(padding.top + 10, Math.min(y - 4, rect.height - padding.bottom - 4)));
        });

        // Time labels
        const labelIndexes = [0, Math.floor(data.length / 2), data.length - 1];
        ctx.textAlign = 'center';
        labelIndexes.forEach(i => {
            const x = padding.left + (i / (data.length - 1)) * width;
            ctx.fillText(fmtKlineTime(data[i].openTime), x, rect.height - 6);
        });

        // Crosshair
        if (klineHoverIndex !== null && data[klineHoverIndex]) {
            const point = data[klineHoverIndex];
            const x = padding.left + (klineHoverIndex / (data.length - 1)) * width;
            const y = padding.top + height - ((point.close - minVal) / (maxVal - minVal)) * height;
            ctx.strokeStyle = 'rgba(240,185,11,0.4)';
            ctx.beginPath();
            ctx.moveTo(x, padding.top);
            ctx.lineTo(x, rect.height - padding.bottom);
            ctx.stroke();
            ctx.fillStyle = '#F0B90B';
            ctx.beginPath();
            ctx.arc(x, y, 3, 0, Math.PI * 2);
            ctx.fill();
        }
    }

    function handleKlineCanvasClick(e) {
        const canvas = $("klineCanvas");
        if (!canvas || !currentKlineData.length) return;
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const index = Math.round((x / rect.width) * (currentKlineData.length - 1));
        const point = currentKlineData[index];
        if (!point) return;
        klineHoverIndex = index;
        updateKlineMeta(point);
        drawKlineChart();

        if (y >= rect.height - 24) {
            loadKlineData(currentKlineSymbol, currentKlineInterval, point.openTime);
        }
    }

    function renderKlineIntervals() {
        const container = $("klineIntervals");
        if (!container) return;
        container.innerHTML = "";
        KLINE_INTERVALS.forEach(interval => {
            const btn = document.createElement("button");
            btn.textContent = interval;
            btn.dataset.interval = interval;
            btn.onclick = () => {
                currentKlineInterval = interval;
                updateKlineIntervalLabel();
                loadKlineData(currentKlineSymbol, currentKlineInterval, currentKlineCenterTime);
            };
            container.appendChild(btn);
        });
        updateKlineIntervalLabel();
    }

    function showKlineModal(symbol, interval) {
        currentKlineSymbol = symbol;
        if (interval) currentKlineInterval = interval;
        const modal = $("klineModal");
        if (!modal) return;

        if (!binanceChecked) {
            checkBinanceAvailability().then(() => {
                if (!binanceAvailable) {
                    showToast(t("kline_unavailable"));
                    return;
                }
                showKlineModal(symbol, interval);
            });
            return;
        }

        if (!binanceAvailable) {
            showToast(t("kline_unavailable"));
            return;
        }

        $("klineSymbol").textContent = symbol;
        modal.style.display = 'flex';
        updateKlineIntervalLabel();
        loadKlineData(symbol, currentKlineInterval, null);

        currentKlinePivotData = null;
        getPivotData(symbol).then(data => {
            currentKlinePivotData = data;
            drawKlineChart();
        }).catch(() => {
            currentKlinePivotData = null;
        });
    }

    function hideKlineModal() {
        const modal = $("klineModal");
        if (modal) modal.style.display = 'none';
        currentKlineSymbol = null;
    }

    function setupKlineModal() {
        const modal = $("klineModal");
        if (!modal) return;

        modal.onclick = (e) => {
            if (e.target === modal) hideKlineModal();
        };

        const canvas = $("klineCanvas");
        if (canvas) {
            canvas.addEventListener('click', handleKlineCanvasClick);
        }

        renderKlineIntervals();
        updateKlineIntervalLabel();
        window.addEventListener('resize', debounce(() => {
            if (modal.style.display !== 'none') {
                drawKlineChart();
            }
        }, 150));
    }

    window.hideKlineModal = hideKlineModal;
    window.showKlineModal = showKlineModal;

    // ==================== 操作菜单 ====================
    function showActionMenu(e, symbol) {
        menuSymbol = symbol;
        const menu = $("actionMenu");
        const showSignalsBtn = document.querySelector('[data-action="signals"]');
        showSignalsBtn.style.display = menuFromRanking ? '' : 'none';

        menu.classList.add("show");
        const x = Math.min(e.clientX, window.innerWidth - 200);
        const y = Math.min(e.clientY, window.innerHeight - 220);
        menu.style.left = x + "px";
        menu.style.top = y + "px";
    }

    function hideActionMenu() {
        $("actionMenu").classList.remove("show");
        menuSymbol = null;
    }

    function copyToClipboard(text) {
        if (navigator.clipboard && window.isSecureContext) {
            return navigator.clipboard.writeText(text);
        }
        return new Promise((resolve, reject) => {
            const textarea = document.createElement('textarea');
            textarea.value = text;
            textarea.style.position = 'fixed';
            textarea.style.left = '-9999px';
            document.body.appendChild(textarea);
            textarea.focus();
            textarea.select();
            try {
                const success = document.execCommand('copy');
                document.body.removeChild(textarea);
                if (success) resolve();
                else reject(new Error('execCommand failed'));
            } catch (err) {
                document.body.removeChild(textarea);
                reject(err);
            }
        });
    }

    function setupActionMenu() {
        document.addEventListener("click", (e) => {
            if (!e.target.closest("#actionMenu")) hideActionMenu();
        });

        document.querySelectorAll(".action-menu-item").forEach(item => {
            item.addEventListener("click", (e) => {
                e.stopPropagation();
                const action = item.dataset.action;
                if (!menuSymbol) return;

                switch (action) {
                    case "trade":
                        jumpToTrade(menuSymbol);
                        break;

                    case "copy":
                        copyToClipboard(menuSymbol)
                            .then(() => showToast(t("toast_copied", { symbol: menuSymbol })))
                            .catch(() => showToast(t("toast_copy_failed")));
                        break;

                    case "filter":
                        const currentSymbol = $("symbol").value.trim().toUpperCase();
                        const exactSymbol = "$" + menuSymbol;
                        if (currentSymbol === menuSymbol || currentSymbol === exactSymbol) {
                            $("symbol").value = "";
                            applyFilters();
                            updateView();
                            showToast(t("toast_filter_cleared"));
                        } else {
                            $("symbol").value = exactSymbol;
                            applyFilters();
                            updateView();
                            showToast(t("toast_filtered", { symbol: menuSymbol }));
                        }
                        break;

                    case "signals":
                        // 切换到 Signals 面板并过滤
                        $("symbol").value = "$" + menuSymbol;
                        currentView = 'signals';
                        document.querySelectorAll(".tab").forEach(t => {
                            t.classList.toggle("active", t.dataset.view === 'signals');
                        });
                        applyFilters();
                        updateView();
                        showToast(t("toast_showing_signals", { symbol: menuSymbol }));
                        break;

                    case "pivots":
                        // 显示枢轴点预览
                        showPivotPreview(menuSymbol);
                        break;

                    case "position":
                        showPositionModal(menuSymbol);
                        break;

                    case "kline":
                        showKlineModal(menuSymbol);
                        break;
                }
                hideActionMenu();
            });
        });
    }

    // ==================== 控件设置 ====================
    function setupLevelBtns() {
        // 过滤 Level 按钮
        document.querySelectorAll("#filterLevels button").forEach(b => {
            b.addEventListener("click", () => {
                const l = b.dataset.level;
                if (selectedLevels.has(l)) {
                    selectedLevels.delete(l);
                    b.classList.remove("active");
                } else {
                    selectedLevels.add(l);
                    b.classList.add("active");
                }
                saveSettings();
                applyFilters();
                updateView();
            });
        });

        // 声音 Level 按钮
        document.querySelectorAll("#soundLevels button").forEach(b => {
            b.addEventListener("click", () => {
                const l = b.dataset.level;
                if (soundLevels.has(l)) {
                    soundLevels.delete(l);
                    b.classList.remove("active");
                } else {
                    soundLevels.add(l);
                    b.classList.add("active");
                }
                saveSettings();
            });
        });

        $("soundEnabled").addEventListener("change", saveSettings);
    }

    function updateSoundLevelBtns() {
        document.querySelectorAll("#soundLevels button").forEach(b => {
            b.classList.toggle("active", soundLevels.has(b.dataset.level));
        });
    }

    function updateFilterLevelBtns() {
        document.querySelectorAll("#filterLevels button").forEach(b => {
            b.classList.toggle("active", selectedLevels.has(b.dataset.level));
        });
    }

    function setupTabs() {
        document.querySelectorAll(".tab").forEach(t => {
            t.addEventListener("click", () => {
                document.querySelectorAll(".tab").forEach(x => x.classList.remove("active"));
                t.classList.add("active");
                currentView = t.dataset.view;
                updateView();
            });
        });
    }

    function setupLanguage() {
        const langSelect = $("langSelect");
        if (!langSelect) return;
        langSelect.onchange = () => setLanguage(langSelect.value);
    }

    function setupFilterToggle() {
        const btn = $("filterToggle");
        if (!btn) return;
        btn.onclick = () => {
            setFilterCollapsed(!filterCollapsed);
        };
    }

    function setupFilters() {
        const updateRankingSortControls = () => {
            const orderEl = $("rankingSortOrder");
            if (orderEl) {
                orderEl.disabled = rankingSort === 'default';
            }
        };
        // Symbol 搜索（防抖，仅前端过滤）
        const debouncedFilter = debounce(() => {
            applyFilters();
            applyPatternFilters();
            updateView();
            // 宽屏模式下同步更新成交额和成交笔数面板
            if (isWidescreen) {
                updateWideVolumePanelLocal();
                updateWideTradesPanelLocal();
            }
        }, 200);

        $("symbol").oninput = debouncedFilter;
        $("period").onchange = () => { saveSettings(); applyFilters(); updateView(); };
        $("direction").onchange = () => { saveSettings(); applyFilters(); updateView(); };
        $("minDiff").oninput = debounce(() => {
            saveSettings();
            applyFilters();
            updateView();
        }, 300);

        // 成交额过滤
        const debouncedVolumeFilter = debounce(() => {
            saveSettings();
            applyFilters();
            updateView();
        }, 300);
        $("minVolume").oninput = debouncedVolumeFilter;
        $("volumeUnit").onchange = debouncedVolumeFilter;

        // Limit 变更需要重新请求后端
        $("limit").onchange = () => {
            saveSettings();
            loadHistory();
        };

        // Pattern 筛选器
        $("patternType").onchange = () => { applyPatternFilters(); updateView(); };
        $("efficiencyRank").onchange = () => { applyPatternFilters(); updateView(); };
        $("patternDirection").onchange = () => { applyPatternFilters(); updateView(); };

        // Ranking 筛选器
        const rankingCompareEl = $("rankingCompare");
        const rankingRefreshEl = $("rankingRefresh");
        const rankingSortEl = $("rankingSort");
        const rankingSortOrderEl = $("rankingSortOrder");
        if (rankingSortEl) rankingSortEl.value = rankingSort;
        if (rankingSortOrderEl) rankingSortOrderEl.value = rankingSortOrder;
        if (rankingCompareEl) {
            rankingCompareEl.onchange = () => {
                rankingCompare = rankingCompareEl.value;
                if (currentView === 'volume' || currentView === 'trades') {
                    loadRanking(currentView).then(() => updateRankingList());
                }
            };
        }
        if (rankingSortEl) {
            rankingSortEl.onchange = () => {
                rankingSort = rankingSortEl.value || 'default';
                updateRankingSortControls();
                if (currentView === 'volume' || currentView === 'trades') {
                    updateRankingList();
                }
            };
        }
        if (rankingSortOrderEl) {
            rankingSortOrderEl.onchange = () => {
                rankingSortOrder = rankingSortOrderEl.value || 'asc';
                if (currentView === 'volume' || currentView === 'trades') {
                    updateRankingList();
                }
            };
        }
        if (rankingRefreshEl) {
            rankingRefreshEl.onclick = () => {
                if (currentView === 'volume' || currentView === 'trades') {
                    loadRanking(currentView).then(() => updateRankingList());
                }
            };
        }

        updateRankingSortControls();

        // 宽屏面板过滤控件事件绑定
        setupWidePanelFilters();
    }

    // 设置宽屏面板过滤控件
    function setupWidePanelFilters() {
        // 成交额面板过滤控件
        const wideVolumeCompareEl = $("wideVolumeCompare");
        const wideVolumeSortEl = $("wideVolumeSort");
        const wideVolumeSortOrderEl = $("wideVolumeSortOrder");

        if (wideVolumeCompareEl) {
            wideVolumeCompareEl.onchange = () => {
                wideVolumeCompare = wideVolumeCompareEl.value;
                updateWideVolumePanel();
            };
        }
        if (wideVolumeSortEl) {
            wideVolumeSortEl.onchange = () => {
                wideVolumeSort = wideVolumeSortEl.value || 'default';
                updateWideVolumePanelLocal();
            };
        }
        if (wideVolumeSortOrderEl) {
            wideVolumeSortOrderEl.onchange = () => {
                wideVolumeSortOrder = wideVolumeSortOrderEl.value || 'asc';
                updateWideVolumePanelLocal();
            };
        }

        // 成交笔数面板过滤控件
        const wideTradesCompareEl = $("wideTradesCompare");
        const wideTradesSortEl = $("wideTradesSort");
        const wideTradesSortOrderEl = $("wideTradesSortOrder");

        if (wideTradesCompareEl) {
            wideTradesCompareEl.onchange = () => {
                wideTradesCompare = wideTradesCompareEl.value;
                updateWideTradesPanel();
            };
        }
        if (wideTradesSortEl) {
            wideTradesSortEl.onchange = () => {
                wideTradesSort = wideTradesSortEl.value || 'default';
                updateWideTradesPanelLocal();
            };
        }
        if (wideTradesSortOrderEl) {
            wideTradesSortOrderEl.onchange = () => {
                wideTradesSortOrder = wideTradesSortOrderEl.value || 'asc';
                updateWideTradesPanelLocal();
            };
        }
    }

    // 仅本地排序更新（不重新加载数据）
    function updateWideVolumePanelLocal() {
        if (!wideVolumeCluster) return;
        const items = rankingData.volume || [];
        const filtered = filterRankingItems(items);
        const sorted = sortRankingItems(filtered, 'volume', wideVolumeSort, wideVolumeSortOrder);
        const rows = sorted.map((item, i) => renderRankingItemFromAPI(item, i, 'volume'));
        wideVolumeCluster.update(rows);
    }

    function updateWideTradesPanelLocal() {
        if (!wideTradesCluster) return;
        const items = rankingData.trades || [];
        const filtered = filterRankingItems(items);
        const sorted = sortRankingItems(filtered, 'trades', wideTradesSort, wideTradesSortOrder);
        const rows = sorted.map((item, i) => renderRankingItemFromAPI(item, i, 'trades'));
        wideTradesCluster.update(rows);
    }

    function setupPatternModal() {
        // 点击背景关闭弹窗
        $("patternModal").onclick = (e) => {
            if (e.target === $("patternModal")) {
                hidePatternModal();
            }
        };

        // Trade 按钮
        $("modalTradeBtn").onclick = () => {
            if (!currentPatternSignal) return;
            jumpToTrade(currentPatternSignal.symbol);
            hidePatternModal();
        };

        // Kline 按钮
        const klineBtn = $("modalKlineBtn");
        if (klineBtn) {
            klineBtn.onclick = () => {
                if (!currentPatternSignal) return;
                showKlineModal(currentPatternSignal.symbol);
                hidePatternModal();
            };
        }

        // Filter 按钮
        $("modalFilterBtn").onclick = () => {
            if (!currentPatternSignal) return;
            $("symbol").value = "$" + currentPatternSignal.symbol;
            applyPatternFilters();
            updateView();
            hidePatternModal();
            showToast(t("toast_filtered", { symbol: currentPatternSignal.symbol }));
        };
    }

    // ==================== 跳转交易 ====================
    function jumpToTrade(symbol) {
        if (!symbol) return;
        if (window.parent !== window) {
            window.parent.postMessage({ type: "jump_symbol", symbol }, "*");
        } else {
            window.open("https://www.binance.com/futures/" + symbol, "_blank");
        }
    }

    // ==================== 前端 Pattern 关联 ====================
    // 在前端查找与 signal 相关的 pattern（60分钟窗口内，前后都查）
    function findRelatedPattern(symbol, signalTime) {
        if (!masterPatterns || masterPatterns.length === 0) return null;

        const windowMs = 60 * 60 * 1000; // 60 minutes
        const signalTs = signalTime.getTime();

        let closest = null;
        let closestDiff = Infinity;

        for (const pat of masterPatterns) {
            if (pat.symbol !== symbol) continue;

            const patTs = new Date(pat.detected_at).getTime();
            const diff = Math.abs(signalTs - patTs);

            if (diff <= windowMs && diff < closestDiff) {
                closestDiff = diff;
                closest = pat;
            }
        }

        if (!closest) return null;

        // 构建 related_pattern 对象
        const pivotUp = true; // 需要从 signal 获取，这里简化处理
        const patternBullish = closest.direction === 'bullish';
        let correlation = 'moderate';
        if (closest.direction !== 'neutral') {
            correlation = (pivotUp && patternBullish) || (!pivotUp && !patternBullish) ? 'strong' : 'weak';
        }

        // 计算时间差
        const diffMs = signalTs - new Date(closest.detected_at).getTime();
        const diffMin = Math.round(Math.abs(diffMs) / 60000);
        const timeDiff = diffMs >= 0
            ? t("time_minutes_ago", { m: diffMin })
            : t("time_in_minutes", { m: diffMin });

        return {
            id: closest.id,
            pattern: closest.pattern,
            pattern_cn: closest.pattern_cn,
            direction: closest.direction,
            confidence: closest.confidence,
            up_percent: closest.up_percent,
            down_percent: closest.down_percent,
            efficiency_rank: closest.efficiency_rank,
            correlation: correlation,
            detected_at: closest.detected_at,
            count: 1,
            time_diff: timeDiff
        };
    }

    // ==================== 数据加载 ====================

    // 获取枢轴点数据（带缓存）
    async function getPivotData(symbol) {
        const now = Date.now();
        const cached = pivotCache.get(symbol);
        if (cached && (now - cached.timestamp) < PIVOT_CACHE_TTL) {
            return cached.data;
        }

        try {
            const r = await fetch(`/api/pivots/${symbol}`);
            if (!r.ok) return null;
            const data = await r.json();
            pivotCache.set(symbol, { data, timestamp: now });
            return data;
        } catch (_) {
            return null;
        }
    }

    // 查找当前价格上下方最近的点位
    function findNearestLevels(pivotData, currentPrice, period) {
        if (!pivotData || !currentPrice) return { above: null, below: null };

        const levels = period === '1w' ? pivotData.weekly : pivotData.daily;
        if (!levels) return { above: null, below: null };

        const allLevels = [
            { name: 'R5', price: levels.r5 },
            { name: 'R4', price: levels.r4 },
            { name: 'R3', price: levels.r3 },
            { name: 'R2', price: levels.r2 },
            { name: 'R1', price: levels.r1 },
            { name: 'PP', price: levels.pp },
            { name: 'S1', price: levels.s1 },
            { name: 'S2', price: levels.s2 },
            { name: 'S3', price: levels.s3 },
            { name: 'S4', price: levels.s4 },
            { name: 'S5', price: levels.s5 },
        ].filter(l => l.price > 0);

        let above = null, below = null;
        let aboveDiff = Infinity, belowDiff = Infinity;

        for (const level of allLevels) {
            const diff = level.price - currentPrice;
            if (diff > 0 && diff < aboveDiff) {
                aboveDiff = diff;
                above = level;
            } else if (diff < 0 && Math.abs(diff) < belowDiff) {
                belowDiff = Math.abs(diff);
                below = level;
            }
        }

        return { above, below };
    }

    // 格式化枢轴点位显示
    function formatPivotLevel(level, currentPrice, period) {
        if (!level || !currentPrice) return '';
        const pct = ((level.price - currentPrice) / currentPrice * 100);
        const pctStr = (pct >= 0 ? '+' : '') + pct.toFixed(1) + '%';
        const colorClass = pct >= 0 ? 'pivot-up' : 'pivot-down';
        const periodLabel = period === '1w' ? t("period_week_short") : t("period_day_short");
        return `<span class="${colorClass}">${periodLabel}:${level.name}(${pctStr})</span>`;
    }

    async function loadHistory() {
        const limit = $("limit").value || 1000;
        $("hint").textContent = t("hint_loading");

        try {
            const r = await fetch(`/api/history?limit=${limit}`);
            if (!r.ok) throw new Error(t("error_http", { status: r.status }));

            const data = await r.json();
            masterSignals = (data || []).sort((a, b) =>
                (new Date(b.triggered_at) || 0) - (new Date(a.triggered_at) || 0)
            );

            // 前端补充关联 pattern（如果后端没有关联）
            for (const sig of masterSignals) {
                if (!sig.related_pattern && masterPatterns.length > 0) {
                    const related = findRelatedPattern(sig.symbol, new Date(sig.triggered_at));
                    if (related) {
                        sig.related_pattern = related;
                    }
                }
            }

            applyFilters();
            updateView();
        } catch (e) {
            $("hint").textContent = t("hint_load_failed", { error: e });
        }
    }

    async function loadPivotStatus() {
        try {
            const r = await fetch("/api/pivot-status");
            if (!r.ok) return;
            const d = await r.json();

            const fmt = p => p ? (
                (p.is_stale ? `<span class="pill stale">${t("status_stale")}</span>` : `<span class="pill fresh">${t("status_ok")}</span>`) +
                " " + fmtDur(p.seconds_until) + " (" + t("label_symbols", { count: p.symbol_count }) + ")"
            ) : "-";

            $("dailyStatus").innerHTML = fmt(d.daily);
            $("weeklyStatus").innerHTML = fmt(d.weekly);
        } catch (_) { }
    }

    async function loadRuntimeStats() {
        try {
            const r = await fetch("/api/runtime");
            if (!r.ok) return;
            const d = await r.json();

            // 头部简要信息
            $("rtGoroutines").textContent = d.goroutines || '-';
            $("rtKlines").textContent = d.kline_symbols || '-';
            // 显示 "交易对/信号数" 格式
            const symbols = d.symbols || 0;
            const signals = d.signals || 0;
            $("rtSignals").textContent = symbols + '/' + signals;
            $("rtHeap").textContent = d.heap_mb ? d.heap_mb.toFixed(0) : '-';

            // 底部详细信息
            const ftSSE = $("ftSSE");
            const ftGoroutines = $("ftGoroutines");
            const ftHeap = $("ftHeap");
            const ftSymbols = $("ftSymbols");
            const ftSignals = $("ftSignals");
            const ftUptime = $("ftUptime");
            const ftVersion = $("ftVersion");

            if (ftSSE) ftSSE.textContent = d.sse_subscribers || 0;
            if (ftGoroutines) ftGoroutines.textContent = d.goroutines || '-';
            if (ftHeap) ftHeap.textContent = d.heap_mb ? d.heap_mb.toFixed(1) + 'MB' : '-';
            if (ftSymbols) ftSymbols.textContent = symbols;
            if (ftSignals) ftSignals.textContent = signals;
            if (ftUptime) ftUptime.textContent = d.uptime || '-';
            if (ftVersion) ftVersion.textContent = d.version ? 'v' + d.version : '';
        } catch (_) { }
    }

    async function loadTickers() {
        try {
            const r = await fetch("/api/tickers");
            if (!r.ok) return;
            const data = await r.json();
            if (data && typeof data === "object") {
                for (const [symbol, ticker] of Object.entries(data)) {
                    tickerData.set(symbol, ticker);
                }
            }
        } catch (_) { }
    }

    async function loadPatterns() {
        try {
            const r = await fetch("/api/patterns?limit=500");
            if (!r.ok) return;
            const data = await r.json();
            masterPatterns = (data || []).sort((a, b) =>
                (new Date(b.detected_at) || 0) - (new Date(a.detected_at) || 0)
            );
            if (currentView === 'patterns') {
                applyPatternFilters();
                updatePatternList();
            }
        } catch (_) { }
    }

    // Load ranking data from backend API
    async function loadRanking(type, customCompare) {
        try {
            const compare = customCompare !== undefined ? customCompare : rankingCompare;
            const compareParam = compare ? `&compare=${compare}` : '';
            const r = await fetch(`/api/ranking/current?type=${type}${compareParam}`);
            if (!r.ok) return;
            const data = await r.json();
            if (data && data.items) {
                rankingData[type] = data.items;
                // Update hint with timestamp info (only for narrow screen mode)
                if (!isWidescreen && data.timestamp) {
                    const ts = new Date(data.timestamp);
                    const compareTs = data.compare_to ? new Date(data.compare_to) : null;
                    const typeLabel = rankingTypeLabel(type);
                    const hint = compareTs
                        ? t("hint_ranking", { type: typeLabel, count: data.items.length, time: fmtRelTime(compareTs) })
                        : t("hint_ranking_simple", { type: typeLabel, count: data.items.length });
                    $("hint").textContent = hint;
                }
            }
        } catch (_) { }
    }

    // Load ranking history for a specific symbol
    async function loadRankingHistory(symbol) {
        try {
            const r = await fetch(`/api/ranking/history/${symbol}`);
            if (!r.ok) return null;
            return await r.json();
        } catch (_) {
            return null;
        }
    }

    // ==================== SSE 连接 ====================
    let tickerUpdatePending = false;
    let sseReconnectTimer = null;
    let sseReconnectDelay = 1000;
    const SSE_MAX_DELAY = 30000;

    function connectSSE() {
        // 清除之前的重连定时器
        if (sseReconnectTimer) {
            clearTimeout(sseReconnectTimer);
            sseReconnectTimer = null;
        }

        let es;
        try {
            es = new EventSource("/api/sse");
        } catch (_) {
            setStatus("disconnected");
            scheduleReconnect();
            return;
        }

        setStatus("reconnecting");

        es.onopen = () => {
            setStatus("connected");
            sseReconnectDelay = 1000; // 重置重连延迟
        };

        es.onerror = () => {
            const state = es.readyState;
            if (state === EventSource.CLOSED) {
                setStatus("disconnected");
                es.close();
                scheduleReconnect();
            } else {
                setStatus("reconnecting");
            }
        };

        // 新信号
        es.addEventListener("signal", e => {
            try {
                const signal = JSON.parse(e.data);
                let deferUpdate = false;

                // 声音提醒
                if (soundLevels.has(signal.level)) playBeep();

                // 合并到主数据
                if (signal && signal.id) {
                    const exists = masterSignals.findIndex(s => s.id === signal.id);
                    if (exists === -1) {
                        // 尝试关联最近的 pattern（前端关联）
                        if (!signal.related_pattern) {
                            const relatedPattern = findRelatedPattern(signal.symbol, new Date(signal.triggered_at));
                            if (relatedPattern) {
                                signal.related_pattern = relatedPattern;
                            }
                        }
                        masterSignals.unshift(signal);
                        // 保持数据量限制
                        const limit = parseInt($("limit").value) || 1000;
                        if (masterSignals.length > limit * 1.2) {
                            masterSignals = masterSignals.slice(0, limit);
                        }

                        if (shouldDeferSignalUpdate()) {
                            pendingSignalCount += 1;
                            updateNewSignalBanner();
                            deferUpdate = true;
                        }
                    }
                }

                if (deferUpdate) return;

                // 重新过滤并更新视图
                applyFilters();
                if (isWidescreen) {
                    updateWideSignalPanel();
                } else if (currentView === 'signals') {
                    updateSignalList();
                }
            } catch (_) { }
        });

        // 形态信号
        es.addEventListener("pattern", e => {
            try {
                const pattern = JSON.parse(e.data);

                // 合并到主数据
                if (pattern && pattern.id) {
                    const exists = masterPatterns.findIndex(p => p.id === pattern.id);
                    if (exists === -1) {
                        masterPatterns.unshift(pattern);
                        // 保持数据量限制
                        if (masterPatterns.length > 1000) {
                            masterPatterns = masterPatterns.slice(0, 1000);
                        }
                    }
                }

                // 重新过滤并更新视图
                if (currentView === 'patterns') {
                    applyPatternFilters();
                    updatePatternList();
                }
            } catch (_) { }
        });

        // Ticker 更新（节流）
        es.addEventListener("ticker", e => {
            try {
                const batch = JSON.parse(e.data);
                if (batch && batch.tickers) {
                    // 更新数据层
                    for (const [symbol, ticker] of Object.entries(batch.tickers)) {
                        tickerData.set(symbol, ticker);
                    }

                    // 节流更新 DOM
                    if (!tickerUpdatePending) {
                        tickerUpdatePending = true;
                        requestAnimationFrame(() => {
                            updateVisibleItems();
                            tickerUpdatePending = false;
                        });
                    }
                }
            } catch (_) { }
        });
    }

    function scheduleReconnect() {
        if (sseReconnectTimer) return;

        sseReconnectTimer = setTimeout(() => {
            sseReconnectTimer = null;
            connectSSE();
        }, sseReconnectDelay);

        // 指数退避，最大 30 秒
        sseReconnectDelay = Math.min(sseReconnectDelay * 1.5, SSE_MAX_DELAY);
    }

    // 只更新可视区域的 DOM
    function updateVisibleItems() {
        // 宽屏模式下更新宽屏信号面板
        if (isWidescreen) {
            document.querySelectorAll("#wideSignalList .item[data-symbol]").forEach(item => {
                updateSignalItemTicker(item);
            });
            // 更新枢轴点预览 Modal（如果打开）
            updatePivotPreviewIfOpen();
            updatePositionModalIfOpen();
            return;
        }

        if (currentView === 'signals') {
            // 更新可视的信号项
            document.querySelectorAll("#signalList .item[data-symbol]").forEach(item => {
                updateSignalItemTicker(item);
            });

            // 更新枢轴点预览 Modal（如果打开）
            updatePivotPreviewIfOpen();
            updatePositionModalIfOpen();
        } else {
            // 排行榜视图：重新计算并更新
            updateRankingList();
        }
    }

    // 更新单个信号项的 ticker 信息
    function updateSignalItemTicker(item) {
        const symbol = item.dataset.symbol;
        const ticker = tickerData.get(symbol);
        if (!ticker) return;

        const signalPrice = parseFloat(item.dataset.price) || 0;

        // 更新价格差异
        const subDiv = item.querySelector(".sub > div:first-child");
        if (subDiv && signalPrice > 0) {
            const diff = ticker.last_price - signalPrice;
            const diffPct = (diff / signalPrice) * 100;
            const diffSign = diff >= 0 ? '+' : '';
            const diffClass = diff >= 0 ? 'up' : 'down';
            subDiv.innerHTML = `${t("label_signal")}: ${fmtPrice(signalPrice)} <span class="price-diff ${diffClass}">${diffSign}${diffPct.toFixed(2)}%</span>`;
        }

        // 更新 ticker 信息
        let priceInfoEl = item.querySelector(".price-info");
        if (!priceInfoEl) {
            priceInfoEl = document.createElement("div");
            priceInfoEl.className = "price-info";
            item.appendChild(priceInfoEl);
        }

        const pctClass = ticker.price_percent >= 0 ? 'up' : 'down';
        priceInfoEl.innerHTML = `
            <span class="price-now">${fmtPrice(ticker.last_price)}</span>
            <span class="price-pct ${pctClass}">${fmtPct(ticker.price_percent)}</span>
            <span class="volume">${fmtVolume(ticker.quote_volume)}</span>
            <span class="trades">${fmtTradeCount(ticker.trade_count)} ${t("label_trades_unit")}</span>
        `;

        // 更新枢轴点位（仅可视项，使用同步版本）
        if (visibleSymbols.has(symbol)) {
            updateItemPivotLevelsSync(item, symbol);
        }
    }

    // 更新枢轴点预览 Modal（实时价格更新）
    function updatePivotPreviewIfOpen() {
        const modal = $("pivotModal");
        if (!modal || modal.style.display === 'none' || !currentPivotPreviewSymbol) return;

        const cached = pivotCache.get(currentPivotPreviewSymbol);
        if (cached && cached.data) {
            renderPivotPreview(cached.data, currentPivotPreviewPeriod);
        }
    }

    // 更新相对时间
    function updateRelTimes() {
        document.querySelectorAll(".time-rel").forEach(el => {
            const item = el.closest(".item");
            if (item && item.dataset.time) {
                el.textContent = fmtRelTime(item.dataset.time);
            } else if (el.dataset.time) {
                el.textContent = fmtRelTime(el.dataset.time);
            }
        });
    }

    // ==================== 声音 ====================
    const playBeep = () => {
        if (!$("soundEnabled").checked) return;
        try {
            const c = new (window.AudioContext || window.webkitAudioContext)();
            const o = c.createOscillator();
            const g = c.createGain();
            o.type = "sine";
            o.frequency.value = 880;
            g.gain.value = 0.08;
            o.connect(g);
            g.connect(c.destination);
            o.start();
            o.stop(c.currentTime + 0.15);
            setTimeout(() => c.close(), 500);
        } catch (_) { }
    };

    // ==================== 初始化 ====================
    // iOS PWA 真实视口高度设置
    let initialVh = 0; // 记录初始视口高度

    function setVh() {
        const vh = (window.visualViewport?.height || window.innerHeight) * 0.01;
        // 首次设置时记录初始高度
        if (initialVh === 0) {
            initialVh = vh;
        }
        document.documentElement.style.setProperty('--vh', `${vh}px`);
    }

    function calcScrollHeight() {
        const headerArea = document.querySelector('.header-area');
        const footerStats = $("footerStats");
        if (!headerArea) return;

        const headerHeight = headerArea.offsetHeight;
        const footerHeight = footerStats ? footerStats.offsetHeight : 40;

        // 使用 visualViewport 获取真实可视高度（iOS PWA 兼容）
        const viewportHeight = window.visualViewport
            ? window.visualViewport.height
            : window.innerHeight;

        // 考虑底部栏高度
        const availableHeight = Math.max(200, viewportHeight - headerHeight - footerHeight - 8);

        $("signalScroll").style.height = availableHeight + 'px';
        $("patternScroll").style.height = availableHeight + 'px';
        $("rankingScroll").style.height = availableHeight + 'px';

        // 通知 Clusterize 重新计算
        if (signalCluster) signalCluster.refresh();
        if (patternCluster) patternCluster.refresh();
        if (rankingCluster) rankingCluster.refresh();
    }

    // iOS 键盘收起后恢复布局
    function handleViewportResize() {
        const currentVh = (window.visualViewport?.height || window.innerHeight) * 0.01;

        // 如果视口高度恢复到接近初始值，说明键盘收起了
        if (initialVh > 0 && currentVh >= initialVh * 0.95) {
            // 使用初始高度，避免键盘影响
            document.documentElement.style.setProperty('--vh', `${initialVh}px`);
        } else {
            setVh();
        }

        calcScrollHeight();
    }

    function init() {
        // 立即设置 --vh 变量（iOS PWA 关键）
        setVh();

        applyI18n();
        loadSettings();
        updateSoundLevelBtns();
        updateFilterLevelBtns();
        setupLevelBtns();
        setupTabs();
        setupLanguage();
        setupFilterToggle();
        setupFilters();
        setupActionMenu();
        setupPatternModal();
        setupRankingModal();
        setupPositionModal();
        setupKlineModal();
        initClusterize();
        setupSignalNewBanner();
        setupSignalScrollTracking();

        // 宽屏模式检测
        checkWidescreenMode();
        window.addEventListener('resize', debounce(checkWidescreenMode, 150));

        // 延迟计算高度，确保 DOM 已渲染
        requestAnimationFrame(() => {
            calcScrollHeight();
        });
        window.addEventListener('resize', debounce(calcScrollHeight, 100));

        // iOS PWA 模式下监听 visualViewport 变化
        if (window.visualViewport) {
            window.visualViewport.addEventListener('resize', debounce(handleViewportResize, 150));
        }

        // 监听输入框 blur 事件，键盘收起后强制恢复布局
        document.addEventListener('focusout', (e) => {
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') {
                // 延迟执行，等待键盘完全收起
                setTimeout(() => {
                    if (initialVh > 0) {
                        document.documentElement.style.setProperty('--vh', `${initialVh}px`);
                    }
                    calcScrollHeight();
                }, 300);
            }
        });

        // Runtime stats 点击刷新
        const runtimeStats = $("runtimeStats");
        if (runtimeStats) {
            runtimeStats.onclick = loadRuntimeStats;
        }

        // Refresh 按钮
        $("refresh").onclick = () => {
            loadHistory();
            loadPatterns();
            loadPivotStatus();
            loadTickers();
            loadRuntimeStats();
        };

        // 初始加载 - 先加载 patterns，再加载 history（确保后端能关联 pattern）
        loadTickers().then(() => {
            loadPatterns().then(() => {
                loadHistory();
            });
            loadPivotStatus();
            loadRuntimeStats();
            connectSSE();
        });

        checkBinanceAvailability();

        // 定时任务
        setInterval(loadPivotStatus, 60000);
        setInterval(loadRuntimeStats, 5000); // 每5秒刷新运行时状态
        setInterval(updateRelTimes, 10000);
    }

    // 启动
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
