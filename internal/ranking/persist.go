package ranking

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	rankingSubDir  = "ranking"
	snapshotsFile  = "snapshots.json"
)

// persistedData is the structure for persisted ranking data.
type persistedData struct {
	Snapshots []*Snapshot `json:"snapshots"`
	SavedAt   time.Time   `json:"saved_at"`
}

// Persist saves the current snapshots to disk.
func (s *Store) Persist() error {
	if s.dataDir == "" {
		return nil // No persistence configured
	}

	s.mu.RLock()
	data := persistedData{
		Snapshots: make([]*Snapshot, len(s.snapshots)),
		SavedAt:   time.Now(),
	}
	copy(data.Snapshots, s.snapshots)
	s.mu.RUnlock()

	// Create directory if needed
	dir := filepath.Join(s.dataDir, rankingSubDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write to temp file first, then rename for atomicity
	filePath := filepath.Join(dir, snapshotsFile)
	tempPath := filePath + ".tmp"

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, filePath)
}

// Load loads snapshots from disk.
func (s *Store) Load() error {
	if s.dataDir == "" {
		return nil // No persistence configured
	}

	filePath := filepath.Join(s.dataDir, rankingSubDir, snapshotsFile)

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No data file yet
		}
		return err
	}

	var data persistedData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Load snapshots and cleanup old ones
	s.snapshots = data.Snapshots
	s.cleanupLocked()

	log.Printf("ranking store: loaded %d snapshots from disk", len(s.snapshots))
	return nil
}
