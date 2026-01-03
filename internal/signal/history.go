package signal

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Period constants for bucket keys
const (
	PeriodDaily  = "1d"
	PeriodWeekly = "1w"
	PeriodOther  = "other"
)

// Default capacity ratios for period buckets
const (
	dailyRatio  = 0.80 // 80% for daily signals
	weeklyRatio = 0.15 // 15% for weekly signals
	otherRatio  = 0.05 // 5% for other signals
)

// periodBucket holds signals for a specific period with independent capacity.
type periodBucket struct {
	mu           sync.RWMutex
	max          int
	signals      []Signal
	symbolsUpper []string

	fileMu    sync.Mutex
	filePath  string
	fileLines int
}

// newPeriodBucket creates a new bucket with the given capacity.
func newPeriodBucket(max int) *periodBucket {
	if max <= 0 {
		max = 1000
	}
	return &periodBucket{max: max}
}

// normalizePeriod converts various period formats to standard bucket keys.
func normalizePeriod(period string) string {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "1d", "d", "daily":
		return PeriodDaily
	case "1w", "w", "weekly":
		return PeriodWeekly
	default:
		return PeriodOther
	}
}

type History struct {
	// Legacy fields for backward compatibility (used during migration)
	mu           sync.RWMutex
	max          int
	signals      []Signal
	symbolsUpper []string

	fileMu    sync.Mutex
	filePath  string
	fileLines int

	// New period-separated storage
	buckets   map[string]*periodBucket // period -> bucket
	bucketsMu sync.RWMutex

	// Capacity configuration
	periodMax  map[string]int // per-period capacity overrides
	defaultMax int            // default capacity for unconfigured periods

	// Persistence configuration
	baseDir    string // directory for period files
	baseName   string // base filename without extension
	separated  bool   // true if using period-separated storage
	migrated   bool   // true if migration has been attempted
}

func NewHistory(max int) *History {
	if max <= 0 {
		max = 10000
	}

	// Calculate default capacity for each period based on ratios
	dailyMax := int(float64(max) * dailyRatio)
	weeklyMax := int(float64(max) * weeklyRatio)
	otherMax := int(float64(max) * otherRatio)

	// Ensure minimum capacity
	if dailyMax < 100 {
		dailyMax = 100
	}
	if weeklyMax < 100 {
		weeklyMax = 100
	}
	if otherMax < 50 {
		otherMax = 50
	}

	periodMax := map[string]int{
		PeriodDaily:  dailyMax,
		PeriodWeekly: weeklyMax,
		PeriodOther:  otherMax,
	}

	buckets := map[string]*periodBucket{
		PeriodDaily:  newPeriodBucket(dailyMax),
		PeriodWeekly: newPeriodBucket(weeklyMax),
		PeriodOther:  newPeriodBucket(otherMax),
	}

	return &History{
		max:        max,
		periodMax:  periodMax,
		defaultMax: otherMax,
		buckets:    buckets,
		separated:  true, // Use separated storage by default
	}
}

func (h *History) EnablePersistence(filePath string) error {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return nil
	}

	// Parse base directory and filename
	h.baseDir = filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	ext := filepath.Ext(baseName)
	h.baseName = strings.TrimSuffix(baseName, ext)

	if err := os.MkdirAll(h.baseDir, 0o755); err != nil {
		return err
	}

	// Check if old unified file exists and needs migration
	if _, err := os.Stat(filePath); err == nil {
		// Old unified file exists - attempt migration
		if err := h.migrateFromUnified(filePath); err != nil {
			log.Printf("signal history migration failed: %v, falling back to unified storage", err)
			// Fall back to legacy unified storage
			h.separated = false
			return h.enableLegacyPersistence(filePath)
		}
		h.migrated = true
	}

	// Enable persistence for each bucket
	h.bucketsMu.Lock()
	for periodKey, bucket := range h.buckets {
		bucketFile := h.getPeriodFilePath(periodKey)
		if err := bucket.enablePersistence(bucketFile); err != nil {
			log.Printf("signal history: failed to enable persistence for period %s: %v", periodKey, err)
		}
	}
	h.bucketsMu.Unlock()

	return nil
}

