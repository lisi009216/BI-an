package ticker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"example.com/binance-pivot-monitor/internal/binance"
	"github.com/gorilla/websocket"
)

// TickerBatch 批量行情更新，用于 SSE 推送
type TickerBatch struct {
	Tickers   map[string]*Ticker `json:"tickers"`
	Timestamp int64              `json:"ts"`
}

// Monitor 监控 ticker 数据并广播
type Monitor struct {
	Store         *Store
	BatchInterval time.Duration // 批量推送间隔，默认 500ms

	mu        sync.RWMutex
	listeners []chan TickerBatch
	pending   map[string]*Ticker // 待推送的变化
}

func NewMonitor(store *Store) *Monitor {
	return &Monitor{
		Store:         store,
		BatchInterval: 500 * time.Millisecond,
		pending:       make(map[string]*Ticker),
	}
}

// Subscribe 订阅批量行情更新
func (m *Monitor) Subscribe(buffer int) chan TickerBatch {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan TickerBatch, buffer)
	m.mu.Lock()
	m.listeners = append(m.listeners, ch)
	m.mu.Unlock()
	return ch
}

// Unsubscribe 取消订阅
func (m *Monitor) Unsubscribe(ch chan TickerBatch) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.listeners {
		if c == ch {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}

// broadcast 广播批量更新
func (m *Monitor) broadcast(batch TickerBatch) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.listeners {
		select {
		case ch <- batch:
		default:
			// 丢弃，避免阻塞
		}
	}
}

// Run 启动 ticker 监控
func (m *Monitor) Run(ctx context.Context) {
	// 启动批量推送协程
	go m.batchPusher(ctx)

	backoff := 1 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}

		conn, _, err := binance.DialTickerArr(ctx)
		if err != nil {
			log.Printf("ticker ws dial failed: %v", err)
			if !sleepContext(ctx, backoff) {
				return
			}
			backoff = minDuration(backoff*2, 30*time.Second)
			continue
		}

		log.Printf("ticker ws connected")
		backoff = 1 * time.Second

		err = m.readLoop(ctx, conn)
		_ = conn.Close()
		if err != nil && ctx.Err() == nil {
			log.Printf("ticker ws read loop exit: %v", err)
		}

		if !sleepContext(ctx, backoff) {
			return
		}
		backoff = minDuration(backoff*2, 30*time.Second)
	}
}

func (m *Monitor) readLoop(ctx context.Context, conn *websocket.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Ping 协程
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-t.C:
				_ = conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
			}
		}
	}()
	defer close(done)

	msgCount := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, b, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 调试：打印前几条原始消息
		if msgCount < 2 {
			log.Printf("ticker raw msg #%d len=%d prefix: %s", msgCount, len(b), string(b[:min(len(b), 300)]))
		}

		var events []binance.TickerEvent
		if err := json.Unmarshal(b, &events); err != nil {
			// 打印前几条解析失败的消息
			if msgCount < 5 {
				log.Printf("ticker unmarshal error: %v, data prefix: %s", err, string(b[:min(len(b), 300)]))
			}
			msgCount++
			continue
		}

		// 调试：打印解析结果
		if msgCount < 2 && len(events) > 0 {
			log.Printf("ticker parsed: count=%d, first=%+v", len(events), events[0])
		}

		for _, ev := range events {
			m.Store.Update(ev.Symbol, ev.LastPrice, ev.PricePercent, ev.TradeCount, ev.QuoteVolume)

			// 记录待推送
			m.mu.Lock()
			m.pending[ev.Symbol] = &Ticker{
				Symbol:       ev.Symbol,
				LastPrice:    ev.LastPrice,
				PricePercent: ev.PricePercent,
				TradeCount:   ev.TradeCount,
				QuoteVolume:  ev.QuoteVolume,
				UpdatedAt:    time.Now().UnixMilli(),
			}
			m.mu.Unlock()
		}

		// 首次成功解析时打印日志
		if msgCount == 0 && len(events) > 0 {
			log.Printf("ticker first batch received: %d symbols", len(events))
		}
		msgCount++
	}
}

// batchPusher 定时批量推送变化的 ticker
func (m *Monitor) batchPusher(ctx context.Context) {
	ticker := time.NewTicker(m.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if len(m.pending) == 0 {
				m.mu.Unlock()
				continue
			}

			batch := TickerBatch{
				Tickers:   m.pending,
				Timestamp: time.Now().UnixMilli(),
			}
			m.pending = make(map[string]*Ticker)
			m.mu.Unlock()

			m.broadcast(batch)
		}
	}
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
