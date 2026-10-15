package drivers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/config"
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
)

func TestGenericPCSDriver_ReadWrite(t *testing.T) {
	// Start mock modbus server
	server := modbus.NewServer(":0")
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start modbus server: %v", err)
	}
	defer server.Stop()

	// Allocate registers
	unitID := byte(1)
	server.AllocateRegisters(unitID, 100)

	// Set some values on the server based on generic profile registers
	// generic profile: ActivePowerSetReg: 0, ActivePowerActReg: 1, ReactivePowerReg: 2, StatusReg: 3
	server.SetRegister(unitID, 1, 500) // Actual Active Power
	server.SetRegister(unitID, 2, 200) // Reactive Power
	server.SetRegister(unitID, 3, 1)   // Status Running

	// Create client
	client := modbus.NewClient(server.Addr(), 1*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	registry := DefaultPCSProfileRegistry()
	profile, _ := registry.Get("generic")

	driver := NewGenericPCSDriver(client, unitID, profile)

	// Test ReadState
	reading, err := driver.ReadState()
	if err != nil {
		t.Fatalf("ReadState failed: %v", err)
	}
	if reading.ActualActivePowerW != 500 {
		t.Errorf("expected 500W, got %d", reading.ActualActivePowerW)
	}
	if reading.ReactivePowerVar != 200 {
		t.Errorf("expected 200VAR, got %d", reading.ReactivePowerVar)
	}
	if reading.Status != devices.PCSRunning {
		t.Errorf("expected Running status, got %v", reading.Status)
	}

	// Test WriteSetpoint
	if err := driver.WriteSetpoint(1000); err != nil {
		t.Fatalf("WriteSetpoint failed: %v", err)
	}

	// Check server value
	val := server.GetRegister(unitID, 0) // generic profile ActivePowerSetReg is 0
	if val != 1000 {
		t.Errorf("expected server to have 1000, got %d", val)
	}

	// Test ReadAlarms
	alarms, err := driver.ReadAlarms()
	if err != nil {
		t.Fatalf("ReadAlarms failed: %v", err)
	}
	if len(alarms) != 0 {
		t.Errorf("expected 0 alarms, got %d", len(alarms))
	}

	// Test Capabilities
	caps := driver.Capabilities()
	if caps.SupportsAlarms {
		t.Errorf("expected SupportsAlarms to be false")
	}
	if !caps.SupportsBidirectional {
		t.Errorf("expected SupportsBidirectional to be true")
	}
	if caps.SupportsReactivePower {
		t.Errorf("expected SupportsReactivePower to be false")
	}

	// Test ProtocolType
	if proto := driver.ProtocolType(); proto != "modbus" {
		t.Errorf("expected ProtocolType 'modbus', got %s", proto)
	}
}

func TestLoadPCSDriver_BuiltIn(t *testing.T) {
	cfg := config.PCSDeviceConfig{Brand: "sungrow", Unit: 2}
	driver, err := LoadPCSDriver(cfg, nil)
	if err != nil {
		t.Fatalf("expected no error loading built-in, got: %v", err)
	}

	if driver.Brand() != "sungrow" {
		t.Errorf("expected brand sungrow, got %s", driver.Brand())
	}
	if driver.CommandMode() != "master" {
		t.Errorf("expected command mode master, got %s", driver.CommandMode())
	}
}

func TestLoadPCSDriver_LiteOn(t *testing.T) {
	cfg := config.PCSDeviceConfig{Brand: "liteon", Unit: 1}
	driver, err := LoadPCSDriver(cfg, nil)
	if err != nil {
		t.Fatalf("expected no error loading built-in liteon, got: %v", err)
	}

	if driver.Brand() != "liteon" {
		t.Errorf("expected brand liteon, got %s", driver.Brand())
	}
	if driver.CommandMode() != "individual" {
		t.Errorf("expected command mode individual, got %s", driver.CommandMode())
	}
}

