package failover

import (
	"encoding/json"
	"net/http"
)

type failoverStatusResponse struct {
	State  string `json:"state"`
	SiteID string `json:"site_id"`
}

func RegisterAPI(mux *http.ServeMux, fm *FailoverManager, hb *HeartbeatPublisher) {
	mux.HandleFunc("/api/failover/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := failoverStatusResponse{
			State:  fm.State().String(),
			SiteID: fm.siteID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/failover/recover", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := fm.RequestRecovery()
		if err != nil {
			http.Error(w, "Failed to request recovery: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status": "recovery_requested"}`))
	})
}
