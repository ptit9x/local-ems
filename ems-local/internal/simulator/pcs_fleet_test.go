package simulator

import (
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/devices/drivers"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
)

func TestPCSFleetSimulator_Create24Units(t *testing.T) {
	reg := drivers.DefaultPCSProfileRegistry()
	profile, err := reg.Get("liteon")
	if err != nil {
		t.Fatalf("failed to get liteon profile: %v", err)
	}

	sim := NewPCSFleetSimulator(":0", 24, profile)
	if err := sim.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}
	defer sim.Stop()

	if sim.Addr() == "" {
		t.Fatal("expected non-empty simulator address")
	}

	units := sim.GetAllUnits()
	if len(units) != 24 {
		t.Fatalf("expected 24 units, got %d", len(units))
	}

	for i := 0; i < 24; i++ {
		u := sim.GetUnit(i)
		expectedUnitID := byte(i + 1)

		if u.UnitID != expectedUnitID {
			t.Errorf("unit[%d]: expected UnitID %d, got %d", i, expectedUnitID, u.UnitID)
		}
		if u.Brand != "liteon" {
			t.Errorf("unit[%d]: expected Brand 'liteon', got %s", i, u.Brand)
		}
		if u.PowerW != 0 {
			t.Errorf("unit[%d]: expected PowerW 0, got %d", i, u.PowerW)
		}
		if u.SetpointW != 0 {
			t.Errorf("unit[%d]: expected SetpointW 0, got %d", i, u.SetpointW)
		}
		if u.Status != 0 {
			t.Errorf("unit[%d]: expected Status 0 (Standby), got %d", i, u.Status)
		}
		if u.FaultCode != 0 {
			t.Errorf("unit[%d]: expected FaultCode 0, got %d", i, u.FaultCode)
		}

		// Verify GetAllUnits slice snapshot matches GetUnit
		if units[i] != u {
			t.Errorf("unit[%d]: GetAllUnits snapshot mismatch: %+v vs %+v", i, units[i], u)
		}
	}

	// Verify out-of-range GetUnit returns empty struct
	outOfBounds := sim.GetUnit(99)
	if outOfBounds.UnitID != 0 {
		t.Errorf("expected empty state for out-of-bounds unit, got %+v", outOfBounds)
	}
	negIndex := sim.GetUnit(-1)
	if negIndex.UnitID != 0 {
		t.Errorf("expected empty state for negative index unit, got %+v", negIndex)
	}
}

func TestPCSFleetSimulator_SetpointRamp(t *testing.T) {
	reg := drivers.DefaultPCSProfileRegistry()
	profile, err := reg.Get("liteon")
	if err != nil {
		t.Fatalf("failed to get liteon profile: %v", err)
	}

	sim := NewPCSFleetSimulator(":0", 24, profile)
	if err := sim.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}
	defer sim.Stop()

	client := modbus.NewClient(sim.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	// Use ProfiledPCSAdapter for Unit 1 (index 0, UnitID 1)
	adapter := drivers.NewProfiledPCSAdapter(client, 1, profile)

	// Write setpoint 1200W (ramp target)
	if err := adapter.WriteActivePower(1200); err != nil {
		t.Fatalf("failed to write active power: %v", err)
	}

	// Tick 1: 0W -> 500W (max ramp 500W/tick)
	sim.Tick()
	u := sim.GetUnit(0)
	if u.PowerW != 500 {
		t.Errorf("tick 1: expected PowerW 500, got %d", u.PowerW)
	}
	if u.Status != 1 {
		t.Errorf("tick 1: expected Status 1 (Running), got %d", u.Status)
	}

	// Verify Modbus register via adapter poll
	if err := adapter.Poll(); err != nil {
		t.Fatalf("tick 1: failed to poll adapter: %v", err)
	}
	if data := adapter.Data(); data.ActualActivePowerW != 500 {
		t.Errorf("tick 1 modbus: expected ActualActivePowerW 500, got %d", data.ActualActivePowerW)
	} else if data.Status != devices.PCSRunning {
		t.Errorf("tick 1 modbus: expected PCSRunning, got %v", data.Status)
	}

	// Tick 2: 500W -> 1000W
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 1000 {
		t.Errorf("tick 2: expected PowerW 1000, got %d", u.PowerW)
	}

	if err := adapter.Poll(); err != nil {
		t.Fatalf("tick 2: failed to poll adapter: %v", err)
	}
	if data := adapter.Data(); data.ActualActivePowerW != 1000 {
		t.Errorf("tick 2 modbus: expected ActualActivePowerW 1000, got %d", data.ActualActivePowerW)
	}

	// Tick 3: 1000W -> 1200W (ramps remaining 200W)
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 1200 {
		t.Errorf("tick 3: expected PowerW 1200, got %d", u.PowerW)
	}

	if err := adapter.Poll(); err != nil {
		t.Fatalf("tick 3: failed to poll adapter: %v", err)
	}
	if data := adapter.Data(); data.ActualActivePowerW != 1200 {
		t.Errorf("tick 3 modbus: expected ActualActivePowerW 1200, got %d", data.ActualActivePowerW)
	}

	// Tick 4: stays at 1200W
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 1200 {
		t.Errorf("tick 4: expected PowerW 1200, got %d", u.PowerW)
	}

	// Ramp down to 0W
	if err := adapter.WriteActivePower(0); err != nil {
		t.Fatalf("failed to write active power 0: %v", err)
	}

	// Tick 5: 1200W -> 700W
	sim.Tick()
	if p := sim.GetUnit(0).PowerW; p != 700 {
		t.Errorf("tick 5: expected PowerW 700, got %d", p)
	}

	// Tick 6: 700W -> 200W
	sim.Tick()
	if p := sim.GetUnit(0).PowerW; p != 200 {
		t.Errorf("tick 6: expected PowerW 200, got %d", p)
	}

	// Tick 7: 200W -> 0W (reaches 0, status returns to Standby)
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 0 {
		t.Errorf("tick 7: expected PowerW 0, got %d", u.PowerW)
	}
	if u.Status != 0 {
		t.Errorf("tick 7: expected Status 0 (Standby), got %d", u.Status)
	}

	if err := adapter.Poll(); err != nil {
		t.Fatalf("tick 7: failed to poll adapter: %v", err)
	}
	if data := adapter.Data(); data.ActualActivePowerW != 0 {
		t.Errorf("tick 7 modbus: expected ActualActivePowerW 0, got %d", data.ActualActivePowerW)
	} else if data.Status != devices.PCSStopped {
		t.Errorf("tick 7 modbus: expected PCSStopped, got %v", data.Status)
	}
}

