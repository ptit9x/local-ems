package sync_test

import (
	"testing"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
	"github.com/vmo/ems-platform/ems-local/internal/sync"
)

func TestCycleAccumulator(t *testing.T) {
	acc := sync.NewCycleAccumulator("site-1")

	for i := 0; i < 300; i++ {
		cr := cycle.CycleResult{
			Timestamp: time.Now(),
			State: controllers.DeviceState{
				SolarPowerW:  1000,
				LoadPowerW:   500,
				GridPowerW:   0,
				BatterySOC:   50.0,
				BatteryTempC: 25.0,
			},
			ResolvedPower: resolver.ResolvedPower{
				Setpoint: 500, // discharge
			},
		}
		acc.Add(cr)
	}

	msg := acc.Flush()

	if msg.CycleCount != 300 {
		t.Errorf("expected 300 cycles, got %d", msg.CycleCount)
	}
	if msg.SiteID != "site-1" {
		t.Errorf("expected site-1, got %s", msg.SiteID)
	}
	if msg.AvgSolarW != 1000 {
		t.Errorf("expected avg solar 1000, got %d", msg.AvgSolarW)
	}
	if msg.AvgLoadW != 500 {
		t.Errorf("expected avg load 500, got %d", msg.AvgLoadW)
	}
	if msg.AvgESSW != 500 {
		t.Errorf("expected avg ess 500, got %d", msg.AvgESSW)
	}
	if msg.AvgSOC != 50.0 {
		t.Errorf("expected avg soc 50, got %f", msg.AvgSOC)
	}
	if msg.MaxTempC != 25.0 {
		t.Errorf("expected max temp 25, got %f", msg.MaxTempC)
	}
	// 300 cycles * 500W = 150000 Ws = 41.6 Wh -> int is 41
	if msg.EnergyDischargeWh != 41 {
		t.Errorf("expected energy discharge 41 Wh, got %d", msg.EnergyDischargeWh)
	}
	if msg.EnergyChargeWh != 0 {
		t.Errorf("expected energy charge 0, got %d", msg.EnergyChargeWh)
	}
}
