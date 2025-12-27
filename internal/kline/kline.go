// Package kline provides virtual K-line data structures and storage for candlestick pattern recognition.
package kline

import (
	"math"
	"time"
)

// Kline represents a single candlestick (K-line) data.
type Kline struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	OpenTime  time.Time `json:"open_time"`
	CloseTime time.Time `json:"close_time"`
	IsClosed  bool      `json:"is_closed"`
}

// Body returns the absolute size of the kline body (|Close - Open|).
func (k *Kline) Body() float64 {
	return math.Abs(k.Close - k.Open)
}

// UpperShadow returns the length of the upper shadow.
func (k *Kline) UpperShadow() float64 {
	if k.Close > k.Open {
		return k.High - k.Close
	}
	return k.High - k.Open
}

// LowerShadow returns the length of the lower shadow.
func (k *Kline) LowerShadow() float64 {
	if k.Close > k.Open {
		return k.Open - k.Low
	}
	return k.Close - k.Low
}

// IsBullish returns true if the kline is bullish (Close > Open).
func (k *Kline) IsBullish() bool {
	return k.Close > k.Open
}

// IsBearish returns true if the kline is bearish (Close < Open).
func (k *Kline) IsBearish() bool {
	return k.Close < k.Open
}

// Range returns the total range of the kline (High - Low).
func (k *Kline) Range() float64 {
	return k.High - k.Low
}

// Clone returns a deep copy of the kline.
func (k *Kline) Clone() Kline {
	return Kline{
		Symbol:    k.Symbol,
		Open:      k.Open,
		High:      k.High,
		Low:       k.Low,
		Close:     k.Close,
		OpenTime:  k.OpenTime,
		CloseTime: k.CloseTime,
		IsClosed:  k.IsClosed,
	}
}
