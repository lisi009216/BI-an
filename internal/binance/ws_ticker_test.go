package binance

import (
	"encoding/json"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestTickerEvent_UnmarshalJSON_NumberFormat(t *testing.T) {
	// 测试数字格式的 JSON
	jsonData := `{
		"s": "BTCUSDT",
		"E": 1234567890123,
		"n": 12345,
		"c": "50000.50",
		"p": "100.25",
		"P": "0.20",
		"h": "51000.00",
		"l": "49000.00",
		"v": "1000.5",
		"q": "50000000.00"
	}`

	var event TickerEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %s, want BTCUSDT", event.Symbol)
	}
	if event.EventTime != 1234567890123 {
		t.Errorf("EventTime = %d, want 1234567890123", event.EventTime)
	}
	if event.TradeCount != 12345 {
		t.Errorf("TradeCount = %d, want 12345", event.TradeCount)
	}
	if event.LastPrice != 50000.50 {
		t.Errorf("LastPrice = %f, want 50000.50", event.LastPrice)
	}
}

func TestTickerEvent_UnmarshalJSON_StringFormat(t *testing.T) {
	// 测试字符串格式的数字（Binance 有时会返回这种格式）
	jsonData := `{
		"s": "ETHUSDT",
		"E": "1234567890123",
		"n": "67890",
		"c": "3000.50",
		"p": "50.25",
		"P": "1.70",
		"h": "3100.00",
		"l": "2900.00",
		"v": "5000.5",
		"q": "15000000.00"
	}`

	var event TickerEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.Symbol != "ETHUSDT" {
		t.Errorf("Symbol = %s, want ETHUSDT", event.Symbol)
	}
	// 关键测试：字符串格式的 EventTime 和 TradeCount 应该被正确解析
	if event.EventTime != 1234567890123 {
		t.Errorf("EventTime = %d, want 1234567890123 (string format)", event.EventTime)
	}
	if event.TradeCount != 67890 {
		t.Errorf("TradeCount = %d, want 67890 (string format)", event.TradeCount)
	}
	if event.LastPrice != 3000.50 {
		t.Errorf("LastPrice = %f, want 3000.50", event.LastPrice)
	}
}

func TestTickerEvent_UnmarshalJSON_MixedFormat(t *testing.T) {
	// 测试混合格式（部分数字，部分字符串）
	jsonData := `{
		"s": "BNBUSDT",
		"E": 1234567890123,
		"n": "99999",
		"c": "500.50",
		"p": "10.25",
		"P": "2.10",
		"h": "510.00",
		"l": "490.00",
		"v": "10000.5",
		"q": "5000000.00"
	}`

	var event TickerEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.EventTime != 1234567890123 {
		t.Errorf("EventTime = %d, want 1234567890123 (number format)", event.EventTime)
	}
	if event.TradeCount != 99999 {
		t.Errorf("TradeCount = %d, want 99999 (string format)", event.TradeCount)
	}
}

func TestTickerEvent_UnmarshalJSON_MissingFields(t *testing.T) {
	// 测试缺失字段
	jsonData := `{
		"s": "BTCUSDT",
		"c": "50000.50"
	}`

	var event TickerEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %s, want BTCUSDT", event.Symbol)
	}
	// 缺失字段应该是默认值 0
	if event.EventTime != 0 {
		t.Errorf("EventTime = %d, want 0 (missing field)", event.EventTime)
	}
	if event.TradeCount != 0 {
		t.Errorf("TradeCount = %d, want 0 (missing field)", event.TradeCount)
	}
}

// Property 5: JSON 数值解析等价性
// *For any* numeric value, parsing it from JSON number format and JSON string format
// should produce the same result.
// **Validates: Requirements 4.1, 4.2, 4.4**

func TestProperty_ParseIntEquivalence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("parseInt produces same result for number and string formats", prop.ForAll(
		func(eventTime, tradeCount int64) bool {
			// 确保值在合理范围内
			if eventTime < 0 {
				eventTime = -eventTime
			}
			if tradeCount < 0 {
				tradeCount = -tradeCount
			}

			// 数字格式
			jsonNumber := []byte(`{"s":"TEST","E":` + string(json.Number(itoa(eventTime))) + `,"n":` + string(json.Number(itoa(tradeCount))) + `}`)

			// 字符串格式
			jsonString := []byte(`{"s":"TEST","E":"` + itoa(eventTime) + `","n":"` + itoa(tradeCount) + `"}`)

			var eventNum, eventStr TickerEvent
			if err := json.Unmarshal(jsonNumber, &eventNum); err != nil {
				return false
			}
			if err := json.Unmarshal(jsonString, &eventStr); err != nil {
				return false
			}

			// 两种格式应该产生相同的结果
			if eventNum.EventTime != eventStr.EventTime {
				t.Logf("EventTime mismatch: number=%d, string=%d", eventNum.EventTime, eventStr.EventTime)
				return false
			}
			if eventNum.TradeCount != eventStr.TradeCount {
				t.Logf("TradeCount mismatch: number=%d, string=%d", eventNum.TradeCount, eventStr.TradeCount)
				return false
			}

			// 值应该正确
			if eventNum.EventTime != eventTime {
				t.Logf("EventTime value mismatch: got=%d, want=%d", eventNum.EventTime, eventTime)
				return false
			}
			if eventNum.TradeCount != tradeCount {
				t.Logf("TradeCount value mismatch: got=%d, want=%d", eventNum.TradeCount, tradeCount)
				return false
			}

			return true
		},
		gen.Int64Range(0, 9999999999999),
		gen.Int64Range(0, 999999999),
	))

	properties.TestingRun(t)
}

// itoa converts int64 to string
func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
