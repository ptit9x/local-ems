// Package devices provides hardware adapters that read/write device data
// through Modbus TCP, translating raw registers into typed values.
package devices

// Register map constants for each device type.
// These define the Modbus register addresses used by the simulator and adapters.

// Meter register map (Grid Meter and Solar Meter share the same layout).
// Unit IDs: GridMeter=1, SolarMeter=2
const (
	MeterRegActivePowerHi  uint16 = 0  // int32: Active Power in watts (hi word)
	MeterRegActivePowerLo  uint16 = 1  // int32: Active Power in watts (lo word)
	MeterRegVoltageL1      uint16 = 2  // uint16: Voltage L1 in 0.1V
	MeterRegVoltageL2      uint16 = 3  // uint16: Voltage L2 in 0.1V
	MeterRegVoltageL3      uint16 = 4  // uint16: Voltage L3 in 0.1V
	MeterRegCurrentL1      uint16 = 5  // uint16: Current L1 in 0.1A
	MeterRegCurrentL2      uint16 = 6  // uint16: Current L2 in 0.1A
	MeterRegCurrentL3      uint16 = 7  // uint16: Current L3 in 0.1A
	MeterRegFrequency      uint16 = 8  // uint16: Frequency in 0.01Hz
	MeterRegisterCount            = 10
)

// Meter unit IDs
const (
	UnitGridMeter  byte = 1
	UnitSolarMeter byte = 2
)

// BMS register map.
// Unit ID: BMS=3
const (
	BMSRegSOC            uint16 = 0  // uint16: SOC in 0.1% (e.g., 500 = 50.0%)
	BMSRegTemperature    uint16 = 1  // int16: Temperature in 0.1°C (e.g., 250 = 25.0°C)
	BMSRegCellVoltageMin uint16 = 2  // uint16: Min cell voltage in mV
	BMSRegCellVoltageMax uint16 = 3  // uint16: Max cell voltage in mV
	BMSRegStatus         uint16 = 4  // uint16: 0=Standby, 1=Charging, 2=Discharging, 3=Fault
	BMSRegCycleCount     uint16 = 5  // uint16: Charge/discharge cycle count
	BMSRegisterCount            = 8
)

const UnitBMS byte = 3

// PCS (Power Conversion System) register map.
// Unit ID: PCS=4
const (
	PCSRegActivePowerSetpointHi uint16 = 0  // int32: Setpoint in watts (hi) — WRITABLE
	PCSRegActivePowerSetpointLo uint16 = 1  // int32: Setpoint in watts (lo) — WRITABLE
	PCSRegActualActivePowerHi   uint16 = 2  // int32: Actual power in watts (hi) — READ
	PCSRegActualActivePowerLo   uint16 = 3  // int32: Actual power in watts (lo) — READ
	PCSRegStatus                uint16 = 4  // uint16: 0=Stopped, 1=Running, 2=Fault
	PCSRegReactivePowerHi       uint16 = 5  // int32: Reactive power in var (hi)
	PCSRegReactivePowerLo       uint16 = 6  // int32: Reactive power in var (lo)
	PCSRegisterCount                   = 8
)

const UnitPCS byte = 4
