package time_of_use

import (
	"time"

	"github.com/vmo/local-ems/internal/engine/controllers"
	"github.com/vmo/local-ems/internal/engine/resolver"
)

// TariffRate represents the electricity tariff category.
type TariffRate int

const (
	OffPeak TariffRate = iota // Cheap — should charge
	MidPeak                   // Normal — no action
	OnPeak                    // Expensive — should discharge
)

// TariffPeriod defines a time range and its tariff rate.
type TariffPeriod struct {
	StartHour int        // 0-23
	EndHour   int        // 0-23 (exclusive, wraps around midnight if Start > End)
	Rate      TariffRate
}

// Config represents the configuration for the TimeOfUse controller.
type Config struct {
	Periods           []TariffPeriod
	MaxChargeW        int     // Max charge power during off-peak (default: 3000)
	MaxDischargeW     int     // Max discharge power during on-peak (default: 5000)
	MinSOCToDischarge float64 // Don't discharge below this SOC (default: 20%)
	Enabled           bool
}

// DefaultConfig returns a typical TOU schedule:
// OffPeak: 22:00-06:00 (night), OnPeak: 09:00-12:00 + 17:00-21:00
func DefaultConfig() Config {
	return Config{
		Periods: []TariffPeriod{
			{StartHour: 22, EndHour: 6, Rate: OffPeak},   // night
			{StartHour: 9, EndHour: 12, Rate: OnPeak},    // morning peak
			{StartHour: 17, EndHour: 21, Rate: OnPeak},   // evening peak
		},
		MaxChargeW:        3000,
		MaxDischargeW:     5000,
		MinSOCToDischarge: 20.0,
		Enabled:           true,
	}
}

// Controller implements the TimeOfUse controller logic.
type Controller struct {
	config  Config
	NowFunc func() time.Time // injectable for testing (default: time.Now)
}

// New creates a new TimeOfUse controller.
func New(config Config) *Controller {
	return &Controller{
		config:  config,
		NowFunc: time.Now,
	}
}

// ID returns the unique identifier for this controller.
func (c *Controller) ID() string {
	return "time_of_use"
}

// IsEnabled returns true if the controller is enabled.
func (c *Controller) IsEnabled() bool {
	return c.config.Enabled
}

// Run is a no-op — TimeOfUse uses DesiredPower to advise the cycle manager.
func (c *Controller) Run(_ controllers.DeviceState, _ *resolver.ConstraintCollector) error {
	return nil
}

// DesiredPower returns the desired power based on current tariff period.
// Positive = discharge (on-peak), negative = charge (off-peak), 0 = no action.
func (c *Controller) DesiredPower(state controllers.DeviceState) int {
	if !c.IsEnabled() {
		return 0
	}

	hour := c.NowFunc().Hour()
	rate := c.currentRate(hour)

	switch rate {
	case OffPeak:
		return -c.config.MaxChargeW // charge during cheap hours
	case OnPeak:
		if state.BatterySOC > c.config.MinSOCToDischarge {
			return c.config.MaxDischargeW // discharge during expensive hours
		}
		return 0 // SOC too low, don't discharge
	default:
		return 0 // mid-peak, no action
	}
}

// currentRate determines the tariff rate for the given hour.
func (c *Controller) currentRate(hour int) TariffRate {
	for _, p := range c.config.Periods {
		if c.inPeriod(hour, p) {
			return p.Rate
		}
	}
	return MidPeak // default if no period matches
}

// inPeriod checks if hour falls within the period (handles midnight wrap).
func (c *Controller) inPeriod(hour int, p TariffPeriod) bool {
	if p.StartHour < p.EndHour {
		return hour >= p.StartHour && hour < p.EndHour
	}
	// Wraps around midnight (e.g., 22:00-06:00)
	return hour >= p.StartHour || hour < p.EndHour
}
