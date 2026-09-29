package balancing

import (
	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// Config represents the configuration for the Balancing controller.
type Config struct {
	Enabled bool
}

// DefaultConfig returns the default balancing configuration.
func DefaultConfig() Config {
	return Config{Enabled: true}
}

// Controller implements self-consumption balancing logic.
// It is the lowest-priority controller and serves as the default fallback:
// solar surplus → charge battery, load deficit → discharge battery.
type Controller struct {
	config Config
}

// New creates a new Balancing controller.
func New(config Config) *Controller {
	return &Controller{config: config}
}

// ID returns the unique identifier for this controller.
func (c *Controller) ID() string {
	return "balancing"
}

// IsEnabled returns true if the controller is enabled.
func (c *Controller) IsEnabled() bool {
	return c.config.Enabled
}

// Run is a no-op for balancing — it uses DesiredPower to advise.
func (c *Controller) Run(_ controllers.DeviceState, _ *resolver.ConstraintCollector) error {
	return nil
}

// DesiredPower returns the desired power to balance self-consumption.
// Positive = discharge (cover load deficit), negative = charge (absorb solar surplus).
func (c *Controller) DesiredPower(state controllers.DeviceState) int {
	if !c.IsEnabled() {
		return 0
	}

	// surplus = solar - load
	// positive surplus → excess solar → should charge (negative power)
	// negative surplus → load deficit → should discharge (positive power)
	surplus := state.SolarPowerW - state.LoadPowerW
	return -surplus
}
