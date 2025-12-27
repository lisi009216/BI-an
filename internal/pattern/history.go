package pattern

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// History stores pattern signal history.
// Storage strategy: memory-first, optional persistence via file.
type History struct {
	mu          sync.RWMutex
	signals     []Signal
	maxSize     int
	filePath    string // Empty means memory-only mode
	persistMode bool
	file        *os.File
	fileLines   int // 跟踪文件行数，用于截断判断
}

// DefaultPatternHistoryMax is the default maximum number of pattern signals to keep.
const DefaultPatternHistoryMax = 1000

// NewHistory creates a new history store.
// filePath: empty string for memory-only mode, non-empty to enable persistence.
func NewHistory(filePath string, maxSize int) (*History, error) {
	// 参数校验：防止负数或零导致 panic
	if maxSize <= 0 {
		log.Printf("WARN: invalid PATTERN_HISTORY_MAX=%d, using default %d", maxSize, DefaultPatternHistoryMax)
		maxSize = DefaultPatternHistoryMax
	}

	h := &History{
		signals:     make([]Signal, 0, maxSize),
		maxSize:     maxSize,
		filePath:    filePath,
		persistMode: filePath != "",
	}

	if h.persistMode {
		// Ensure directory exists
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Load existing history
		if err := h.load(); err != nil {
			// Log warning but continue - file might not exist yet
		}

		// Open file for appending
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		h.file = f
	}

	return h, nil
}

// load reads existing signals from file.
func (h *History) load() error {
	f, err := os.Open(h.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var signals []Signal
	lines := 0

	for scanner.Scan() {
		lines++
		var sig Signal
		if err := json.Unmarshal(scanner.Bytes(), &sig); err != nil {
			continue // Skip invalid lines
		}
		signals = append(signals, sig)
	}

	// Keep only the most recent maxSize signals
	if len(signals) > h.maxSize {
		signals = signals[len(signals)-h.maxSize:]
	}

	h.signals = signals
	h.fileLines = lines
	return scanner.Err()
}

// Add adds a signal to history.
// If persistence is enabled, writes to file synchronously.
func (h *History) Add(sig Signal) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to memory
	h.signals = append(h.signals, sig)

	// Maintain max size
	if len(h.signals) > h.maxSize {
		h.signals = h.signals[len(h.signals)-h.maxSize:]
	}

	// Persist if enabled
	if h.persistMode && h.file != nil {
		data, err := json.Marshal(sig)
		if err != nil {
			return err
		}
		if _, err := h.file.Write(append(data, '\n')); err != nil {
			return err
		}
		h.fileLines++

		// 每 100 条检查一次，文件行数超过 maxSize*2 时触发截断
		if h.fileLines%100 == 0 && h.fileLines > h.maxSize*2 {
			oldLines := h.fileLines
			if err := h.compact(); err != nil {
				log.Printf("WARN: pattern history compact failed: %v", err)
				// 继续运行，不中断
			} else {
				log.Printf("pattern history compacted: %d -> %d lines", oldLines, h.fileLines)
			}
		}
	}

	return nil
}

// Recent returns the most recent signals.
func (h *History) Recent(limit int) []Signal {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.signals) {
		limit = len(h.signals)
	}

	// Return most recent (from end)
	start := len(h.signals) - limit
	result := make([]Signal, limit)
	copy(result, h.signals[start:])

	// Reverse to get newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// QueryOptions defines options for querying history.
type QueryOptions struct {
	Symbol    string
	Pattern   PatternType
	Direction Direction
	Limit     int
	Since     time.Time
}

// Query queries signals with filtering options.
func (h *History) Query(opts QueryOptions) []Signal {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []Signal

	// Iterate from newest to oldest
	for i := len(h.signals) - 1; i >= 0; i-- {
		sig := h.signals[i]

		// Apply filters
		if opts.Symbol != "" && sig.Symbol != opts.Symbol {
			continue
		}
		if opts.Pattern != "" && sig.Pattern != opts.Pattern {
			continue
		}
		if opts.Direction != "" && sig.Direction != opts.Direction {
			continue
		}
		if !opts.Since.IsZero() && sig.DetectedAt.Before(opts.Since) {
			continue
		}

		result = append(result, sig)

		if opts.Limit > 0 && len(result) >= opts.Limit {
			break
		}
	}

	return result
}

// IsPersistent returns whether persistence is enabled.
func (h *History) IsPersistent() bool {
	return h.persistMode
}

// Count returns the number of signals in memory.
func (h *History) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.signals)
}

// Close closes the history file if open.
func (h *History) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

// compact 截断历史文件，只保留最新的 maxSize 条记录
// 参考 internal/signal/history.go 的实现
func (h *History) compact() error {
	if !h.persistMode || h.filePath == "" {
		return nil
	}

	// 保存旧文件句柄，以便失败时恢复
	oldFile := h.file
	h.file = nil

	// 创建临时文件
	tmp := h.filePath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		// 恢复旧文件句柄
		h.file = oldFile
		return err
	}

	// 写入最新的记录
	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	for _, sig := range h.signals {
		if err := enc.Encode(sig); err != nil {
			bw.Flush()
			f.Close()
			os.Remove(tmp)
			h.file = oldFile
			return err
		}
	}

	if err := bw.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		h.file = oldFile
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		h.file = oldFile
		return err
	}

	// 关闭旧文件句柄（在原子替换前）
	if oldFile != nil {
		oldFile.Close()
	}

	// 原子替换
	if err := os.Rename(tmp, h.filePath); err != nil {
		os.Remove(tmp)
		// 尝试重新打开原文件
		if newFile, openErr := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); openErr == nil {
			h.file = newFile
		}
		return err
	}

	// 重新打开文件用于追加
	newFile, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// 尝试恢复，但文件已被替换，只能记录错误
		return err
	}
	h.file = newFile
	h.fileLines = len(h.signals)

	return nil
}

// QueryBySymbolAndTime finds patterns for a symbol within a time window around a reference time.
// Returns patterns sorted by time proximity (closest first).
func (h *History) QueryBySymbolAndTime(symbol string, refTime time.Time, window time.Duration) []Signal {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []Signal

	for _, sig := range h.signals {
		if sig.Symbol != symbol {
			continue
		}

		// Check if within time window
		diff := refTime.Sub(sig.DetectedAt)
		if diff < 0 {
			diff = -diff
		}
		if diff <= window {
			result = append(result, sig)
		}
	}

	// Sort by time proximity (closest to refTime first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			diffI := refTime.Sub(result[i].DetectedAt)
			if diffI < 0 {
				diffI = -diffI
			}
			diffJ := refTime.Sub(result[j].DetectedAt)
			if diffJ < 0 {
				diffJ = -diffJ
			}
			if diffJ < diffI {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
