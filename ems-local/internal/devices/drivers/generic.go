package drivers

import (
	"github.com/vmo/ems-platform/ems-local/internal/devices"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
)

// GenericPCSDriver wraps a ProfiledPCSAdapter and implements the PCSDriver interface.
type GenericPCSDriver struct {
	adapter *ProfiledPCSAdapter
}

// NewGenericPCSDriver creates a new GenericPCSDriver.
func NewGenericPCSDriver(client *modbus.Client, unitID byte, profile PCSProfile) *GenericPCSDriver {
	return &GenericPCSDriver{
		adapter: NewProfiledPCSAdapter(client, unitID, profile),
	}
}

// ReadState reads the current PCS state from hardware.
func (d *GenericPCSDriver) ReadState() (devices.PCSReading, error) {
	if err := d.adapter.Poll(); err != nil {
		return devices.PCSReading{}, err
	}
	return d.adapter.Data(), nil
}

// WriteSetpoint writes the active power setpoint.
func (d *GenericPCSDriver) WriteSetpoint(powerW int) error {
	return d.adapter.WriteActivePower(powerW)
}

// CommandMode returns "master" or "individual".
func (d *GenericPCSDriver) CommandMode() string {
	return d.adapter.CommandMode()
}

// Brand returns the driver brand identifier.
func (d *GenericPCSDriver) Brand() string {
	return d.adapter.Brand()
}

// ReadAlarms reads active alarms/faults from PCS. Generic driver does not support alarms.
func (d *GenericPCSDriver) ReadAlarms() ([]Alarm, error) {
	return []Alarm{}, nil
}

// Capabilities returns what operations this driver supports.
func (d *GenericPCSDriver) Capabilities() PCSCapabilities {
	return PCSCapabilities{
		SupportsReactivePower: false,
		SupportsAlarms:        false,
		SupportsBidirectional: true,
		MaxPowerW:             0,
		MinPowerW:             0,
	}
}

// ProtocolType returns "modbus" or "api".
func (d *GenericPCSDriver) ProtocolType() string {
	return "modbus"
}

// Close releases any resources.
func (d *GenericPCSDriver) Close() error {
	// ProfiledPCSAdapter does not manage the modbus client lifecycle.
	return nil
}
