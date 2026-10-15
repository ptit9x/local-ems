package config

import (
	"os"
	"testing"
)

func TestDefaultDev(t *testing.T) {
	cfg := DefaultDev()

	if len(cfg.Devices.BMS) != 2 {
		t.Errorf("expected 2 BMS, got %d", len(cfg.Devices.BMS))
	}
	if len(cfg.Devices.PCS) != 2 {
		t.Errorf("expected 2 PCS, got %d", len(cfg.Devices.PCS))
	}
	if len(cfg.Devices.Meter) != 2 {
		t.Errorf("expected 2 Meter, got %d", len(cfg.Devices.Meter))
	}
	if len(cfg.Devices.EVCharger) != 2 {
		t.Errorf("expected 2 EVCharger, got %d", len(cfg.Devices.EVCharger))
	}
	if err := cfg.validate(); err != nil {
		t.Errorf("default dev config should be valid: %v", err)
	}
	if cfg.Site.ID == "" {
		t.Error("expected site ID to be set")
	}
}

func TestDefaultDev_Topology(t *testing.T) {
	cfg := DefaultDev()

	if cfg.Topology.PCCMeter != "meter-0" {
		t.Errorf("expected pcc_meter=meter-0, got %s", cfg.Topology.PCCMeter)
	}
	if len(cfg.Topology.BESSGroups) != 2 {
		t.Errorf("expected 2 BESS groups, got %d", len(cfg.Topology.BESSGroups))
	}
}

func TestDefaultDev_SyncDefaults(t *testing.T) {
	cfg := DefaultDev()

	if cfg.Sync.Enabled {
		t.Error("expected sync disabled in dev")
	}
	if cfg.Sync.IntervalSeconds != 300 {
		t.Errorf("expected interval 300, got %d", cfg.Sync.IntervalSeconds)
	}
	if cfg.Sync.BufferDays != 7 {
		t.Errorf("expected buffer 7 days, got %d", cfg.Sync.BufferDays)
	}
}

func TestDefaultDev_StorageDefaults(t *testing.T) {
	cfg := DefaultDev()

	if cfg.Storage.RetentionDays != 7 {
		t.Errorf("expected retention 7 days, got %d", cfg.Storage.RetentionDays)
	}
	if !cfg.Storage.WALMode {
		t.Error("expected WAL mode enabled in dev")
	}
}

func TestLoadYAML(t *testing.T) {
	yaml := `
simulate: true
devices:
  bms:
    - { id: bms-0, host: 127.0.0.1, port: 5020, unit: 1 }
    - { id: bms-1, host: 127.0.0.1, port: 5020, unit: 2 }
  pcs:
    - { id: pcs-0, host: 127.0.0.1, port: 5020, unit: 3 }
    - { id: pcs-1, host: 127.0.0.1, port: 5020, unit: 4 }
  meter:
    - { id: meter-0, host: 127.0.0.1, port: 5020, unit: 5 }
    - { id: meter-1, host: 127.0.0.1, port: 5020, unit: 6 }
polling:
  fast_ms: 200
  standard_ms: 1000
  slow_ms: 5000
`
	f, err := os.CreateTemp("", "ems-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.Simulate {
		t.Error("expected simulate=true")
	}
	if len(cfg.Devices.BMS) != 2 {
		t.Errorf("expected 2 BMS, got %d", len(cfg.Devices.BMS))
	}
	if cfg.Devices.BMS[0].ID != "bms-0" {
		t.Errorf("expected bms-0, got %s", cfg.Devices.BMS[0].ID)
	}
	if cfg.FastInterval().Milliseconds() != 200 {
		t.Errorf("expected 200ms fast interval, got %v", cfg.FastInterval())
	}
	// Check that defaults were applied
	if cfg.Site.ID != "default-site" {
		t.Errorf("expected default site ID, got %s", cfg.Site.ID)
	}
	if cfg.Storage.RetentionDays != 7 {
		t.Errorf("expected default retention 7 days, got %d", cfg.Storage.RetentionDays)
	}
}

func TestValidation_DuplicateID(t *testing.T) {
	cfg := DefaultDev()
	cfg.Devices.BMS[1].ID = cfg.Devices.BMS[0].ID // duplicate
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for duplicate ID")
	}
}

