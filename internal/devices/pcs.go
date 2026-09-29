package devices

import (
	"fmt"

	"github.com/vmo/local-ems/internal/modbus"
)

// PCSReading holds parsed values from the Power Conversion System.
type PCSReading struct {
	ActivePowerSetpointW int       // Current setpoint in watts
	ActualActivePowerW   int       // Actual output power in watts
	ReactivePowerVar     int       // Reactive power in var
	Status               PCSStatus
}

// PCSStatus represents the PCS operating state.
type PCSStatus int

const (
	PCSStopped PCSStatus = 0
	PCSRunning PCSStatus = 1
	PCSFault   PCSStatus = 2
)

// String returns a human-readable PCS status.
func (s PCSStatus) String() string {
	switch s {
	case PCSStopped:
		return "Stopped"
	case PCSRunning:
		return "Running"
	case PCSFault:
		return "Fault"
	default:
		return "Unknown"
	}
}

// PCSAdapter reads PCS data and writes power setpoints via Modbus.
type PCSAdapter struct {
	client *modbus.Client
	unitID byte
}

// NewPCSAdapter creates a new PCS adapter.
func NewPCSAdapter(client *modbus.Client, unitID byte) *PCSAdapter {
	return &PCSAdapter{
		client: client,
		unitID: unitID,
	}
}

// Read fetches all PCS registers and returns a parsed PCSReading.
func (p *PCSAdapter) Read() (PCSReading, error) {
	regs, err := p.client.ReadHoldingRegisters(p.unitID, PCSRegActivePowerSetpointHi, 7)
	if err != nil {
		return PCSReading{}, fmt.Errorf("pcs read: %w", err)
	}

	if len(regs) < 7 {
		return PCSReading{}, fmt.Errorf("pcs: expected 7 registers, got %d", len(regs))
	}

	return PCSReading{
		ActivePowerSetpointW: int(modbus.Int32FromRegisters(regs[0], regs[1])),
		ActualActivePowerW:   int(modbus.Int32FromRegisters(regs[2], regs[3])),
		Status:               PCSStatus(regs[4]),
		ReactivePowerVar:     int(modbus.Int32FromRegisters(regs[5], regs[6])),
	}, nil
}

// WriteSetpoint writes the active power setpoint to the PCS.
// Positive = discharge, negative = charge.
func (p *PCSAdapter) WriteSetpoint(powerW int) error {
	return p.client.WriteInt32(p.unitID, PCSRegActivePowerSetpointHi, int32(powerW))
}
