// Package ev_charging implements Dynamic Load Management for EV Chargers
// within the constraint-based controller chain architecture.
package ev_charging

import (
	"sync"

	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
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
//   - If uncontrolled grid import approaches maxSitePower, constrain BESS
//     charging to zero (site is at capacity, charging would push over limit).
func (c *EVChargingController) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	c.mu.RLock()
	enabled := c.enabled
	maxSite := c.maxSitePower
	c.mu.RUnlock()

	if !enabled {
		return nil
	}

	// Uncontrolled grid (what grid would be without ESS)
	uncontrolledGrid := state.GridPowerW + state.ESSActivePowerW

	if uncontrolledGrid >= maxSite {
		// Site is at or over capacity — block BESS charging so we don't add to the load
		zero := 0
		collector.Add(resolver.Constraint{
			MaxChargePower: &zero,
			Source:         c.ID(),
		})
	}

	return nil
}

// DesiredPower returns the suggested power setpoint.
// When site is overloaded and battery has capacity, suggest discharging
// to shave the peak and prevent breaker trip.
func (c *EVChargingController) DesiredPower(state controllers.DeviceState) int {
	c.mu.RLock()
	enabled := c.enabled
	maxSite := c.maxSitePower
	c.mu.RUnlock()

	if !enabled {
		return 0
	}

	uncontrolledGrid := state.GridPowerW + state.ESSActivePowerW

	if uncontrolledGrid > maxSite && state.BatterySOC > 20.0 {
		// Suggest discharging to shave the excess (positive = discharge)
		return uncontrolledGrid - maxSite
	}

	return 0
}
