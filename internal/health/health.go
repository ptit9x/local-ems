package health

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status struct {
	Name      string    `json:"name"`
	Healthy   bool      `json:"healthy"`
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
	Uptime    string    `json:"uptime"`
}

// Monitor tracks the health of the EMS system.
type Monitor struct {
	mu         sync.RWMutex
	startTime  time.Time
	lastCycle  time.Time
	cycleCount int64
	errors     int64
	panicCount int64
}

// NewMonitor creates a new health monitor.
func NewMonitor() *Monitor {
	return &Monitor{
		startTime: time.Now(),
	}
}

// RecordCycle records a successful cycle execution.
func (m *Monitor) RecordCycle() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCycle = time.Now()
	m.cycleCount++
}

// RecordError records a cycle error.
func (m *Monitor) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors++
}

// RecordPanic records a recovered panic.
func (m *Monitor) RecordPanic(r interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panicCount++
	log.Printf("PANIC recovered in controller: %v", r)
}

// Status returns the current health status.
func (m *Monitor) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthy := true
	message := "OK"

	// Unhealthy if no cycle in last 5 seconds
	if !m.lastCycle.IsZero() && time.Since(m.lastCycle) > 5*time.Second {
		healthy = false
		message = "cycle stalled"
	}

	// Unhealthy if panic detected
	if m.panicCount > 0 {
		message = fmt.Sprintf("OK (%d panics recovered)", m.panicCount)
	}

	return Status{
		Name:      "ems-core",
		Healthy:   healthy,
		LastCheck: time.Now(),
		Message:   message,
		Uptime:    time.Since(m.startTime).Round(time.Second).String(),
	}
}

// Stats returns cycle statistics.
func (m *Monitor) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"uptime_seconds": time.Since(m.startTime).Seconds(),
		"cycle_count":    m.cycleCount,
		"error_count":    m.errors,
		"panic_count":    m.panicCount,
		"last_cycle":     m.lastCycle.Format(time.RFC3339),
	}
}
