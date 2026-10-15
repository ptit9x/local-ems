package sync

import (
	"encoding/json"
	"net/http"

	"github.com/vmo/ems-platform/ems-local/internal/store"
)

// RegisterAPI adds sync status endpoints to an HTTP mux.
func RegisterAPI(mux *http.ServeMux, provider StatusProvider, store *store.Store) {
	mux.HandleFunc("/api/sync/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := provider.Status()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/sync/disk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		usage, err := store.DiskUsage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(usage)
	})
}
