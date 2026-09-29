package devices

import (
	"fmt"
	"sync"

	"github.com/vmo/local-ems/internal/modbus"
)

const (
	EVChargerUnitID     byte   = 5
	EVCStatus           uint16 = 0 // Status (0-5)
	EVCActivePower      uint16 = 1 // Active Power (int16, W)
	EVCEnergyDelivered  uint16 = 2 // Cumulative Energy (uint16, Wh)
	EVCMaxPowerLimit    uint16 = 3 // Power limit setting (uint16, W)
	EVCVehicleConnected uint16 = 4 // 0=No, 1=Yes
	EVCCurrentL1        uint16 = 5 // Phase L1 Current (uint16, 0.1A)
	EVCCurrentL2        uint16 = 6 // Phase L2 Current (uint16, 0.1A)
	EVCCurrentL3        uint16 = 7 // Phase L3 Current (uint16, 0.1A)
	EVCCommand          uint16 = 8 // Command register (0=Stop, 1=Start)
)

type EVChargerStatus int

const (
	StatusAvailable     EVChargerStatus = 0
	StatusCharging      EVChargerStatus = 1
	StatusSuspendedEV   EVChargerStatus = 2
	StatusSuspendedEVSE EVChargerStatus = 3
	StatusFinishing     EVChargerStatus = 4
	StatusFault         EVChargerStatus = 5
)

type EVChargerData struct {
	Status           EVChargerStatus
	ActivePower      int // W
	EnergyDelivered  int // Wh
	MaxPowerLimit    int // W
	VehicleConnected bool
	CurrentL1        float64 // A
	CurrentL2        float64 // A
	CurrentL3        float64 // A
}

type EVChargerAdapter struct {
	client *modbus.Client
	unitID byte
	mu     sync.RWMutex
	data   EVChargerData
}

func NewEVChargerAdapter(client *modbus.Client, unitID byte) *EVChargerAdapter {
	return &EVChargerAdapter{
		client: client,
		unitID: unitID,
	}
}

func (e *EVChargerAdapter) Poll() error {
	regs, err := e.client.ReadHoldingRegisters(e.unitID, EVCStatus, 8)
	if err != nil {
		return fmt.Errorf("reading ev charger registers: %w", err)
	}
	if len(regs) < 8 {
		return fmt.Errorf("reading ev charger registers: expected 8 registers, got %d", len(regs))
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.data.Status = EVChargerStatus(regs[0])
	e.data.ActivePower = int(int16(regs[1]))
	e.data.EnergyDelivered = int(regs[2])
	e.data.MaxPowerLimit = int(regs[3])
	e.data.VehicleConnected = regs[4] == 1
	e.data.CurrentL1 = float64(regs[5]) / 10.0
	e.data.CurrentL2 = float64(regs[6]) / 10.0
	e.data.CurrentL3 = float64(regs[7]) / 10.0

	return nil
}

func (e *EVChargerAdapter) Data() EVChargerData {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.data
}

func (e *EVChargerAdapter) WriteMaxPower(watts int) error {
	err := e.client.WriteMultipleRegisters(e.unitID, EVCMaxPowerLimit, []uint16{uint16(watts)})
	if err != nil {
		return fmt.Errorf("writing max power limit: %w", err)
	}
	return nil
}

func (e *EVChargerAdapter) WriteStartStop(start bool) error {
	val := uint16(0)
	if start {
		val = 1
	}
	err := e.client.WriteMultipleRegisters(e.unitID, EVCCommand, []uint16{val})
	if err != nil {
		return fmt.Errorf("writing start/stop command: %w", err)
	}
	return nil
}
