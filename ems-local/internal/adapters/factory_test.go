package adapters

import (
	"testing"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/ocpp"
)

func testFactory() *Factory {
	factory := NewFactory(ocpp.NewCentralSystem(":0", nil))
	factory.ConnectOnCreate = false
	return factory
}

func TestFactoryBuildSelectsConfiguredAdapters(t *testing.T) {
	cfg := &config.Config{Devices: config.DevicesConfig{
		Meter: []config.MeterDeviceConfig{{ID: "meter-pcc", Host: "127.0.0.1", Port: 502, Unit: 1, Role: "grid"}},
		BMS:   []config.BMSDeviceConfig{{ID: "bms-1", Host: "127.0.0.1", Port: 502, Unit: 2}},
		PCS:   []config.PCSDeviceConfig{{ID: "pcs-1", Host: "127.0.0.1", Port: 502, Unit: 3, Brand: "liteon", CommandMode: "individual"}},
		EVCharger: []config.EVChargerDeviceConfig{
			{ID: "ev-modbus", Host: "127.0.0.1", Port: 502, Unit: 4, Protocol: "modbus"},
			{ID: "ev-ocpp", Protocol: "ocpp"},
		},
	}}
	registry, err := testFactory().Build(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	if got := len(registry.All()); got != 5 {
		t.Fatalf("expected 5 adapters, got %d", got)
	}
	if got := len(registry.GetByType(devices.DeviceTypeEVCharger)); got != 2 {
		t.Fatalf("expected 2 EV adapters, got %d", got)
	}
	pcs, ok := registry.Get("pcs-1")
	if !ok {
		t.Fatal("PCS not registered")
	}
	adapter := pcs.(devices.SnapshotDevice)
	if adapter.Vendor() != "liteon" || adapter.ProtocolType() != "modbus" {
		t.Fatalf("unexpected PCS adapter metadata: vendor=%s protocol=%s", adapter.Vendor(), adapter.ProtocolType())
	}
	ev, ok := registry.Get("ev-ocpp")
	if !ok || ev.Protocol() != "ocpp" {
		t.Fatalf("OCPP adapter not selected: %#v", ev)
	}
}

func TestFactoryRejectsUnsupportedProtocol(t *testing.T) {
	_, err := testFactory().Create(devices.DeviceConfig{ID: "meter", Type: devices.DeviceTypeMeter, Host: "127.0.0.1", Port: 502, UnitID: 1, Protocol: "bacnet"})
	if err == nil {
		t.Fatal("expected unsupported protocol error")
	}
}

func TestFactoryRequiresOCPPCentralSystem(t *testing.T) {
	factory := NewFactory(nil)
	_, err := factory.Create(devices.DeviceConfig{ID: "ev-1", Type: devices.DeviceTypeEVCharger, Protocol: "ocpp"})
	if err == nil {
		t.Fatal("expected missing OCPP central system error")
	}
}

func TestManagedAdapterTracksPollFailures(t *testing.T) {
	adapter, err := testFactory().Create(devices.DeviceConfig{ID: "meter", Type: devices.DeviceTypeMeter, Host: "127.0.0.1", Port: 502, UnitID: 1, Protocol: "modbus"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := adapter.Poll(); err == nil {
			t.Fatal("expected disconnected poll to fail")
		}
	}
	if adapter.Status() != devices.DeviceStatusFault || adapter.Health().ConsecutiveErrs != 3 {
		t.Fatalf("expected fault after 3 failures: status=%s health=%+v", adapter.Status(), adapter.Health())
	}
}

func TestFactoryBuildIsTransactional(t *testing.T) {
	cfg := &config.Config{Devices: config.DevicesConfig{Meter: []config.MeterDeviceConfig{
		{ID: "duplicate", Host: "127.0.0.1", Port: 502, Unit: 1},
		{ID: "duplicate", Host: "127.0.0.1", Port: 502, Unit: 2},
	}}}
	if _, err := testFactory().Build(cfg); err == nil {
		t.Fatal("expected duplicate registration failure")
	}
}
