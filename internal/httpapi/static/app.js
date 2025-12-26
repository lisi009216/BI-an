// Pivot Monitor Frontend
(function () {
    'use strict';

    const $ = id => document.getElementById(id);

    // 格式化函数
    const fmtRelTime = v => {
        try {
            const d = new Date(v);
            if (isNaN(d)) return String(v);
            const now = Date.now(), diff = Math.floor((now - d.getTime()) / 1000);
            if (diff < 0) return "just now";
            if (diff < 60) return diff + "s ago";
            if (diff < 3600) return Math.floor(diff / 60) + "m ago";
            if (diff < 28800) return Math.floor(diff / 3600) + "h " + Math.floor((diff % 3600) / 60) + "m ago";
            return d.toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", hour12: false });
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

    const fmtDur = s => {
        if (s < 0) return "now";
        const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60);
        return h > 0 ? h + "h " + m + "m" : m + "m";
    };

    // 状态管理
    let selectedLevels = new Set();
    let soundLevels = new Set(["R4", "R5", "S4", "S5"]);
    let allSignals = [];
    let tickerData = new Map();
    let currentView = 'signals';
    let menuSymbol = null;
    let symbolRanking = { volume: new Map(), trades: new Map() }; // symbol -> rank

    const STORAGE_KEY_SOUND_LEVELS = "pivot_sound_levels";
    const STORAGE_KEY_SOUND_ENABLED = "pivot_sound_enabled";

    const setStatus = s => {
        const e = $("status");
        e.textContent = s || "unknown";
        e.classList.remove("connected", "reconnecting", "disconnected");
        if (s) e.classList.add(s);
    };

    const showToast = (msg, duration = 2000) => {
        const t = $("toast");
        t.textContent = msg;
        t.classList.add("show");
        setTimeout(() => t.classList.remove("show"), duration);
    };

    function loadSoundSettings() {
        try {
            const saved = localStorage.getItem(STORAGE_KEY_SOUND_LEVELS);
            if (saved) soundLevels = new Set(JSON.parse(saved));
            const enabled = localStorage.getItem(STORAGE_KEY_SOUND_ENABLED);
            if (enabled !== null) $("soundEnabled").checked = enabled === "true";
        } catch (_) { }
    }

    function saveSoundSettings() {
        try {
            localStorage.setItem(STORAGE_KEY_SOUND_LEVELS, JSON.stringify(Array.from(soundLevels)));
            localStorage.setItem(STORAGE_KEY_SOUND_ENABLED, $("soundEnabled").checked);
        } catch (_) { }
    }

    function updateSoundLevelBtns() {
        document.querySelectorAll("#soundLevels button").forEach(b => {
            b.classList.toggle("active", soundLevels.has(b.dataset.level));
        });
    }

    function setupLevelBtns() {
        document.querySelectorAll("#filterLevels button").forEach(b => {
            b.addEventListener("click", () => {
                const l = b.dataset.level;
                if (selectedLevels.has(l)) { selectedLevels.delete(l); b.classList.remove("active"); }
                else { selectedLevels.add(l); b.classList.add("active"); }
                loadHistory();
            });
        });
        document.querySelectorAll("#soundLevels button").forEach(b => {
            b.addEventListener("click", () => {
                const l = b.dataset.level;
                if (soundLevels.has(l)) { soundLevels.delete(l); b.classList.remove("active"); }
                else { soundLevels.add(l); b.classList.add("active"); }
                saveSoundSettings();
            });
        });
        $("soundEnabled").addEventListener("change", saveSoundSettings);
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

    function updateView() {
        const list = $("list");
        const ranking = $("ranking");
        if (currentView === 'signals') {
            list.style.display = '';
            ranking.classList.remove("show");
            render(allSignals);
        } else {
            list.style.display = 'none';
            ranking.classList.add("show");
            renderRanking();
        }
    }

    const filters = () => ({
        symbol: $("symbol").value.trim(),
        period: $("period").value,
        levels: Array.from(selectedLevels),
        direction: $("direction").value,
        limit: 400
    });

    const matchF = (s, f) => {
        if (!s) return false;
        if (f.symbol && !String(s.symbol || "").toUpperCase().includes(f.symbol.toUpperCase())) return false;
        if (f.period && s.period !== f.period) return false;
        if (f.levels.length && !f.levels.includes(s.level)) return false;
        if (f.direction && s.direction !== f.direction) return false;
        return true;
    };

    const buildQ = f => {
        const q = new URLSearchParams();
        if (f.symbol) q.set("symbol", f.symbol);
        if (f.period) q.set("period", f.period);
        f.levels.forEach(l => q.append("level", l));
        if (f.direction) q.set("direction", f.direction);
        q.set("limit", f.limit);
        return q.toString();
    };

    const sortD = l => l.sort((a, b) => (new Date(b.triggered_at) || 0) - (new Date(a.triggered_at) || 0));

    const merge = (b, i) => {
        const m = new Map();
        (b || []).forEach(s => s && s.id && m.set(s.id, s));
        (i || []).forEach(s => s && s.id && m.set(s.id, s));
        const o = Array.from(m.values());
        sortD(o);
        return o;
    };

    // 计算排行榜（基于所有信号中的交易对）
    function computeRanking() {
        const signalSymbols = new Set(allSignals.map(s => s.symbol));
        const items = [];
        for (const symbol of signalSymbols) {
            const ticker = tickerData.get(symbol);
            if (ticker) {
                items.push({ symbol, volume: ticker.quote_volume || 0, trades: ticker.trade_count || 0 });
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
    }

    // 渲染信号列表（包含排行信息）
    function render(signals) {
        const e = $("list");
        e.innerHTML = "";

        if (!signals || !signals.length) {
            e.innerHTML = '<div class="item" style="cursor:default">No signals</div>';
            return;
        }

        // 先计算排行
        computeRanking();

        const frag = document.createDocumentFragment();
        signals.forEach(s => {
            const ticker = tickerData.get(s.symbol);
            const item = document.createElement("div");
            item.className = "item";
            item.dataset.time = s.triggered_at;
            item.dataset.symbol = s.symbol;
            item.dataset.price = s.price;

            // 排行信息
            const volRank = symbolRanking.volume.get(s.symbol);
            const tradeRank = symbolRanking.trades.get(s.symbol);
            let rankHtml = '';
            if (volRank || tradeRank) {
                const vr = volRank ? `<span class="rank-badge vol" title="Volume Rank">#${volRank}V</span>` : '';
                const tr = tradeRank ? `<span class="rank-badge trd" title="Trades Rank">#${tradeRank}T</span>` : '';
                rankHtml = `<div class="ranks">${vr}${tr}</div>`;
            }

            // 价格差异（相对信号）
            let diffHtml = '';
            let tickerHtml = '';
            if (ticker) {
                const diff = ticker.last_price - s.price;
                const diffPct = s.price > 0 ? ((diff / s.price) * 100) : 0;
                const diffSign = diff >= 0 ? '+' : '';
                const diffClass = diff >= 0 ? 'up' : 'down';
                diffHtml = `<span class="price-diff ${diffClass}">${diffSign}${diffPct.toFixed(2)}%</span>`;

                const pctClass = ticker.price_percent >= 0 ? 'up' : 'down';
                tickerHtml = `
          <div class="price-info">
            <span class="price-now">${fmtPrice(ticker.last_price)}</span>
            <span class="price-pct ${pctClass}">${fmtPct(ticker.price_percent)}</span>
            <span class="volume">${fmtVolume(ticker.quote_volume)}</span>
            <span class="trades">${fmtTradeCount(ticker.trade_count)} trades</span>
          </div>
        `;
            }

            item.innerHTML = `
        <div class="top">
          <div class="sym">${s.symbol} ${rankHtml}</div>
          <div class="tags">
            <span class="tag">${s.period}</span>
            <span class="tag">${s.level}</span>
            <span class="tag ${s.direction}">${s.direction}</span>
          </div>
        </div>
        <div class="sub">
          <div>Signal: ${fmtPrice(s.price)} ${diffHtml}</div>
          <div class="muted time-rel">${fmtRelTime(s.triggered_at)}</div>
        </div>
        ${tickerHtml}
      `;

            item.onclick = (e) => { e.preventDefault(); e.stopPropagation(); showActionMenu(e, s.symbol); };
            frag.appendChild(item);
        });
        e.appendChild(frag);
    }

    function showActionMenu(e, symbol) {
        menuSymbol = symbol;
        const menu = $("actionMenu");
        menu.classList.add("show");
        const x = Math.min(e.clientX, window.innerWidth - 180);
        const y = Math.min(e.clientY, window.innerHeight - 150);
        menu.style.left = x + "px";
        menu.style.top = y + "px";
    }

    function hideActionMenu() {
        $("actionMenu").classList.remove("show");
        menuSymbol = null;
    }

    // 复制到剪贴板（兼容多种环境）
    function copyToClipboard(text) {
        // 优先使用 Clipboard API
        if (navigator.clipboard && window.isSecureContext) {
            return navigator.clipboard.writeText(text);
        }
        // 备用方案：使用 execCommand
        return new Promise((resolve, reject) => {
            const textarea = document.createElement('textarea');
            textarea.value = text;
            textarea.style.position = 'fixed';
            textarea.style.left = '-9999px';
            textarea.style.top = '-9999px';
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
                        if (window.parent !== window) {
                            window.parent.postMessage({ type: "jump_symbol", symbol: menuSymbol }, "*");
                        } else {
                            window.open("https://www.binance.com/futures/" + menuSymbol, "_blank");
                        }
                        break;
                    case "copy":
                        copyToClipboard(menuSymbol)
                            .then(() => showToast("Copied: " + menuSymbol))
                            .catch(() => showToast("Copy failed"));
                        break;
                    case "filter":
                        if ($("symbol").value.trim().toUpperCase() === menuSymbol) {
                            $("symbol").value = "";
                            loadHistory();
                            showToast("Filter cleared");
                        } else {
                            $("symbol").value = menuSymbol;
                            loadHistory();
                            showToast("Filtered: " + menuSymbol);
                        }
                        break;
                }
                hideActionMenu();
            });
        });
    }

    // 排行榜渲染（不限制数量）
    function renderRanking() {
        const type = currentView;
        computeRanking();

        const signalSymbols = new Set(allSignals.map(s => s.symbol));
        if (signalSymbols.size === 0) {
            $("rankingTitle").textContent = type === 'volume' ? '24h Volume Ranking' : '24h Trades Ranking';
            $("rankingList").innerHTML = '<div style="padding:20px;text-align:center;color:var(--text-tertiary)">No signal data</div>';
            return;
        }

        const items = [];
        for (const symbol of signalSymbols) {
            const ticker = tickerData.get(symbol);
            if (ticker) items.push({ symbol, volume: ticker.quote_volume || 0, trades: ticker.trade_count || 0 });
        }

        if (items.length === 0) {
            $("rankingTitle").textContent = type === 'volume' ? '24h Volume Ranking' : '24h Trades Ranking';
            $("rankingList").innerHTML = '<div style="padding:20px;text-align:center;color:var(--text-tertiary)">Waiting for ticker data...</div>';
            return;
        }

        if (type === 'volume') items.sort((a, b) => b.volume - a.volume);
        else items.sort((a, b) => b.trades - a.trades);

        renderRankingList(items, type);
    }

    function renderRankingList(items, type) {
        $("rankingTitle").textContent = type === 'volume' ? '24h Volume Ranking (USDT)' : '24h Trades Ranking';
        const list = $("rankingList");
        const frag = document.createDocumentFragment();

        items.forEach((item, i) => {
            const el = document.createElement("div");
            el.className = "ranking-item";
            el.dataset.symbol = item.symbol;
            const rankClass = i < 3 ? 'ranking-rank top3' : 'ranking-rank';
            const value = type === 'volume' ? fmtVolume(item.volume) : fmtTradeCount(item.trades) + ' trades';

            el.innerHTML = `
        <span class="${rankClass}">#${i + 1}</span>
        <span class="ranking-symbol">${item.symbol}</span>
        <span class="ranking-value">${value}</span>
      `;
            el.onclick = (e) => { e.preventDefault(); showActionMenu(e, item.symbol); };
            frag.appendChild(el);
        });

        list.innerHTML = '';
        list.appendChild(frag);
    }

    // 更新价格信息（高性能局部更新）
    function updateSignalPrices() {
        if (currentView !== 'signals') {
            if (currentView === 'volume' || currentView === 'trades') renderRanking();
            return;
        }

        document.querySelectorAll(".item[data-symbol]").forEach(item => {
            const symbol = item.dataset.symbol;
            const ticker = tickerData.get(symbol);
            if (!ticker) return;

            const signalPrice = parseFloat(item.dataset.price) || 0;

            // 更新价格差异
            const subDiv = item.querySelector(".sub > div:first-child");
            if (subDiv) {
                const diff = ticker.last_price - signalPrice;
                const diffPct = signalPrice > 0 ? ((diff / signalPrice) * 100) : 0;
                const diffSign = diff >= 0 ? '+' : '';
                const diffClass = diff >= 0 ? 'up' : 'down';
                subDiv.innerHTML = `Signal: ${fmtPrice(signalPrice)} <span class="price-diff ${diffClass}">${diffSign}${diffPct.toFixed(2)}%</span>`;
            }

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
        <span class="trades">${fmtTradeCount(ticker.trade_count)} trades</span>
      `;
        });
    }

    function updateRelTimes() {
        document.querySelectorAll(".time-rel").forEach(el => {
            const item = el.closest(".item");
            if (item && item.dataset.time) el.textContent = fmtRelTime(item.dataset.time);
        });
    }

    async function loadHistory() {
        const f = filters();
        $("hint").textContent = "Loading...";
        try {
            const r = await fetch("/api/history?" + buildQ(f));
            if (!r.ok) throw new Error("http " + r.status);
            allSignals = merge([], await r.json());
            updateView();
            $("hint").textContent = "Signals: " + allSignals.length;
        } catch (e) {
            $("hint").textContent = "Load failed: " + e;
        }
    }

    async function loadPivotStatus() {
        try {
            const r = await fetch("/api/pivot-status");
            if (!r.ok) return;
            const d = await r.json();
            const fmt = p => p ? ((p.is_stale ? '<span class="pill stale">STALE</span>' : '<span class="pill fresh">OK</span>') + " " + fmtDur(p.seconds_until) + " (" + p.symbol_count + " symbols)") : "-";
            $("dailyStatus").innerHTML = fmt(d.daily);
            $("weeklyStatus").innerHTML = fmt(d.weekly);
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

    const playBeep = () => {
        if (!$("soundEnabled").checked) return;
        try {
            const c = new (window.AudioContext || window.webkitAudioContext)();
            const o = c.createOscillator();
            const g = c.createGain();
            o.type = "sine"; o.frequency.value = 880; g.gain.value = 0.08;
            o.connect(g); g.connect(c.destination);
            o.start(); o.stop(c.currentTime + 0.15);
            setTimeout(() => c.close(), 500);
        } catch (_) { }
    };

    function connectSSE() {
        let es;
        try { es = new EventSource("/api/sse"); }
        catch (_) { setStatus("disconnected"); return; }

        setStatus("reconnecting");
        es.onopen = () => setStatus("connected");
        es.onerror = () => setStatus(es.readyState === 2 ? "disconnected" : "reconnecting");

        es.addEventListener("signal", e => {
            try {
                const s = JSON.parse(e.data);
                if (soundLevels.has(s.level)) playBeep();
                if (matchF(s, filters())) {
                    allSignals = merge(allSignals, [s]);
                    updateView();
                    $("hint").textContent = "Signals: " + allSignals.length;
                }
            } catch (_) { }
        });

        es.addEventListener("ticker", e => {
            try {
                const batch = JSON.parse(e.data);
                if (batch && batch.tickers) {
                    for (const [symbol, ticker] of Object.entries(batch.tickers)) {
                        tickerData.set(symbol, ticker);
                    }
                    requestAnimationFrame(updateSignalPrices);
                }
            } catch (_) { }
        });
    }

    const debounce = (fn, ms) => { let t; return () => { clearTimeout(t); t = setTimeout(fn, ms); }; };

    const dr = debounce(loadHistory, 300);
    $("refresh").onclick = () => { loadHistory(); loadPivotStatus(); loadTickers(); };
    $("symbol").oninput = dr;
    $("period").onchange = loadHistory;
    $("direction").onchange = loadHistory;

    setupLevelBtns();
    setupTabs();
    setupActionMenu();
    loadSoundSettings();
    updateSoundLevelBtns();

    loadTickers().then(() => { loadHistory(); loadPivotStatus(); connectSSE(); });
    setInterval(loadPivotStatus, 60000);
    setInterval(updateRelTimes, 10000);
})();
