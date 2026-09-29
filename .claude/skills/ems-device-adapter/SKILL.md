---
name: ems-device-adapter
description: "Use when creating a new device adapter for the Local EMS project. Also use when adding hardware support, implementing a Modbus device, or scaffolding a device driver."
metadata:
  category: technique
  author: richard
  triggers: new device, add device, device adapter, modbus device, hardware adapter,
    sensor adapter, inverter adapter, meter adapter
risk: low
source: local
---

# EMS Device Adapter Generator

Generates a complete device adapter for the Local EMS project, following the existing adapter patterns (Meter, BMS, PCS, EVCharger).

## When to Use

- Adding support for a new hardware device type
- Creating a Modbus adapter for a new sensor/inverter/meter
- User says "add device", "new adapter", "support new hardware"

## Architecture Rules

> [!CAUTION]
> Device adapters ONLY read/write Modbus registers. They do NOT make control decisions.
> Control logic belongs in **controllers**, not adapters.

### Device Interface Hierarchy

```
Device (base — read-only)
├── ID() string
├── Type() DeviceType
├── Protocol() string
├── Status() DeviceStatus
└── Poll() error

ControllableDevice (writable)
├── Device
└── Write(cmd Command) error

ReconfigurableDevice (dynamic config)
├── Device
└── Reconfigure(newConfig DeviceConfig) error
```

### Device Types (existing)

| Type | Const | Description |
|------|-------|-------------|
| BMS | `DeviceTypeBMS` | Battery Management System |
| PCS | `DeviceTypePCS` | Power Conversion System |
| Meter | `DeviceTypeMeter` | Electricity Meter |
| EVCharger | `DeviceTypeEVCharger` | EV Charger |

### Data Flow

```
Modbus Registers → Adapter.Poll()/Read() → Typed Reading struct → DeviceManager → DeviceState
```

### Existing Adapter Patterns

| Adapter | File | Pattern |
|---------|------|---------|
| MeterAdapter | `meter.go` | Read-only, returns `MeterReading` |
| BMSAdapter | `bms.go` | Read-only, returns `BMSReading` |
| PCSAdapter | `pcs.go` | Read + Write, `WriteSetpoint()` |
| EVChargerAdapter | `evcharger.go` | Read + Write + `sync.RWMutex` for concurrent data |

## Generation Steps

### Step 1: Define Register Map in `registers.go`

Add register constants following existing pattern:

```go
// <DeviceName> register map.
// Unit ID: <DeviceName>=<N>
const (
	<Prefix>Reg<FieldName1>  uint16 = 0  // <type>: <Description> (<unit>)
	<Prefix>Reg<FieldName2>  uint16 = 1  // <type>: <Description> (<unit>)
	// ... more registers
	<Prefix>RegisterCount           = <N>
)

const Unit<DeviceName> byte = <N>
```

**Register naming rules:**
- Prefix: 3-letter device abbreviation (e.g., `BMS`, `PCS`, `EVC`, `MTR`)
- Hi/Lo suffix for int32 split across 2 registers
- Always document: data type, description, unit, scale factor

### Step 2: Create Reading Struct and Adapter

Create `internal/devices/<device_name>.go`:

