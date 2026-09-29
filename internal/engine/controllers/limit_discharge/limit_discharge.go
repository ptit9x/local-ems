package limit_discharge

import (
	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// Config represents the configuration for the LimitDischarge controller.
type Config struct {
	MinSOC         float64 // Below this: no discharge (default: 15.0)
	ForceChargeSOC float64 // Below this: force charge (default: 10.0)
	ForceChargeW   int     // Force charge power in watts (default: 2000)
	Enabled        bool
}

// DefaultConfig returns the default configuration for the LimitDischarge controller.
func DefaultConfig() Config {
	return Config{
		MinSOC:         15.0,
		ForceChargeSOC: 10.0,
		ForceChargeW:   2000,
		Enabled:        true,
	}
}

// Controller implements the LimitDischarge controller logic.
type Controller struct {
	config Config
}

// New creates a new LimitDischarge controller.
func New(config Config) *Controller {
	return &Controller{
		config: config,
	}
}

// ID returns the unique identifier for this controller.
func (c *Controller) ID() string {
	return "limit_discharge"
}

// IsEnabled returns true if the controller is enabled.
func (c *Controller) IsEnabled() bool {
	return c.config.Enabled
}

// Run executes the controller logic and adds constraints if necessary.
func (c *Controller) Run(state controllers.DeviceState, collector *resolver.ConstraintCollector) error {
	if !c.IsEnabled() {
		return nil
	}

	var maxDischargePower *int
	var forcePower *int

	if state.BatterySOC < c.config.ForceChargeSOC {
		zero := 0
		maxDischargePower = &zero
		force := -c.config.ForceChargeW
		forcePower = &force
	} else if state.BatterySOC < c.config.MinSOC {
		zero := 0
		maxDischargePower = &zero
	} else {
		// SOC >= MinSOC, no constraints added
		return nil
	}

	collector.Add(resolver.Constraint{
		MaxDischargePower: maxDischargePower,
		ForcePower:        forcePower,
		Source:            c.ID(),
	})

	return nil
}
