package aggregator

import (
	"testing"
)

func TestAggregator_Basic(t *testing.T) {
	agg := New(2, 2, 2)

	if agg.BMSCount() != 2 || agg.PCSCount() != 2 || agg.MeterCount() != 2 {
		t.Error("wrong counts")
	}

	// Update devices
	agg.UpdateBMS(0, BMSState{ID: "bms-0", SOC: 80, CapacityWh: 5000, MaxChargeW: 2500, MaxDischargeW: 5000, TempC: 28, MinCellMV: 3200, MaxCellMV: 3400, Online: true})
	agg.UpdateBMS(1, BMSState{ID: "bms-1", SOC: 60, CapacityWh: 5000, MaxChargeW: 2500, MaxDischargeW: 5000, TempC: 32, MinCellMV: 3100, MaxCellMV: 3450, Online: true})

	agg.UpdatePCS(0, PCSState{ID: "pcs-0", ActivePowerW: 1000, MaxPowerW: 5000, Status: PCSRunning, Online: true})
	agg.UpdatePCS(1, PCSState{ID: "pcs-1", ActivePowerW: 500, MaxPowerW: 5000, Status: PCSRunning, Online: true})

	agg.UpdateMeter(0, MeterState{ID: "meter-0", ActivePowerW: 3000, Online: true})
	agg.UpdateMeter(1, MeterState{ID: "meter-1", ActivePowerW: 100, Online: true})

	state := agg.Aggregate()

	// SOC = weighted average: (80*5000 + 60*5000) / 10000 = 70
	if state.TotalSOC != 70 {
		t.Errorf("expected SOC 70, got %.1f", state.TotalSOC)
	}

	// Capacity = 5000 + 5000 = 10000
	if state.TotalCapacityWh != 10000 {
		t.Errorf("expected capacity 10000, got %d", state.TotalCapacityWh)
	}

	// Max limits = sum
	if state.MaxChargeW != 5000 {
		t.Errorf("expected MaxChargeW 5000, got %d", state.MaxChargeW)
	}
	if state.MaxDischargeW != 10000 {
		t.Errorf("expected MaxDischargeW 10000, got %d", state.MaxDischargeW)
	}

	// Cell voltage: min=3100, max=3450
	if state.MinCellVoltMV != 3100 {
		t.Errorf("expected MinCellVoltMV 3100, got %d", state.MinCellVoltMV)
	}
	if state.MaxCellVoltMV != 3450 {
		t.Errorf("expected MaxCellVoltMV 3450, got %d", state.MaxCellVoltMV)
	}

	// Max temp = 32
	if state.MaxCellTempC != 32 {
		t.Errorf("expected MaxCellTempC 32, got %.1f", state.MaxCellTempC)
	}

	// PCS total = 1000 + 500 = 1500
	if state.TotalActivePowerW != 1500 {
		t.Errorf("expected TotalActivePowerW 1500, got %d", state.TotalActivePowerW)
	}

	// Grid = 3000 (first online meter)
	if state.GridPowerW != 3000 {
		t.Errorf("expected GridPowerW 3000, got %d", state.GridPowerW)
	}

	// Counts
	if state.BMSOnline != 2 {
		t.Errorf("expected 2 BMS online, got %d", state.BMSOnline)
	}
	if state.PCSRunning != 2 {
		t.Errorf("expected 2 PCS running, got %d", state.PCSRunning)
	}
	if state.MeterOnline != 2 {
		t.Errorf("expected 2 meter online, got %d", state.MeterOnline)
	}
}

func TestAggregator_OfflineDevices(t *testing.T) {
	agg := New(2, 2, 1)

	agg.UpdateBMS(0, BMSState{ID: "bms-0", SOC: 80, CapacityWh: 5000, Online: true, MinCellMV: 3200, MaxCellMV: 3400})
	agg.UpdateBMS(1, BMSState{ID: "bms-1", SOC: 20, CapacityWh: 5000, Online: false}) // offline

	agg.UpdatePCS(0, PCSState{ID: "pcs-0", ActivePowerW: 1000, Status: PCSRunning, Online: true})
	agg.UpdatePCS(1, PCSState{ID: "pcs-1", Status: PCSFault, Online: true}) // fault but online

	agg.UpdateMeter(0, MeterState{ID: "meter-0", ActivePowerW: 2000, Online: true})

	state := agg.Aggregate()

	// Only bms-0 online, SOC should be 80 (not weighted with offline bms-1)
	if state.TotalSOC != 80 {
		t.Errorf("expected SOC 80 (only bms-0), got %.1f", state.TotalSOC)
	}
	if state.BMSOnline != 1 {
		t.Errorf("expected 1 BMS online, got %d", state.BMSOnline)
	}

	// PCS: pcs-0 running, pcs-1 online but fault
	if state.PCSOnline != 2 {
		t.Errorf("expected 2 PCS online, got %d", state.PCSOnline)
	}
	if state.PCSRunning != 1 {
		t.Errorf("expected 1 PCS running, got %d", state.PCSRunning)
	}
}

func TestAggregator_SafetyAlarms(t *testing.T) {
	agg := New(1, 1, 1)

	agg.UpdateBMS(0, BMSState{ID: "bms-0", SOC: 50, CapacityWh: 5000, Online: true,
		MinCellMV: 2700, MaxCellMV: 3700, TempC: 56}) // all alarms!
	agg.UpdatePCS(0, PCSState{ID: "pcs-0", Online: true, Status: PCSRunning})
	agg.UpdateMeter(0, MeterState{ID: "meter-0", Online: true})

	state := agg.Aggregate()

	if !state.HasAlarm {
		t.Error("expected alarm for cell overvolt/undervolt/overtemp")
	}
	if len(state.Alarms) < 3 {
		t.Errorf("expected at least 3 alarms, got %d: %v", len(state.Alarms), state.Alarms)
	}
}

func TestAggregator_NoBMSOnline(t *testing.T) {
	agg := New(1, 1, 1)

	agg.UpdateBMS(0, BMSState{ID: "bms-0", Online: false})
	agg.UpdatePCS(0, PCSState{ID: "pcs-0", Online: true})
	agg.UpdateMeter(0, MeterState{ID: "meter-0", Online: true})

	state := agg.Aggregate()

	if !state.HasAlarm {
		t.Error("expected alarm when no BMS online")
	}
	if state.TotalSOC != 0 {
		t.Errorf("expected SOC 0 with no BMS, got %.1f", state.TotalSOC)
	}
}
