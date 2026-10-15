package devices

import (
	"testing"
)

type mockDevice struct {
	id      string
	devType DeviceType
	closed  bool
}

func (m *mockDevice) ID() string           { return m.id }
func (m *mockDevice) Type() DeviceType     { return m.devType }
func (m *mockDevice) Protocol() string     { return "modbus-tcp" }
func (m *mockDevice) Status() DeviceStatus { return DeviceStatusOnline }
func (m *mockDevice) Poll() error          { return nil }
func (m *mockDevice) Close() error         { m.closed = true; return nil }

type mockReconfigurableDevice struct {
	mockDevice
	reconfigured bool
}

func (m *mockReconfigurableDevice) Reconfigure(newConfig DeviceConfig) error {
	m.reconfigured = true
	return nil
}

func TestDeviceRegistry_Register(t *testing.T) {
	r := NewDeviceRegistry()
	dev := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	err := r.Register(dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, exists := r.Get("dev1")
	if !exists {
		t.Fatalf("expected device to exist")
	}
	if got.ID() != "dev1" {
		t.Errorf("expected ID dev1, got %s", got.ID())
	}
}

func TestDeviceRegistry_RegisterDuplicate(t *testing.T) {
	r := NewDeviceRegistry()
	dev := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	_ = r.Register(dev)
	err := r.Register(dev)
	if err == nil {
		t.Fatalf("expected error on duplicate register, got nil")
	}
}

func TestDeviceRegistry_Unregister(t *testing.T) {
	r := NewDeviceRegistry()
	dev := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	_ = r.Register(dev)
	
	err := r.Unregister("dev1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	_, exists := r.Get("dev1")
	if exists {
		t.Fatalf("expected device to be removed")
	}
}

func TestDeviceRegistry_GetByType(t *testing.T) {
	r := NewDeviceRegistry()
	dev1 := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	dev2 := &mockDevice{id: "dev2", devType: DeviceTypePCS}
	dev3 := &mockDevice{id: "dev3", devType: DeviceTypeBMS}
	
	_ = r.Register(dev1)
	_ = r.Register(dev2)
	_ = r.Register(dev3)
	
	bmsDevices := r.GetByType(DeviceTypeBMS)
	if len(bmsDevices) != 2 {
		t.Errorf("expected 2 BMS devices, got %d", len(bmsDevices))
	}
}

func TestDeviceRegistry_All(t *testing.T) {
	r := NewDeviceRegistry()
	dev1 := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	dev2 := &mockDevice{id: "dev2", devType: DeviceTypePCS}
	
	_ = r.Register(dev1)
	_ = r.Register(dev2)
	
	all := r.All()
	if len(all) != 2 {
		t.Errorf("expected 2 devices, got %d", len(all))
	}
}

func TestDeviceRegistry_Reconfigure(t *testing.T) {
	r := NewDeviceRegistry()
	dev := &mockReconfigurableDevice{mockDevice: mockDevice{id: "dev1", devType: DeviceTypeBMS}}
	_ = r.Register(dev)
	
	err := r.Reconfigure("dev1", DeviceConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dev.reconfigured {
		t.Errorf("expected device to be reconfigured")
	}
}

func TestDeviceRegistry_Close(t *testing.T) {
	r := NewDeviceRegistry()
	dev1 := &mockDevice{id: "dev1", devType: DeviceTypeBMS}
	dev2 := &mockDevice{id: "dev2", devType: DeviceTypePCS}
	
	_ = r.Register(dev1)
	_ = r.Register(dev2)
	
	r.Close()
	
	if !dev1.closed {
		t.Errorf("expected dev1 to be closed")
	}
	if !dev2.closed {
		t.Errorf("expected dev2 to be closed")
	}
}
