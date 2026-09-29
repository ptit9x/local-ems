package devices

import (
	"testing"
	"time"

	"github.com/vmo/local-ems/internal/modbus"
)

// setupTestServer creates a Modbus server with all device registers allocated
// and returns the server and a connected client.
func setupTestServer(t *testing.T) (*modbus.Server, *modbus.Client) {
	t.Helper()

	srv := modbus.NewServer(":0")
	srv.AllocateRegisters(UnitGridMeter, MeterRegisterCount)
	srv.AllocateRegisters(UnitSolarMeter, MeterRegisterCount)
	srv.AllocateRegisters(UnitBMS, BMSRegisterCount)
	srv.AllocateRegisters(UnitPCS, PCSRegisterCount)

	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start modbus server: %v", err)
	}

	client := modbus.NewClient(srv.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		srv.Stop()
		t.Fatalf("failed to connect modbus client: %v", err)
	}

	return srv, client
}

func teardown(srv *modbus.Server, client *modbus.Client) {
	client.Close()
	srv.Stop()
}

// --- Meter Adapter Tests ---

func TestMeterAdapter_ReadGridMeter(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// Set grid meter registers: Active Power = -1500W (selling to grid)
	srv.SetInt32(UnitGridMeter, MeterRegActivePowerHi, -1500)
	srv.SetRegister(UnitGridMeter, MeterRegVoltageL1, 2300) // 230.0V
	srv.SetRegister(UnitGridMeter, MeterRegVoltageL2, 2310) // 231.0V
	srv.SetRegister(UnitGridMeter, MeterRegVoltageL3, 2290) // 229.0V
	srv.SetRegister(UnitGridMeter, MeterRegCurrentL1, 65)   // 6.5A
	srv.SetRegister(UnitGridMeter, MeterRegCurrentL2, 72)   // 7.2A
	srv.SetRegister(UnitGridMeter, MeterRegCurrentL3, 58)   // 5.8A
	srv.SetRegister(UnitGridMeter, MeterRegFrequency, 5000) // 50.00Hz

	adapter := NewMeterAdapter(client, UnitGridMeter, "Grid")
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.ActivePowerW != -1500 {
		t.Errorf("ActivePowerW: got %d, want -1500", reading.ActivePowerW)
	}
	if reading.VoltageL1V != 230.0 {
		t.Errorf("VoltageL1V: got %.1f, want 230.0", reading.VoltageL1V)
	}
	if reading.VoltageL2V != 231.0 {
		t.Errorf("VoltageL2V: got %.1f, want 231.0", reading.VoltageL2V)
	}
	if reading.CurrentL1A != 6.5 {
		t.Errorf("CurrentL1A: got %.1f, want 6.5", reading.CurrentL1A)
	}
	if reading.FrequencyHz != 50.0 {
		t.Errorf("FrequencyHz: got %.2f, want 50.00", reading.FrequencyHz)
	}
}

func TestMeterAdapter_ReadSolarMeter(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// Solar producing 8500W
	srv.SetInt32(UnitSolarMeter, MeterRegActivePowerHi, 8500)
	srv.SetRegister(UnitSolarMeter, MeterRegVoltageL1, 2320)
	srv.SetRegister(UnitSolarMeter, MeterRegFrequency, 5001)

	adapter := NewMeterAdapter(client, UnitSolarMeter, "Solar")
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.ActivePowerW != 8500 {
		t.Errorf("ActivePowerW: got %d, want 8500", reading.ActivePowerW)
	}
	if reading.VoltageL1V != 232.0 {
		t.Errorf("VoltageL1V: got %.1f, want 232.0", reading.VoltageL1V)
	}
}

func TestMeterAdapter_ReadZeroPower(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// All zeros (night time, no solar)
	adapter := NewMeterAdapter(client, UnitSolarMeter, "Solar")
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.ActivePowerW != 0 {
		t.Errorf("ActivePowerW: got %d, want 0", reading.ActivePowerW)
	}
}

