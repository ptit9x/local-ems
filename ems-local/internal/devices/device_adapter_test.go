package devices

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestHealthTracker_InitialState(t *testing.T) {
	tracker := NewHealthTracker()

	health := tracker.Health()
	if health.Connected {
		t.Errorf("expected Connected to be false initially, got true")
	}
	if health.ErrorCount != 0 {
		t.Errorf("expected ErrorCount to be 0, got %d", health.ErrorCount)
	}
	if health.ConsecutiveErrs != 0 {
		t.Errorf("expected ConsecutiveErrs to be 0, got %d", health.ConsecutiveErrs)
	}
	if health.LastError != "" {
		t.Errorf("expected LastError to be empty, got %q", health.LastError)
	}
	if !health.LastSeen.IsZero() {
		t.Errorf("expected LastSeen to be zero time, got %v", health.LastSeen)
	}
	if tracker.IsHealthy() {
		t.Errorf("expected IsHealthy() to be false initially")
	}
}

func TestHealthTracker_RecordSuccess(t *testing.T) {
	tracker := NewHealthTracker()

	before := time.Now()
	tracker.RecordSuccess()
	after := time.Now()

	health := tracker.Health()
	if !health.Connected {
		t.Errorf("expected Connected to be true after RecordSuccess")
	}
	if health.ConsecutiveErrs != 0 {
		t.Errorf("expected ConsecutiveErrs to be 0, got %d", health.ConsecutiveErrs)
	}
	if health.LastSeen.Before(before) || health.LastSeen.After(after) {
		t.Errorf("expected LastSeen to be between %v and %v, got %v", before, after, health.LastSeen)
	}
	if !tracker.IsHealthy() {
		t.Errorf("expected IsHealthy() to be true after RecordSuccess")
	}
}

func TestHealthTracker_RecordError(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	err1 := errors.New("connection refused")
	tracker.RecordError(err1)

	health := tracker.Health()
	if health.ErrorCount != 1 {
		t.Errorf("expected ErrorCount to be 1, got %d", health.ErrorCount)
	}
	if health.ConsecutiveErrs != 1 {
		t.Errorf("expected ConsecutiveErrs to be 1, got %d", health.ConsecutiveErrs)
	}
	if health.LastError != "connection refused" {
		t.Errorf("expected LastError to be 'connection refused', got %q", health.LastError)
	}

	err2 := errors.New("timeout reading registers")
	tracker.RecordError(err2)

	health = tracker.Health()
	if health.ErrorCount != 2 {
		t.Errorf("expected ErrorCount to be 2, got %d", health.ErrorCount)
	}
	if health.ConsecutiveErrs != 2 {
		t.Errorf("expected ConsecutiveErrs to be 2, got %d", health.ConsecutiveErrs)
	}
	if health.LastError != "timeout reading registers" {
		t.Errorf("expected LastError to be 'timeout reading registers', got %q", health.LastError)
	}
}

func TestHealthTracker_IsHealthy_Threshold(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	// 0 consecutive errors -> healthy
	if !tracker.IsHealthy() {
		t.Errorf("expected healthy with 0 consecutive errors")
	}

	// 1 consecutive error -> still healthy (< 3)
	tracker.RecordError(errors.New("err 1"))
	if !tracker.IsHealthy() {
		t.Errorf("expected healthy with 1 consecutive error (< 3)")
	}

	// 2 consecutive errors -> still healthy (< 3)
	tracker.RecordError(errors.New("err 2"))
	if !tracker.IsHealthy() {
		t.Errorf("expected healthy with 2 consecutive errors (< 3)")
	}

	// 3 consecutive errors -> unhealthy (not < 3)
	tracker.RecordError(errors.New("err 3"))
	if tracker.IsHealthy() {
		t.Errorf("expected unhealthy with 3 consecutive errors")
	}

	// 4 consecutive errors -> unhealthy
	tracker.RecordError(errors.New("err 4"))
	if tracker.IsHealthy() {
		t.Errorf("expected unhealthy with 4 consecutive errors")
	}
}

func TestHealthTracker_Recovery(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	for i := 0; i < 3; i++ {
		tracker.RecordError(errors.New("transient error"))
	}

	if tracker.IsHealthy() {
		t.Fatalf("expected tracker to be unhealthy after 3 errors")
	}

	// Recovery via RecordSuccess
	tracker.RecordSuccess()

	if !tracker.IsHealthy() {
		t.Errorf("expected tracker to recover and be healthy after RecordSuccess")
	}

	health := tracker.Health()
	if health.ConsecutiveErrs != 0 {
		t.Errorf("expected ConsecutiveErrs reset to 0, got %d", health.ConsecutiveErrs)
	}
	if health.ErrorCount != 3 {
		t.Errorf("expected cumulative ErrorCount to remain 3, got %d", health.ErrorCount)
	}
}

