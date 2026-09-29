package ev_charging

import (
	"testing"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
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
	err := c.Run(controllers.DeviceState{GridPowerW: 8000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	constraints := col.Constraints()
	if len(constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(constraints))
	}
	if constraints[0].MaxDischargePower == nil || *constraints[0].MaxDischargePower != 0 {
		t.Errorf("expected maxDischarge 0")
	}
}

func TestEVCharging_SolarExcess(t *testing.T) {
	c := New(10000)
	c.SetEVTotalPower(3000)
	col := resolver.NewConstraintCollector()
	err := c.Run(controllers.DeviceState{GridPowerW: -1000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	constraints := col.Constraints()
	if len(constraints) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(constraints))
	}
	if constraints[0].MaxDischargePower == nil || *constraints[0].MaxDischargePower != 3000 {
		t.Errorf("expected maxDischarge 3000")
	}
}

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
	c.SetEVTotalPower(5000)

	// Test high SOC, overloaded site -> Should suggest charging
	stateHighSOC := controllers.DeviceState{GridPowerW: 6000, BatterySOC: 25.0}
	dp := c.DesiredPower(stateHighSOC)
	if dp >= 0 {
		t.Errorf("expected negative desired power for charging, got %d", dp)
	}

	// Test low SOC, overloaded site -> Should not suggest charging
	stateLowSOC := controllers.DeviceState{GridPowerW: 6000, BatterySOC: 15.0}
	dp2 := c.DesiredPower(stateLowSOC)
	if dp2 != 0 {
		t.Errorf("expected 0 desired power for low SOC, got %d", dp2)
	}
}

func TestEVCharging_ZeroEVPower(t *testing.T) {
	c := New(10000)
	c.SetEVTotalPower(0)
	col := resolver.NewConstraintCollector()
	err := c.Run(controllers.DeviceState{GridPowerW: -1000}, col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Constraints()) != 0 {
		t.Errorf("expected no constraints when EV power is 0")
	}
}
