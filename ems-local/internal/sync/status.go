package sync

import "time"

// SyncStatus represents the current state of cloud synchronization.
type SyncStatus struct {
	Connected      bool      `json:"connected"`
	LastSyncTime   time.Time `json:"last_sync_time"`
	PendingCount   int       `json:"pending_count"`
	OldestUnsynced time.Time `json:"oldest_unsynced"`
	TotalSynced    int64     `json:"total_synced"`
	ConnectionInfo string    `json:"connection_info,omitempty"`
}

// StatusProvider provides sync status information.
type StatusProvider interface {
	Status() SyncStatus
}
