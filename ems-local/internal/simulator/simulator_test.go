package simulator

import (
	"testing"
)

func TestSolarSinePattern(t *testing.T) {
	sim := New(DefaultConfig())

	// tick=0 (midnight): solar should be ~0
	state := sim.NextState()
	if state.SolarPowerW > 100 {
		t.Errorf("expected ~0 solar at midnight, got %d", state.SolarPowerW)
	}

	// Advance to noon (tick 43200): solar should be near peak
	sim2 := New(DefaultConfig())
	sim2.tick = 43200 - 1
	stateNoon := sim2.NextState()
	if stateNoon.SolarPowerW < sim2.config.SolarPeakW*90/100 {
		t.Errorf("expected near-peak solar at noon, got %d (peak=%d)",
			stateNoon.SolarPowerW, sim2.config.SolarPeakW)
	}
}

func TestLoadInRange(t *testing.T) {
	sim := New(DefaultConfig())
	cfg := sim.config

	for i := 0; i < 100; i++ {
		state := sim.NextState()
		minLoad := 500 // clamped minimum
		maxLoad := cfg.LoadBaseW + cfg.LoadVarianceW
		if state.LoadPowerW < minLoad || state.LoadPowerW > maxLoad {
			t.Errorf("load %dW out of range [%d, %d]", state.LoadPowerW, minLoad, maxLoad)
		}
	}
}

func TestSOCUpdatesOnCharge(t *testing.T) {
	sim := New(Config{
		InitialSOC:        50.0,
		SolarPeakW:        0,
		LoadBaseW:          1000,
		LoadVarianceW:      0,
		BatteryCapacityWh: 100, // small capacity for visible changes
		MaxChargeRateW:    10000,
		MaxDischargeRateW: 10000,
	})

	initialSOC := sim.SOC()
	sim.ApplyPower(-5000) // charge 5kW

	if sim.SOC() <= initialSOC {
		t.Errorf("SOC should increase after charging: was %.2f, now %.2f",
			initialSOC, sim.SOC())
	}
}

func TestSOCUpdatesOnDischarge(t *testing.T) {
	sim := New(Config{
		InitialSOC:        50.0,
		SolarPeakW:        0,
		LoadBaseW:          1000,
		LoadVarianceW:      0,
		BatteryCapacityWh: 100,
		MaxChargeRateW:    10000,
		MaxDischargeRateW: 10000,
	})

	initialSOC := sim.SOC()
	sim.ApplyPower(5000) // discharge 5kW

	if sim.SOC() >= initialSOC {
		t.Errorf("SOC should decrease after discharging: was %.2f, now %.2f",
			initialSOC, sim.SOC())
	}
}

func TestSOCClampsTo0_100(t *testing.T) {
	sim := New(Config{
		InitialSOC:        99.9,
		BatteryCapacityWh: 1, // tiny capacity
		MaxChargeRateW:    100000,
		MaxDischargeRateW: 100000,
	})

	// Massive charge should not exceed 100%
	sim.ApplyPower(-100000)
	if sim.SOC() > 100.0 {
		t.Errorf("SOC should be clamped to 100, got %.2f", sim.SOC())
	}

	// Reset and discharge
	sim2 := New(Config{
		InitialSOC:        0.1,
		BatteryCapacityWh: 1,
		MaxChargeRateW:    100000,
		MaxDischargeRateW: 100000,
	})
	sim2.ApplyPower(100000)
	if sim2.SOC() < 0.0 {
		t.Errorf("SOC should be clamped to 0, got %.2f", sim2.SOC())
	}
}

func TestDefaultConfigProducesValidState(t *testing.T) {
	sim := New(DefaultConfig())
	state := sim.NextState()

	if state.BatterySOC < 0 || state.BatterySOC > 100 {
		t.Errorf("invalid SOC: %.2f", state.BatterySOC)
	}
	if state.BatteryTempC < 20 || state.BatteryTempC > 30 {
		t.Errorf("unexpected temperature: %.2f", state.BatteryTempC)
	}
	if state.CellVoltageMinMV < 3000 || state.CellVoltageMaxMV > 4500 {
		t.Errorf("unexpected cell voltage: min=%d max=%d",
			state.CellVoltageMinMV, state.CellVoltageMaxMV)
	}
	if state.LoadPowerW <= 0 {
		t.Errorf("load should be positive, got %d", state.LoadPowerW)
	}
}