// --- BMS Adapter Tests ---

func TestBMSAdapter_Read(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// SOC = 72.5%, Temp = 28.3°C, CellMin = 3650mV, CellMax = 3720mV
	srv.SetRegister(UnitBMS, BMSRegSOC, 725)
	srv.SetRegister(UnitBMS, BMSRegTemperature, 283)
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMin, 3650)
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMax, 3720)
	srv.SetRegister(UnitBMS, BMSRegStatus, uint16(BMSDischarging))
	srv.SetRegister(UnitBMS, BMSRegCycleCount, 150)

	adapter := NewBMSAdapter(client, UnitBMS)
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.SOC != 72.5 {
		t.Errorf("SOC: got %.1f, want 72.5", reading.SOC)
	}
	if reading.TemperatureC != 28.3 {
		t.Errorf("TemperatureC: got %.1f, want 28.3", reading.TemperatureC)
	}
	if reading.CellVoltageMinMV != 3650 {
		t.Errorf("CellVoltageMinMV: got %d, want 3650", reading.CellVoltageMinMV)
	}
	if reading.CellVoltageMaxMV != 3720 {
		t.Errorf("CellVoltageMaxMV: got %d, want 3720", reading.CellVoltageMaxMV)
	}
	if reading.Status != BMSDischarging {
		t.Errorf("Status: got %v, want Discharging", reading.Status)
	}
	if reading.CycleCount != 150 {
		t.Errorf("CycleCount: got %d, want 150", reading.CycleCount)
	}
}

func TestBMSAdapter_ReadLowSOC(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	srv.SetRegister(UnitBMS, BMSRegSOC, 95) // 9.5%
	srv.SetRegister(UnitBMS, BMSRegTemperature, 350) // 35.0°C
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMin, 3100)
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMax, 3250)
	srv.SetRegister(UnitBMS, BMSRegStatus, uint16(BMSStandby))

	adapter := NewBMSAdapter(client, UnitBMS)
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.SOC != 9.5 {
		t.Errorf("SOC: got %.1f, want 9.5", reading.SOC)
	}
	if reading.TemperatureC != 35.0 {
		t.Errorf("TemperatureC: got %.1f, want 35.0", reading.TemperatureC)
	}
}

// --- PCS Adapter Tests ---

func TestPCSAdapter_Read(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// PCS discharging 3000W
	srv.SetInt32(UnitPCS, PCSRegActivePowerSetpointHi, 3000)
	srv.SetInt32(UnitPCS, PCSRegActualActivePowerHi, 2950) // slight lag
	srv.SetRegister(UnitPCS, PCSRegStatus, uint16(PCSRunning))
	srv.SetInt32(UnitPCS, PCSRegReactivePowerHi, 100)

	adapter := NewPCSAdapter(client, UnitPCS)
	reading, err := adapter.Read()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.ActivePowerSetpointW != 3000 {
		t.Errorf("ActivePowerSetpointW: got %d, want 3000", reading.ActivePowerSetpointW)
	}
	if reading.ActualActivePowerW != 2950 {
		t.Errorf("ActualActivePowerW: got %d, want 2950", reading.ActualActivePowerW)
	}
	if reading.Status != PCSRunning {
		t.Errorf("Status: got %v, want Running", reading.Status)
	}
	if reading.ReactivePowerVar != 100 {
		t.Errorf("ReactivePowerVar: got %d, want 100", reading.ReactivePowerVar)
	}
}

func TestPCSAdapter_WriteSetpoint(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	adapter := NewPCSAdapter(client, UnitPCS)

	// Write discharge setpoint
	if err := adapter.WriteSetpoint(5000); err != nil {
		t.Fatalf("WriteSetpoint: unexpected error: %v", err)
	}

	// Verify via server registers
	got := srv.GetInt32(UnitPCS, PCSRegActivePowerSetpointHi)
	if got != 5000 {
		t.Errorf("setpoint register: got %d, want 5000", got)
	}
}

