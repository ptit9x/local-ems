// Package aggregator combines state from multiple BMS, PCS, and Meter devices
// into a unified SystemState for the decision engine.
package aggregator

import (
	"sync"
	"time"
)

// BMSState represents the state of a single BMS unit.
type BMSState struct {
	ID            string    `json:"id"`
	SOC           float64   `json:"soc"`            // 0-100%
	TotalVoltageV float64   `json:"total_voltage_v"` // rack voltage
	CurrentA      float64   `json:"current_a"`
	TempC         float64   `json:"temp_c"`          // max cell temp
	MinCellMV     int       `json:"min_cell_mv"`
	MaxCellMV     int       `json:"max_cell_mv"`
	CapacityWh    int       `json:"capacity_wh"`
	MaxChargeW    int       `json:"max_charge_w"`
	MaxDischargeW int       `json:"max_discharge_w"`
	Online        bool      `json:"online"`
	AlarmFlags    uint32    `json:"alarm_flags"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PCSState represents the state of a single PCS unit.
type PCSState struct {
	ID            string    `json:"id"`
	ActivePowerW  int       `json:"active_power_w"`   // current output (+ discharge, - charge)
	SetpointW     int       `json:"setpoint_w"`       // commanded setpoint
	MaxPowerW     int       `json:"max_power_w"`      // rated capacity
	Status        PCSStatus `json:"status"`
	FaultCode     uint16    `json:"fault_code"`
	Online        bool      `json:"online"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PCSStatus represents PCS operational status.
type PCSStatus int

const (
	PCSStandby PCSStatus = iota
	PCSRunning
	PCSFault
	PCSOffline
)

// MeterState represents the state of a power meter.
type MeterState struct {
	ID           string    `json:"id"`
	ActivePowerW int       `json:"active_power_w"` // grid power (+ import, - export)
	VoltageV     float64   `json:"voltage_v"`
	CurrentA     float64   `json:"current_a"`
	FrequencyHz  float64   `json:"frequency_hz"`
	EnergyWhIn   int64     `json:"energy_wh_in"`   // total import
	EnergyWhOut  int64     `json:"energy_wh_out"`  // total export
	Online       bool      `json:"online"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EVChargerState represents the state of a single EV charger.
type EVChargerState struct {
	ID               string    `json:"id"`
	Status           int       `json:"status"`            // 0=Available,1=Charging,2=SuspendedEV,3=SuspendedEVSE,4=Finishing,5=Fault
	ActivePowerW     int       `json:"active_power_w"`    // current charging power
	EnergyDelivered  int       `json:"energy_delivered"`  // cumulative Wh in session
	MaxPowerLimit    int       `json:"max_power_limit"`   // current power limit setting
	VehicleConnected bool      `json:"vehicle_connected"`
	CurrentL1        float64   `json:"current_l1"`
	CurrentL2        float64   `json:"current_l2"`
	CurrentL3        float64   `json:"current_l3"`
	Online           bool      `json:"online"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SystemState is the aggregated state used by the decision engine.
type SystemState struct {
	// Aggregated values
	TotalSOC          float64 `json:"total_soc"`
	TotalCapacityWh   int     `json:"total_capacity_wh"`
	TotalActivePowerW int     `json:"total_active_power_w"` // sum PCS
	MaxCellTempC      float64 `json:"max_cell_temp_c"`
	MinCellVoltMV     int     `json:"min_cell_volt_mv"`
	MaxCellVoltMV     int     `json:"max_cell_volt_mv"`
	MaxChargeW        int     `json:"max_charge_w"`         // sum BMS limits
	MaxDischargeW     int     `json:"max_discharge_w"`      // sum BMS limits
	GridPowerW        int     `json:"grid_power_w"`         // from meter
	SolarPowerW       int     `json:"solar_power_w"`        // from meter or calculated
	LoadPowerW        int     `json:"load_power_w"`         // calculated
	AvgBatteryTempC   float64 `json:"avg_battery_temp_c"`

	// Per-device detail
	BMS       []BMSState       `json:"bms"`
	PCS       []PCSState       `json:"pcs"`
	Meter     []MeterState     `json:"meter"`
	EVCharger []EVChargerState `json:"ev_charger"`

	// Counts
	BMSOnline       int `json:"bms_online"`
	PCSOnline       int `json:"pcs_online"`
	PCSRunning      int `json:"pcs_running"`
	MeterOnline     int `json:"meter_online"`
	EVChargerOnline int `json:"ev_charger_online"`
	EVCharging      int `json:"ev_charging"` // number actively charging

	// EV Charger aggregation
	TotalEVPowerW int `json:"total_ev_power_w"` // sum of all EV charger power

	// Alarms
	HasAlarm bool     `json:"has_alarm"`
	Alarms   []string `json:"alarms,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

// Aggregator combines data from multiple devices into a unified state.
type Aggregator struct {
	bms   []BMSState
	pcs   []PCSState
	meter []MeterState
	ev    []EVChargerState
	mu    sync.RWMutex
}

// New creates an Aggregator sized for the given device counts.
func New(bmsCount, pcsCount, meterCount int, evCount ...int) *Aggregator {
	nEV := 0
	if len(evCount) > 0 {
		nEV = evCount[0]
	}
	return &Aggregator{
		bms:   make([]BMSState, bmsCount),
		pcs:   make([]PCSState, pcsCount),
		meter: make([]MeterState, meterCount),
		ev:    make([]EVChargerState, nEV),
	}
}

// UpdateBMS updates the state of a specific BMS by index.
func (a *Aggregator) UpdateBMS(index int, state BMSState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index >= 0 && index < len(a.bms) {
		a.bms[index] = state
	}
}

// UpdatePCS updates the state of a specific PCS by index.
func (a *Aggregator) UpdatePCS(index int, state PCSState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index >= 0 && index < len(a.pcs) {
		a.pcs[index] = state
	}
}

// UpdateMeter updates the state of a specific meter by index.
func (a *Aggregator) UpdateMeter(index int, state MeterState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index >= 0 && index < len(a.meter) {
		a.meter[index] = state
	}
}

// UpdateEVCharger updates the state of a specific EV charger by index.
func (a *Aggregator) UpdateEVCharger(index int, state EVChargerState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index >= 0 && index < len(a.ev) {
		a.ev[index] = state
	}
}

// Aggregate computes the unified SystemState from all device states.
func (a *Aggregator) Aggregate() SystemState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	s := SystemState{
		Timestamp:   time.Now(),
		MinCellVoltMV: 99999,
	}

	// Copy per-device data
	s.BMS = make([]BMSState, len(a.bms))
	copy(s.BMS, a.bms)
	s.PCS = make([]PCSState, len(a.pcs))
	copy(s.PCS, a.pcs)
	s.Meter = make([]MeterState, len(a.meter))
	copy(s.Meter, a.meter)
	s.EVCharger = make([]EVChargerState, len(a.ev))
	copy(s.EVCharger, a.ev)

	// BMS aggregation
	var totalCapWh int
	var totalSOCWeighted float64
	var totalTempSum float64
	var bmsOnline int

	for _, b := range a.bms {
		if !b.Online {
			continue
		}
		bmsOnline++
		totalCapWh += b.CapacityWh
		totalSOCWeighted += b.SOC * float64(b.CapacityWh)
		s.MaxChargeW += b.MaxChargeW
		s.MaxDischargeW += b.MaxDischargeW
		totalTempSum += b.TempC

		if b.MaxCellMV > s.MaxCellVoltMV {
			s.MaxCellVoltMV = b.MaxCellMV
		}
		if b.MinCellMV < s.MinCellVoltMV && b.MinCellMV > 0 {
			s.MinCellVoltMV = b.MinCellMV
		}
		if b.TempC > s.MaxCellTempC {
			s.MaxCellTempC = b.TempC
		}

		// Alarm checks
		if b.AlarmFlags != 0 {
			s.HasAlarm = true
			s.Alarms = append(s.Alarms, b.ID+": alarm flags active")
		}
	}

	s.BMSOnline = bmsOnline
	s.TotalCapacityWh = totalCapWh
	if totalCapWh > 0 {
		s.TotalSOC = totalSOCWeighted / float64(totalCapWh)
	}
	if bmsOnline > 0 {
		s.AvgBatteryTempC = totalTempSum / float64(bmsOnline)
	}
	if s.MinCellVoltMV == 99999 {
		s.MinCellVoltMV = 0
	}

	// PCS aggregation
	for _, p := range a.pcs {
		if !p.Online {
			continue
		}
		s.PCSOnline++
		if p.Status == PCSRunning {
			s.PCSRunning++
		}
		s.TotalActivePowerW += p.ActivePowerW
	}

	// Meter aggregation (use first online meter for grid power)
	for _, m := range a.meter {
		if !m.Online {
			continue
		}
		s.MeterOnline++
		if s.GridPowerW == 0 {
			s.GridPowerW = m.ActivePowerW
		}
	}

	// Calculated load = grid + ess_discharge + solar
	s.LoadPowerW = s.GridPowerW + s.TotalActivePowerW + s.SolarPowerW

	// EV Charger aggregation
	for _, ev := range a.ev {
		if !ev.Online {
			continue
		}
		s.EVChargerOnline++
		s.TotalEVPowerW += ev.ActivePowerW
		if ev.Status == 1 { // Charging
			s.EVCharging++
		}
	}

	// Safety alarms
	if s.MaxCellTempC > 55 {
		s.HasAlarm = true
		s.Alarms = append(s.Alarms, "CRITICAL: cell temperature > 55°C")
	}
	if s.MaxCellVoltMV > 3650 {
		s.HasAlarm = true
		s.Alarms = append(s.Alarms, "CRITICAL: cell overvoltage > 3.65V")
	}
	if s.MinCellVoltMV > 0 && s.MinCellVoltMV < 2800 {
		s.HasAlarm = true
		s.Alarms = append(s.Alarms, "CRITICAL: cell undervoltage < 2.8V")
	}
	if bmsOnline == 0 {
		s.HasAlarm = true
		s.Alarms = append(s.Alarms, "CRITICAL: no BMS online")
	}

	return s
}

// BMSCount returns the number of BMS units.
func (a *Aggregator) BMSCount() int { return len(a.bms) }

// PCSCount returns the number of PCS units.
func (a *Aggregator) PCSCount() int { return len(a.pcs) }

// MeterCount returns the number of meters.
func (a *Aggregator) MeterCount() int { return len(a.meter) }

// EVChargerCount returns the number of EV chargers.
func (a *Aggregator) EVChargerCount() int { return len(a.ev) }
