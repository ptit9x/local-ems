package store_test

import (
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/store"
)

func TestRetentionPolicy(t *testing.T) {
	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer s.Close()

	now := time.Now()

	// Insert 10 records: 5 older (will be synced), 5 newer (will be unsynced)
	for i := 0; i < 10; i++ {
		record := store.CycleRecord{
			Timestamp: now.Add(-time.Duration(20-i) * time.Hour), // older records first
			SolarW:    i * 100,
		}
		if err := s.Insert(record); err != nil {
			t.Fatalf("failed to insert: %v", err)
		}
	}

	count, err := s.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 10 {
		t.Fatalf("expected 10 records, got %d", count)
	}

	// Mark the 5 oldest records as synced
	syncTime := now.Add(-15*time.Hour - 30*time.Minute) // Covers records at -20h through -16h (5 records)
	if err := s.MarkCyclesSynced(syncTime); err != nil {
		t.Fatalf("failed to mark synced: %v", err)
	}

	// PruneSynced with 0 days (meaning prune everything that is synced and < now)
	deleted, err := s.PruneSynced(0)
	if err != nil {
		t.Fatalf("failed to prune: %v", err)
	}
	if deleted != 5 {
		t.Fatalf("expected to delete 5 synced records, deleted %d", deleted)
	}

	count, err = s.Count()
	if err != nil {
		t.Fatalf("failed to count after prune: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5 unsynced records to remain, got %d", count)
	}
}