func TestPCSAdapter_WriteChargeSetpoint(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	adapter := NewPCSAdapter(client, UnitPCS)

	// Write charge setpoint (negative)
	if err := adapter.WriteSetpoint(-3000); err != nil {
		t.Fatalf("WriteSetpoint: unexpected error: %v", err)
	}

	got := srv.GetInt32(UnitPCS, PCSRegActivePowerSetpointHi)
	if got != -3000 {
		t.Errorf("setpoint register: got %d, want -3000", got)
	}
}

// --- Device Manager Integration Test ---

func TestManager_ReadState(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	// Setup: Grid buying 2000W, Solar producing 5000W, BMS at 65%, PCS discharging 1000W
	srv.SetInt32(UnitGridMeter, MeterRegActivePowerHi, 2000)
	srv.SetInt32(UnitSolarMeter, MeterRegActivePowerHi, 5000)
	srv.SetRegister(UnitBMS, BMSRegSOC, 650) // 65.0%
	srv.SetRegister(UnitBMS, BMSRegTemperature, 260) // 26.0°C
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMin, 3600)
	srv.SetRegister(UnitBMS, BMSRegCellVoltageMax, 3700)
	srv.SetRegister(UnitBMS, BMSRegStatus, uint16(BMSDischarging))
	srv.SetInt32(UnitPCS, PCSRegActualActivePowerHi, 1000)
	srv.SetRegister(UnitPCS, PCSRegStatus, uint16(PCSRunning))

	gridAdapter := NewMeterAdapter(client, UnitGridMeter, "Grid")
	solarAdapter := NewMeterAdapter(client, UnitSolarMeter, "Solar")
	bmsAdapter := NewBMSAdapter(client, UnitBMS)
	pcsAdapter := NewPCSAdapter(client, UnitPCS)

	mgr := NewManager(gridAdapter, solarAdapter, bmsAdapter, pcsAdapter)
	state, err := mgr.ReadState()
	if err != nil {
		t.Fatalf("ReadState: unexpected error: %v", err)
	}

	if state.GridPowerW != 2000 {
		t.Errorf("GridPowerW: got %d, want 2000", state.GridPowerW)
	}
	if state.SolarPowerW != 5000 {
		t.Errorf("SolarPowerW: got %d, want 5000", state.SolarPowerW)
	}
	if state.BatterySOC != 65.0 {
		t.Errorf("BatterySOC: got %.1f, want 65.0", state.BatterySOC)
	}
	if state.BatteryTempC != 26.0 {
		t.Errorf("BatteryTempC: got %.1f, want 26.0", state.BatteryTempC)
	}
	if state.ESSActivePowerW != 1000 {
		t.Errorf("ESSActivePowerW: got %d, want 1000", state.ESSActivePowerW)
	}

	// Load = Grid + Solar - ESS = 2000 + 5000 - 1000 = 6000W
	if state.LoadPowerW != 6000 {
		t.Errorf("LoadPowerW: got %d, want 6000", state.LoadPowerW)
	}
}

func TestManager_WriteSetpoint(t *testing.T) {
	srv, client := setupTestServer(t)
	defer teardown(srv, client)

	gridAdapter := NewMeterAdapter(client, UnitGridMeter, "Grid")
	solarAdapter := NewMeterAdapter(client, UnitSolarMeter, "Solar")
	bmsAdapter := NewBMSAdapter(client, UnitBMS)
	pcsAdapter := NewPCSAdapter(client, UnitPCS)

	mgr := NewManager(gridAdapter, solarAdapter, bmsAdapter, pcsAdapter)

	if err := mgr.WriteSetpoint(-2500); err != nil {
		t.Fatalf("WriteSetpoint: unexpected error: %v", err)
	}

	got := srv.GetInt32(UnitPCS, PCSRegActivePowerSetpointHi)
	if got != -2500 {
		t.Errorf("PCS setpoint: got %d, want -2500", got)
	}
}
