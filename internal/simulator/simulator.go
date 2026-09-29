package simulator

import (
	"math"
	"math/rand"

	"github.com/vmo/local-ems/internal/engine/controllers"
)

const dayTicks = 86400 // 1 tick = 1 second, full day

// Config holds simulator parameters.
type Config struct {
	InitialSOC        float64 // Initial battery SOC % (default: 50.0)
	SolarPeakW        int     // Peak solar output watts (default: 10000)
	LoadBaseW          int     // Base load watts (default: 5000)
	LoadVarianceW      int     // Load random variance watts (default: 3000)
	BatteryCapacityWh int     // Battery capacity in Wh (default: 5000000 = 5MWh)
	MaxChargeRateW    int     // Max charge rate watts (default: 50000)
	MaxDischargeRateW int     // Max discharge rate watts (default: 50000)
}

// DefaultConfig returns the default simulator configuration for a 5MWh BESS.
func DefaultConfig() Config {
	return Config{
		InitialSOC:        50.0,
		SolarPeakW:        10000,
		LoadBaseW:          5000,
		LoadVarianceW:      3000,
		BatteryCapacityWh: 5000000,
		MaxChargeRateW:    50000,
		MaxDischargeRateW: 50000,
	}
}

// Simulator generates fake device data for testing the EMS without real hardware.
type Simulator struct {
	tick           int
	battSOC        float64
	essActivePower int // last applied ESS power (for grid calculation)
	config         Config
	rng            *rand.Rand
}

// New creates a new Simulator with the given configuration.
func New(cfg Config) *Simulator {
	return &Simulator{
		tick:    0,
		battSOC: cfg.InitialSOC,
		config:  cfg,
		rng:     rand.New(rand.NewSource(42)), // deterministic for reproducibility
	}
}

// NextState generates the next simulated DeviceState.
func (s *Simulator) NextState() controllers.DeviceState {
	solarW := s.solarPower()
	loadW := s.loadPower()

	// Grid = what's needed from grid: load - solar + ESS (positive = buying)
	// If ESS is discharging (positive), it reduces grid demand
	// If ESS is charging (negative), it increases grid demand
	gridW := loadW - solarW + s.essActivePower

	// Battery temperature: 25°C base + small noise
	tempC := 25.0 + (s.rng.Float64()-0.5)*2.0

	// Cell voltage: linear interpolation from SOC
	// 3200mV at 0%, 4200mV at 100%
	cellV := int(3200 + s.battSOC/100.0*1000)

	state := controllers.DeviceState{
		GridPowerW:       gridW,
		SolarPowerW:      solarW,
		LoadPowerW:       loadW,
		BatterySOC:       s.battSOC,
		BatteryTempC:     tempC,
		ESSActivePowerW:  s.essActivePower,
		CellVoltageMinMV: cellV - 10, // slight variance
		CellVoltageMaxMV: cellV + 10,
	}

	s.tick++
	return state
}

// ApplyPower applies the resolved ESS power setpoint and updates SOC.
// Positive = discharge, negative = charge.
func (s *Simulator) ApplyPower(setpointW int) {
	// Clamp to hardware limits
	if setpointW > s.config.MaxDischargeRateW {
		setpointW = s.config.MaxDischargeRateW
	}
	if setpointW < -s.config.MaxChargeRateW {
		setpointW = -s.config.MaxChargeRateW
	}

	s.essActivePower = setpointW

	// Update SOC: deltaSOC = (power * 1s) / capacity * 100
	// Discharge (positive power) decreases SOC
	// Charge (negative power) increases SOC
	deltaWh := float64(setpointW) / 3600.0 // 1 tick = 1 second
	deltaSOC := deltaWh / float64(s.config.BatteryCapacityWh) * 100.0
	s.battSOC -= deltaSOC // subtract because discharge is positive

	// Clamp SOC to 0-100%
	if s.battSOC > 100.0 {
		s.battSOC = 100.0
	}
	if s.battSOC < 0.0 {
		s.battSOC = 0.0
	}
}

// solarPower generates a sine wave simulating solar output.
// Peak at tick=43200 (noon if tick starts at midnight), 0 at night.
func (s *Simulator) solarPower() int {
	tickInDay := s.tick % dayTicks
	// Sine wave: 0 at 0h, peak at 12h, 0 at 24h
	// Solar hours roughly 6h-18h (ticks 21600-64800)
	angle := math.Pi * float64(tickInDay) / float64(dayTicks)
	power := float64(s.config.SolarPeakW) * math.Sin(angle)
	if power < 0 {
		return 0
	}
	return int(power)
}

// loadPower generates random load with base + variance.
func (s *Simulator) loadPower() int {
	variance := 0
	if s.config.LoadVarianceW > 0 {
		variance = s.rng.Intn(s.config.LoadVarianceW) - s.config.LoadVarianceW/2
	}
	load := s.config.LoadBaseW + variance
	if load < 500 {
		load = 500 // minimum load
	}
	return load
}

// SOC returns the current battery state of charge.
func (s *Simulator) SOC() float64 {
	return s.battSOC
}

// Tick returns the current tick count.
func (s *Simulator) Tick() int {
	return s.tick
}
