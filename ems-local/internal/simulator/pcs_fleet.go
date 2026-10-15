package simulator

import (
	"math"
	"sync"

	"github.com/vmo/ems-platform/ems-local/internal/devices/drivers"
	"github.com/vmo/ems-platform/ems-local/internal/modbus"
)

// PCSUnitState represents the simulated state of a single PCS unit.
type PCSUnitState struct {
	UnitID    byte
	Brand     string
	PowerW    int    // Current actual power output (follows setpoint with lag)
	SetpointW int    // Last received setpoint from EMS
	Status    int    // 0=Standby, 1=Running, 2=Fault
	FaultCode uint16 // Active fault code (0 if none)
}

// PCSFleetSimulator manages N simulated PCS units, each as a separate
// Modbus slave on a shared Modbus server.
type PCSFleetSimulator struct {
	server  *modbus.Server
	units   []*PCSUnitState
	mu      sync.RWMutex
	profile drivers.PCSProfile // register layout for all units
}

// NewPCSFleetSimulator creates a new PCS fleet simulator managing count units.
// addr specifies the Modbus TCP server address (e.g. ":0" or "127.0.0.1:502").
func NewPCSFleetSimulator(addr string, count int, profile drivers.PCSProfile) *PCSFleetSimulator {
	server := modbus.NewServer(addr)

	// Determine register allocation size based on profile.RegisterCount, ensuring
	// enough space for offset register maps (such as Lite-On starting at 40001).
	regCount := int(profile.RegisterCount)
	maxAddr := profile.ActivePowerSetReg
	if profile.ActivePowerActReg > maxAddr {
		maxAddr = profile.ActivePowerActReg
	}
	if profile.ReactivePowerReg > maxAddr {
		maxAddr = profile.ReactivePowerReg
	}
	if profile.StatusReg > maxAddr {
		maxAddr = profile.StatusReg
	}
	if int(maxAddr)+int(profile.RegisterCount) > regCount {
		regCount = int(maxAddr) + int(profile.RegisterCount)
	}

	sim := &PCSFleetSimulator{
		server:  server,
		units:   make([]*PCSUnitState, count),
		profile: profile,
	}

	for i := 0; i < count; i++ {
		unitID := byte(i + 1)
		server.AllocateRegisters(unitID, regCount)

		unit := &PCSUnitState{
			UnitID:    unitID,
			Brand:     profile.Brand,
			PowerW:    0,
			SetpointW: 0,
			Status:    0, // Standby
			FaultCode: 0,
		}
		sim.units[i] = unit

		// Initialize Modbus registers
		server.SetRegister(unitID, profile.ActivePowerSetReg, 0)
		sim.writeRegisters(unit)
	}

	return sim
}

// Start begins listening for Modbus TCP connections.
func (s *PCSFleetSimulator) Start() error {
	return s.server.Start()
}

// Stop shuts down the Modbus TCP server.
func (s *PCSFleetSimulator) Stop() {
	_ = s.server.Stop()
}

// Addr returns the Modbus server address.
func (s *PCSFleetSimulator) Addr() string {
	return s.server.Addr()
}

// Tick simulates one time step:
// - For each unit: actual power ramps toward setpoint (max 500W/tick)
// - If status is Fault, power drops to 0
// - Updates Modbus registers from internal state
func (s *PCSFleetSimulator) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	scale := s.profile.ScaleFactor
	if scale == 0 {
		scale = 1.0
	}

	const maxRampW = 500

	for _, unit := range s.units {
		// Read setpoint register written by EMS
		rawSetpoint := s.server.GetRegister(unit.UnitID, s.profile.ActivePowerSetReg)
		unit.SetpointW = int(float64(int16(rawSetpoint)) * scale)

		if unit.Status == 2 {
			// Fault: power drops to 0
			unit.PowerW = 0
		} else {
			// Ramp actual power toward setpoint (max 500W/tick)
			diff := unit.SetpointW - unit.PowerW
			if diff > maxRampW {
				unit.PowerW += maxRampW
			} else if diff < -maxRampW {
				unit.PowerW -= maxRampW
			} else {
				unit.PowerW = unit.SetpointW
			}

			if unit.PowerW != 0 || unit.SetpointW != 0 {
				unit.Status = 1 // Running
			} else {
				unit.Status = 0 // Standby
			}
		}

		s.writeRegisters(unit)
	}
}

// SetFault injects a fault on a specific unit.
func (s *PCSFleetSimulator) SetFault(unitIndex int, faultCode uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if unitIndex < 0 || unitIndex >= len(s.units) {
		return
	}

	unit := s.units[unitIndex]
	unit.Status = 2 // Fault
	unit.FaultCode = faultCode
	unit.PowerW = 0
	s.writeRegisters(unit)
}

// ClearFault clears the fault on a specific unit and returns it to Standby.
func (s *PCSFleetSimulator) ClearFault(unitIndex int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if unitIndex < 0 || unitIndex >= len(s.units) {
		return
	}

	unit := s.units[unitIndex]
	unit.Status = 0 // Standby
	unit.FaultCode = 0
	s.writeRegisters(unit)
}

// GetUnit returns a snapshot copy of a single unit's state.
func (s *PCSFleetSimulator) GetUnit(unitIndex int) PCSUnitState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if unitIndex < 0 || unitIndex >= len(s.units) {
		return PCSUnitState{}
	}
	return *s.units[unitIndex]
}

// GetAllUnits returns a snapshot copy of all unit states.
func (s *PCSFleetSimulator) GetAllUnits() []PCSUnitState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]PCSUnitState, len(s.units))
	for i, u := range s.units {
		result[i] = *u
	}
	return result
}

// SetSetpoint sets the active power setpoint for a specific unit.
func (s *PCSFleetSimulator) SetSetpoint(unitIndex int, setpointW int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if unitIndex < 0 || unitIndex >= len(s.units) {
		return
	}

	unit := s.units[unitIndex]
	unit.SetpointW = setpointW

	scale := s.profile.ScaleFactor
	if scale == 0 {
		scale = 1.0
	}
	rawSetpoint := uint16(int16(math.Round(float64(setpointW) / scale)))
	s.server.SetRegister(unit.UnitID, s.profile.ActivePowerSetReg, rawSetpoint)
}

// Server returns the underlying Modbus server.
func (s *PCSFleetSimulator) Server() *modbus.Server {
	return s.server
}

// writeRegisters updates the Modbus registers for a unit based on its current state.
func (s *PCSFleetSimulator) writeRegisters(unit *PCSUnitState) {
	scale := s.profile.ScaleFactor
	if scale == 0 {
		scale = 1.0
	}

	rawPower := uint16(int16(math.Round(float64(unit.PowerW) / scale)))
	s.server.SetRegister(unit.UnitID, s.profile.ActivePowerActReg, rawPower)
	s.server.SetRegister(unit.UnitID, s.profile.StatusReg, uint16(unit.Status))
	s.server.SetRegister(unit.UnitID, s.profile.ReactivePowerReg, 0)
}
