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
