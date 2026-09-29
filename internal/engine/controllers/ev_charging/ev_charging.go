// Package ev_charging implements Dynamic Load Management for EV Chargers
// within the constraint-based controller chain architecture.
package ev_charging

import (
	"sync"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// EVChargingController manages EV Charger power within grid capacity limits.
// It adds constraints to prevent BESS discharge when the site is at capacity,
// and allows battery-assisted EV charging when solar excess is available.
//
// Priority: P3 (same level as peak_shaving and time_of_use).
type EVChargingController struct {
	mu           sync.RWMutex
	maxSitePower int  // Maximum allowed total site import power (W) at PCC
	evTotalPower int  // Current total EV charging power (W), updated externally
	enabled      bool
}

// New creates a new EVChargingController with the given site power limit.
func New(maxSitePower int) *EVChargingController {
	return &EVChargingController{
		maxSitePower: maxSitePower,
		enabled:      true,
	}
}

// SetEVTotalPower updates the current total EV charging power.
// Called by the cycle manager each cycle with aggregated EV charger data.
func (c *EVChargingController) SetEVTotalPower(watts int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evTotalPower = watts
}

// SetEnabled enables or disables this controller.
func (c *EVChargingController) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// ID returns the unique identifier for this controller.
func (c *EVChargingController) ID() string {
	return "ev_charging"
}

// IsEnabled returns whether this controller is currently active.
func (c *EVChargingController) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// Run executes the EV charging control logic.
//
// Logic:
//   - If grid import + EV charging approaches maxSitePower, constrain BESS
//     discharge to zero (site is at capacity, discharging would push over limit).
//   - If grid is exporting (solar excess) and EV chargers are active, allow
//     BESS discharge up to EV charging power (feed solar via battery to EVs).
func (c *EVChargingController) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	c.mu.RLock()
	enabled := c.enabled
	evPower := c.evTotalPower
	maxSite := c.maxSitePower
	c.mu.RUnlock()

	if !enabled {
		return nil
	}

	available := maxSite - (state.GridPowerW + evPower)

	if available <= 0 {
		// Site is at or over capacity — block BESS discharge
		zero := 0
		collector.Add(resolver.Constraint{
			MaxDischargePower: &zero,
			Source:            "ev_charging",
		})
	} else if state.GridPowerW < 0 && evPower > 0 {
		// Grid is exporting while EVs are charging — allow discharge
		// up to EV power so battery can assist EV charging with solar excess
		maxDischarge := evPower
		collector.Add(resolver.Constraint{
			MaxDischargePower: &maxDischarge,
			Source:            "ev_charging",
		})
	}

	return nil
}

// DesiredPower returns the suggested power setpoint.
// When site is overloaded and battery has capacity, suggest charging
// to absorb excess and prevent breaker trip.
func (c *EVChargingController) DesiredPower(state controllers.DeviceState) int {
	c.mu.RLock()
	enabled := c.enabled
	evPower := c.evTotalPower
	maxSite := c.maxSitePower
	c.mu.RUnlock()

	if !enabled {
		return 0
	}

	available := maxSite - (state.GridPowerW + evPower)

	if available <= 0 && state.BatterySOC > 20.0 {
		// Suggest charging to absorb excess (available is negative)
		return available
	}

	return 0
}
