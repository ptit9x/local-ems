package peak_shaving

import (
	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// Config represents the configuration for the PeakShaving controller.
type Config struct {
	PeakThresholdW int  // Grid import threshold in watts (default: 50000 = 50kW)
	Enabled        bool
}

// DefaultConfig returns the default peak shaving configuration.
func DefaultConfig() Config {
	return Config{
		PeakThresholdW: 50000,
		Enabled:        true,
	}
}

// Controller implements the PeakShaving controller logic.
// It does not add constraints — it advises the cycle manager on desired
// discharge power to shave peaks above the threshold.
type Controller struct {
	config Config
}

// New creates a new PeakShaving controller.
func New(config Config) *Controller {
	return &Controller{config: config}
}

// ID returns the unique identifier for this controller.
func (c *Controller) ID() string {
	return "peak_shaving"
}

// IsEnabled returns true if the controller is enabled.
func (c *Controller) IsEnabled() bool {
	return c.config.Enabled
}

// Run is a no-op for peak shaving — it uses DesiredPower instead.
func (c *Controller) Run(_ controllers.DeviceState, _ *resolver.ConstraintCollector) error {
	return nil
}

// DesiredPower returns the desired discharge power to shave the peak.
// Returns positive watts for discharge, 0 if no peak detected.
func (c *Controller) DesiredPower(state controllers.DeviceState) int {
	if !c.IsEnabled() {
		return 0
	}
	if state.GridPowerW > c.config.PeakThresholdW {
		return state.GridPowerW - c.config.PeakThresholdW
	}
	return 0
}
