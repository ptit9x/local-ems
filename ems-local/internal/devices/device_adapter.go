package devices

import (
	"sync"
	"time"
)

// DeviceAdapter is the high-level interface for all expandable device adapters.
// It extends the base Device interface with protocol-aware capabilities.
type DeviceAdapter interface {
	Device
	// ProtocolType returns the communication protocol ("modbus", "api", "ocpp").
	ProtocolType() string
	// Vendor returns the device vendor/brand name.
	Vendor() string
	// Health returns the current health status of the device connection.
	Health() DeviceHealth
}

// DeviceHealth represents the health/connection status of a device.
type DeviceHealth struct {
	Connected       bool
	LastSeen        time.Time
	ErrorCount      int
	ConsecutiveErrs int
	LastError       string
}

// HealthTracker tracks connection health for any device.
type HealthTracker struct {
	mu              sync.RWMutex
	connected       bool
	lastSeen        time.Time
	errorCount      int
	consecutiveErrs int
	lastError       string
}

// NewHealthTracker creates a new HealthTracker instance.
func NewHealthTracker() *HealthTracker {
	return &HealthTracker{}
}

// RecordSuccess marks connection as healthy, updates the last seen timestamp,
// and resets consecutive errors.
func (h *HealthTracker) RecordSuccess() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connected = true
	h.lastSeen = time.Now()
	h.consecutiveErrs = 0
}

// RecordError increments error counts and updates the last error message.
func (h *HealthTracker) RecordError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.errorCount++
	h.consecutiveErrs++
	if err != nil {
		h.lastError = err.Error()
	}
}

// Health returns current health snapshot.
func (h *HealthTracker) Health() DeviceHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return DeviceHealth{
		Connected:       h.connected,
		LastSeen:        h.lastSeen,
		ErrorCount:      h.errorCount,
		ConsecutiveErrs: h.consecutiveErrs,
		LastError:       h.lastError,
	}
}

// IsHealthy returns true if connected and consecutiveErrs < 3.
func (h *HealthTracker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.connected && h.consecutiveErrs < 3
}

// SetConnected explicitly sets the connection status.
func (h *HealthTracker) SetConnected(connected bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connected = connected
}

// Reset resets all health tracking metrics to their initial state.
func (h *HealthTracker) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connected = false
	h.lastSeen = time.Time{}
	h.errorCount = 0
	h.consecutiveErrs = 0
	h.lastError = ""
}
