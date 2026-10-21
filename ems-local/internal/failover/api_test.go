package failover

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	emsSync "github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestAPI(t *testing.T) {
	pub := emsSync.NewStubPublisher()
	fm := NewFailoverManager(pub, "test-site", "ems")
	hb := NewHeartbeatPublisher(pub, "test-site", "ems", 0)

	mux := http.NewServeMux()
	RegisterAPI(mux, fm, hb)

	// Test GET /api/failover/status
	req, err := http.NewRequest(http.MethodGet, "/api/failover/status", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	var resp failoverStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.State != "LOCAL_ACTIVE" {
		t.Errorf("expected state LOCAL_ACTIVE, got %s", resp.State)
	}
	if resp.SiteID != "test-site" {
		t.Errorf("expected site_id test-site, got %s", resp.SiteID)
	}

	// Test POST /api/failover/recover
	reqPost, err := http.NewRequest(http.MethodPost, "/api/failover/recover", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rrPost := httptest.NewRecorder()
	mux.ServeHTTP(rrPost, reqPost)

	if rrPost.Code != http.StatusAccepted {
		t.Errorf("expected status Accepted, got %v", rrPost.Code)
	}

	if fm.State() != FailoverRecovery {
		t.Errorf("expected state to be recovery, got %v", fm.State())
	}
}
