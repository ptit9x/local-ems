package devices

import (
	"fmt"

	"github.com/vmo/local-ems/internal/engine/controllers"
)

// Manager aggregates all device adapters and produces a unified DeviceState.
type Manager struct {
	gridMeter  *MeterAdapter
	solarMeter *MeterAdapter
	bms        *BMSAdapter
	pcs        *PCSAdapter
}

// NewManager creates a new device manager with all adapters.
func NewManager(gridMeter, solarMeter *MeterAdapter, bms *BMSAdapter, pcs *PCSAdapter) *Manager {
	return &Manager{
		gridMeter:  gridMeter,
		solarMeter: solarMeter,
		bms:        bms,
		pcs:        pcs,
	}
}

// ReadState reads all devices and composes a DeviceState for the engine.
func (m *Manager) ReadState() (controllers.DeviceState, error) {
	grid, err := m.gridMeter.Read()
	if err != nil {
		return controllers.DeviceState{}, fmt.Errorf("grid meter: %w", err)
	}

	solar, err := m.solarMeter.Read()
	if err != nil {
		return controllers.DeviceState{}, fmt.Errorf("solar meter: %w", err)
	}

	bms, err := m.bms.Read()
	if err != nil {
		return controllers.DeviceState{}, fmt.Errorf("bms: %w", err)
	}

	pcs, err := m.pcs.Read()
	if err != nil {
		return controllers.DeviceState{}, fmt.Errorf("pcs: %w", err)
	}

	// Load = grid import + solar - ESS discharge
	// But more accurately: Load = what the building is consuming
	// Grid = Load - Solar + ESS  =>  Load = Grid + Solar - ESS
	loadW := grid.ActivePowerW + solar.ActivePowerW - pcs.ActualActivePowerW

	return controllers.DeviceState{
		GridPowerW:       grid.ActivePowerW,
		SolarPowerW:      solar.ActivePowerW,
		LoadPowerW:       loadW,
		BatterySOC:       bms.SOC,
		BatteryTempC:     bms.TemperatureC,
		ESSActivePowerW:  pcs.ActualActivePowerW,
		CellVoltageMinMV: bms.CellVoltageMinMV,
		CellVoltageMaxMV: bms.CellVoltageMaxMV,
	}, nil
}

// WriteSetpoint writes the active power setpoint to the PCS.
func (m *Manager) WriteSetpoint(powerW int) error {
	return m.pcs.WriteSetpoint(powerW)
}