func TestPCSFleetSimulator_FaultInjection(t *testing.T) {
	reg := drivers.DefaultPCSProfileRegistry()
	profile, err := reg.Get("liteon")
	if err != nil {
		t.Fatalf("failed to get liteon profile: %v", err)
	}

	sim := NewPCSFleetSimulator(":0", 24, profile)
	if err := sim.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}
	defer sim.Stop()

	client := modbus.NewClient(sim.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	adapter := drivers.NewProfiledPCSAdapter(client, 1, profile)

	// Ramp unit 0 to 1000W
	if err := adapter.WriteActivePower(1000); err != nil {
		t.Fatalf("failed to write setpoint: %v", err)
	}
	sim.Tick() // 500W
	sim.Tick() // 1000W

	if u := sim.GetUnit(0); u.PowerW != 1000 || u.Status != 1 {
		t.Fatalf("expected unit running at 1000W, got power=%d status=%d", u.PowerW, u.Status)
	}

	// Inject fault on unit 0
	faultCode := uint16(0x002A)
	sim.SetFault(0, faultCode)

	// Verify power drops to 0 and status is Fault
	u := sim.GetUnit(0)
	if u.PowerW != 0 {
		t.Errorf("expected PowerW 0 immediately on fault, got %d", u.PowerW)
	}
	if u.Status != 2 {
		t.Errorf("expected Status 2 (Fault), got %d", u.Status)
	}
	if u.FaultCode != faultCode {
		t.Errorf("expected FaultCode 0x%04X, got 0x%04X", faultCode, u.FaultCode)
	}

	// Call Tick() to verify power remains 0 during fault
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 0 {
		t.Errorf("expected PowerW 0 after tick in fault, got %d", u.PowerW)
	}
	if u.Status != 2 {
		t.Errorf("expected Status 2 after tick in fault, got %d", u.Status)
	}

	// Verify Modbus registers via adapter
	if err := adapter.Poll(); err != nil {
		t.Fatalf("failed to poll adapter during fault: %v", err)
	}
	if data := adapter.Data(); data.ActualActivePowerW != 0 {
		t.Errorf("modbus during fault: expected power 0, got %d", data.ActualActivePowerW)
	} else if data.Status != devices.PCSFault {
		t.Errorf("modbus during fault: expected PCSFault (2), got %v", data.Status)
	}

	// Writing setpoint while in fault should not resume power
	if err := adapter.WriteActivePower(1500); err != nil {
		t.Fatalf("failed to write setpoint during fault: %v", err)
	}
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 0 {
		t.Errorf("expected power to remain 0 when setpoint sent during fault, got %d", u.PowerW)
	}
	if u.Status != 2 {
		t.Errorf("expected status to remain 2 during fault, got %d", u.Status)
	}

	// Clear fault: returns to Standby
	sim.ClearFault(0)
	u = sim.GetUnit(0)
	if u.Status != 0 {
		t.Errorf("expected Status 0 (Standby) after ClearFault, got %d", u.Status)
	}
	if u.FaultCode != 0 {
		t.Errorf("expected FaultCode 0 after ClearFault, got %d", u.FaultCode)
	}

	// Next tick allows ramping towards setpoint (1500W)
	sim.Tick()
	u = sim.GetUnit(0)
	if u.PowerW != 500 {
		t.Errorf("expected PowerW 500 after recovery, got %d", u.PowerW)
	}
	if u.Status != 1 {
		t.Errorf("expected Status 1 (Running) after recovery, got %d", u.Status)
	}
}

