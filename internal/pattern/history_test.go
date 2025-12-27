package pattern

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestHistory_MemoryOnly(t *testing.T) {
	h, err := NewHistory("", 100)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}

	if h.IsPersistent() {
		t.Error("Expected memory-only mode")
	}

	// Add signals
	klineTime := time.Now()
	sig1 := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime)
	sig2 := NewSignal("ETHUSDT", PatternEngulfing, DirectionBearish, 80, klineTime)

	if err := h.Add(sig1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := h.Add(sig2); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if h.Count() != 2 {
		t.Errorf("Count = %d, want 2", h.Count())
	}

	// Recent
	recent := h.Recent(10)
	if len(recent) != 2 {
		t.Errorf("Recent length = %d, want 2", len(recent))
	}
	// Newest first
	if recent[0].Symbol != "ETHUSDT" {
		t.Errorf("Recent[0].Symbol = %s, want ETHUSDT", recent[0].Symbol)
	}
}

func TestHistory_MaxSize(t *testing.T) {
	h, _ := NewHistory("", 3)

	klineTime := time.Now()
	for i := 0; i < 5; i++ {
		sig := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime.Add(time.Duration(i)*time.Minute))
		h.Add(sig)
	}

	if h.Count() != 3 {
		t.Errorf("Count = %d, want 3", h.Count())
	}
}

func TestHistory_Query(t *testing.T) {
	h, _ := NewHistory("", 100)

	klineTime := time.Now()
	h.Add(NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime))
	h.Add(NewSignal("ETHUSDT", PatternEngulfing, DirectionBearish, 80, klineTime))
	h.Add(NewSignal("BTCUSDT", PatternShootingStar, DirectionBearish, 70, klineTime))

	// Query by symbol
	results := h.Query(QueryOptions{Symbol: "BTCUSDT"})
	if len(results) != 2 {
		t.Errorf("Query by symbol: got %d, want 2", len(results))
	}

	// Query by pattern
	results = h.Query(QueryOptions{Pattern: PatternEngulfing})
	if len(results) != 1 {
		t.Errorf("Query by pattern: got %d, want 1", len(results))
	}

	// Query by direction
	results = h.Query(QueryOptions{Direction: DirectionBearish})
	if len(results) != 2 {
		t.Errorf("Query by direction: got %d, want 2", len(results))
	}

	// Query with limit
	results = h.Query(QueryOptions{Limit: 1})
	if len(results) != 1 {
		t.Errorf("Query with limit: got %d, want 1", len(results))
	}
}

func TestHistory_Persistence(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pattern_history_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "history.jsonl")

	// Create history and add signals
	h1, err := NewHistory(filePath, 100)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}

	if !h1.IsPersistent() {
		t.Error("Expected persistent mode")
	}

	klineTime := time.Now().Truncate(time.Second) // Truncate for comparison
	sig1 := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime)
	sig1.DetectedAt = sig1.DetectedAt.Truncate(time.Second)

	if err := h1.Add(sig1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	h1.Close()

	// Create new history and verify loaded
	h2, err := NewHistory(filePath, 100)
	if err != nil {
		t.Fatalf("NewHistory (reload) failed: %v", err)
	}
	defer h2.Close()

	if h2.Count() != 1 {
		t.Errorf("Reloaded count = %d, want 1", h2.Count())
	}

	recent := h2.Recent(1)
	if len(recent) != 1 {
		t.Fatalf("Recent length = %d, want 1", len(recent))
	}

	loaded := recent[0]
	if loaded.Symbol != sig1.Symbol {
		t.Errorf("Loaded Symbol = %s, want %s", loaded.Symbol, sig1.Symbol)
	}
	if loaded.Pattern != sig1.Pattern {
		t.Errorf("Loaded Pattern = %s, want %s", loaded.Pattern, sig1.Pattern)
	}
	if loaded.Direction != sig1.Direction {
		t.Errorf("Loaded Direction = %s, want %s", loaded.Direction, sig1.Direction)
	}
}

