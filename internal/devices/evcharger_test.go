package devices

import (
	"testing"
	"time"

	"github.com/vmo/local-ems/internal/modbus"
)

func startTestServer(t *testing.T) (*modbus.Server, *modbus.Client) {
	t.Helper()
	srv := modbus.NewServer(":0")
	if err := srv.Start(); err != nil {
		t.Fatalf("starting test server: %v", err)
	}
	client := modbus.NewClient(srv.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		srv.Stop()
		t.Fatalf("connecting client: %v", err)
	}
	return srv, client
}

func TestEVChargerAdapter_Poll(t *testing.T) {
	srv, client := startTestServer(t)
	defer srv.Stop()

	// Allocate registers for EV charger unit ID (must be done before SetRegister)
	srv.AllocateRegisters(EVChargerUnitID, 16) // 16 registers to cover all EV charger data + command

	// Setup initial register values
	srv.SetRegister(EVChargerUnitID, EVCStatus, uint16(StatusCharging))
	srv.SetRegister(EVChargerUnitID, EVCActivePower, 10000)   // 10000 W
	srv.SetRegister(EVChargerUnitID, EVCEnergyDelivered, 500) // 500 Wh
	srv.SetRegister(EVChargerUnitID, EVCMaxPowerLimit, 22000) // 22000 W
	srv.SetRegister(EVChargerUnitID, EVCVehicleConnected, 1)  // Connected
	srv.SetRegister(EVChargerUnitID, EVCCurrentL1, 320)       // 32.0 A
	srv.SetRegister(EVChargerUnitID, EVCCurrentL2, 310)       // 31.0 A
	srv.SetRegister(EVChargerUnitID, EVCCurrentL3, 300)       // 30.0 A

	adapter := NewEVChargerAdapter(client, EVChargerUnitID)
	if err := adapter.Poll(); err != nil {
		t.Fatalf("poll failed: %v", err)
	}

	data := adapter.Data()

	if data.Status != StatusCharging {
		t.Errorf("expected StatusCharging, got %d", data.Status)
	}
	if data.ActivePower != 10000 {
		t.Errorf("expected ActivePower 10000, got %d", data.ActivePower)
	}
	if data.EnergyDelivered != 500 {
		t.Errorf("expected EnergyDelivered 500, got %d", data.EnergyDelivered)
	}
	if data.MaxPowerLimit != 22000 {
		t.Errorf("expected MaxPowerLimit 22000, got %d", data.MaxPowerLimit)
	}
	if !data.VehicleConnected {
		t.Error("expected VehicleConnected true")
	}
	if data.CurrentL1 != 32.0 {
		t.Errorf("expected CurrentL1 32.0, got %.1f", data.CurrentL1)
	}
	if data.CurrentL2 != 31.0 {
		t.Errorf("expected CurrentL2 31.0, got %.1f", data.CurrentL2)
	}
	if data.CurrentL3 != 30.0 {
		t.Errorf("expected CurrentL3 30.0, got %.1f", data.CurrentL3)
	}
}

func TestEVChargerAdapter_WriteMaxPower(t *testing.T) {
	srv, client := startTestServer(t)
	defer srv.Stop()

	srv.AllocateRegisters(EVChargerUnitID, 16)
	adapter := NewEVChargerAdapter(client, EVChargerUnitID)

	if err := adapter.WriteMaxPower(7000); err != nil {
		t.Fatalf("WriteMaxPower failed: %v", err)
	}

	// Read back the register
	val := srv.GetRegister(EVChargerUnitID, EVCMaxPowerLimit)
	if val != 7000 {
		t.Errorf("expected register value 7000, got %d", val)
	}
}

func TestEVChargerAdapter_WriteStartStop(t *testing.T) {
	srv, client := startTestServer(t)
	defer srv.Stop()

	srv.AllocateRegisters(EVChargerUnitID, 16)
	adapter := NewEVChargerAdapter(client, EVChargerUnitID)

	// Start
	if err := adapter.WriteStartStop(true); err != nil {
		t.Fatalf("WriteStartStop(true) failed: %v", err)
	}
	val := srv.GetRegister(EVChargerUnitID, EVCCommand)
	if val != 1 {
		t.Errorf("expected command register 1 (start), got %d", val)
	}

	// Stop
	if err := adapter.WriteStartStop(false); err != nil {
		t.Fatalf("WriteStartStop(false) failed: %v", err)
	}
	val = srv.GetRegister(EVChargerUnitID, EVCCommand)
	if val != 0 {
		t.Errorf("expected command register 0 (stop), got %d", val)
	}
}

func TestEVChargerAdapter_FaultStatus(t *testing.T) {
	srv, client := startTestServer(t)
	defer srv.Stop()

	srv.AllocateRegisters(EVChargerUnitID, 16)
	srv.SetRegister(EVChargerUnitID, EVCStatus, uint16(StatusFault))

	adapter := NewEVChargerAdapter(client, EVChargerUnitID)
	if err := adapter.Poll(); err != nil {
		t.Fatalf("poll failed: %v", err)
	}

	if adapter.Data().Status != StatusFault {
		t.Errorf("expected StatusFault (5), got %d", adapter.Data().Status)
	}
}

func TestEVChargerAdapter_VehicleConnected(t *testing.T) {
	srv, client := startTestServer(t)
	defer srv.Stop()

	srv.AllocateRegisters(EVChargerUnitID, 16)

	// Not connected
	srv.SetRegister(EVChargerUnitID, EVCVehicleConnected, 0)
	adapter := NewEVChargerAdapter(client, EVChargerUnitID)
	if err := adapter.Poll(); err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if adapter.Data().VehicleConnected {
		t.Error("expected VehicleConnected false")
	}

	// Connected
	srv.SetRegister(EVChargerUnitID, EVCVehicleConnected, 1)
	if err := adapter.Poll(); err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if !adapter.Data().VehicleConnected {
		t.Error("expected VehicleConnected true")
	}
}
