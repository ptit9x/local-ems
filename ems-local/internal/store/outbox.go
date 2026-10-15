package store

import (
	"database/sql"
	"fmt"
	"time"
)

// OutboxRecord represents a message to be synced to the cloud.
type OutboxRecord struct {
	ID        int64
	Payload   string
	CreatedAt time.Time
	SyncedAt  *time.Time
	Status    string
}

// InsertOutbox adds a new message to the outbox.
func (s *Store) InsertOutbox(payload string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO sync_outbox (payload, created_at, status)
		VALUES (?, ?, 'pending')`,
		payload, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

// GetPendingOutbox returns up to limit pending records.
func (s *Store) GetPendingOutbox(limit int) ([]OutboxRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query(`
		SELECT id, payload, created_at, synced_at, status
		FROM sync_outbox
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer rows.Close()

	var records []OutboxRecord
	for rows.Next() {
		var r OutboxRecord
		var ts string
		var syncedAtPtr *string
		if err := rows.Scan(&r.ID, &r.Payload, &ts, &syncedAtPtr, &r.Status); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		
		t, _ := time.Parse(time.RFC3339, ts)
		r.CreatedAt = t
		
		if syncedAtPtr != nil {
			syncedTime, _ := time.Parse(time.RFC3339, *syncedAtPtr)
			r.SyncedAt = &syncedTime
		}
		
		records = append(records, r)
	}

	return records, nil
}

// MarkOutboxSent marks a record as sent.
func (s *Store) MarkOutboxSent(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		UPDATE sync_outbox
		SET status = 'sent', synced_at = ?
		WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("mark outbox sent: %w", err)
	}
	return nil
}

// MarkOutboxAcked marks a record as acked.
func (s *Store) MarkOutboxAcked(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		UPDATE sync_outbox
		SET status = 'acked'
		WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("mark outbox acked: %w", err)
	}
	return nil
}

// PendingCount returns the number of pending outbox records.
func (s *Store) PendingCount() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sync_outbox WHERE status = 'pending'`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count pending outbox: %w", err)
	}
	return count, nil
}

// OldestUnsynced returns the timestamp of the oldest unsynced (not acked) record.
func (s *Store) OldestUnsynced() (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ts string
	err := s.db.QueryRow(`
		SELECT created_at FROM sync_outbox
		WHERE status != 'acked'
		ORDER BY created_at ASC LIMIT 1`).Scan(&ts)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("oldest unsynced query: %w", err)
	}
	
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}
	return t, nil
}

// PruneOutboxSynced deletes acked records older than the given duration.
func (s *Store) PruneOutboxSynced(olderThan time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().UTC().Add(-olderThan).Format(time.RFC3339)
	result, err := s.db.Exec(`
		DELETE FROM sync_outbox 
		WHERE status = 'acked' AND created_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune outbox: %w", err)
	}
	return result.RowsAffected()
}