```go
package devices

import (
	"fmt"
	"sync"

	"github.com/vmo/local-ems/internal/modbus"
)

// <DeviceName>Reading holds parsed values from the <device description>.
type <DeviceName>Reading struct {
	// Fields parsed from Modbus registers
	// Use engineering units (W, V, A, °C, %)
	// Document scale factor if any (e.g., "in 0.1V")
}

// <DeviceName>Status represents the device operating state.
type <DeviceName>Status int

const (
	<DeviceName>StatusNormal  <DeviceName>Status = 0
	<DeviceName>StatusFault   <DeviceName>Status = 1
	// ... add states matching device spec
)

// <DeviceName>Adapter reads data from a <device> via Modbus.
type <DeviceName>Adapter struct {
	client *modbus.Client
	unitID byte
	mu     sync.RWMutex   // Use if Poll() stores data for concurrent access
	data   <DeviceName>Reading
}

// New<DeviceName>Adapter creates a new adapter.
func New<DeviceName>Adapter(client *modbus.Client, unitID byte) *<DeviceName>Adapter {
	return &<DeviceName>Adapter{
		client: client,
		unitID: unitID,
	}
}

// Poll reads all registers and updates internal data.
func (a *<DeviceName>Adapter) Poll() error {
	regs, err := a.client.ReadHoldingRegisters(a.unitID, <Prefix>Reg<First>, <Prefix>RegisterCount)
	if err != nil {
		return fmt.Errorf("reading <device> registers: %w", err)
	}
	if len(regs) < int(<Prefix>RegisterCount) {
		return fmt.Errorf("reading <device> registers: expected %d, got %d",
			<Prefix>RegisterCount, len(regs))
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Parse registers into typed fields
	// Use modbus.Int32FromRegisters(hi, lo) for int32 values
	// Use float64(regs[N]) / ScaleFactor for scaled values
	// Use int16(regs[N]) for signed 16-bit values

	return nil
}

// Data returns the current reading (thread-safe).
func (a *<DeviceName>Adapter) Data() <DeviceName>Reading {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data
}

// For ControllableDevice, add Write methods:
//
// func (a *<DeviceName>Adapter) Write<Parameter>(value int) error {
//     err := a.client.WriteMultipleRegisters(a.unitID, <Prefix>Reg<Param>, []uint16{uint16(value)})
//     if err != nil {
//         return fmt.Errorf("writing <parameter>: %w", err)
//     }
//     return nil
// }
```

### Step 3: Add DeviceType constant

In `device.go`, add:

```go
DeviceType<Name> DeviceType = "<Name>"
```

### Step 4: Create Unit Tests

Create `internal/devices/<device_name>_test.go`:

```go
package devices

import (
	"testing"
)

func Test<DeviceName>Reading_Defaults(t *testing.T) {
	// Verify Reading struct zero values are safe
	var r <DeviceName>Reading
	// Assert default values are sensible
}

func Test<DeviceName>Status_String(t *testing.T) {
	// Verify status string representations
}
```

### Step 5: Update YAML Config Schema

In `config/dev.yaml`, add device section:

```yaml
devices:
  <device_name>:
    - id: <device>-0
      host: 127.0.0.1
      port: 5020
      unit: <N>
      protocol: modbus
```

### Step 6: Register in Simulator (if needed)

Add simulated registers for the new device in `internal/simulator/simulator.go`.

### Step 7: Run Tests

```bash
go test ./internal/devices/... -v
go test ./... -v
```

## Checklist

- [ ] Register constants added to `registers.go` with documentation
- [ ] Reading struct with engineering units
- [ ] Adapter with `Poll()` → parse registers → typed fields
- [ ] Thread-safe with `sync.RWMutex` if data accessed concurrently
- [ ] Error wrapping: `fmt.Errorf("context: %w", err)`
- [ ] DeviceType constant added to `device.go`
- [ ] Write methods for ControllableDevice (if writable)
- [ ] Unit tests covering read parsing, error cases
- [ ] YAML config updated for dev and production
- [ ] Simulator registers added (if simulate mode needed)
- [ ] All tests pass: `go test ./...`

## Register Parsing Patterns (from existing code)

```go
// int32 from 2 registers (hi, lo):
value := int(modbus.Int32FromRegisters(regs[hiIdx], regs[loIdx]))

// Scaled uint16 (0.1V, 0.01Hz, etc.):
voltage := float64(regs[idx]) / 10.0
freq := float64(regs[idx]) / 100.0

// Signed int16:
temp := float64(int16(regs[idx])) / 10.0

// Boolean from uint16:
connected := regs[idx] == 1

// Status enum:
status := DeviceStatus(regs[idx])
```
