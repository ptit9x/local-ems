package devices

import (
	"fmt"

	"github.com/vmo/local-ems/internal/modbus"
)

// BMSReading holds parsed values from the Battery Management System.
type BMSReading struct {
	SOC            float64 // State of charge 0-100%
	TemperatureC   float64 // Battery temperature in °C
	CellVoltageMinMV int   // Minimum cell voltage in millivolts
	CellVoltageMaxMV int   // Maximum cell voltage in millivolts
	Status         BMSStatus
	CycleCount     int
}

// BMSStatus represents the BMS operating state.
type BMSStatus int

const (
	BMSStandby     BMSStatus = 0
	BMSCharging    BMSStatus = 1
	BMSDischarging BMSStatus = 2
	BMSFault       BMSStatus = 3
)

// String returns a human-readable BMS status.
func (s BMSStatus) String() string {
	switch s {
	case BMSStandby:
		return "Standby"
	case BMSCharging:
		return "Charging"
	case BMSDischarging:
		return "Discharging"
	case BMSFault:
		return "Fault"
	default:
		return "Unknown"
	}
}

// BMSAdapter reads battery management system data from a Modbus device.
type BMSAdapter struct {
	client *modbus.Client
	unitID byte
}

// NewBMSAdapter creates a new BMS adapter.
func NewBMSAdapter(client *modbus.Client, unitID byte) *BMSAdapter {
	return &BMSAdapter{
		client: client,
		unitID: unitID,
	}
}

// Read fetches all BMS registers and returns a parsed BMSReading.
func (b *BMSAdapter) Read() (BMSReading, error) {
	regs, err := b.client.ReadHoldingRegisters(b.unitID, BMSRegSOC, 6)
	if err != nil {
		return BMSReading{}, fmt.Errorf("bms read: %w", err)
	}

	if len(regs) < 6 {
		return BMSReading{}, fmt.Errorf("bms: expected 6 registers, got %d", len(regs))
	}

	return BMSReading{
		SOC:              float64(regs[BMSRegSOC]) / 10.0,
		TemperatureC:     float64(int16(regs[BMSRegTemperature])) / 10.0,
		CellVoltageMinMV: int(regs[BMSRegCellVoltageMin]),
		CellVoltageMaxMV: int(regs[BMSRegCellVoltageMax]),
		Status:           BMSStatus(regs[BMSRegStatus]),
		CycleCount:       int(regs[BMSRegCycleCount]),
	}, nil
}