func TestValidation_NoBMS(t *testing.T) {
	cfg := DefaultDev()
	cfg.Devices.BMS = nil
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for no BMS")
	}
}

func TestValidation_TopologyInvalidMeter(t *testing.T) {
	cfg := DefaultDev()
	cfg.Topology.PCCMeter = "nonexistent-meter"
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for invalid topology meter reference")
	}
}

func TestValidation_SyncEnabledNoMQTT(t *testing.T) {
	cfg := DefaultDev()
	cfg.Sync.Enabled = true
	cfg.Sync.MQTTBroker = ""
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error when sync enabled without MQTT broker")
	}
}

func TestValidation_EVChargerConfigs(t *testing.T) {
	cfg := DefaultDev()
	cfg.Devices.EVCharger = append(cfg.Devices.EVCharger, EVChargerDeviceConfig{
		ID: "", Host: "127.0.0.1", Port: 5020, Unit: 9,
	})
	if err := cfg.validate(); err == nil {
		t.Error("expected validation error for EV charger without ID")
	}
}

func TestApplyDefaults_PCSBrand(t *testing.T) {
	cfg := &Config{
		Devices: DevicesConfig{
			BMS:   []BMSDeviceConfig{{ID: "bms-0", Host: "127.0.0.1", Port: 5020, Unit: 1}},
			PCS:   []PCSDeviceConfig{{ID: "pcs-0", Host: "127.0.0.1", Port: 5020, Unit: 3}},
			Meter: []MeterDeviceConfig{{ID: "m-0", Host: "127.0.0.1", Port: 5020, Unit: 5}},
		},
	}
	cfg.applyDefaults()

	if cfg.Devices.PCS[0].Brand != "generic" {
		t.Errorf("expected default brand 'generic', got %s", cfg.Devices.PCS[0].Brand)
	}
	if cfg.Devices.PCS[0].CommandMode != "individual" {
		t.Errorf("expected default command_mode 'individual', got %s", cfg.Devices.PCS[0].CommandMode)
	}
}

func TestDefaultIntervals(t *testing.T) {
	cfg := &Config{} // zero values
	if cfg.FastInterval().Milliseconds() != 200 {
		t.Errorf("default fast should be 200ms")
	}
	if cfg.StandardInterval().Milliseconds() != 1000 {
		t.Errorf("default standard should be 1000ms")
	}
	if cfg.SlowInterval().Milliseconds() != 5000 {
		t.Errorf("default slow should be 5000ms")
	}
}

