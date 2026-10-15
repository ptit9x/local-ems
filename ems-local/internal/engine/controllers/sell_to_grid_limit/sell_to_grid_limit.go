package sell_to_grid_limit

import (
	"github.com/vmo/ems-platform/ems-local/internal/engine/controllers"
	"github.com/vmo/ems-platform/ems-local/internal/engine/resolver"
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

// Run calculates the maximum allowed ESS discharge that respects the sell-to-grid limit
// and adds it as a constraint.
//
// Grid power convention: positive = buying from grid, negative = selling to grid.
func (c *Controller) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	if !c.IsEnabled() {
		return nil
	}

	// Calculate uncontrolled grid power (what grid import would be without ESS)
	// Positive GridPower means import, positive ESSActivePower means discharge (which reduces import)
	uncontrolledGrid := state.GridPowerW + state.ESSActivePowerW

	// The maximum allowed discharge is the uncontrolled grid import plus the allowed export limit.
	// Any discharge above this will cause grid export to exceed MaxSellToGridPower.
	maxDischarge := uncontrolledGrid + c.config.MaxSellToGridPower

	collector.Add(resolver.Constraint{
		MaxDischargePower: &maxDischarge,
		Source:            c.ID(),
	})

	return nil
}
