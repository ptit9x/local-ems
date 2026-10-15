package drivers

import (
	"time"

	"github.com/vmo/ems-platform/ems-local/internal/devices"
)

// Alarm represents a PCS alarm/fault event.
type Alarm struct {
	Code        uint16
	Description string
	Severity    AlarmSeverity
	Timestamp   time.Time
	Active      bool
}

// AlarmSeverity represents the severity level of an alarm.
type AlarmSeverity int

const (
	AlarmInfo     AlarmSeverity = 0
	AlarmWarning  AlarmSeverity = 1
	AlarmFault    AlarmSeverity = 2
	AlarmCritical AlarmSeverity = 3
)

// String returns a human-readable representation of AlarmSeverity.
func (s AlarmSeverity) String() string {
	switch s {
	case AlarmInfo:
		return "Info"
	case AlarmWarning:
		return "Warning"
	case AlarmFault:
		return "Fault"
	case AlarmCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// PCSCapabilities describes what operations a PCS driver supports.
type PCSCapabilities struct {
	SupportsReactivePower bool
	SupportsAlarms        bool
	SupportsBidirectional bool
	MaxPowerW             int
	MinPowerW             int
}

// PCSDriver abstracts PCS communication for different brands.
type PCSDriver interface {
	// ReadState reads the current PCS state from hardware.
	ReadState() (devices.PCSReading, error)
	// WriteSetpoint writes the active power setpoint.
	WriteSetpoint(powerW int) error
	// CommandMode returns "master" or "individual".
	CommandMode() string
	// Brand returns the driver brand identifier.
	Brand() string
	// ReadAlarms reads active alarms/faults from PCS.
	ReadAlarms() ([]Alarm, error)
	// Capabilities returns what this driver supports.
	Capabilities() PCSCapabilities
	// ProtocolType returns "modbus" or "api".
	ProtocolType() string
	// Close releases any resources.
	Close() error
}
