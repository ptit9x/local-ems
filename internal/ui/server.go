// Package ui provides the REST API and WebSocket server for the Local HMI.
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vmo/local-ems/internal/devices/aggregator"
	"github.com/vmo/local-ems/internal/engine/cycle"
	"github.com/vmo/local-ems/internal/health"
	"github.com/vmo/local-ems/internal/store"
)

// Server is the HTTP server for the Local EMS UI.
type Server struct {
	store      *store.Store
	monitor    *health.Monitor
	agg        *aggregator.Aggregator
	mux        *http.ServeMux
	httpServer *http.Server
	wsHub      *WSHub
	auth       AuthConfig
	sessions   *sessionStore
	lastResult *cycle.CycleResult
	mu         sync.RWMutex
}

// NewServer creates a new UI server.
func NewServer(addr string, st *store.Store, mon *health.Monitor, auth AuthConfig) *Server {
	s := &Server{
		store:    st,
		monitor:  mon,
		mux:      http.NewServeMux(),
		wsHub:    NewWSHub(),
		auth:     auth,
		sessions: newSessionStore(),
	}

	// Dashboard routes
	s.mux.HandleFunc("/live", s.servePage("live"))
	s.mux.HandleFunc("/history", s.servePage("history"))
	s.mux.HandleFunc("/settings/config", s.servePage("config"))
	s.mux.HandleFunc("/settings/health", s.servePage("health"))
	s.mux.HandleFunc("/settings/about", s.servePage("about"))
	s.mux.HandleFunc("/user", s.servePage("user"))
	s.mux.HandleFunc("/component/", s.handleComponent)

	// Root redirects to live page
	s.mux.HandleFunc("/", s.handleRoot)

	// API endpoints
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/history", s.handleHistory)
	s.mux.HandleFunc("/api/history/date", s.handleHistoryByDate)
	s.mux.HandleFunc("/api/history/summary", s.handleHistorySummary)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/devices", s.handleDevices)
	s.mux.HandleFunc("/ws", s.wsHub.HandleWS)

	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        authMiddleware(s.mux, s.auth, s.sessions),
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s
}

// SetAggregator sets the device aggregator for per-device API.
func (s *Server) SetAggregator(agg *aggregator.Aggregator) {
	s.agg = agg
}

// Start begins the HTTP server and WebSocket hub.
func (s *Server) Start() error {
	go s.wsHub.Run()
	log.Printf("UI server starting on %s", s.httpServer.Addr)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("UI server error: %v", err)
		}
	}()
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.wsHub.Stop()
	return s.httpServer.Shutdown(ctx)
}

// BroadcastCycleResult sends a cycle result to all WebSocket clients.
func (s *Server) BroadcastCycleResult(cr cycle.CycleResult) {
	s.mu.Lock()
	s.lastResult = &cr
	s.mu.Unlock()

	msg := cycleToJSON(cr)
	s.wsHub.Broadcast(msg)
}

// --- HTTP Handlers ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	lr := s.lastResult
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if lr == nil {
		json.NewEncoder(w).Encode(map[string]string{"status": "waiting_for_first_cycle"})
		return
	}

	json.NewEncoder(w).Encode(cycleToMap(*lr))
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	records, err := s.store.QueryRecent(300)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(records)
}

func (s *Server) handleHistoryByDate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	records, err := s.store.QueryByDate(date)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(records)
}

func (s *Server) handleHistorySummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	summary, err := s.store.DailySummary(date)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(summary)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := s.monitor.Status()
	stats := s.monitor.Stats()

	resp := map[string]interface{}{
		"status": status,
		"stats":  stats,
	}

	if !status.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.agg == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "no aggregator"})
		return
	}

	state := s.agg.Aggregate()
	json.NewEncoder(w).Encode(state)
}

func (s *Server) handleComponent(w http.ResponseWriter, r *http.Request) {
	// Extract component ID from /device/0/component/{id}
	compID := strings.TrimPrefix(r.URL.Path, "/component/")
	if compID == "" {
		http.Redirect(w, r, "/live", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.Replace(componentHTML, "/*COMP_ID*/", "'"+compID+"'", 1)
	fmt.Fprint(w, html)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/live", http.StatusFound)
}

func (s *Server) servePage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Inject the active page into the HTML via a JS variable
		html := strings.Replace(dashboardHTML, "/*ACTIVE_PAGE*/", "'"+page+"'", 1)
		fmt.Fprint(w, html)
	}
}

// --- Helpers ---

func cycleToMap(cr cycle.CycleResult) map[string]interface{} {
	return map[string]interface{}{
		"timestamp":   cr.Timestamp.Format(time.RFC3339),
		"solar_w":     cr.State.SolarPowerW,
		"load_w":      cr.State.LoadPowerW,
		"grid_w":      cr.State.GridPowerW,
		"ess_w":       cr.ResolvedPower.Setpoint,
		"desired_w":   cr.DesiredPower,
		"soc":         cr.State.BatterySOC,
		"temp_c":      cr.State.BatteryTempC,
		"clamped":     cr.ResolvedPower.Clamped,
		"constraints": cr.ResolvedPower.AppliedConstraints,
		"cycle_ms":    cr.CycleDurationMs,
	}
}

func cycleToJSON(cr cycle.CycleResult) []byte {
	data, _ := json.Marshal(cycleToMap(cr))
	return data
}
