package binance

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// TickerEvent 24小时行情数据
type TickerEvent struct {
	Symbol       string  // 交易对
	LastPrice    float64 // 最新成交价格
	PriceChange  float64 // 24小时价格变化
	PricePercent float64 // 24小时价格变化(百分比)
	High         float64 // 24小时内最高成交价
	Low          float64 // 24小时内最低成交价
	Volume       float64 // 24小时内成交量
	QuoteVolume  float64 // 24小时内成交额
	TradeCount   int64   // 24小时内成交数
	EventTime    int64   // 事件时间
}

func (e *TickerEvent) UnmarshalJSON(data []byte) error {
	// 使用 Decoder 配合 UseNumber 来处理所有数字类型
	// 这样无论是 "123" 还是 123 都能正确解析
	var raw map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	// 辅助函数：解析为 float64
	parseFloat := func(key string) float64 {
		v, ok := raw[key]
		if !ok {
			return 0
		}
		switch val := v.(type) {
		case json.Number:
			f, _ := val.Float64()
			return f
		case string:
			f, _ := json.Number(val).Float64()
			return f
		case float64:
			return val
		}
		return 0
	}

	// 辅助函数：解析为 int64
	parseInt := func(key string) int64 {
		v, ok := raw[key]
		if !ok {
			return 0
		}
		switch val := v.(type) {
		case json.Number:
			i, _ := val.Int64()
			return i
		case float64:
			return int64(val)
		}
		return 0
	}

	// 解析字符串字段
	if v, ok := raw["s"].(string); ok {
		e.Symbol = v
	}

	// 解析数字字段
	e.EventTime = parseInt("E")
	e.TradeCount = parseInt("n")
	e.LastPrice = parseFloat("c")
	e.PriceChange = parseFloat("p")
	e.PricePercent = parseFloat("P")
	e.High = parseFloat("h")
	e.Low = parseFloat("l")
	e.Volume = parseFloat("v")
	e.QuoteVolume = parseFloat("q")

	return nil
}

// DialTickerArr 订阅所有交易对的24小时行情
func DialTickerArr(ctx context.Context) (*websocket.Conn, *http.Response, error) {
	d := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
	}
	url := FStreamWSBaseURL + "/!ticker@arr"
	return d.DialContext(ctx, url, nil)
}