// Property test: History persistence round-trip
func TestProperty_HistoryPersistenceRoundTrip(t *testing.T) {
	properties := gopter.NewProperties(nil)

	patternTypes := []PatternType{PatternHammer, PatternEngulfing, PatternMorningStar, PatternShootingStar}
	directions := []Direction{DirectionBullish, DirectionBearish}

	properties.Property("Signals survive persistence round-trip", prop.ForAll(
		func(symbolIdx, patternIdx, directionIdx, confidence int) bool {
			// Create temp file
			tmpDir, err := os.MkdirTemp("", "history_prop_test")
			if err != nil {
				return false
			}
			defer os.RemoveAll(tmpDir)

			filePath := filepath.Join(tmpDir, "history.jsonl")

			symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
			symbol := symbols[symbolIdx%len(symbols)]
			pattern := patternTypes[patternIdx%len(patternTypes)]
			direction := directions[directionIdx%len(directions)]
			conf := confidence % 101

			klineTime := time.Now().Truncate(time.Second)

			// Create and add signal
			h1, err := NewHistory(filePath, 100)
			if err != nil {
				return false
			}

			original := NewSignal(symbol, pattern, direction, conf, klineTime)
			original.DetectedAt = original.DetectedAt.Truncate(time.Second)

			if err := h1.Add(original); err != nil {
				h1.Close()
				return false
			}
			h1.Close()

			// Reload and verify
			h2, err := NewHistory(filePath, 100)
			if err != nil {
				return false
			}
			defer h2.Close()

			if h2.Count() != 1 {
				return false
			}

			loaded := h2.Recent(1)[0]

			// Verify all fields match
			if loaded.Symbol != original.Symbol {
				return false
			}
			if loaded.Pattern != original.Pattern {
				return false
			}
			if loaded.Direction != original.Direction {
				return false
			}
			if loaded.Confidence != original.Confidence {
				return false
			}
			if loaded.UpPercent != original.UpPercent {
				return false
			}
			if loaded.DownPercent != original.DownPercent {
				return false
			}

			return true
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}


// Property 3: 参数校验防护
// *For any* negative or zero value passed to NewHistory, the system should not panic
// and should use a positive default value.
// **Validates: Requirements 2.1**

func TestNewHistory_NegativeMaxSize(t *testing.T) {
	// 测试负数不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewHistory panicked with negative maxSize: %v", r)
		}
	}()

	h, err := NewHistory("", -10)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}
	if h == nil {
		t.Error("NewHistory returned nil")
	}

	// 验证使用了默认值
	if h.maxSize != DefaultPatternHistoryMax {
		t.Errorf("maxSize = %d, want default %d", h.maxSize, DefaultPatternHistoryMax)
	}
}

func TestNewHistory_ZeroMaxSize(t *testing.T) {
	// 测试零不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewHistory panicked with zero maxSize: %v", r)
		}
	}()

	h, err := NewHistory("", 0)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}
	if h == nil {
		t.Error("NewHistory returned nil")
	}

	// 验证使用了默认值
	if h.maxSize != DefaultPatternHistoryMax {
		t.Errorf("maxSize = %d, want default %d", h.maxSize, DefaultPatternHistoryMax)
	}
}

func TestNewHistory_ValidMaxSize(t *testing.T) {
	h, err := NewHistory("", 500)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}
	if h == nil {
		t.Error("NewHistory returned nil")
	}

	// 验证使用了传入的值
	if h.maxSize != 500 {
		t.Errorf("maxSize = %d, want 500", h.maxSize)
	}
}

func TestProperty_NewHistoryNeverPanics(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("NewHistory never panics with any maxSize value", prop.ForAll(
		func(maxSize int) bool {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewHistory panicked with maxSize=%d: %v", maxSize, r)
				}
			}()

			h, err := NewHistory("", maxSize)
			if err != nil {
				return false
			}
			if h == nil {
				return false
			}

			// maxSize 应该总是正数
			if h.maxSize <= 0 {
				return false
			}

			return true
		},
		gen.IntRange(-1000, 1000),
	))

	properties.TestingRun(t)
}


// Property 4: 文件截断保留最新记录
// *For any* sequence of pattern signals written to history, after truncation,
// the file should contain exactly the most recent maxSize entries in chronological order.
// **Validates: Requirements 3.2, 3.3**

