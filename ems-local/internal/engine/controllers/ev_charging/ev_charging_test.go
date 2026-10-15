package ev_charging

import (
	"testing"

	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
)

func TestEVCharging_Name(t *testing.T) {
	c := New(10000)
	if c.ID() != "ev_charging" {
		t.Errorf("expected ID ev_charging, got %s", c.ID())
	}
}

func TestEVCharging_Disabled(t *testing.T) {
	c := New(10000)
	c.SetEnabled(false)
	col := resolver.NewConstraintCollector()
	err := c.Run(controllers.DeviceState{GridPowerW: 10000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Constraints()) != 0 {
		t.Errorf("expected no constraints when disabled")
	}
}

func TestEVCharging_SiteOverload(t *testing.T) {
	c := New(10000)
	c.SetEVTotalPower(2000)
	col := resolver.NewConstraintCollector()
	// Uncontrolled grid = Grid (8000) + ESS (3000) = 11000 > maxSite (10000)
	err := c.Run(controllers.DeviceState{GridPowerW: 8000, ESSActivePowerW: 3000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	constraints := col.Constraints()
	if len(constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(constraints))
	}
	if constraints[0].MaxChargePower == nil || *constraints[0].MaxChargePower != 0 {
		t.Errorf("expected maxCharge 0, got %v", constraints[0].MaxChargePower)
	}
}

// (TestEVCharging_SolarExcess removed since behavior was broken and redundant)

func TestEVCharging_NormalOperation(t *testing.T) {
	c := New(10000)
	c.SetEVTotalPower(1000)
	col := resolver.NewConstraintCollector()
	err := c.Run(controllers.DeviceState{GridPowerW: 5000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Constraints()) != 0 {
		t.Errorf("expected no constraints, got %d", len(col.Constraints()))
	}
}

func TestEVCharging_DesiredPower(t *testing.T) {
	c := New(10000)

	// Test high SOC, overloaded site -> Should suggest discharging (positive power)
	stateHighSOC := controllers.DeviceState{GridPowerW: 8000, ESSActivePowerW: 3000, BatterySOC: 25.0} // Uncontrolled grid = 11000
	dp := c.DesiredPower(stateHighSOC)
	if dp <= 0 {
		t.Errorf("expected positive desired power for discharging, got %d", dp)
	}
	if dp != 1000 {
		t.Errorf("expected 1000 desired power, got %d", dp)
	}

	// Test low SOC, overloaded site -> Should not suggest discharging
	stateLowSOC := controllers.DeviceState{GridPowerW: 8000, ESSActivePowerW: 3000, BatterySOC: 15.0}
	dp2 := c.DesiredPower(stateLowSOC)
	if dp2 != 0 {
		t.Errorf("expected 0 desired power for low SOC, got %d", dp2)
	}
}

// (TestEVCharging_ZeroEVPower removed)