func TestLoadPCSDriver_CustomYAML(t *testing.T) {
	dir := filepath.Join("config", "pcs_profiles")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	yamlContent := []byte(`
brand: "test_brand"
registers:
  active_power_setpoint: 100
  active_power_actual: 102
  reactive_power: 104
  status: 110
  register_count: 11
scale_factor: 10.0
command_mode: "individual"
status_map:
  0: 0
  1: 1
  2: 2
description: "Test brand"
`)
	filePath := filepath.Join(dir, "test_brand.yaml")
	if err := os.WriteFile(filePath, yamlContent, 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}
	defer os.Remove(filePath)

	cfg := config.PCSDeviceConfig{Brand: "test_brand", Unit: 5}
	driver, err := LoadPCSDriver(cfg, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if driver.Brand() != "test_brand" {
		t.Errorf("expected test_brand, got %s", driver.Brand())
	}
	if driver.CommandMode() != "individual" {
		t.Errorf("expected individual, got %s", driver.CommandMode())
	}
}

func TestDifferentBrandsRegisterMaps(t *testing.T) {
	registry := DefaultPCSProfileRegistry()
	generic, _ := registry.Get("generic")
	sungrow, _ := registry.Get("sungrow")

	if generic.ActivePowerActReg == sungrow.ActivePowerActReg {
		t.Errorf("expected different active power registers, got same: %d", generic.ActivePowerActReg)
	}
}

func TestGenericPCSDriver_InterfaceCompliance(t *testing.T) {
	var _ PCSDriver = (*GenericPCSDriver)(nil)
}

func TestGenericPCSDriver_CapabilitiesAndAlarms(t *testing.T) {
	driver := NewGenericPCSDriver(nil, 1, PCSProfile{Brand: "test"})

	// ProtocolType
	if proto := driver.ProtocolType(); proto != "modbus" {
		t.Errorf("expected protocol 'modbus', got '%s'", proto)
	}

	// ReadAlarms
	alarms, err := driver.ReadAlarms()
	if err != nil {
		t.Fatalf("unexpected error from ReadAlarms: %v", err)
	}
	if len(alarms) != 0 {
		t.Errorf("expected 0 alarms, got %d", len(alarms))
	}

	// Capabilities
	caps := driver.Capabilities()
	if caps.SupportsAlarms {
		t.Errorf("expected SupportsAlarms=false")
	}
	if !caps.SupportsBidirectional {
		t.Errorf("expected SupportsBidirectional=true")
	}
	if caps.SupportsReactivePower {
		t.Errorf("expected SupportsReactivePower=false")
	}
}

func TestAlarmTypes(t *testing.T) {
	now := time.Now()
	alarm := Alarm{
		Code:        101,
		Description: "Overvoltage fault",
		Severity:    AlarmCritical,
		Timestamp:   now,
		Active:      true,
	}

	if alarm.Code != 101 {
		t.Errorf("expected Code 101, got %d", alarm.Code)
	}
	if alarm.Description != "Overvoltage fault" {
		t.Errorf("expected Description 'Overvoltage fault', got '%s'", alarm.Description)
	}
	if alarm.Severity != AlarmCritical {
		t.Errorf("expected AlarmCritical, got %v", alarm.Severity)
	}
	if !alarm.Active {
		t.Errorf("expected Active=true")
	}
	if !alarm.Timestamp.Equal(now) {
		t.Errorf("expected Timestamp %v, got %v", now, alarm.Timestamp)
	}

	// Test severity strings
	tests := []struct {
		severity AlarmSeverity
		expected string
	}{
		{AlarmInfo, "Info"},
		{AlarmWarning, "Warning"},
		{AlarmFault, "Fault"},
		{AlarmCritical, "Critical"},
		{AlarmSeverity(99), "Unknown"},
	}

	for _, tc := range tests {
		if tc.severity.String() != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, tc.severity.String())
		}
	}
}