func TestPCSFleetSimulator_IndividualControl(t *testing.T) {
	reg := drivers.DefaultPCSProfileRegistry()
	profile, err := reg.Get("liteon")
	if err != nil {
		t.Fatalf("failed to get liteon profile: %v", err)
	}

	sim := NewPCSFleetSimulator(":0", 24, profile)
	if err := sim.Start(); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}
	defer sim.Stop()

	client := modbus.NewClient(sim.Addr(), 2*time.Second)
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	// Create adapters for multiple units
	adapter1 := drivers.NewProfiledPCSAdapter(client, 1, profile)
	adapter2 := drivers.NewProfiledPCSAdapter(client, 2, profile)
	adapter3 := drivers.NewProfiledPCSAdapter(client, 3, profile)
	adapter24 := drivers.NewProfiledPCSAdapter(client, 24, profile)

	// Set distinct setpoints across units
	// Unit 1: 1000W discharge
	// Unit 2: 2000W discharge
	// Unit 3: -1000W charge
	// Unit 4 (index 3): unconfigured, stays 0W Standby
	// Unit 24: 500W discharge
	if err := adapter1.WriteActivePower(1000); err != nil {
		t.Fatalf("failed to write unit 1 setpoint: %v", err)
	}
	if err := adapter2.WriteActivePower(2000); err != nil {
		t.Fatalf("failed to write unit 2 setpoint: %v", err)
	}
	if err := adapter3.WriteActivePower(-1000); err != nil {
		t.Fatalf("failed to write unit 3 setpoint: %v", err)
	}
	if err := adapter24.WriteActivePower(500); err != nil {
		t.Fatalf("failed to write unit 24 setpoint: %v", err)
	}

	// Tick 1 (ramp 500W)
	sim.Tick()
	if p := sim.GetUnit(0).PowerW; p != 500 {
		t.Errorf("tick 1 unit 1: expected 500W, got %d", p)
	}
	if p := sim.GetUnit(1).PowerW; p != 500 {
		t.Errorf("tick 1 unit 2: expected 500W, got %d", p)
	}
	if p := sim.GetUnit(2).PowerW; p != -500 {
		t.Errorf("tick 1 unit 3: expected -500W, got %d", p)
	}
	if p := sim.GetUnit(3).PowerW; p != 0 {
		t.Errorf("tick 1 unit 4: expected 0W, got %d", p)
	}
	if p := sim.GetUnit(23).PowerW; p != 500 {
		t.Errorf("tick 1 unit 24: expected 500W, got %d", p)
	}

	// Tick 2 (ramp another 500W)
	sim.Tick()
	u1 := sim.GetUnit(0)
	u2 := sim.GetUnit(1)
	u3 := sim.GetUnit(2)
	u4 := sim.GetUnit(3)
	u24 := sim.GetUnit(23)

	if u1.PowerW != 1000 || u1.Status != 1 {
		t.Errorf("tick 2 unit 1: expected 1000W Running, got power=%d status=%d", u1.PowerW, u1.Status)
	}
	if u2.PowerW != 1000 || u2.Status != 1 {
		t.Errorf("tick 2 unit 2: expected 1000W Running, got power=%d status=%d", u2.PowerW, u2.Status)
	}
	if u3.PowerW != -1000 || u3.Status != 1 {
		t.Errorf("tick 2 unit 3: expected -1000W Running, got power=%d status=%d", u3.PowerW, u3.Status)
	}
	if u4.PowerW != 0 || u4.Status != 0 {
		t.Errorf("tick 2 unit 4: expected 0W Standby, got power=%d status=%d", u4.PowerW, u4.Status)
	}
	if u24.PowerW != 500 || u24.Status != 1 {
		t.Errorf("tick 2 unit 24: expected 500W Running, got power=%d status=%d", u24.PowerW, u24.Status)
	}

	// Tick 3 and 4: Unit 2 ramps to 2000W
	sim.Tick()
	sim.Tick()
	if p := sim.GetUnit(1).PowerW; p != 2000 {
		t.Errorf("tick 4 unit 2: expected 2000W, got %d", p)
	}

	// Verify all units over Modbus TCP via their respective adapters
	if err := adapter1.Poll(); err != nil {
		t.Fatalf("failed to poll unit 1: %v", err)
	}
	if err := adapter2.Poll(); err != nil {
		t.Fatalf("failed to poll unit 2: %v", err)
	}
	if err := adapter3.Poll(); err != nil {
		t.Fatalf("failed to poll unit 3: %v", err)
	}
	if err := adapter24.Poll(); err != nil {
		t.Fatalf("failed to poll unit 24: %v", err)
	}

	if d := adapter1.Data().ActualActivePowerW; d != 1000 {
		t.Errorf("adapter 1: expected 1000W, got %d", d)
	}
	if d := adapter2.Data().ActualActivePowerW; d != 2000 {
		t.Errorf("adapter 2: expected 2000W, got %d", d)
	}
	if d := adapter3.Data().ActualActivePowerW; d != -1000 {
		t.Errorf("adapter 3: expected -1000W, got %d", d)
	}
	if d := adapter24.Data().ActualActivePowerW; d != 500 {
		t.Errorf("adapter 24: expected 500W, got %d", d)
	}
}
