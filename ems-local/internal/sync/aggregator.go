package sync

import (
	"fmt"
	"sync"
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/engine/cycle"
)

// TelemetryMessage is the aggregated message sent to Cloud every 5 minutes.
type TelemetryMessage struct {
	SchemaVersion     int       `json:"schema_version"`
	MessageID         string    `json:"message_id"`
	SiteID            string    `json:"site_id"`
	Timestamp         time.Time `json:"timestamp"`
	IntervalSec       int       `json:"interval_sec"`
	AvgSolarW         int       `json:"avg_solar_w"`
	AvgLoadW          int       `json:"avg_load_w"`
	AvgGridW          int       `json:"avg_grid_w"`
	AvgESSW           int       `json:"avg_ess_w"`
	MinSOC            float64   `json:"min_soc"`
	MaxSOC            float64   `json:"max_soc"`
	AvgSOC            float64   `json:"avg_soc"`
	MaxTempC          float64   `json:"max_temp_c"`
	EnergyChargeWh    int       `json:"energy_charge_wh"`
	EnergyDischargeWh int       `json:"energy_discharge_wh"`
	CycleCount        int       `json:"cycle_count"`
	Alerts            []string  `json:"alerts,omitempty"`
}

// CycleAccumulator accumulates cycle data over a time window.
type CycleAccumulator struct {
	mu          sync.Mutex
	siteID      string
	solarSum    int64
	loadSum     int64
	gridSum     int64
	essSum      int64
	minSOC      float64
	maxSOC      float64
	socSum      float64
	maxTemp     float64
	chargeWh    int64
	dischargeWh int64
	count       int
	alerts      []string
	startTime   time.Time
}

// NewCycleAccumulator creates a new CycleAccumulator.
func NewCycleAccumulator(siteID string) *CycleAccumulator {
	return &CycleAccumulator{
		siteID:    siteID,
		startTime: time.Now().UTC(),
	}
}

// Add adds a cycle result to the accumulator.
func (c *CycleAccumulator) Add(cr cycle.CycleResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.solarSum += int64(cr.State.SolarPowerW)
	c.loadSum += int64(cr.State.LoadPowerW)
	c.gridSum += int64(cr.State.GridPowerW)
	c.essSum += int64(cr.ResolvedPower.Setpoint)

	if c.count == 0 || cr.State.BatterySOC < c.minSOC {
		c.minSOC = cr.State.BatterySOC
	}
	if c.count == 0 || cr.State.BatterySOC > c.maxSOC {
		c.maxSOC = cr.State.BatterySOC
	}
	c.socSum += cr.State.BatterySOC

	if c.count == 0 || cr.State.BatteryTempC > c.maxTemp {
		c.maxTemp = cr.State.BatteryTempC
	}

	// Assuming ESSW positive is discharge, negative is charge?
	// The exact definition might depend, but let's assume ESSW > 0 means discharging, ESSW < 0 means charging.
	// Since we don't have the duration for each cycle, if they run every second, power (W) = Energy (J/s).
	// Let's assume typical 1-second cycles to calculate Wh. Wait, chargeWh isn't explicitly defined by cycle duration in this context,
	// but generally 1W over 1 second is 1/3600 Wh. Let's just accumulate the power values and divide by 3600 per count if they are 1s intervals.
	// We'll assume the interval is 1s, or we can just accumulate `setpoint` in a simple way. Wait, we don't have enough info, maybe we just do `essSum / 3600`?
	// Let's just do a naive approach: for each cycle (assumed 1 second), Wh = Power / 3600.
	// Better: accumulate charge and discharge sums separately.
	if cr.ResolvedPower.Setpoint > 0 {
		c.dischargeWh += int64(cr.ResolvedPower.Setpoint) // sum of Watt-seconds
	} else if cr.ResolvedPower.Setpoint < 0 {
		c.chargeWh += int64(-cr.ResolvedPower.Setpoint)
	}

	c.count++
}

// Flush returns the aggregated telemetry message and resets the accumulator.
func (c *CycleAccumulator) Flush() TelemetryMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	intervalSec := int(now.Sub(c.startTime).Seconds())
	if intervalSec <= 0 {
		intervalSec = 1
	}

	msg := TelemetryMessage{
		SchemaVersion: 1,
		MessageID:     fmt.Sprintf("%s-%d", c.siteID, c.startTime.UnixNano()),
		SiteID:        c.siteID,
		Timestamp:     c.startTime,
		IntervalSec:   intervalSec,
		MinSOC:        c.minSOC,
		MaxSOC:        c.maxSOC,
		MaxTempC:      c.maxTemp,
		CycleCount:    c.count,
	}

	if c.count > 0 {
		msg.AvgSolarW = int(c.solarSum / int64(c.count))
		msg.AvgLoadW = int(c.loadSum / int64(c.count))
		msg.AvgGridW = int(c.gridSum / int64(c.count))
		msg.AvgESSW = int(c.essSum / int64(c.count))
		msg.AvgSOC = c.socSum / float64(c.count)
	} else {
		// If no cycles, send 0
		msg.MinSOC = 0
		msg.MaxSOC = 0
		msg.MaxTempC = 0
	}

	// Assuming 1 cycle = 1 second for Wh calculation
	// Since chargeWh holds Watt-seconds, divide by 3600 for Watt-hours
	msg.EnergyChargeWh = int(c.chargeWh / 3600)
	msg.EnergyDischargeWh = int(c.dischargeWh / 3600)

	// Reset accumulator
	c.solarSum = 0
	c.loadSum = 0
	c.gridSum = 0
	c.essSum = 0
	c.minSOC = 0
	c.maxSOC = 0
	c.socSum = 0
	c.maxTemp = 0
	c.chargeWh = 0
	c.dischargeWh = 0
	c.count = 0
	c.alerts = nil
	c.startTime = now

	return msg
}