func TestHistory_FileCompaction(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pattern_history_compact_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "history.jsonl")
	maxSize := 10

	h, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}

	// 写入超过 maxSize*2 条记录以触发截断
	// 截断在 fileLines%100==0 && fileLines > maxSize*2 时触发
	// 所以我们需要写入至少 100 条记录
	numSignals := 100
	klineTime := time.Now()

	for i := 0; i < numSignals; i++ {
		sig := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime.Add(time.Duration(i)*time.Minute))
		if err := h.Add(sig); err != nil {
			t.Fatalf("Add failed at %d: %v", i, err)
		}
	}

	// 验证内存中只有 maxSize 条记录
	if h.Count() != maxSize {
		t.Errorf("Memory count = %d, want %d", h.Count(), maxSize)
	}

	// 关闭并重新打开，验证文件被截断
	h.Close()

	h2, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory (reload) failed: %v", err)
	}
	defer h2.Close()

	// 验证重新加载后只有 maxSize 条记录
	if h2.Count() != maxSize {
		t.Errorf("Reloaded count = %d, want %d", h2.Count(), maxSize)
	}

	// 验证是最新的记录（最后 maxSize 条）
	recent := h2.Recent(maxSize)
	// recent[0] 应该是最新的，即第 numSignals-1 条
	expectedLastMinute := time.Duration(numSignals-1) * time.Minute
	expectedFirstMinute := time.Duration(numSignals-maxSize) * time.Minute

	// 验证最新的记录
	if recent[0].KlineTime.Sub(klineTime).Round(time.Minute) != expectedLastMinute {
		t.Errorf("Newest signal time offset = %v, want %v",
			recent[0].KlineTime.Sub(klineTime).Round(time.Minute), expectedLastMinute)
	}

	// 验证最旧的记录（在 recent 中是最后一个）
	if recent[maxSize-1].KlineTime.Sub(klineTime).Round(time.Minute) != expectedFirstMinute {
		t.Errorf("Oldest signal time offset = %v, want %v",
			recent[maxSize-1].KlineTime.Sub(klineTime).Round(time.Minute), expectedFirstMinute)
	}
}

func TestHistory_CompactPreservesOrder(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pattern_history_order_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "history.jsonl")
	maxSize := 5

	h, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}

	// 写入足够多的记录触发截断
	numSignals := 100
	klineTime := time.Now()

	for i := 0; i < numSignals; i++ {
		sig := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, i%100, klineTime.Add(time.Duration(i)*time.Minute))
		h.Add(sig)
	}

	h.Close()

	// 重新加载
	h2, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory (reload) failed: %v", err)
	}
	defer h2.Close()

	// 验证顺序：Recent 返回的是从新到旧
	recent := h2.Recent(maxSize)

	// 验证 confidence 值是递减的（因为 Recent 返回从新到旧）
	for i := 0; i < len(recent)-1; i++ {
		// recent[i] 应该比 recent[i+1] 更新
		if recent[i].KlineTime.Before(recent[i+1].KlineTime) {
			t.Errorf("Order violation at %d: %v should be after %v",
				i, recent[i].KlineTime, recent[i+1].KlineTime)
		}
	}
}

func TestHistory_FileLineTracking(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pattern_history_lines_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "history.jsonl")
	maxSize := 100

	h, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory failed: %v", err)
	}

	// 初始行数应该是 0
	if h.fileLines != 0 {
		t.Errorf("Initial fileLines = %d, want 0", h.fileLines)
	}

	// 写入 50 条记录
	klineTime := time.Now()
	for i := 0; i < 50; i++ {
		sig := NewSignal("BTCUSDT", PatternHammer, DirectionBullish, 75, klineTime)
		h.Add(sig)
	}

	// 行数应该是 50
	if h.fileLines != 50 {
		t.Errorf("After 50 adds, fileLines = %d, want 50", h.fileLines)
	}

	h.Close()

	// 重新加载，验证行数被正确读取
	h2, err := NewHistory(filePath, maxSize)
	if err != nil {
		t.Fatalf("NewHistory (reload) failed: %v", err)
	}
	defer h2.Close()

	if h2.fileLines != 50 {
		t.Errorf("Reloaded fileLines = %d, want 50", h2.fileLines)
	}
}