func TestHealthTracker_RecordError_NilError(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	// Recording nil error should increment count without panicking
	tracker.RecordError(nil)

	health := tracker.Health()
	if health.ErrorCount != 1 {
		t.Errorf("expected ErrorCount to be 1, got %d", health.ErrorCount)
	}
	if health.ConsecutiveErrs != 1 {
		t.Errorf("expected ConsecutiveErrs to be 1, got %d", health.ConsecutiveErrs)
	}
	if health.LastError != "" {
		t.Errorf("expected LastError to remain empty when err is nil, got %q", health.LastError)
	}
}

func TestHealthTracker_SetConnected(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	if !tracker.IsHealthy() {
		t.Fatalf("expected healthy")
	}

	// Explicitly set connected to false
	tracker.SetConnected(false)
	if tracker.Health().Connected {
		t.Errorf("expected Connected to be false")
	}
	if tracker.IsHealthy() {
		t.Errorf("expected IsHealthy() to be false when Connected is false")
	}

	// Explicitly set connected to true
	tracker.SetConnected(true)
	if !tracker.Health().Connected {
		t.Errorf("expected Connected to be true")
	}
	if !tracker.IsHealthy() {
		t.Errorf("expected IsHealthy() to be true when Connected is true and consecutive errors < 3")
	}
}

func TestHealthTracker_Reset(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()
	tracker.RecordError(errors.New("some error"))

	tracker.Reset()

	health := tracker.Health()
	if health.Connected {
		t.Errorf("expected Connected to be false after Reset")
	}
	if health.ErrorCount != 0 {
		t.Errorf("expected ErrorCount to be 0 after Reset, got %d", health.ErrorCount)
	}
	if health.ConsecutiveErrs != 0 {
		t.Errorf("expected ConsecutiveErrs to be 0 after Reset, got %d", health.ConsecutiveErrs)
	}
	if health.LastError != "" {
		t.Errorf("expected LastError to be empty after Reset, got %q", health.LastError)
	}
	if !health.LastSeen.IsZero() {
		t.Errorf("expected LastSeen to be zero after Reset, got %v", health.LastSeen)
	}
	if tracker.IsHealthy() {
		t.Errorf("expected IsHealthy() to be false after Reset")
	}
}

func TestHealthTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewHealthTracker()
	var wg sync.WaitGroup

	iterations := 500

	// Writer goroutines recording success
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tracker.RecordSuccess()
		}
	}()

	// Writer goroutines recording errors
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			tracker.RecordError(errors.New("concurrent error"))
		}
	}()

	// Reader goroutines checking health
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.Health()
		}
	}()

	// Reader goroutines checking is healthy
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = tracker.IsHealthy()
		}
	}()

	wg.Wait()
}

// mockDeviceAdapter verifies that a type can implement DeviceAdapter and satisfies Device interface.
type mockDeviceAdapter struct {
	mockDevice
	protocolType string
	vendor       string
	tracker      *HealthTracker
}

func (m *mockDeviceAdapter) ProtocolType() string {
	return m.protocolType
}

func (m *mockDeviceAdapter) Vendor() string {
	return m.vendor
}

func (m *mockDeviceAdapter) Health() DeviceHealth {
	return m.tracker.Health()
}

func TestDeviceAdapter_InterfaceImplementation(t *testing.T) {
	tracker := NewHealthTracker()
	tracker.RecordSuccess()

	adapter := &mockDeviceAdapter{
		mockDevice:   mockDevice{id: "pcs-01", devType: DeviceTypePCS},
		protocolType: "modbus",
		vendor:       "sungrow",
		tracker:      tracker,
	}

	// Verify DeviceAdapter implements Device
	var dev Device = adapter
	if dev.ID() != "pcs-01" {
		t.Errorf("expected ID 'pcs-01', got %s", dev.ID())
	}
	if dev.Type() != DeviceTypePCS {
		t.Errorf("expected Type PCS, got %s", dev.Type())
	}

	// Verify DeviceAdapter methods
	var devAdapter DeviceAdapter = adapter
	if devAdapter.ProtocolType() != "modbus" {
		t.Errorf("expected ProtocolType 'modbus', got %s", devAdapter.ProtocolType())
	}
	if devAdapter.Vendor() != "sungrow" {
		t.Errorf("expected Vendor 'sungrow', got %s", devAdapter.Vendor())
	}
	h := devAdapter.Health()
	if !h.Connected {
		t.Errorf("expected Health.Connected to be true")
	}

	// Verify registration in DeviceRegistry works with DeviceAdapter
	registry := NewDeviceRegistry()
	if err := registry.Register(adapter); err != nil {
		t.Fatalf("failed to register DeviceAdapter in DeviceRegistry: %v", err)
	}

	retrieved, exists := registry.Get("pcs-01")
	if !exists {
		t.Fatalf("expected device to exist in registry")
	}
	if retrieved.ID() != "pcs-01" {
		t.Errorf("expected retrieved device ID 'pcs-01', got %s", retrieved.ID())
	}

	retrievedAdapter, ok := retrieved.(DeviceAdapter)
	if !ok {
		t.Fatalf("expected retrieved device to implement DeviceAdapter")
	}
	if retrievedAdapter.Vendor() != "sungrow" {
		t.Errorf("expected Vendor 'sungrow', got %s", retrievedAdapter.Vendor())
	}
}