// getPeriodFilePath returns the file path for a specific period bucket.
func (h *History) getPeriodFilePath(periodKey string) string {
	return filepath.Join(h.baseDir, h.baseName+"_"+periodKey+".jsonl")
}

// enablePersistence enables persistence for a single bucket.
func (b *periodBucket) enablePersistence(filePath string) error {
	b.fileMu.Lock()
	defer b.fileMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Create empty file
			f2, err2 := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err2 != nil {
				return err2
			}
			_ = f2.Close()
			b.filePath = filePath
			b.fileLines = 0
			return nil
		}
		return err
	}
	defer f.Close()

	// Load existing signals from file
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	capHint := 1024
	if b.max < capHint {
		capHint = b.max
	}
	loaded := make([]Signal, 0, capHint)
	lines := 0
	for scanner.Scan() {
		lines++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s Signal
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			continue
		}
		loaded = append(loaded, s)
		if len(loaded) > b.max {
			loaded = loaded[len(loaded)-b.max:]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	loadedUpper := make([]string, len(loaded))
	for i := range loaded {
		loadedUpper[i] = strings.ToUpper(loaded[i].Symbol)
	}

	b.mu.Lock()
	b.signals = loaded
	b.symbolsUpper = loadedUpper
	b.mu.Unlock()

	b.filePath = filePath
	b.fileLines = lines

	// Compact if needed
	if b.fileLines > b.max*2 {
		snapshot := make([]Signal, len(loaded))
		copy(snapshot, loaded)
		if err := b.compactFile(snapshot); err == nil {
			b.fileLines = len(snapshot)
		}
	}

	return nil
}

// migrateFromUnified migrates signals from the old unified file to period-separated files.
func (h *History) migrateFromUnified(unifiedPath string) error {
	f, err := os.Open(unifiedPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read all signals from unified file
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	// Group signals by period
	signalsByPeriod := make(map[string][]Signal)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s Signal
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			continue
		}
		periodKey := normalizePeriod(s.Period)
		signalsByPeriod[periodKey] = append(signalsByPeriod[periodKey], s)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Write signals to period-specific files
	for periodKey, signals := range signalsByPeriod {
		periodFile := h.getPeriodFilePath(periodKey)

		// Ensure bucket exists
		h.bucketsMu.Lock()
		if _, ok := h.buckets[periodKey]; !ok {
			h.buckets[periodKey] = newPeriodBucket(h.defaultMax)
		}
		bucket := h.buckets[periodKey]
		h.bucketsMu.Unlock()

		// Trim to bucket capacity
		if len(signals) > bucket.max {
			signals = signals[len(signals)-bucket.max:]
		}

		// Write to file
		tmpFile := periodFile + ".tmp"
		fw, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		bw := bufio.NewWriter(fw)
		enc := json.NewEncoder(bw)
		for _, s := range signals {
			if err := enc.Encode(s); err != nil {
				_ = bw.Flush()
				_ = fw.Close()
				return err
			}
		}
		if err := bw.Flush(); err != nil {
			_ = fw.Close()
			return err
		}
		if err := fw.Close(); err != nil {
			return err
		}
		if err := os.Rename(tmpFile, periodFile); err != nil {
			return err
		}

		log.Printf("signal history: migrated %d signals to %s", len(signals), periodFile)
	}

	// Rename old unified file
	migratedPath := unifiedPath + ".migrated"
	if err := os.Rename(unifiedPath, migratedPath); err != nil {
		log.Printf("signal history: failed to rename old file: %v", err)
		// Not a fatal error - migration succeeded
	} else {
		log.Printf("signal history: renamed old file to %s", migratedPath)
	}

	return nil
}

