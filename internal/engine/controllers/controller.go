package controllers

import "github.com/vmo/local-ems/internal/engine/resolver"

// DeviceState holds current readings from all devices (read-only process image).
type DeviceState struct {
	GridPowerW       int     // Grid meter: positive=buying, negative=selling
	SolarPowerW      int     // Solar meter: positive=generating
	LoadPowerW       int     // Load meter: positive=consuming
	BatterySOC       float64 // Battery SOC 0-100%
	BatteryTempC     float64 // Battery temperature in Celsius
	ESSActivePowerW  int     // Current ESS power: positive=discharge, negative=charge
	CellVoltageMinMV int     // Min cell voltage in millivolts
	CellVoltageMaxMV int     // Max cell voltage in millivolts
}

// Controller is the interface all energy controllers must implement.
type Controller interface {
	// ID returns the unique identifier for this controller.
	ID() string
	// Run reads the current device state and adds constraints to the collector.
	Run(state DeviceState, collector *resolver.ConstraintCollector) error
	// IsEnabled returns whether this controller is currently active.
	IsEnabled() bool
}

// PowerAdvisor is optionally implemented by controllers that suggest a desired
// power setpoint. The cycle manager collects these and picks the highest-priority
// non-zero value to feed into the resolver.
type PowerAdvisor interface {
	DesiredPower(state DeviceState) int
}
