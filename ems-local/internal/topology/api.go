package topology

import (
	"encoding/json"
	"net/http"
)

func RegisterAPI(mux *http.ServeMux, mgr *Manager) {
	mux.HandleFunc("GET /api/topology", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mgr.Get())
	})

	mux.HandleFunc("PUT /api/topology", func(w http.ResponseWriter, r *http.Request) {
		var topo Topology
		if err := json.NewDecoder(r.Body).Decode(&topo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := mgr.Update(topo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mgr.Get())
	})

	mux.HandleFunc("GET /api/topology/devices", func(w http.ResponseWriter, r *http.Request) {
		mgr.mu.RLock()
		defer mgr.mu.RUnlock()
		devices := make([]DeviceInfo, 0, len(mgr.devices))
		for _, d := range mgr.devices {
			devices = append(devices, d)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(devices)
	})

	mux.HandleFunc("POST /api/topology/validate", func(w http.ResponseWriter, r *http.Request) {
		var topo Topology
		if err := json.NewDecoder(r.Body).Decode(&topo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mgr.mu.RLock()
		// We temporarily simulate an update to validate against current devices
		tempMgr := &Manager{
			topology: topo,
			devices:  mgr.devices,
		}
		mgr.mu.RUnlock()

		if err := tempMgr.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"valid"}`))
	})
}
