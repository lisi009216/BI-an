package kline

import (
	"testing"
	"time"
)

func TestKline_Body(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		close    float64
		expected float64
	}{
		{"bullish", 100.0, 110.0, 10.0},
		{"bearish", 110.0, 100.0, 10.0},
		{"doji", 100.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{Open: tt.open, Close: tt.close}
			if got := k.Body(); got != tt.expected {
				t.Errorf("Body() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_UpperShadow(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		high     float64
		close    float64
		expected float64
	}{
		{"bullish", 100.0, 115.0, 110.0, 5.0},
		{"bearish", 110.0, 115.0, 100.0, 5.0},
		{"no upper shadow bullish", 100.0, 110.0, 110.0, 0.0},
		{"no upper shadow bearish", 110.0, 110.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{Open: tt.open, High: tt.high, Close: tt.close}
			if got := k.UpperShadow(); got != tt.expected {
				t.Errorf("UpperShadow() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_LowerShadow(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		low      float64
		close    float64
		expected float64
	}{
		{"bullish", 100.0, 95.0, 110.0, 5.0},
		{"bearish", 110.0, 95.0, 100.0, 5.0},
		{"no lower shadow bullish", 100.0, 100.0, 110.0, 0.0},
		{"no lower shadow bearish", 110.0, 100.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{Open: tt.open, Low: tt.low, Close: tt.close}
			if got := k.LowerShadow(); got != tt.expected {
				t.Errorf("LowerShadow() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_IsBullish(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		close    float64
		expected bool
	}{
		{"bullish", 100.0, 110.0, true},
		{"bearish", 110.0, 100.0, false},
		{"doji", 100.0, 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{Open: tt.open, Close: tt.close}
			if got := k.IsBullish(); got != tt.expected {
				t.Errorf("IsBullish() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_IsBearish(t *testing.T) {
	tests := []struct {
		name     string
		open     float64
		close    float64
		expected bool
	}{
		{"bullish", 100.0, 110.0, false},
		{"bearish", 110.0, 100.0, true},
		{"doji", 100.0, 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{Open: tt.open, Close: tt.close}
			if got := k.IsBearish(); got != tt.expected {
				t.Errorf("IsBearish() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_Range(t *testing.T) {
	tests := []struct {
		name     string
		high     float64
		low      float64
		expected float64
	}{
		{"normal", 115.0, 95.0, 20.0},
		{"zero range", 100.0, 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kline{High: tt.high, Low: tt.low}
			if got := k.Range(); got != tt.expected {
				t.Errorf("Range() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestKline_Clone(t *testing.T) {
	now := time.Now()
	original := &Kline{
		Symbol:    "BTCUSDT",
		Open:      100.0,
		High:      115.0,
		Low:       95.0,
		Close:     110.0,
		OpenTime:  now,
		CloseTime: now.Add(5 * time.Minute),
		IsClosed:  true,
	}

	clone := original.Clone()

	// Verify all fields are copied
	if clone.Symbol != original.Symbol {
		t.Errorf("Clone Symbol = %v, want %v", clone.Symbol, original.Symbol)
	}
	if clone.Open != original.Open {
		t.Errorf("Clone Open = %v, want %v", clone.Open, original.Open)
	}
	if clone.High != original.High {
		t.Errorf("Clone High = %v, want %v", clone.High, original.High)
	}
	if clone.Low != original.Low {
		t.Errorf("Clone Low = %v, want %v", clone.Low, original.Low)
	}
	if clone.Close != original.Close {
		t.Errorf("Clone Close = %v, want %v", clone.Close, original.Close)
	}
	if !clone.OpenTime.Equal(original.OpenTime) {
		t.Errorf("Clone OpenTime = %v, want %v", clone.OpenTime, original.OpenTime)
	}
	if !clone.CloseTime.Equal(original.CloseTime) {
		t.Errorf("Clone CloseTime = %v, want %v", clone.CloseTime, original.CloseTime)
	}
	if clone.IsClosed != original.IsClosed {
		t.Errorf("Clone IsClosed = %v, want %v", clone.IsClosed, original.IsClosed)
	}

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clone.Close = 200.0
	if original.Close == clone.Close {
		t.Error("Clone is not a deep copy - modifying clone affected original")
	}
}