// enableLegacyPersistence enables the old unified persistence mode (fallback).
func (h *History) enableLegacyPersistence(filePath string) error {
	h.fileMu.Lock()
	defer h.fileMu.Unlock()

	f, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f2, err2 := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err2 != nil {
				return err2
			}
			_ = f2.Close()
			h.filePath = filePath
			h.fileLines = 0
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	capHint := 1024
	if h.max < capHint {
		capHint = h.max
	}
	loaded := make([]Signal, 0, capHint)
	lines := 0
	for scanner.Scan() {
		lines++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s Signal
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			continue
		}
		loaded = append(loaded, s)
		if len(loaded) > h.max {
			loaded = loaded[len(loaded)-h.max:]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	loadedUpper := make([]string, len(loaded))
	for i := range loaded {
		loadedUpper[i] = strings.ToUpper(loaded[i].Symbol)
	}

	h.mu.Lock()
	h.signals = loaded
	h.symbolsUpper = loadedUpper
	h.mu.Unlock()

	h.filePath = filePath
	h.fileLines = lines

	if h.fileLines > h.max*2 {
		snapshot := make([]Signal, len(loaded))
		copy(snapshot, loaded)
		if err := h.compactLocked(snapshot); err == nil {
			h.fileLines = len(snapshot)
		}
	}

	return nil
}

func (h *History) Add(s Signal) {
	// Use period-separated storage
	if h.separated {
		h.addToBucket(s)
		return
	}

	// Legacy unified storage (fallback mode)
	if h.filePath == "" {
		upper := strings.ToUpper(s.Symbol)
		h.mu.Lock()
		h.signals = append(h.signals, s)
		h.symbolsUpper = append(h.symbolsUpper, upper)
		if len(h.signals) > h.max {
			h.signals = h.signals[len(h.signals)-h.max:]
			h.symbolsUpper = h.symbolsUpper[len(h.symbolsUpper)-h.max:]
		}
		h.mu.Unlock()
		return
	}

	h.fileMu.Lock()
	defer h.fileMu.Unlock()
	upper := strings.ToUpper(s.Symbol)

	h.mu.Lock()
	h.signals = append(h.signals, s)
	h.symbolsUpper = append(h.symbolsUpper, upper)
	if len(h.signals) > h.max {
		h.signals = h.signals[len(h.signals)-h.max:]
		h.symbolsUpper = h.symbolsUpper[len(h.symbolsUpper)-h.max:]
	}
	h.mu.Unlock()

	if err := h.appendLocked(s); err == nil {
		h.fileLines += 1
		if h.fileLines > h.max*2 {
			h.mu.RLock()
			snapshot := make([]Signal, len(h.signals))
			copy(snapshot, h.signals)
			h.mu.RUnlock()
			if err := h.compactLocked(snapshot); err == nil {
				h.fileLines = len(snapshot)
			}
		}
	}
}

// addToBucket adds a signal to the appropriate period bucket.
func (h *History) addToBucket(s Signal) {
	periodKey := normalizePeriod(s.Period)

	h.bucketsMu.RLock()
	bucket, ok := h.buckets[periodKey]
	h.bucketsMu.RUnlock()

	if !ok {
		// Create bucket for unknown period on demand
		h.bucketsMu.Lock()
		bucket, ok = h.buckets[periodKey]
		if !ok {
			bucket = newPeriodBucket(h.defaultMax)
			h.buckets[periodKey] = bucket
		}
		h.bucketsMu.Unlock()
	}

	upper := strings.ToUpper(s.Symbol)

	// Memory-only mode
	if bucket.filePath == "" {
		bucket.mu.Lock()
		bucket.signals = append(bucket.signals, s)
		bucket.symbolsUpper = append(bucket.symbolsUpper, upper)
		if len(bucket.signals) > bucket.max {
			bucket.signals = bucket.signals[len(bucket.signals)-bucket.max:]
			bucket.symbolsUpper = bucket.symbolsUpper[len(bucket.symbolsUpper)-bucket.max:]
		}
		bucket.mu.Unlock()
		return
	}

	// Persistence mode
	bucket.fileMu.Lock()
	defer bucket.fileMu.Unlock()

	bucket.mu.Lock()
	bucket.signals = append(bucket.signals, s)
	bucket.symbolsUpper = append(bucket.symbolsUpper, upper)
	if len(bucket.signals) > bucket.max {
		bucket.signals = bucket.signals[len(bucket.signals)-bucket.max:]
		bucket.symbolsUpper = bucket.symbolsUpper[len(bucket.symbolsUpper)-bucket.max:]
	}
	bucket.mu.Unlock()

	if err := bucket.appendToFile(s); err == nil {
		bucket.fileLines++
		if bucket.fileLines > bucket.max*2 {
			bucket.mu.RLock()
			snapshot := make([]Signal, len(bucket.signals))
			copy(snapshot, bucket.signals)
			bucket.mu.RUnlock()
			if err := bucket.compactFile(snapshot); err == nil {
				bucket.fileLines = len(snapshot)
			}
		}
	}
}

// appendToFile appends a signal to the bucket's file.
func (b *periodBucket) appendToFile(s Signal) error {
	f, err := os.OpenFile(b.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(s); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// compactFile compacts the bucket's file with the given snapshot.
func (b *periodBucket) compactFile(snapshot []Signal) error {
	tmp := b.filePath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	for _, s := range snapshot {
		if err := enc.Encode(s); err != nil {
			_ = bw.Flush()
			_ = f.Close()
			return err
		}
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, b.filePath)
}

func (h *History) appendLocked(s Signal) error {
	f, err := os.OpenFile(h.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(s); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func (h *History) compactLocked(snapshot []Signal) error {
	tmp := h.filePath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	for _, s := range snapshot {
		if err := enc.Encode(s); err != nil {
			_ = bw.Flush()
			_ = f.Close()
			return err
		}
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, h.filePath)
}

func (h *History) Query(symbolContains, period, level, direction, source string, limit int) []Signal {
	if limit <= 0 {
		limit = 200
	}
	if limit > 4000 {
		limit = 4000
	}

	// Use period-separated query
	if h.separated {
		return h.queryFromBuckets(symbolContains, period, level, direction, source, limit)
	}

	// Legacy unified query
	symbolContains = strings.TrimSpace(symbolContains)
	period = strings.ToLower(strings.TrimSpace(period))
	level = strings.TrimSpace(level)
	direction = strings.ToLower(strings.TrimSpace(direction))
	source = strings.TrimSpace(source)
	symbolContainsUpper := strings.ToUpper(symbolContains)

	var levelSet map[string]struct{}
	if level != "" {
		if strings.Contains(level, ",") {
			levelSet = make(map[string]struct{})
			for _, p := range strings.Split(level, ",") {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				levelSet[strings.ToUpper(p)] = struct{}{}
			}
			level = ""
		} else {
			level = strings.ToUpper(level)
		}
	}

	h.mu.RLock()
	res := make([]Signal, 0, limit)
	for i := len(h.signals) - 1; i >= 0 && len(res) < limit; i-- {
		s := h.signals[i]
		if symbolContainsUpper != "" {
			if !strings.Contains(h.symbolsUpper[i], symbolContainsUpper) {
				continue
			}
		}
		if period != "" && s.Period != period {
			continue
		}
		if level != "" && s.Level != level {
			continue
		}
		if levelSet != nil {
			if _, ok := levelSet[s.Level]; !ok {
				continue
			}
		}
		if direction != "" && s.Direction != direction {
			continue
		}
		if source != "" && !strings.EqualFold(s.Source, source) {
			continue
		}
		res = append(res, s)
	}
	h.mu.RUnlock()
	return res
}

// queryFromBuckets queries signals from period-separated buckets.
func (h *History) queryFromBuckets(symbolContains, period, level, direction, source string, limit int) []Signal {
	symbolContains = strings.TrimSpace(symbolContains)
	period = strings.ToLower(strings.TrimSpace(period))
	level = strings.TrimSpace(level)
	direction = strings.ToLower(strings.TrimSpace(direction))
	source = strings.TrimSpace(source)
	symbolContainsUpper := strings.ToUpper(symbolContains)

	var levelSet map[string]struct{}
	if level != "" {
		if strings.Contains(level, ",") {
			levelSet = make(map[string]struct{})
			for _, p := range strings.Split(level, ",") {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				levelSet[strings.ToUpper(p)] = struct{}{}
			}
			level = ""
		} else {
			level = strings.ToUpper(level)
		}
	}

	// Determine which buckets to query
	var bucketsToQuery []*periodBucket
	var periodKey string
	h.bucketsMu.RLock()
	if period != "" {
		// Query only the specific period bucket
		periodKey = normalizePeriod(period)
		if bucket, ok := h.buckets[periodKey]; ok {
			bucketsToQuery = []*periodBucket{bucket}
		}
	} else {
		// Query all buckets
		for _, bucket := range h.buckets {
			bucketsToQuery = append(bucketsToQuery, bucket)
		}
	}
	h.bucketsMu.RUnlock()

	if len(bucketsToQuery) == 0 {
		return []Signal{}
	}

	// Collect matching signals from all relevant buckets
	var allMatches []Signal
	for _, bucket := range bucketsToQuery {
		bucket.mu.RLock()
		for i := len(bucket.signals) - 1; i >= 0; i-- {
			s := bucket.signals[i]
			if symbolContainsUpper != "" {
				if !strings.Contains(bucket.symbolsUpper[i], symbolContainsUpper) {
					continue
				}
			}
			// Period filter: when querying with period, check normalized period matches
			// (bucket selection already filters, but signals may have different period strings)
			if periodKey != "" && normalizePeriod(s.Period) != periodKey {
				continue
			}
			if level != "" && s.Level != level {
				continue
			}
			if levelSet != nil {
				if _, ok := levelSet[s.Level]; !ok {
					continue
				}
			}
			if direction != "" && s.Direction != direction {
				continue
			}
			if source != "" && !strings.EqualFold(s.Source, source) {
				continue
			}
			allMatches = append(allMatches, s)
		}
		bucket.mu.RUnlock()
	}

	// Sort by triggered_at descending (newest first)
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].TriggeredAt.After(allMatches[j].TriggeredAt)
	})

	// Apply limit
	if len(allMatches) > limit {
		allMatches = allMatches[:limit]
	}

	return allMatches
}

// Count returns the number of signals in history.
func (h *History) Count() int {
	// Use period-separated count
	if h.separated {
		total := 0
		h.bucketsMu.RLock()
		for _, bucket := range h.buckets {
			bucket.mu.RLock()
			total += len(bucket.signals)
			bucket.mu.RUnlock()
		}
		h.bucketsMu.RUnlock()
		return total
	}

	// Legacy unified count
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.signals)
}

// SymbolCount returns the number of unique symbols in history.
func (h *History) SymbolCount() int {
	// Use period-separated count
	if h.separated {
		seen := make(map[string]struct{})
		h.bucketsMu.RLock()
		for _, bucket := range h.buckets {
			bucket.mu.RLock()
			for _, s := range bucket.signals {
				seen[s.Symbol] = struct{}{}
			}
			bucket.mu.RUnlock()
		}
		h.bucketsMu.RUnlock()
		return len(seen)
	}

	// Legacy unified count
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[string]struct{})
	for _, s := range h.signals {
		seen[s.Symbol] = struct{}{}
	}
	return len(seen)
}
