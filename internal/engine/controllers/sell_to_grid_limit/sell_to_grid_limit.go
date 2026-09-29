package sell_to_grid_limit

import (
	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// Config represents the configuration for the SellToGridLimit controller.
type Config struct {
	MaxSellToGridPower int  // Max watts allowed to sell to grid (default: 0 for zero-export)
	Enabled            bool
}

// DefaultConfig returns the default zero-export configuration.
func DefaultConfig() Config {
	return Config{
		MaxSellToGridPower: 0,
		Enabled:            true,
	}
}

// Controller implements the SellToGridLimit controller logic.
// Based on OpenEMS ControllerEssSellToGridLimitImpl.
type Controller struct {
	config Config
}

// New creates a new SellToGridLimit controller.
func New(config Config) *Controller {
	return &Controller{config: config}
}

// ID returns the unique identifier for this controller.
func (c *Controller) ID() string {
	return "sell_to_grid_limit"
}

// IsEnabled returns true if the controller is enabled.
func (c *Controller) IsEnabled() bool {
	return c.config.Enabled
}

// Run checks if sell-to-grid power exceeds the configured limit and adds
// constraints to reduce ESS discharge or increase charge accordingly.
//
// Grid power convention: positive = buying from grid, negative = selling to grid.
func (c *Controller) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	if !c.IsEnabled() {
		return nil
	}

	// Calculate current sell-to-grid power (positive value)
	currentSellPower := -state.GridPowerW // negative grid = selling, flip sign
	if currentSellPower <= c.config.MaxSellToGridPower {
		// Not exceeding limit, no constraint needed
		return nil
	}

	// Excess power that needs to be absorbed
	excessPower := currentSellPower - c.config.MaxSellToGridPower

	// Calculate new ESS power limit: reduce discharge (or increase charge) by excessPower
	// ESS positive = discharge, negative = charge
	newESSLimit := state.ESSActivePowerW - excessPower

	if newESSLimit >= 0 {
		// Still discharging but at reduced rate
		collector.Add(resolver.Constraint{
			MaxDischargePower: &newESSLimit,
			Source:            c.ID(),
		})
	} else {
		// Need to charge to absorb excess
		zero := 0
		chargePower := -newESSLimit // make positive for MaxChargePower (not used here)
		forcePower := -chargePower  // negative = charge
		collector.Add(resolver.Constraint{
			MaxDischargePower: &zero,
			ForcePower:        &forcePower,
			Source:            c.ID(),
		})
	}

	return nil
}