func TestLoadCDISingaporeConfig(t *testing.T) {
	path := "config/cdi-singapore.yaml"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "../../config/cdi-singapore.yaml"
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load cdi-singapore.yaml: %v", err)
	}

	if cfg.Site.ID != "cdi-singapore" {
		t.Errorf("expected site id 'cdi-singapore', got %s", cfg.Site.ID)
	}
	if cfg.Site.Name != "CDI Singapore BESS Site" {
		t.Errorf("expected site name 'CDI Singapore BESS Site', got %s", cfg.Site.Name)
	}
	if !cfg.Simulate {
		t.Errorf("expected simulate=true")
	}

	// 4 BMS units, each on separate unit IDs
	if len(cfg.Devices.BMS) != 4 {
		t.Fatalf("expected 4 BMS, got %d", len(cfg.Devices.BMS))
	}
	bmsUnits := make(map[int]bool)
	for _, b := range cfg.Devices.BMS {
		if b.Host != "127.0.0.1" || b.Port != 5020 {
			t.Errorf("BMS %s: expected 127.0.0.1:5020, got %s:%d", b.ID, b.Host, b.Port)
		}
		if bmsUnits[b.Unit] {
			t.Errorf("duplicate BMS unit ID: %d", b.Unit)
		}
		bmsUnits[b.Unit] = true
	}

	// 24 PCS units, ALL with brand: "liteon", command_mode: "individual"
	if len(cfg.Devices.PCS) != 24 {
		t.Fatalf("expected 24 PCS, got %d", len(cfg.Devices.PCS))
	}
	for _, p := range cfg.Devices.PCS {
		if p.Host != "127.0.0.1" || p.Port != 5020 {
			t.Errorf("PCS %s: expected 127.0.0.1:5020, got %s:%d", p.ID, p.Host, p.Port)
		}
		if p.Brand != "liteon" {
			t.Errorf("PCS %s: expected brand 'liteon', got %s", p.ID, p.Brand)
		}
		if p.CommandMode != "individual" {
			t.Errorf("PCS %s: expected command_mode 'individual', got %s", p.ID, p.CommandMode)
		}
	}

	// 3 Power Meters (meter-pcc, meter-solar, meter-load)
	if len(cfg.Devices.Meter) != 3 {
		t.Fatalf("expected 3 meters, got %d", len(cfg.Devices.Meter))
	}
	meterIDs := map[string]bool{"meter-pcc": false, "meter-solar": false, "meter-load": false}
	for _, m := range cfg.Devices.Meter {
		if m.Host != "127.0.0.1" || m.Port != 5020 {
			t.Errorf("Meter %s: expected 127.0.0.1:5020, got %s:%d", m.ID, m.Host, m.Port)
		}
		if _, ok := meterIDs[m.ID]; !ok {
			t.Errorf("unexpected meter ID: %s", m.ID)
		}
		meterIDs[m.ID] = true
	}
	for id, found := range meterIDs {
		if !found {
			t.Errorf("expected meter %s not found", id)
		}
	}

	// 6 EV Chargers (ev-0 to ev-5), protocol: "ocpp"
	if len(cfg.Devices.EVCharger) != 6 {
		t.Fatalf("expected 6 EV chargers, got %d", len(cfg.Devices.EVCharger))
	}
	for _, ev := range cfg.Devices.EVCharger {
		if ev.Host != "127.0.0.1" || ev.Port != 5020 {
			t.Errorf("EV %s: expected 127.0.0.1:5020, got %s:%d", ev.ID, ev.Host, ev.Port)
		}
		if ev.Protocol != "ocpp" {
			t.Errorf("EV %s: expected protocol 'ocpp', got %s", ev.ID, ev.Protocol)
		}
	}

	// Topology checks
	if cfg.Topology.PCCMeter != "meter-pcc" {
		t.Errorf("expected pcc_meter 'meter-pcc', got %s", cfg.Topology.PCCMeter)
	}
	if cfg.Topology.SolarMeter != "meter-solar" {
		t.Errorf("expected solar_meter 'meter-solar', got %s", cfg.Topology.SolarMeter)
	}
	if len(cfg.Topology.BESSGroups) != 4 {
		t.Fatalf("expected 4 BESS groups, got %d", len(cfg.Topology.BESSGroups))
	}
	for _, g := range cfg.Topology.BESSGroups {
		if len(g.BMS) != 1 {
			t.Errorf("group %s: expected 1 BMS, got %d", g.Name, len(g.BMS))
		}
		if len(g.PCS) != 6 {
			t.Errorf("group %s: expected 6 PCS, got %d", g.Name, len(g.PCS))
		}
		if g.CommandMode != "individual" {
			t.Errorf("group %s: expected command_mode 'individual', got %s", g.Name, g.CommandMode)
		}
	}

	// Polling
	if cfg.Polling.FastMs != 200 || cfg.Polling.StandardMs != 1000 || cfg.Polling.SlowMs != 5000 {
		t.Errorf("unexpected polling: %+v", cfg.Polling)
	}

	// Engine
	if cfg.Engine.CycleMs != 1000 || cfg.Engine.Distribution != "equal" {
		t.Errorf("unexpected engine: %+v", cfg.Engine)
	}

	// Storage
	if cfg.Storage.Driver != "sqlite" || cfg.Storage.DSN != "/var/lib/ems/cdi-singapore.db" || cfg.Storage.RetentionDays != 7 || !cfg.Storage.WALMode {
		t.Errorf("unexpected storage: %+v", cfg.Storage)
	}

	// Sync
	if cfg.Sync.Enabled {
		t.Errorf("expected sync disabled")
	}

	// UI
	if cfg.UI.Port != 8080 {
		t.Errorf("expected UI port 8080, got %d", cfg.UI.Port)
	}
}

