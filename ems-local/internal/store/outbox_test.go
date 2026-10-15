package store_test

import (
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/store"
)

func TestOutboxLifecycle(t *testing.T) {
	dbPath := ":memory:"
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Insert 2 records
	if err := s.InsertOutbox("payload1"); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if err := s.InsertOutbox("payload2"); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Check pending count
	count, err := s.PendingCount()
	if err != nil || count != 2 {
		t.Errorf("expected 2 pending, got %d (err: %v)", count, err)
	}

	// Get pending
	records, err := s.GetPendingOutbox(10)
	if err != nil || len(records) != 2 {
		t.Fatalf("expected 2 pending records, got %d (err: %v)", len(records), err)
	}
	
	if records[0].Payload != "payload1" {
		t.Errorf("expected payload1, got %s", records[0].Payload)
	}

	// Mark sent
	if err := s.MarkOutboxSent(records[0].ID); err != nil {
		t.Fatalf("mark sent failed: %v", err)
	}

	// Now pending count should be 1
	count, _ = s.PendingCount()
	if count != 1 {
		t.Errorf("expected 1 pending after mark sent, got %d", count)
	}

	// Mark acked
	if err := s.MarkOutboxAcked(records[0].ID); err != nil {
		t.Fatalf("mark acked failed: %v", err)
	}

	// Oldest unsynced should now point to records[1]
	oldest, err := s.OldestUnsynced()
	if err != nil || oldest.IsZero() {
		t.Errorf("oldest unsynced failed: %v, %v", oldest, err)
	}
}

func TestPruneOutboxSynced(t *testing.T) {
	s, _ := store.New(":memory:")
	defer s.Close()

	s.InsertOutbox("p1")
	s.InsertOutbox("p2")

	records, _ := s.GetPendingOutbox(10)
	s.MarkOutboxAcked(records[0].ID) // p1 acked

	// Prune older than 1 hour (won't prune because it's new)
	pruned, _ := s.PruneOutboxSynced(time.Hour)
	if pruned != 0 {
		t.Errorf("expected 0 pruned, got %d", pruned)
	}

	// Prune older than 0 (will prune acked)
	pruned, _ = s.PruneOutboxSynced(0) // wait, since it's 0, time.Now() > created_at
	// wait, it might prune it if created_at is slightly before Now(). Let's use -1s
	pruned, _ = s.PruneOutboxSynced(-time.Second) 
	if pruned != 1 {
		t.Errorf("expected 1 pruned, got %d", pruned)
	}
}
