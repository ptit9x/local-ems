package sync_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/store"
	"github.com/vmo/ems-platform/ems-local/internal/sync"
)

type mockProvider struct {
	status sync.SyncStatus
}

func (m *mockProvider) Status() sync.SyncStatus {
	return m.status
}

func TestSyncStatusAPI(t *testing.T) {
	now := time.Now()
	provider := &mockProvider{
		status: sync.SyncStatus{
			Connected:      true,
			LastSyncTime:   now,
			PendingCount:   5,
			OldestUnsynced: now.Add(-1 * time.Hour),
			TotalSynced:    100,
			ConnectionInfo: "mock",
		},
	}

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer s.Close()

	mux := http.NewServeMux()
	sync.RegisterAPI(mux, provider, s)

	req, err := http.NewRequest("GET", "/api/sync/status", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response sync.SyncStatus
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.PendingCount != 5 {
		t.Errorf("expected PendingCount 5, got %d", response.PendingCount)
	}
	if !response.Connected {
		t.Errorf("expected Connected true, got false")
	}
}

func TestSyncDiskAPI(t *testing.T) {
	provider := &mockProvider{}

	s, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer s.Close()

	mux := http.NewServeMux()
	sync.RegisterAPI(mux, provider, s)

	req, err := http.NewRequest("GET", "/api/sync/disk", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response store.DiskUsage
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Memory DB returns 0 bytes
	if response.DatabaseBytes != 0 {
		t.Errorf("expected DatabaseBytes 0, got %d", response.DatabaseBytes)
	}
}
