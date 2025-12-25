package ticker

import (
	"sync"
	"time"
)

// Ticker 精简的行情数据，用于前端显示
type Ticker struct {
	Symbol       string  `json:"symbol"`
	LastPrice    float64 `json:"last_price"`
	PricePercent float64 `json:"price_percent"` // 24h 涨跌幅
	TradeCount   int64   `json:"trade_count"`   // 24h 成交数
	QuoteVolume  float64 `json:"quote_volume"`  // 24h 成交额(USDT)
	UpdatedAt    int64   `json:"updated_at"`    // 更新时间戳(ms)
}

// Store 存储所有交易对的行情数据
type Store struct {
	mu      sync.RWMutex
	tickers map[string]*Ticker
}

func NewStore() *Store {
	return &Store{
		tickers: make(map[string]*Ticker),
	}
}

// Update 更新单个交易对的行情
func (s *Store) Update(symbol string, lastPrice, pricePercent float64, tradeCount int64, quoteVolume float64) {
	s.mu.Lock()
	s.tickers[symbol] = &Ticker{
		Symbol:       symbol,
		LastPrice:    lastPrice,
		PricePercent: pricePercent,
		TradeCount:   tradeCount,
		QuoteVolume:  quoteVolume,
		UpdatedAt:    time.Now().UnixMilli(),
	}
	s.mu.Unlock()
}

// Get 获取单个交易对的行情
func (s *Store) Get(symbol string) (*Ticker, bool) {
	s.mu.RLock()
	t, ok := s.tickers[symbol]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	// 返回副本
	copy := *t
	return &copy, true
}

// GetAll 获取所有交易对的行情
func (s *Store) GetAll() map[string]*Ticker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*Ticker, len(s.tickers))
	for k, v := range s.tickers {
		copy := *v
		result[k] = &copy
	}
	return result
}

// GetBySymbols 获取指定交易对的行情
func (s *Store) GetBySymbols(symbols []string) map[string]*Ticker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*Ticker, len(symbols))
	for _, sym := range symbols {
		if t, ok := s.tickers[sym]; ok {
			copy := *t
			result[sym] = &copy
		}
	}
	return result
}

// Count 返回存储的交易对数量
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tickers)
}
