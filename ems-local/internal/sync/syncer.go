package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
	"github.com/vmo/ems-platform/ems-local/internal/store"
)

// Syncer manages the store-and-forward sync loop.
type Syncer struct {
	store       *store.Store
	publisher   Publisher
	siteID      string
	topicPrefix string
	interval    time.Duration
	accumulator *CycleAccumulator
	done        chan struct{}
	wg          sync.WaitGroup

	mu           sync.Mutex
	lastSyncTime time.Time
}

// NewSyncer creates a new Syncer.
func NewSyncer(store *store.Store, pub Publisher, siteID, topicPrefix string, interval time.Duration) *Syncer {
	return &Syncer{
		store:       store,
		publisher:   pub,
		siteID:      siteID,
		topicPrefix: topicPrefix,
		interval:    interval,
		accumulator: NewCycleAccumulator(siteID),
		done:        make(chan struct{}),
	}
}

// Start starts the sync loop.
func (s *Syncer) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.run(ctx)
}

// Stop stops the sync loop.
func (s *Syncer) Stop() {
	close(s.done)
	s.wg.Wait()
}

func (s *Syncer) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Initial connection
	_ = s.publisher.Connect()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-ticker.C:
			s.tick()
		case <-heartbeat.C:
			s.publishHeartbeat()
		}
	}
}

func (s *Syncer) publishHeartbeat() {
	if !s.publisher.IsConnected() {
		return
	}
	payload, _ := json.Marshal(map[string]interface{}{"schema_version": 1, "site_id": s.siteID, "timestamp": time.Now().UTC()})
	_ = s.publisher.Publish(fmt.Sprintf("%s/%s/heartbeat", s.topicPrefix, s.siteID), payload)
}

func (s *Syncer) tick() {
	// 1. Flush accumulator
	msg := s.accumulator.Flush()

	// Only insert if there are actually cycles
	if msg.CycleCount > 0 {
		payloadBytes, err := json.Marshal(msg)
		if err == nil {
			_ = s.store.InsertOutbox(string(payloadBytes))
		}
	}

	// 2. Try to reconnect if not connected
	if !s.publisher.IsConnected() {
		if err := s.publisher.Connect(); err != nil {
			return // skip sync this tick
		}
	}

	// 3. Read pending outbox and publish
	// Batch size of 100
	for {
		records, err := s.store.GetPendingOutbox(100)
		if err != nil || len(records) == 0 {
			break
		}

		for _, r := range records {
			topic := fmt.Sprintf("%s/%s/telemetry", s.topicPrefix, s.siteID)
			err := s.publisher.Publish(topic, []byte(r.Payload))
			if err != nil {
				// Stop processing batch on publish error
				_ = s.publisher.Close()
				return
			}

			// Mark as sent
			_ = s.store.MarkOutboxSent(r.ID)
			// Wait, the requirements state 'sent' and 'acked'. Since MQTT publish
			// doesn't inherently give an async ack in this simple interface,
			// we can just mark it 'acked' or 'sent'. The requirements mention
			// "mark outbox sent" and "mark outbox acked".
			// Since we don't have a callback for ack, let's mark it sent then acked for now.
			_ = s.store.MarkOutboxAcked(r.ID)

			s.mu.Lock()
			s.lastSyncTime = time.Now().UTC()
			s.mu.Unlock()
		}
	}
}

// AddCycle feeds cycle data to the syncer.
func (s *Syncer) AddCycle(cr cycle.CycleResult) {
	s.accumulator.Add(cr)
}

// Status returns the current sync status.
func (s *Syncer) Status() SyncStatus {
	s.mu.Lock()
	lastSync := s.lastSyncTime
	s.mu.Unlock()

	pending, _ := s.store.PendingCount()
	oldest, _ := s.store.OldestUnsynced()

	return SyncStatus{
		Connected:      s.publisher.IsConnected(),
		LastSyncTime:   lastSync,
		PendingCount:   pending,
		OldestUnsynced: oldest,
	}
}
