package devices

import (
	"fmt"

	"github.com/vmo/local-ems/internal/modbus"
)

// MeterReading holds parsed values from an electricity meter.
type MeterReading struct {
	ActivePowerW int     // Active power in watts (positive=import, negative=export)
	VoltageL1V   float64 // Phase L1 voltage in volts
	VoltageL2V   float64 // Phase L2 voltage in volts
	VoltageL3V   float64 // Phase L3 voltage in volts
	CurrentL1A   float64 // Phase L1 current in amps
	CurrentL2A   float64 // Phase L2 current in amps
	CurrentL3A   float64 // Phase L3 current in amps
	FrequencyHz  float64 // Grid frequency in Hz
}

// MeterAdapter reads meter data from a Modbus device.
type MeterAdapter struct {
	client *modbus.Client
	unitID byte
	name   string
}

// NewMeterAdapter creates a new meter adapter.
func NewMeterAdapter(client *modbus.Client, unitID byte, name string) *MeterAdapter {
	return &MeterAdapter{
		client: client,
		unitID: unitID,
		name:   name,
	}
}

// Read fetches all meter registers and returns a parsed MeterReading.
func (m *MeterAdapter) Read() (MeterReading, error) {
	regs, err := m.client.ReadHoldingRegisters(m.unitID, MeterRegActivePowerHi, MeterRegisterCount)
	if err != nil {
		return MeterReading{}, fmt.Errorf("meter %s read: %w", m.name, err)
	}

	if len(regs) < int(MeterRegisterCount) {
		return MeterReading{}, fmt.Errorf("meter %s: expected %d registers, got %d",
			m.name, MeterRegisterCount, len(regs))
	}

	return MeterReading{
		ActivePowerW: int(modbus.Int32FromRegisters(regs[MeterRegActivePowerHi], regs[MeterRegActivePowerLo])),
		VoltageL1V:   float64(regs[MeterRegVoltageL1]) / 10.0,
		VoltageL2V:   float64(regs[MeterRegVoltageL2]) / 10.0,
		VoltageL3V:   float64(regs[MeterRegVoltageL3]) / 10.0,
		CurrentL1A:   float64(regs[MeterRegCurrentL1]) / 10.0,
		CurrentL2A:   float64(regs[MeterRegCurrentL2]) / 10.0,
		CurrentL3A:   float64(regs[MeterRegCurrentL3]) / 10.0,
		FrequencyHz:  float64(regs[MeterRegFrequency]) / 100.0,
	}, nil
}

// Name returns the meter's descriptive name.
func (m *MeterAdapter) Name() string {
	return m.name
}
